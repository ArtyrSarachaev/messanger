package config

import "fmt"

type Config struct {
	Kafka      Server `json:"kafka"`
	HttpServer Server `json:"http_server"`
	Postgres   `json:"postgres"`
	WSServer   Server `json:"websocket_server"`
}

type Postgres struct {
	Host                 string `json:"host"`
	Port                 int64  `json:"port"`
	UserName             string `json:"user"`
	Password             string `json:"password"`
	NameDB               string `json:"name_db"`
	PathToCreateDatabase string `json:"path_to_create_db"`
	MaxOpenConns         int64  `json:"max_open_conns"`
	MaxIdleConns         int64  `json:"max_idle_conns"`
	LifetimeConn         int64  `json:"lifetime_conn"`
}

type Server struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

func (p Postgres) СreateConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		p.Host,
		p.Port,
		p.UserName,
		p.Password,
		p.NameDB)
}
