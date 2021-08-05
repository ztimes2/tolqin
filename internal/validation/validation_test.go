package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	err := NewError("test")
	var vErr *Error
	assert.ErrorAs(t, err, &vErr)
	assert.Equal(t, "test", vErr.Field)
}
