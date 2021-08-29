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
			expectedErrFn: testutil.IsValidationError("latitude"),
		},
		{
			name: "return error for invalid latitude",
			coord: Coordinates{
				Latitude:  91,
				Longitude: 0,
			},
			expectedErrFn: testutil.IsValidationError("latitude"),
		},
		{
			name: "return error for invalid longitude",
			coord: Coordinates{
				Latitude:  0,
				Longitude: -181,
			},
			expectedErrFn: testutil.IsValidationError("longitude"),
		},
		{
			name: "return error for invalid longitude",
			coord: Coordinates{
				Latitude:  0,
				Longitude: 181,
			},
			expectedErrFn: testutil.IsValidationError("longitude"),
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

func TestBounds_Validate(t *testing.T) {
	tests := []struct {
		name          string
		bounds        Bounds
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error for invalid north-east latitude",
			bounds: Bounds{
				NorthEast: Coordinates{
					Latitude:  91,
					Longitude: 180,
				},
				SouthWest: Coordinates{
					Latitude:  -90,
					Longitude: -180,
				},
			},
			expectedErrFn: testutil.IsValidationError("north-east coordinates"),
		},
		{
			name: "return error for invalid north-east longitude",
			bounds: Bounds{
				NorthEast: Coordinates{
					Latitude:  90,
					Longitude: 181,
				},
				SouthWest: Coordinates{
					Latitude:  -90,
					Longitude: -180,
				},
			},
			expectedErrFn: testutil.IsValidationError("north-east coordinates"),
		},
		{
			name: "return error for invalid south-west latitude",
			bounds: Bounds{
				NorthEast: Coordinates{
					Latitude:  90,
					Longitude: 180,
				},
				SouthWest: Coordinates{
					Latitude:  -91,
					Longitude: -180,
				},
			},
			expectedErrFn: testutil.IsValidationError("south-west coordinates"),
		},
		{
			name: "return error for invalid south-west longitude",
			bounds: Bounds{
				NorthEast: Coordinates{
					Latitude:  90,
					Longitude: 180,
				},
				SouthWest: Coordinates{
					Latitude:  -90,
					Longitude: -181,
				},
			},
			expectedErrFn: testutil.IsValidationError("south-west coordinates"),
		},
		{
			name: "return no error",
			bounds: Bounds{
				NorthEast: Coordinates{
					Latitude:  90,
					Longitude: 180,
				},
				SouthWest: Coordinates{
					Latitude:  -90,
					Longitude: -180,
				},
			},
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.bounds.Validate()
			test.expectedErrFn(t, err)
		})
	}
}
