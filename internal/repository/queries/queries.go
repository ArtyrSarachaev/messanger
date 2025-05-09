package queries

import (
	_ "embed"
)

// Вот магическая штука embed, она при билде файлик читает и кладёт в переменную сразу.
// Но для серьёзных штук лучше не использовать. Безопасность превыше всего ☝️
var (
	//go:embed sql/insert-user.sql
	InsertUser string
)
