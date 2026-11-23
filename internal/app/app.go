package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"messanger/internal/auth"
	"messanger/internal/broker"
	httpserver "messanger/internal/http"
	"messanger/internal/logic"
	"messanger/internal/repository"
	"messanger/internal/websocket"
	"messanger/pkg/config"
	"messanger/pkg/postgres"

	"golang.org/x/sync/errgroup"
)

const shutdownTimeout = 10 * time.Second

func Run(cfg config.Config, log *slog.Logger) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.Postgres)
	if err != nil {
		return err
	}
	defer pool.Close()

	tokens := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.TTL)

	userLogic := logic.NewUserLogic(repository.NewUserRepository(pool), tokens)
	messageLogic := logic.NewMessageLogic(repository.NewMessageRepository(pool), userLogic)

	producer := broker.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic)
	defer closeWithLog(log, "kafka producer", producer.Close)

	consumer := broker.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID, messageLogic, log)
	defer closeWithLog(log, "kafka consumer", consumer.Close)

	httpSrv := httpserver.NewServer(userLogic, messageLogic, tokens)
	wsSrv := websocket.NewServer(userLogic, producer, tokens, log)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Info("kafka consumer started", "topic", cfg.Kafka.Topic)
		return consumer.Run(gCtx)
	})

	g.Go(func() error {
		log.Info("http server started", "addr", cfg.HttpServer.Address())
		return ignoreServerClosed(httpSrv.Start(cfg.HttpServer.Address()))
	})

	g.Go(func() error {
		log.Info("websocket server started", "addr", cfg.WSServer.Address())
		return ignoreServerClosed(wsSrv.Start(cfg.WSServer.Address()))
	})

	// Graceful shutdown: по сигналу или при падении любого компонента.
	g.Go(func() error {
		<-gCtx.Done()
		log.Info("shutting down")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		return errors.Join(
			httpSrv.Shutdown(shutdownCtx),
			wsSrv.Shutdown(shutdownCtx),
		)
	})

	if err := g.Wait(); err != nil {
		return err
	}

	log.Info("stopped gracefully")
	return nil
}

func ignoreServerClosed(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func closeWithLog(log *slog.Logger, name string, closeFn func() error) {
	if err := closeFn(); err != nil {
		log.Warn("close failed", "component", name, "error", err)
	}
}

func NewLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
