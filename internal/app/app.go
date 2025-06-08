package app

import (
	"context"
	kafka "messanger/internal/broker"
	httpHandler "messanger/internal/http"
	"messanger/internal/logic"
	"messanger/internal/repository"
	"messanger/internal/websocket"
	"messanger/pkg/config"
	"messanger/pkg/postgres"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func Run(config config.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := zap.NewExample()

	// db pool
	poolPG, err := postgres.NewPool(ctx, config)
	if err != nil {
		log.Sugar().Fatalf("cant create pool connections for postgres: %v", err)
	}

	defer poolPG.Pool.Close()

	//repository
	userRepository := repository.NewUserRepository(poolPG.Pool)
	messageRepository := repository.NewMessageRepository(poolPG.Pool)

	//logic
	userLogic := logic.NewUserLogic(userRepository)
	messageLogic := logic.NewMessageLogic(messageRepository, userLogic)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()

	g, gCtx := errgroup.WithContext(ctx)

	//broker
	kafkaWriter := kafka.NewWriter(config.Kafka.Host + ":" + config.Kafka.Port)
	g.Go(func() error {
		kafkaReader := kafka.NewReader(ctx, config.Kafka.Host+":"+config.Kafka.Port, messageLogic)
		go func() {
			<-gCtx.Done()
			_ = kafkaReader.Shutdown(context.Background())
		}()
		log.Info("kafka reader is starting")
		return kafkaReader.Start(ctx)
	})

	//servers
	g.Go(func() error {
		httpServer := httpHandler.Server(userLogic, messageLogic)
		go func() {
			<-gCtx.Done()
			_ = httpServer.Shutdown(context.Background())
		}()
		log.Info("http server is starting")
		return httpServer.Start(config.HttpServer.Host + ":" + config.HttpServer.Port)
	})

	g.Go(func() error {
		wsServer := websocket.Server(kafkaWriter, userLogic)
		go func() {
			<-gCtx.Done()
			_ = wsServer.Shutdown(context.Background())
		}()
		log.Info("web socket server is starting")
		return wsServer.Start(config.WSServer.Host + ":" + config.WSServer.Port)
	})

	if err := g.Wait(); err != nil {
		log.Sugar().Errorf("Server error: %v", err)
	}
	log.Info("All servers is stopped gracefully")
}
