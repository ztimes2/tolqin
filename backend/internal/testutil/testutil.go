package testutil

import (
	"github.com/stretchr/testify/assert"
)

func IsError(target error) assert.ErrorAssertionFunc {
	return func(tt assert.TestingT, e error, i ...interface{}) bool {
		return assert.ErrorIs(tt, e, target, i...)
	}
}
