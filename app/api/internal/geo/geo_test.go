package geo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/testutil"
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

func TestLocation_Sanitize(t *testing.T) {
	l := Location{
		CountryCode: " kz ",
		Locality:    " Locality 1 ",
	}
	assert.Equal(
		t,
		Location{
			CountryCode: "kz",
			Locality:    "Locality 1",
		},
		l.Sanitize(),
	)
}

func TestLocation_Validate(t *testing.T) {
	tests := []struct {
		name          string
		location      Location
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error for empty locality",
			location: Location{
				Locality:    "",
				CountryCode: "kz",
				Coordinates: Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedErrFn: testutil.IsValidationError("locality"),
		},
		{
			name: "return error for invalid country code",
			location: Location{
				Locality:    "Locality 1",
				CountryCode: "zz",
				Coordinates: Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedErrFn: testutil.IsValidationError("country code"),
		},
		{
			name: "return error for invalid latitude",
			location: Location{
				Locality:    "Locality",
				CountryCode: "kz",
				Coordinates: Coordinates{
					Latitude:  -91,
					Longitude: 3.21,
				},
			},
			expectedErrFn: testutil.IsValidationError("latitude"),
		},
		{
			name: "return error for invalid longitude",
			location: Location{
				Locality:    "Locality",
				CountryCode: "kz",
				Coordinates: Coordinates{
					Latitude:  1.23,
					Longitude: 181,
				},
			},
			expectedErrFn: testutil.IsValidationError("longitude"),
		},
		{
			name: "return no error",
			location: Location{
				Locality:    "Locality",
				CountryCode: "kz",
				Coordinates: Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.location.Validate()
			test.expectedErrFn(t, err)
		})
	}
}
