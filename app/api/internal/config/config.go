package config

import (
	"context"
	"time"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend"
	"github.com/heetch/confita/backend/env"
	"github.com/ztimes2/tolqin/app/api/pkg/dotenv"
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

type Nominatim struct {
	BaseURL string        `config:"NOMINATIM_BASE_URL,required"`
	Timeout time.Duration `config:"NOMINATIM_TIMEOUT"`
}

func Load(cfg interface{}) error {
	backends := []backend.Backend{
		env.NewBackend(),
		dotenv.NewBackend(),
	}

	return confita.NewLoader(backends...).Load(context.Background(), cfg)
}
