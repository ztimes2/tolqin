package config

import (
	"context"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
)

const (
	defaultLogLevel = "info"
	defaultLogFormat = "json"
)

type Config struct {
	ServerPort string `config:"SERVER_PORT,required"`
	LogLevel string `config:"LOG_LEVEL"`
	LogFormat string `config:"LOG_FORMAT"`
}

func Load() (Config, error) {
	loader := confita.NewLoader(
		env.NewBackend(),
		newDotEnv(),
	)

	c := Config{
		LogLevel: defaultLogLevel,
		LogFormat: defaultLogFormat,
	}

	if err := loader.Load(context.Background(), &c); err != nil {
		return Config{}, err
	}

	return c, nil
}
