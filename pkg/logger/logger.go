package logger

import (
	"context"
	"os"

	env "messanger/pkg/environment"

	"github.com/sirupsen/logrus"
)

/*
	Logrus норм выбор, но самый пиздатый сейчас это go.uber.org/zap
*/

const (
	loggerKey = "logger"
)

// New лучше все зависимости передавать явно, в таком коде гораздо проще разбираться
// Так что получи env перед созданием логгера и передай в конструктор
func New(ctx context.Context) *logrus.Logger {
	var log = logrus.New()

	switch env.GetEnv() {
	case env.Local:
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

// LoggerFromContext -> FromContext. В коде будет красиво logger.FromContext
// А это у вас такая практика класть логгер в контекст? Или ты где-то увидел
func LoggerFromContext(ctx context.Context) *logrus.Logger {
	if logger, ok := ctx.Value(loggerKey).(*logrus.Logger); ok {
		return logger
	}
	return logrus.StandardLogger()
}
