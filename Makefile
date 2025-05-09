# после двоеточия, прям на строчке с командой, перечисляются другие команды,
# которые надо запустить перед текущей
# с новой строки можно просто писать shell скрипт обычный

# all команда срабатывает, если запускаешь make без параметров
all: build test

# ^ альтернатива
#all:
#	make build
#	make test

# init-* скрипты установки внешних зависимостей проекта для разработки
init-windows:
	scoop install migrate

init-macos:
	brew install golang-migrate

test:
	go test internal/...

build:
	go build -o main.out cmd/main.go

run:
	go run cmd/main.go -config=./configs/local.json
# пример с ENV переменными(на винде скорее всего не работает): env CONFIG_PATH=./configs/local.json go run cmd/main.go

migrate-up:
	migrate -database postgresql://user_pg:password_pg@localhost:5433/db_pg?sslmode=disable -path ./db/migrations up

migrate-down:
	migrate -database postgresql://user_pg:password_pg@localhost:5433/db_pg?sslmode=disable -path ./db/migrations down