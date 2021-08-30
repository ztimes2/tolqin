package psqlutil

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

const (
	CodeInvalidTextRepresentation = "22P02"
)

const (
	driverName = "postgres"

	sslModeNameDisable = "disable"
)

type Config struct {
	Host         string
	Port         string
	Username     string
	Password     string
	DatabaseName string
	SSLMode      SSLMode
}

func (c Config) String() string {
	entries := []string{
		"host=" + c.Host,
		"port=" + c.Port,
		"dbname=" + c.DatabaseName,
	}
	if c.SSLMode != SSLModeUndefined {
		entries = append(entries, "sslmode="+c.SSLMode.String())
	}
	if c.Username != "" {
		entries = append(entries, "user="+c.Username)
	}
	if c.Password != "" {
		entries = append(entries, "password="+c.Password)
	}
	return strings.Join(entries, " ")
}

type SSLMode int

const (
	SSLModeUndefined SSLMode = iota
	SSLModeDisabled
)

func NewSSLMode(s string) SSLMode {
	switch s {
	case sslModeNameDisable:
		return SSLModeDisabled
	default:
		return SSLModeUndefined
	}
}

func (s SSLMode) String() string {
	switch s {
	case SSLModeDisabled:
		return sslModeNameDisable
	default:
		return ""
	}
}

func NewDB(cfg Config) (*sqlx.DB, error) {
	return sqlx.Open(driverName, cfg.String())
}

func WrapDB(db *sql.DB) *sqlx.DB {
	return sqlx.NewDb(db, driverName)
}

func NewQueryBuilder() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

func Wildcard(s string) string {
	return "%" + s + "%"
}

func Between(key string, min, max float64) sq.Sqlizer {
	return sq.Expr(fmt.Sprintf("%s BETWEEN ? AND ?", key), min, max)
}

func IsInvalidTextRepresenationError(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	if pqErr.Code != CodeInvalidTextRepresentation {
		return false
	}
	return true
}
