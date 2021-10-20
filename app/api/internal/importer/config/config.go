package config

import (
	"context"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
	"github.com/heetch/confita/backend/env"
	"github.com/sirupsen/logrus"
	"github.com/ztimes2/tolqin/app/api/pkg/dotenv"
	"github.com/ztimes2/tolqin/app/api/pkg/log"
)

const (
	defaultBatchSize = 100
)

type Config struct {
	Database
	Logger

	BatchSize int    `config:"BATCH_SIZE"`
	CSVFile   string `config:"CSV_FILE"`
}

type Database struct {
	Host     string `config:"DB_HOST,required"`
	Port     string `config:"DB_PORT,required"`
	Username string `config:"DB_USERNAME"`
	Password string `config:"DB_PASSWORD"`
	Name     string `config:"DB_NAME,required"`
	SSLMode  string `config:"DB_SSLMODE"`
}

type Logger struct {
	LogLevel  string `config:"LOG_LEVEL"`
	LogFormat string `config:"LOG_FORMAT"`
}

func New() (Config, error) {
	cfg := Config{
		Logger: Logger{
			LogLevel:  logrus.InfoLevel.String(),
			LogFormat: log.FormatJSON,
		},
		BatchSize: defaultBatchSize,
	}

	backends := []backend.Backend{
		env.NewBackend(),
		dotenv.NewBackend(),
	}

	if err := confita.NewLoader(backends...).Load(context.Background(), cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
