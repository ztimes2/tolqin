package testutil

import (
	"github.com/stretchr/testify/assert"
	"github.com/ztimes2/tolqin/app/api/pkg/valerra"
)

// IsError returns github.com/stretchr/testify package's assert.ErrorAssertionFunc
// that checks if at least one of the errors in the chain matches target.
func IsError(target error) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, i ...interface{}) bool {
		return assert.Error(t, err) && assert.ErrorIs(t, err, target, i...)
	}
}

// AreValidationErrors returns github.com/stretchr/testify package's assert.ErrorAssertionFunc
// that checks if at least one of the errors in the chain is of type *valerra.Errors
// and it contains the given targets.
func AreValidationErrors(targets ...error) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, i ...interface{}) bool {
		var vErr *valerra.Errors
		return assert.Error(t, err) &&
			assert.ErrorAs(t, err, &vErr) &&
			assert.NotNil(t, vErr) &&
			assert.Equal(t, targets, vErr.Errors())
	}
}
