package api

import (
	"github.com/ztimes2/tolqin/app/api/internal/config"
)

const (
	defaultLogLevel  = "info"
	defaultLogFormat = "json"
)

type Config struct {
	config.Database
	config.Logger
	config.Nominatim

	ServerPort string `config:"SERVER_PORT,required"`
}

func New() (Config, error) {
	cfg := Config{
		Logger: config.Logger{
			LogLevel:  defaultLogLevel,
			LogFormat: defaultLogFormat,
		},
	}

	if err := config.Load(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
