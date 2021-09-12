package config

import (
	"context"

	"github.com/heetch/confita/backend"
	"github.com/joho/godotenv"
)

// DotEnv implements Backend interface of the github.com/heetch/confita/backend
// package and provides functionality for loading and fetching configuration
// variables from a .env file.
type DotEnv struct {
	values map[string]string
}

// NewDotEnv initializes a new dotEnv.
func NewDotEnv() *DotEnv {
	values, _ := godotenv.Read()
	return &DotEnv{
		values: values,
	}
}

// Get fetches a configuration variable by its key from a .env file.
func (d DotEnv) Get(ctx context.Context, key string) ([]byte, error) {
	if d.values == nil {
		return nil, backend.ErrNotFound
	}

	value, ok := d.values[key]
	if !ok {
		return nil, backend.ErrNotFound
	}

	return []byte(value), nil
}

// Name returns a name of this specific Backend's implementation.
func (d DotEnv) Name() string {
	return ".env"
}
