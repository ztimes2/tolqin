package importer

import (
	"github.com/ztimes2/tolqin/app/api/internal/config"
)

const (
	defaultBatchSize = 100
	defaultLogLevel  = "info"
	defaultLogFormat = "json"
)

type Config struct {
	config.Database
	config.Logger

	BatchSize int    `config:"BATCH_SIZE"`
	CSVFile   string `config:"CSV_FILE"`
}

func New() (Config, error) {
	cfg := Config{
		Logger: config.Logger{
			LogLevel:  defaultLogLevel,
			LogFormat: defaultLogFormat,
		},
		BatchSize: defaultBatchSize,
	}

	if err := config.Load(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
