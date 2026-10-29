package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AppConfig
	DatabaseConfig
	ServerConfig
}

type AppConfig struct {
	LogLevel   string `env:"LOG_LEVEL" env-default:"info"`
	PrivateKey string `env:"ED25519_PRIVATE_KEY" env-required:"true"`
	Version    int    `env:"VERSION" env-default:"1"`
	AdminToken string `env:"ADMIN_TOKEN" env-required:"true"`
	KeyGroups  int    `env:"KEY_GROUPS" env-default:"4"`
	GroupsLen  int    `env:"KEY_GROUPS_LEN" env-default:"5"`
}

type DatabaseConfig struct {
	ConnString        string        `env:"CONNECTION_STRING" env-required:"true"`
	MaxConns          int32         `env:"MAX_CONNS" env-default:"25"`
	MinConns          int32         `env:"MIN_CONNS" env-default:"5"`
	MaxConnLifeTime   time.Duration `env:"MAX_CONN_LIFETIME" env-default:"1h"`
	MaxConnIdleTime   time.Duration `env:"MAX_CONN_IDLETIME" env-default:"30m"`
	HealthCheckPeriod time.Duration `env:"HEALTH_CHECK_PERIOD" env-default:"1m"`
	ConnectionTimeout time.Duration `env:"CONNECTION_TIMEOUT" env-default:"5s"`
}

type ServerConfig struct {
	Port            string        `env:"PORT" env-default:"8080"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" env-default:"2s"`
}

func New() (*Config, error) {
	var cfg Config

	if _, err := os.Stat(".env"); err == nil {
		if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
			return nil, fmt.Errorf("failed to read .env file: %w", err)
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("failed to read environment variables: %w", err)
		}
	}

	return &cfg, nil
}
