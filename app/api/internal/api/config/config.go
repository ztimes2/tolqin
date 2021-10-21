package config

import (
	"context"
	"time"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
	"github.com/heetch/confita/backend/env"
	"github.com/sirupsen/logrus"
	"github.com/ztimes2/tolqin/app/api/pkg/dotenv"
	"github.com/ztimes2/tolqin/app/api/pkg/log"
)

type Config struct {
	Database
	Logger
	Nominatim

	ServerPort string `config:"SERVER_PORT,required"`

	JWTSigningKey string        `config:"JWT_SIGNING_KEY,required"`
	JWTExpiry     time.Duration `config:"JWT_EXPIRY,required"`
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

type Nominatim struct {
	BaseURL string        `config:"NOMINATIM_BASE_URL,required"`
	Timeout time.Duration `config:"NOMINATIM_TIMEOUT"`
}

func Load() (Config, error) {
	cfg := Config{
		Logger: Logger{
			LogLevel:  logrus.InfoLevel.String(),
			LogFormat: log.FormatJSON,
		},
	}

	backends := []backend.Backend{
		env.NewBackend(),
		dotenv.NewBackend(),
	}

	if err := confita.NewLoader(backends...).Load(context.Background(), &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
