package testutil

import (
	"github.com/stretchr/testify/assert"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/valerra"
)

func IsError(target error) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, i ...interface{}) bool {
		return assert.Error(t, err) && assert.ErrorIs(t, err, target, i...)
	}
}

func AreValidationErrors(targets ...error) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, i ...interface{}) bool {
		var vErr *valerra.Errors
		return assert.Error(t, err) &&
			assert.ErrorAs(t, err, &vErr) &&
			assert.NotNil(t, vErr) &&
			assert.Equal(t, targets, vErr.Errors())
	}
}
