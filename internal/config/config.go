package config

import (
	"context"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
)

const (
	defaultLogLevel  = "info"
	defaultLogFormat = "json"
)

type Config struct {
	ServerPort string `config:"SERVER_PORT,required"`
	LogLevel   string `config:"LOG_LEVEL"`
	LogFormat  string `config:"LOG_FORMAT"`

	DatabaseHost     string `config:"DB_HOST,required"`
	DatabasePort     string `config:"DB_PORT,required"`
	DatabaseUsername string `config:"DB_USERNAME"`
	DatabasePassword string `config:"DB_PASSWORD"`
	DatabaseName     string `config:"DB_NAME,required"`
	DatabaseSSLMode  string `config:"DB_SSLMODE"`
}

func Load() (Config, error) {
	loader := confita.NewLoader(
		env.NewBackend(),
		newDotEnv(),
	)

	c := Config{
		LogLevel:  defaultLogLevel,
		LogFormat: defaultLogFormat,
	}

	if err := loader.Load(context.Background(), &c); err != nil {
		return Config{}, err
	}

	return c, nil
}
