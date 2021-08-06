package geo

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
