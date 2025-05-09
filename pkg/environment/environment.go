package environment

// Лучше всё-таки назвать пакет env
// И доп заметка. Никогда не используй проверку env на дев прод и локал в бизнес-логике
// Это окей только на уровне инициализации или старта приложения
// Поэтому проверка env по контексту тоже не имеет смысла
// Вообще в принципе история dev, prod и local это спорная тема, это я бы голосом обсудил

import (
	"context"
	"os"
	"strings"
)

type Env string

const (
	Local Env = "local"
	Dev   Env = "dev"
	Prod  Env = "prod"
)

const (
	kafkaConfigKey = "kafkaConfig"
	envKey         = "ENV"

	// Пакет у тебя отвечает за environment чисто
	// тут не место для userIDKey и kafkaConfigKey, точно
	// не нужно зоны ответственности смешивать
	userIDKey = "user_id"
)

func GetUserId(ctx context.Context) int64 {
	// зачем приводишь сначала к float64 а потом к int64?
	return int64(ctx.Value(userIDKey).(float64))
}

func IsLocal() bool {
	// Не пон зачем вообще тут strings.Compare, он же обычно для сортировки юзается
	// строки можно просто через == сравнивать
	return GetEnv() == Local
}

func IsDev() bool {
	return GetEnv() == Dev
}

func IsProd() bool {
	return GetEnv() == Prod
}

func GetEnv() Env {
	env := os.Getenv(envKey)

	switch strings.ToLower(env) {
	case "local":
		return Local
	case "dev":
		return Dev
	case "prod":
		return Prod
	default:
		return Local
	}
}
