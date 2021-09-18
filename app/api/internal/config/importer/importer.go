package importer

import (
	"github.com/sirupsen/logrus"
	"github.com/ztimes2/tolqin/app/api/internal/config"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/log"
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
			LogLevel:  logrus.InfoLevel.String(),
			LogFormat: log.FormatJSON,
		},
		BatchSize: defaultBatchSize,
	}

	if err := config.Load(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
