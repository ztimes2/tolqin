package testutil

import (
	"github.com/stretchr/testify/assert"
)

func IsError(target error) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, i ...interface{}) bool {
		return assert.ErrorIs(t, err, target, i...)
	}
}
