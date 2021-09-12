package surfing

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/testutil"
)

type mockSpotStore struct {
	mock.Mock
}

func newMockSpotStore() *mockSpotStore {
	return &mockSpotStore{}
}

func (m *mockSpotStore) Spot(id string) (Spot, error) {
	args := m.Called(id)
	return args.Get(0).(Spot), args.Error(1)
}

func (m *mockSpotStore) Spots(p SpotsParams) ([]Spot, error) {
	args := m.Called(p)
	return args.Get(0).([]Spot), args.Error(1)
}

func TestService_Spot(t *testing.T) {
	tests := []struct {
		name          string
		spotStore     SpotStore
		id            string
		expectedSpot  Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during spot store failure",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("Spot", "1").
					Return(Spot{}, errors.New("something went wrong"))
				return m
			}(),
			id:            "1",
			expectedSpot:  Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return spot using sanitized id without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("Spot", "1").
					Return(
						Spot{
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
			expectedSpot: Spot{
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
						Spot{
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
			expectedSpot: Spot{
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
		expectedSpots  []Spot
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
			expectedErrFn: testutil.IsValidationError("country code"),
		},
		{
			name:      "return error for invalid query",
			spotStore: newMockSpotStore(),
			params: SpotsParams{
				Limit:       20,
				Offset:      0,
				CountryCode: "kz",
				Query:       testutil.RepeatRune('a', 101),
			},
			expectedSpots: nil,
			expectedErrFn: testutil.IsValidationError("query"),
		},
		{
			name:      "return error for invalid north-east coordinates",
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
			expectedErrFn: testutil.IsValidationError("north-east coordinates"),
		},
		{
			name:      "return error for invalid south-west coordinates",
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
			expectedErrFn: testutil.IsValidationError("south-west coordinates"),
		},
		{
			name: "return error during spot spore failure",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("Spots", SpotsParams{
						Limit:  20,
						Offset: 0,
					}).
					Return(([]Spot)(nil), errors.New("something went wrong"))
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
					On("Spots", SpotsParams{
						Limit:       10,
						Offset:      0,
						CountryCode: "kz",
						Query:       "query",
					}).
					Return(
						[]Spot{
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
				Query:       " query ",
			},
			expectedSpots: []Spot{
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
					On("Spots", SpotsParams{
						Limit:       20,
						Offset:      3,
						CountryCode: "kz",
					}).
					Return(
						[]Spot{
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
			expectedSpots: []Spot{
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
