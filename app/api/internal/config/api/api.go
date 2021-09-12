package api

import (
	"context"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
	"github.com/ztimes2/tolqin/app/api/internal/config"
)

type Config struct {
	config.Database
	config.Logger
	config.Nominatim

	ServerPort string `config:"SERVER_PORT,required"`
}

func Load() (Config, error) {
	loader := confita.NewLoader(
		env.NewBackend(),
		config.NewDotEnv(),
	)

	cfg := Config{
		Logger: config.Logger{
			LogLevel:  config.DefaultLogLevel,
			LogFormat: config.DefaultLogFormat,
		},
	}

	if err := loader.Load(context.Background(), &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
