package geo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ztimes2/tolqin/internal/testutil"
)

func TestCoordinates_Validate(t *testing.T) {
	tests := []struct {
		name          string
		coord         Coordinates
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error for invalid latitude",
			coord: Coordinates{
				Latitude:  -91,
				Longitude: 0,
			},
			expectedErrFn: testutil.IsValidationError("coordinates"),
		},
		{
			name: "return error for invalid latitude",
			coord: Coordinates{
				Latitude:  91,
				Longitude: 0,
			},
			expectedErrFn: testutil.IsValidationError("coordinates"),
		},
		{
			name: "return error for invalid longitude",
			coord: Coordinates{
				Latitude:  0,
				Longitude: -181,
			},
			expectedErrFn: testutil.IsValidationError("coordinates"),
		},
		{
			name: "return error for invalid longitude",
			coord: Coordinates{
				Latitude:  0,
				Longitude: 181,
			},
			expectedErrFn: testutil.IsValidationError("coordinates"),
		},
		{
			name: "return no error",
			coord: Coordinates{
				Latitude:  2,
				Longitude: 70,
			},
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.coord.Validate()
			test.expectedErrFn(t, err)
		})
	}
}
