package config

import (
	"fmt"
	"net"

	"github.com/caarlos0/env/v10"
)

type (
	Config struct {
		GRPC struct {
			Port        string `env:"GRPC_PORT" envDefault:"50051"`
			GatewayPort string `env:"GRPC_GATEWAY_PORT" envDefault:"8080"`
		}

		PG struct {
			Host       string `env:"POSTGRES_HOST" envDefault:"localhost"`
			Port       string `env:"POSTGRES_PORT" envDefault:"5432"`
			DB         string `env:"POSTGRES_DB" envDefault:"urls"`
			User       string `env:"POSTGRES_USER" envDefault:"user"`
			Password   string `env:"POSTGRES_PASSWORD" envDefault:"12345"`
			MaxConn    string `env:"POSTGRES_MAX_CONN" envDefault:"10"`
			AccessType string `env:"ACCESS_TYPE" envDefault:"POSTGRES"`
		}
	}
)

func (c *Config) ConstructPostgresURL() string {
	hostPort := net.JoinHostPort(c.PG.Host, c.PG.Port)
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable&pool_max_conns=%s",
		c.PG.User,
		c.PG.Password,
		hostPort,
		c.PG.DB,
		c.PG.MaxConn,
	)
}

func New() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return &cfg, err
}
