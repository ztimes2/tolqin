package geo

import (
	"encoding/json"
	"io/ioutil"
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
		name      string
		longitude float64
		expected  bool
	}{
		{
			name:      "return false for longitude less than -180",
			longitude: -181,
			expected:  false,
		},
		{
			name:      "return false for longitude greater than 181",
			longitude: 181,
			expected:  false,
		},
		{
			name:      "return true for longitude between -90 and 90",
			longitude: 180,
			expected:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := IsLongitude(test.longitude)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestIsCountry(t *testing.T) {
	tests := []struct {
		name         string
		codesFn      func(t *testing.T) []string
		expectedBool bool
	}{
		{
			name: "return true for valid iso-2 country codes",
			codesFn: func(t *testing.T) []string {
				b, err := ioutil.ReadFile("testdata/countries.json")
				assert.NoError(t, err)

				c := make(map[string]string)
				err = json.Unmarshal(b, &c)
				assert.NoError(t, err)

				var codes []string
				for code := range c {
					codes = append(codes, code)
				}

				return codes
			},
			expectedBool: true,
		},
		{
			name: "return false for invalid iso-2 codes",
			codesFn: func(_ *testing.T) []string {
				return []string{
					"long",
					"",
					"zz",
					"12",
				}
			},
			expectedBool: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, c := range test.codesFn(t) {
				ok := IsCountry(c)
				assert.Equal(t, test.expectedBool, ok)
			}
		})
	}
}
