package surfing

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/geo"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/surf"
	"github.com/ztimes2/tolqin/app/api/pkg/strutil"
	"github.com/ztimes2/tolqin/app/api/pkg/testutil"
)

type mockSpotStore struct {
	mock.Mock
}

func newMockSpotStore() *mockSpotStore {
	return &mockSpotStore{}
}

func (m *mockSpotStore) Spot(id string) (surf.Spot, error) {
	args := m.Called(id)
	return args.Get(0).(surf.Spot), args.Error(1)
}

func (m *mockSpotStore) Spots(p surf.SpotsParams) ([]surf.Spot, error) {
	args := m.Called(p)
	return args.Get(0).([]surf.Spot), args.Error(1)
}

func TestService_Spot(t *testing.T) {
	tests := []struct {
		name          string
		spotStore     SpotStore
		id            string
		expectedSpot  surf.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name:          "return error for invalid spot id",
			spotStore:     newMockSpotStore(),
			id:            "",
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidSpotID),
		},
		{
			name: "return error during spot store failure",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("Spot", "1").
					Return(surf.Spot{}, errors.New("something went wrong"))
				return m
			}(),
			id:            "1",
			expectedSpot:  surf.Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return spot using sanitized id without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("Spot", "1").
					Return(
						surf.Spot{
							Location: geo.Location{
								Coordinates: geo.Coordinates{
									Latitude:  1.23,
									Longitude: 3.21,
								},
								Locality:    "Locality 1",
								CountryCode: "kz",
							},
							ID:        "1",
							Name:      "Spot 1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			id: " 1 ",
			expectedSpot: surf.Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "kz",
				},
				ID:        "1",
				Name:      "Spot 1",
				CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
			},
			expectedErrFn: assert.NoError,
		},
		{
			name: "return spot without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("Spot", "1").
					Return(
						surf.Spot{
							Location: geo.Location{
								Coordinates: geo.Coordinates{
									Latitude:  1.23,
									Longitude: 3.21,
								},
								Locality:    "Locality 1",
								CountryCode: "kz",
							},
							ID:        "1",
							Name:      "Spot 1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			id: "1",
			expectedSpot: surf.Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "kz",
				},
				ID:        "1",
				Name:      "Spot 1",
				CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
			},
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewService(test.spotStore)

			spot, err := s.Spot(test.id)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpot, spot)
		})
	}
}

