package api

import (
	"github.com/sirupsen/logrus"
	"github.com/ztimes2/tolqin/app/api/internal/config"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/log"
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
			LogLevel:  logrus.InfoLevel.String(),
			LogFormat: log.FormatJSON,
		},
	}

	if err := config.Load(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
