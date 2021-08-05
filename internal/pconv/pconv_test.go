package pconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	s := String("test")
	assert.Equal(t, "test", *s)
}
