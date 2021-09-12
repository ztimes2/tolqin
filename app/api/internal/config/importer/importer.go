package importer

import (
	"context"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
	"github.com/ztimes2/tolqin/internal/config"
)

const (
	defaultBatchSize = 100
)

type Config struct {
	config.Database
	config.Logger

	BatchSize int    `config:"BATCH_SIZE"`
	CSVFile   string `config:"CSV_FILE"`
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
		BatchSize: defaultBatchSize,
	}

	if err := loader.Load(context.Background(), &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
