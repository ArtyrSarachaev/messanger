package app

import (
	"context"
	"errors"
	"messanger/internal/broker"
	httpHandler "messanger/internal/http"
	"messanger/internal/logic"
	"messanger/internal/repository"
	"messanger/internal/websocket"
	"messanger/pkg/config"
	"messanger/pkg/logger"
	"messanger/pkg/postgres"
	"net/http"
)

func Run(ctx context.Context, config config.Config) {
	log := logger.New(ctx)

	// db pool
	poolPG, err := postgres.NewPool(ctx, config)
	if err != nil {
		// по идее то же самое, что и log.Error(err) + panic(err)
		log.Fatal(err)
	}
	var isExist bool
	if isExist, err = poolPG.CheckIsTableExists(ctx); err != nil {
		log.Error(err)
		panic(err)
	}
	if !isExist {
		err = poolPG.CreateTables(ctx, config.Postgres.PathToCreateDatabase)
		if err != nil {
			log.Error(err)
			panic(err)
		}
	}

	defer poolPG.Pool.Close()

	//repository
	userRepository := repository.NewUserRepository(poolPG.Pool)

	//broker
	kafkaWriter := broker.NewMessageKafkaWriter(config.Kafka.Host + ":" + config.Kafka.Port)
	broker.StartMessageKafkaReader(ctx, config.Kafka.Host+":"+config.Kafka.Port)

	//cache

	//logic
	userLogic := logic.NewUserLogic(userRepository)

	//servers
	httpServer := httpHandler.StartHttpServer(userLogic)
	var errHttp error
	go func() {
		err = httpServer.Start(config.HttpServer.Host + ":" + config.HttpServer.Port)
		if err != http.ErrServerClosed {
			log.Errorf("cant start http server, with error: %v", err)
			errHttp = errors.New("cant start http server, with error: " + err.Error())
			return
		}
	}()
	if errHttp != nil {
		panic(errHttp)
	}

	wsServer := websocket.StartWSServer(kafkaWriter)
	var errWS error
	go func() {
		err = wsServer.Start(config.WSServer.Host + ":" + config.WSServer.Port)
		if err != http.ErrServerClosed {
			log.Errorf("cant start websocket server, with error: %v", err)
			errWS = errors.New("cant start websocket server, with error: " + err.Error())
			return
		}
	}()
	if errWS != nil {
		panic(errWS)
	}

	select {}
}