func TestService_Spots(t *testing.T) {
	tests := []struct {
		name           string
		spotStore      SpotStore
		locationSource geo.LocationSource
		params         SpotsParams
		expectedSpots  []surf.Spot
		expectedErrFn  assert.ErrorAssertionFunc
	}{
		{
			name:      "return error for invalid country code",
			spotStore: newMockSpotStore(),
			params: SpotsParams{
				Limit:       20,
				Offset:      0,
				CountryCode: "invalid",
			},
			expectedSpots: nil,
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidCountryCode),
		},
		{
			name:      "return error for invalid query",
			spotStore: newMockSpotStore(),
			params: SpotsParams{
				Limit:       20,
				Offset:      0,
				CountryCode: "kz",
				SearchQuery: strutil.RepeatRune('a', 101),
			},
			expectedSpots: nil,
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidSearchQuery),
		},
		{
			name:      "return error for invalid north-east latitude",
			spotStore: newMockSpotStore(),
			params: SpotsParams{
				Limit:  20,
				Offset: 0,
				Bounds: &geo.Bounds{
					NorthEast: geo.Coordinates{
						Latitude:  91,
						Longitude: 180,
					},
					SouthWest: geo.Coordinates{
						Latitude:  -90,
						Longitude: -180,
					},
				},
			},
			expectedSpots: nil,
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidNorthEastLatitude),
		},
		{
			name:      "return error for invalid north-east longitude",
			spotStore: newMockSpotStore(),
			params: SpotsParams{
				Limit:  20,
				Offset: 0,
				Bounds: &geo.Bounds{
					NorthEast: geo.Coordinates{
						Latitude:  90,
						Longitude: 181,
					},
					SouthWest: geo.Coordinates{
						Latitude:  -90,
						Longitude: -180,
					},
				},
			},
			expectedSpots: nil,
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidNorthEastLongitude),
		},
		{
			name:      "return error for invalid south-west latitude",
			spotStore: newMockSpotStore(),
			params: SpotsParams{
				Limit:  20,
				Offset: 0,
				Bounds: &geo.Bounds{
					NorthEast: geo.Coordinates{
						Latitude:  90,
						Longitude: 180,
					},
					SouthWest: geo.Coordinates{
						Latitude:  -91,
						Longitude: -180,
					},
				},
			},
			expectedSpots: nil,
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidSouthWestLatitude),
		},
		{
			name:      "return error for invalid south-west longitude",
			spotStore: newMockSpotStore(),
			params: SpotsParams{
				Limit:  20,
				Offset: 0,
				Bounds: &geo.Bounds{
					NorthEast: geo.Coordinates{
						Latitude:  90,
						Longitude: 180,
					},
					SouthWest: geo.Coordinates{
						Latitude:  -90,
						Longitude: -181,
					},
				},
			},
			expectedSpots: nil,
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidSouthWestLongitude),
		},
		{
			name: "return error during spot spore failure",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("Spots", surf.SpotsParams{
						Limit:  20,
						Offset: 0,
					}).
					Return(([]surf.Spot)(nil), errors.New("something went wrong"))
				return m
			}(),
			params: SpotsParams{
				Limit:  20,
				Offset: 0,
			},
			expectedSpots: nil,
			expectedErrFn: assert.Error,
		},
		{
			name: "return spots using sanitized params without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("Spots", surf.SpotsParams{
						Limit:       10,
						Offset:      0,
						CountryCode: "kz",
						SearchQuery: surf.SpotSearchQuery{
							Query: "query",
						},
					}).
					Return(
						[]surf.Spot{
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  1.23,
										Longitude: 3.21,
									},
									Locality:    "Locality 1",
									CountryCode: "kz",
								},
								ID:        "1",
								Name:      "Spot 1",
								CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
							},
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  1.23,
										Longitude: 3.21,
									},
									Locality:    "Locality 2",
									CountryCode: "kz",
								},
								ID:        "2",
								Name:      "Spot 2",
								CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
							},
						},
						nil,
					)
				return m
			}(),
			params: SpotsParams{
				Limit:       0,
				Offset:      -1,
				CountryCode: " kz ",
				SearchQuery: " query ",
			},
			expectedSpots: []surf.Spot{
				{
					Location: geo.Location{
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
						Locality:    "Locality 1",
						CountryCode: "kz",
					},
					ID:        "1",
					Name:      "Spot 1",
					CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
				},
				{
					Location: geo.Location{
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
						Locality:    "Locality 2",
						CountryCode: "kz",
					},
					ID:        "2",
					Name:      "Spot 2",
					CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
				},
			},
			expectedErrFn: assert.NoError,
		},
		{
			name: "return spots without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("Spots", surf.SpotsParams{
						Limit:       20,
						Offset:      3,
						CountryCode: "kz",
					}).
					Return(
						[]surf.Spot{
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  1.23,
										Longitude: 3.21,
									},
									Locality:    "Locality 1",
									CountryCode: "kz",
								},
								ID:        "1",
								Name:      "Spot 1",
								CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
							},
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  1.23,
										Longitude: 3.21,
									},
									Locality:    "Locality 2",
									CountryCode: "kz",
								},
								ID:        "2",
								Name:      "Spot 2",
								CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
							},
						},
						nil,
					)
				return m
			}(),
			params: SpotsParams{
				Limit:       20,
				Offset:      3,
				CountryCode: "kz",
			},
			expectedSpots: []surf.Spot{
				{
					Location: geo.Location{
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
						Locality:    "Locality 1",
						CountryCode: "kz",
					},
					ID:        "1",
					Name:      "Spot 1",
					CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
				},
				{
					Location: geo.Location{
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
						Locality:    "Locality 2",
						CountryCode: "kz",
					},
					ID:        "2",
					Name:      "Spot 2",
					CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
				},
			},
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewService(test.spotStore)

			spots, err := s.Spots(test.params)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpots, spots)
		})
	}
}
