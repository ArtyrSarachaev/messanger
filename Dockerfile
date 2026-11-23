FROM golang:1.24-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/messanger ./cmd \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./migrations

FROM alpine:3.20

RUN adduser -D -u 10001 app
WORKDIR /app

COPY --from=build /out/messanger /out/migrate ./
COPY configs ./configs

ENV CONFIG_PATH=/app/configs/local.json
USER app

EXPOSE 8000 8001
ENTRYPOINT ["/app/messanger"]
