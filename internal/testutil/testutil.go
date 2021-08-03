package testutil

import (
	"github.com/stretchr/testify/assert"
	"github.com/ztimes2/tolqin/internal/validation"
)

func IsError(target error) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, i ...interface{}) bool {
		return assert.Error(t, err) && assert.ErrorIs(t, err, target, i...)
	}
}

func IsValidationError(field string) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, i ...interface{}) bool {
		var vErr *validation.Error
		return assert.Error(t, err) &&
			assert.ErrorAs(t, err, &vErr) &&
			assert.NotNil(t, vErr) &&
			assert.Equal(t, field, vErr.Field)
	}
}
