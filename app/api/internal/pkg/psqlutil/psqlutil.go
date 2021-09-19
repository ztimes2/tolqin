package psqlutil

import (
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

const (
	// DriverNamePQ is github.com/lib/pq package's driver name.
	DriverNamePQ = "postgres"

	// DriverNameSQLMock is github.com/DATA-DOG/go-sqlmock package's driver name.
	DriverNameSQLMock = "sqlmock"
)

const (
	sslModeNameDisable = "disable"
)

// NewDB opens a new github.com/jmoiron/sqlx *sqlx.DB using the given database
// driver name and configuration.
//
// The caller is expected to register a PostgreSQL driver to the standard database/sql
// package prior to envoking this function.
func NewDB(driverName string, cfg Config) (*sqlx.DB, error) {
	return sqlx.Open(driverName, cfg.String())
}

// Config holds configuration for connecting to a PostgreSQL database.
type Config struct {
	Host         string
	Port         string
	Username     string
	Password     string
	DatabaseName string
	SSLMode      SSLMode
}

// String returns the confiration as a DSN string.
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

// SSLMode represents PostgreSQL's SSL mode.
type SSLMode int

const (
	// SSLModeUndefined is used as blank SSL mode.
	SSLModeUndefined SSLMode = iota

	// SSLModeDisabled is used to disable SSL mode.
	//
	// As per PostgreSQL's documentation:
	//	"I don't care about security, and I don't want to pay the overhead of encryption."
	SSLModeDisabled

	// TODO add more SSL modes if necessary.
)

// NewSSLMode parses SSLMode from the given string.
//
// The accepted value is only "disable" so far. Any other values return SSLModeUndefined.
func NewSSLMode(s string) SSLMode {
	switch s {
	case sslModeNameDisable:
		return SSLModeDisabled
	default:
		return SSLModeUndefined
	}
}

// String returns string representation of the SSLMode.
func (s SSLMode) String() string {
	switch s {
	case SSLModeDisabled:
		return sslModeNameDisable
	default:
		return ""
	}
}

// NewQueryBuilder returns a new github.com/Masterminds/squirrel query builder for
// PostgreSQL.
func NewQueryBuilder() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

// Wildcard surrounds the given string with PostgreSQL's wildcards.
func Wildcard(s string) string {
	return "%" + s + "%"
}

// Between returns a github.com/Masterminds/squirrel expression for PostgreSQL's
// BETWEEN clause using the given arguments.
func Between(key string, min, max float64) sq.Sqlizer {
	return sq.Expr(fmt.Sprintf("%s BETWEEN ? AND ?", key), min, max)
}

// CastAsVarchar returns PostgreSQL's CAST clause for casting the given key as
// VARCHAR.
func CastAsVarchar(key string) string {
	return fmt.Sprintf("CAST(%s AS VARCHAR)", key)
}
