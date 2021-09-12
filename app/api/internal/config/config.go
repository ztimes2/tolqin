package config

import (
	"time"
)

const (
	DefaultLogLevel  = "info"
	DefaultLogFormat = "json"
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
