package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	tests := []struct {
		name              string
		err               *Error
		expectedErrString string
		expectedField     string
		expectedDesc      string
	}{
		{
			name:              "create error with default description",
			err:               NewError("test field"),
			expectedErrString: `invalid field: "test field"`,
			expectedField:     "test field",
			expectedDesc:      "Invalid test field.",
		},
		{
			name:              "create error with custom description",
			err:               NewError("test field", WithDescription("Test field is not valid.")),
			expectedErrString: `invalid field: "test field"`,
			expectedField:     "test field",
			expectedDesc:      "Test field is not valid.",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedErrString, test.err.Error())
			assert.Equal(t, test.expectedField, test.err.Field())
			assert.Equal(t, test.expectedDesc, test.err.Description())
		})
	}
}
