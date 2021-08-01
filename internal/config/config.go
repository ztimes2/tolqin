package config

import (
	"context"
	"time"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/env"
)

const (
	defaultLogLevel  = "info"
	defaultLogFormat = "json"

	defaultBatchSize = 100
)

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

type API struct {
	Database
	Logger
	Nominatim

	ServerPort string `config:"SERVER_PORT,required"`
}

func LoadAPI() (API, error) {
	loader := confita.NewLoader(
		env.NewBackend(),
		newDotEnv(),
	)

	cfg := API{
		Logger: Logger{
			LogLevel:  defaultLogLevel,
			LogFormat: defaultLogFormat,
		},
	}

	if err := loader.Load(context.Background(), &cfg); err != nil {
		return API{}, err
	}

	return cfg, nil
}

type Importer struct {
	Database
	Logger

	BatchSize int    `config:"BATCH_SIZE"`
	CSVFile   string `config:"CSV_FILE"`
}

type Nominatim struct {
	BaseURL string        `config:"NOMINATIM_BASE_URL,required"`
	Timeout time.Duration `config:"NOMINATIM_TIMEOUT"`
}

func LoadImporter() (Importer, error) {
	loader := confita.NewLoader(
		env.NewBackend(),
		newDotEnv(),
	)

	cfg := Importer{
		Logger: Logger{
			LogLevel:  defaultLogLevel,
			LogFormat: defaultLogFormat,
		},
		BatchSize: defaultBatchSize,
	}

	if err := loader.Load(context.Background(), &cfg); err != nil {
		return Importer{}, err
	}

	return cfg, nil
}
