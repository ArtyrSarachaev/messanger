package config

import (
	"fmt"
	"time"
)

type Config struct {
	Kafka      Kafka    `json:"kafka"`
	HttpServer Server   `json:"http_server"`
	WSServer   Server   `json:"websocket_server"`
	Postgres   Postgres `json:"postgres"`
	JWT        JWT      `json:"jwt"`
}

type Kafka struct {
	Brokers []string `json:"brokers"`
	Topic   string   `json:"topic"`
	GroupID string   `json:"group_id"`
}

type Postgres struct {
	Host     string `json:"host"`
	Port     int64  `json:"port"`
	UserName string `json:"user"`
	Password string `json:"password"`
	NameDB   string `json:"name_db"`
	MaxConns int32  `json:"max_conns"`
	MinConns int32  `json:"min_conns"`
}

type Server struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type JWT struct {
	Secret string        `json:"-"`
	TTL    time.Duration `json:"-"`
}

func (s Server) Address() string {
	return s.Host + ":" + s.Port
}

// DSN возвращает строку подключения к PostgreSQL.
func (p Postgres) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		p.UserName,
		p.Password,
		p.Host,
		p.Port,
		p.NameDB)
}
