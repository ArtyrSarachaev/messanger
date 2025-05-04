package environment

import (
	"context"
	"strings"
)

const (
	kafkaConfigKey = "kafkaConfig"
	envKey         = "env"
	userIDKey      = "user_id"

	envLocalType = "local"
	envDevType   = "dev"
	envProdType  = "prod"
)

func GetUserId(ctx context.Context) int64 {
	return int64(ctx.Value(userIDKey).(float64))
}

func IsLocal(ctx context.Context) bool {
	return strings.Compare(ctx.Value(envKey).(string), envLocalType) == 0
}

func IsDev(ctx context.Context) bool {
	return strings.Compare(ctx.Value(envKey).(string), envDevType) == 0
}

func IsProd(ctx context.Context) bool {
	return strings.Compare(ctx.Value(envKey).(string), envProdType) == 0
}

func GetEnv(ctx context.Context) string {
	if val, ok := ctx.Value(envKey).(string); ok {
		return val
	}
	return ""
}
