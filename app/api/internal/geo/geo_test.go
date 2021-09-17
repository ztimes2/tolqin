package geo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLatitude(t *testing.T) {
	tests := []struct {
		name     string
		latitude float64
		expected bool
	}{
		{
			name:     "return false for latitude less than -90",
			latitude: -91,
			expected: false,
		},
		{
			name:     "return false for latitude greater than 90",
			latitude: 91,
			expected: false,
		},
		{
			name:     "return true for latitude between -90 and 90",
			latitude: 90,
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := IsLatitude(test.latitude)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestIsLongitude(t *testing.T) {
	tests := []struct {
		name     string
		longitude float64
		expected bool
	}{
		{
			name:     "return false for longitude less than -180",
			longitude: -181,
			expected: false,
		},
		{
			name:     "return false for longitude greater than 181",
			longitude: 181,
			expected: false,
		},
		{
			name:     "return true for longitude between -90 and 90",
			longitude: 180,
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := IsLongitude(test.longitude)
			assert.Equal(t, test.expected, actual)
		})
	}
}
