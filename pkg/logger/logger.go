package logger

import (
	"context"
	"os"

	env "messanger/pkg/environment"

	"github.com/sirupsen/logrus"
)

const (
	envLocalType = "local"

	loggerKey = "logger"
)

func New(ctx context.Context) *logrus.Logger {
	var log = logrus.New()

	switch env.GetEnv(ctx) {
	case envLocalType:
		log.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "@timestamp",
				logrus.FieldKeyLevel: "@level",
				logrus.FieldKeyMsg:   "@message",
			},
		})
		log.SetOutput(os.Stdout)
	default:
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "@timestamp",
				logrus.FieldKeyLevel: "@level",
				logrus.FieldKeyMsg:   "@message",
			},
		})
	}
	ctx = context.WithValue(ctx, loggerKey, log)
	return log
}

func LoggerFromContext(ctx context.Context) *logrus.Logger {
	if logger, ok := ctx.Value(loggerKey).(*logrus.Logger); ok {
		return logger
	}
	return logrus.StandardLogger()
}
