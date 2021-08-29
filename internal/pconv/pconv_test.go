package pconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	s := String("test")
	assert.Equal(t, "test", *s)
}

func TestFloat64(t *testing.T) {
	f := Float64(1.23)
	assert.Equal(t, 1.23, *f)
}
