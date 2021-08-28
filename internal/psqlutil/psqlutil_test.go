package psqlutil

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSSLMode(t *testing.T) {
	tests := []struct {
		name            string
		s               string
		expectedSSLMode SSLMode
	}{
		{
			name:            "return ssl mode disabled",
			s:               sslModeNameDisable,
			expectedSSLMode: SSLModeDisabled,
		},
		{
			name:            "return ssl mode undefined for unknown string",
			s:               "unknown",
			expectedSSLMode: SSLModeUndefined,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sslMode := NewSSLMode(test.s)
			assert.Equal(t, test.expectedSSLMode, sslMode)
		})
	}
}

func TestSSLMode_String(t *testing.T) {
	tests := []struct {
		name      string
		sslMode   SSLMode
		expectedS string
	}{
		{
			name:      "return ssl mode disabled string",
			sslMode:   SSLModeDisabled,
			expectedS: sslModeNameDisable,
		},
		{
			name:      "return empty string for ssl mode undefined",
			sslMode:   SSLModeUndefined,
			expectedS: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.sslMode.String()
			assert.Equal(t, test.expectedS, s)
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name              string
		s                 string
		expectedSQLString sql.NullString
	}{
		{
			name:              "return invalid sql string for empty string",
			s:                 "",
			expectedSQLString: sql.NullString{},
		},
		{
			name:              "return valid sql string",
			s:                 "test",
			expectedSQLString: sql.NullString{String: "test", Valid: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sqlString := String(test.s)
			assert.Equal(t, test.expectedSQLString, sqlString)
		})
	}
}

func TestConfig_String(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		expectedS string
	}{
		{
			name: "return string with all entries",
			config: Config{
				Host:         "host",
				Port:         "port",
				DatabaseName: "dbname",
				Username:     "user",
				Password:     "password",
				SSLMode:      SSLModeDisabled,
			},
			expectedS: "host=host port=port dbname=dbname sslmode=disable user=user password=password",
		},
		{
			name: "return string without ssl mode",
			config: Config{
				Host:         "host",
				Port:         "port",
				DatabaseName: "dbname",
				Username:     "user",
				Password:     "password",
			},
			expectedS: "host=host port=port dbname=dbname user=user password=password",
		},
		{
			name: "return string without username",
			config: Config{
				Host:         "host",
				Port:         "port",
				DatabaseName: "dbname",
				Password:     "password",
				SSLMode:      SSLModeDisabled,
			},
			expectedS: "host=host port=port dbname=dbname sslmode=disable password=password",
		},
		{
			name: "return string without password",
			config: Config{
				Host:         "host",
				Port:         "port",
				DatabaseName: "dbname",
				Username:     "user",
				SSLMode:      SSLModeDisabled,
			},
			expectedS: "host=host port=port dbname=dbname sslmode=disable user=user",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := test.config.String()
			assert.Equal(t, test.expectedS, s)
		})
	}
}