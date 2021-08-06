package surfing

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ztimes2/tolqin/internal/geo"
	"github.com/ztimes2/tolqin/internal/pconv"
	"github.com/ztimes2/tolqin/internal/testutil"
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

func (m *mockSpotStore) CreateSpot(p CreateLocalizedSpotParams) (Spot, error) {
	args := m.Called(p)
	return args.Get(0).(Spot), args.Error(1)
}

func (m *mockSpotStore) UpdateSpot(p UpdateLocalizedSpotParams) (Spot, error) {
	args := m.Called(p)
	return args.Get(0).(Spot), args.Error(1)
}

func (m *mockSpotStore) DeleteSpot(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type mockLocationSource struct {
	mock.Mock
}

func newMockLocationSource() *mockLocationSource {
	return &mockLocationSource{}
}

func (m *mockLocationSource) Location(c geo.Coordinates) (geo.Location, error) {
	args := m.Called(c)
	return args.Get(0).(geo.Location), args.Error(1)
}

func TestService_Spot(t *testing.T) {
	tests := []struct {
		name           string
		spotStore      SpotStore
		locationSource geo.LocationSource
		id             string
		expectedSpot   Spot
		expectedErrFn  assert.ErrorAssertionFunc
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
			locationSource: newMockLocationSource(),
			id:             "1",
			expectedSpot:   Spot{},
			expectedErrFn:  assert.Error,
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
								CountryCode: "Country code 1",
							},
							ID:        "1",
							Name:      "Spot 1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			locationSource: newMockLocationSource(),
			id:             " 1 ",
			expectedSpot: Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "Country code 1",
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
								CountryCode: "Country code 1",
							},
							ID:        "1",
							Name:      "Spot 1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			locationSource: newMockLocationSource(),
			id:             "1",
			expectedSpot: Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "Country code 1",
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
			s := NewService(test.spotStore, test.locationSource)

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
			locationSource: newMockLocationSource(),
			params: SpotsParams{
				Limit:  20,
				Offset: 0,
			},
			expectedSpots: nil,
			expectedErrFn: assert.Error,
		},
		{
			name:           "return error for invalid north west latitude",
			spotStore:      newMockSpotStore(),
			locationSource: newMockLocationSource(),
			params: SpotsParams{
				Limit:  20,
				Offset: 0,
				CoordinateRange: &geo.CoordinateRange{
					NorthWest: geo.Coordinates{
						Latitude:  -91,
						Longitude: 180,
					},
					SouthEast: geo.Coordinates{
						Latitude:  90,
						Longitude: 180,
					},
				},
			},
			expectedSpots: nil,
			expectedErrFn: testutil.IsValidationError("north west latitude"),
		},
		{
			name:           "return error for invalid north west longitude",
			spotStore:      newMockSpotStore(),
			locationSource: newMockLocationSource(),
			params: SpotsParams{
				Limit:  20,
				Offset: 0,
				CoordinateRange: &geo.CoordinateRange{
					NorthWest: geo.Coordinates{
						Latitude:  90,
						Longitude: 181,
					},
					SouthEast: geo.Coordinates{
						Latitude:  90,
						Longitude: 180,
					},
				},
			},
			expectedSpots: nil,
			expectedErrFn: testutil.IsValidationError("north west longitude"),
		},
		{
			name:           "return error for invalid south east latitude",
			spotStore:      newMockSpotStore(),
			locationSource: newMockLocationSource(),
			params: SpotsParams{
				Limit:  20,
				Offset: 0,
				CoordinateRange: &geo.CoordinateRange{
					NorthWest: geo.Coordinates{
						Latitude:  90,
						Longitude: 180,
					},
					SouthEast: geo.Coordinates{
						Latitude:  91,
						Longitude: 180,
					},
				},
			},
			expectedSpots: nil,
			expectedErrFn: testutil.IsValidationError("south east latitude"),
		},
		{
			name:           "return error for invalid south east longitude",
			spotStore:      newMockSpotStore(),
			locationSource: newMockLocationSource(),
			params: SpotsParams{
				Limit:  20,
				Offset: 0,
				CoordinateRange: &geo.CoordinateRange{
					NorthWest: geo.Coordinates{
						Latitude:  90,
						Longitude: 180,
					},
					SouthEast: geo.Coordinates{
						Latitude:  90,
						Longitude: -181,
					},
				},
			},
			expectedSpots: nil,
			expectedErrFn: testutil.IsValidationError("south east longitude"),
		},
		{
			name: "return spots using sanitized params without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("Spots", SpotsParams{
						Limit:       10,
						Offset:      0,
						CountryCode: "Country code 1",
						CoordinateRange: &geo.CoordinateRange{
							NorthWest: geo.Coordinates{
								Latitude:  90,
								Longitude: 180,
							},
							SouthEast: geo.Coordinates{
								Latitude:  -90,
								Longitude: -180,
							},
						},
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
									CountryCode: "Country code 1",
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
									CountryCode: "Country code 1",
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
			locationSource: newMockLocationSource(),
			params: SpotsParams{
				Limit:       0,
				Offset:      -1,
				CountryCode: " Country code 1 ",
				CoordinateRange: &geo.CoordinateRange{
					NorthWest: geo.Coordinates{
						Latitude:  90,
						Longitude: 180,
					},
					SouthEast: geo.Coordinates{
						Latitude:  -90,
						Longitude: -180,
					},
				},
			},
			expectedSpots: []Spot{
				{
					Location: geo.Location{
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
						Locality:    "Locality 1",
						CountryCode: "Country code 1",
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
						CountryCode: "Country code 1",
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
						CountryCode: "Country code 1",
						CoordinateRange: &geo.CoordinateRange{
							NorthWest: geo.Coordinates{
								Latitude:  90,
								Longitude: 180,
							},
							SouthEast: geo.Coordinates{
								Latitude:  -90,
								Longitude: -180,
							},
						},
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
									CountryCode: "Country code 1",
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
									CountryCode: "Country code 1",
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
			locationSource: newMockLocationSource(),
			params: SpotsParams{
				Limit:       20,
				Offset:      3,
				CountryCode: "Country code 1",
				CoordinateRange: &geo.CoordinateRange{
					NorthWest: geo.Coordinates{
						Latitude:  90,
						Longitude: 180,
					},
					SouthEast: geo.Coordinates{
						Latitude:  -90,
						Longitude: -180,
					},
				},
			},
			expectedSpots: []Spot{
				{
					Location: geo.Location{
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
						Locality:    "Locality 1",
						CountryCode: "Country code 1",
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
						CountryCode: "Country code 1",
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
			s := NewService(test.spotStore, test.locationSource)

			spots, err := s.Spots(test.params)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpots, spots)
		})
	}
}

func TestService_CreateSpot(t *testing.T) {
	tests := []struct {
		name           string
		spotStore      SpotStore
		locationSource geo.LocationSource
		params         CreateSpotParams
		expectedSpot   Spot
		expectedErrFn  assert.ErrorAssertionFunc
	}{
		{
			name:           "return error for invalid name",
			spotStore:      newMockSpotStore(),
			locationSource: newMockLocationSource(),
			params: CreateSpotParams{
				Name: "",
				Coordinates: geo.Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("name"),
		},
		{
			name:           "return error for invalid latitude",
			spotStore:      newMockSpotStore(),
			locationSource: newMockLocationSource(),
			params: CreateSpotParams{
				Name: "Spot 1",
				Coordinates: geo.Coordinates{
					Latitude:  -91,
					Longitude: 3.21,
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("latitude"),
		},
		{
			name:           "return error for invalid longitide",
			spotStore:      newMockSpotStore(),
			locationSource: newMockLocationSource(),
			params: CreateSpotParams{
				Name: "Spot 1",
				Coordinates: geo.Coordinates{
					Latitude:  1.23,
					Longitude: -181,
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("longitude"),
		},
		{
			name:      "return error during location source failure",
			spotStore: newMockSpotStore(),
			locationSource: func() geo.LocationSource {
				m := newMockLocationSource()
				m.
					On("Location", geo.Coordinates{Latitude: 1.23, Longitude: 3.21}).
					Return(geo.Location{}, errors.New("something went wrong"))
				return m
			}(),
			params: CreateSpotParams{
				Name: "Spot 1",
				Coordinates: geo.Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return error during spot store failure",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("CreateSpot", CreateLocalizedSpotParams{
						Location: geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "Country code 1",
						},
						Name: "Spot 1",
					}).
					Return(Spot{}, errors.New("something went wrong"))
				return m
			}(),
			locationSource: func() geo.LocationSource {
				m := newMockLocationSource()
				m.
					On("Location", geo.Coordinates{Latitude: 1.23, Longitude: 3.21}).
					Return(
						geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "Country code 1",
						},
						nil,
					)
				return m
			}(),
			params: CreateSpotParams{
				Name: "Spot 1",
				Coordinates: geo.Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return non-localized spot without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("CreateSpot", CreateLocalizedSpotParams{
						Location: geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
						},
						Name: "Spot 1",
					}).
					Return(
						Spot{
							Location: geo.Location{
								Coordinates: geo.Coordinates{
									Latitude:  1.23,
									Longitude: 3.21,
								},
							},
							Name:      "Spot 1",
							ID:        "1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			locationSource: func() geo.LocationSource {
				m := newMockLocationSource()
				m.
					On("Location", geo.Coordinates{Latitude: 1.23, Longitude: 3.21}).
					Return(geo.Location{}, geo.ErrLocationNotFound)
				return m
			}(),
			params: CreateSpotParams{
				Name: "Spot 1",
				Coordinates: geo.Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedSpot: Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
				},
				Name:      "Spot 1",
				ID:        "1",
				CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
			},
			expectedErrFn: assert.NoError,
		},
		{
			name: "return spot using sanitized params without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("CreateSpot", CreateLocalizedSpotParams{
						Location: geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "Country code 1",
						},
						Name: "Spot 1",
					}).
					Return(
						Spot{
							Location: geo.Location{
								Coordinates: geo.Coordinates{
									Latitude:  1.23,
									Longitude: 3.21,
								},
								Locality:    "Locality 1",
								CountryCode: "Country code 1",
							},
							Name:      "Spot 1",
							ID:        "1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			locationSource: func() geo.LocationSource {
				m := newMockLocationSource()
				m.
					On("Location", geo.Coordinates{Latitude: 1.23, Longitude: 3.21}).
					Return(
						geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "Country code 1",
						},
						nil,
					)
				return m
			}(),
			params: CreateSpotParams{
				Name: "  Spot 1  ",
				Coordinates: geo.Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedSpot: Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "Country code 1",
				},
				Name:      "Spot 1",
				ID:        "1",
				CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
			},
			expectedErrFn: assert.NoError,
		},
		{
			name: "return spot without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("CreateSpot", CreateLocalizedSpotParams{
						Location: geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "Country code 1",
						},
						Name: "Spot 1",
					}).
					Return(
						Spot{
							Location: geo.Location{
								Coordinates: geo.Coordinates{
									Latitude:  1.23,
									Longitude: 3.21,
								},
								Locality:    "Locality 1",
								CountryCode: "Country code 1",
							},
							Name:      "Spot 1",
							ID:        "1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			locationSource: func() geo.LocationSource {
				m := newMockLocationSource()
				m.
					On("Location", geo.Coordinates{Latitude: 1.23, Longitude: 3.21}).
					Return(
						geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "Country code 1",
						},
						nil,
					)
				return m
			}(),
			params: CreateSpotParams{
				Name: "Spot 1",
				Coordinates: geo.Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedSpot: Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "Country code 1",
				},
				Name:      "Spot 1",
				ID:        "1",
				CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
			},
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewService(test.spotStore, test.locationSource)

			spot, err := s.CreateSpot(test.params)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpot, spot)
		})
	}
}

func TestService_UpdateSpot(t *testing.T) {
	tests := []struct {
		name           string
		spotStore      SpotStore
		locationSource geo.LocationSource
		params         UpdateSpotParams
		expectedSpot   Spot
		expectedErrFn  assert.ErrorAssertionFunc
	}{
		{
			name:           "return error for invalid id",
			spotStore:      newMockSpotStore(),
			locationSource: newMockLocationSource(),
			params: UpdateSpotParams{
				ID:   "",
				Name: pconv.String("Spot 1"),
				Coordinates: &geo.Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("id"),
		},
		{
			name:           "return error for invalid name",
			spotStore:      newMockSpotStore(),
			locationSource: newMockLocationSource(),
			params: UpdateSpotParams{
				ID:   "1",
				Name: pconv.String(""),
				Coordinates: &geo.Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("name"),
		},
		{
			name:           "return error for invalid latitude",
			spotStore:      newMockSpotStore(),
			locationSource: newMockLocationSource(),
			params: UpdateSpotParams{
				ID:   "1",
				Name: pconv.String("Spot 1"),
				Coordinates: &geo.Coordinates{
					Latitude:  -91,
					Longitude: 3.21,
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("latitude"),
		},
		{
			name:           "return error for invalid longitude",
			spotStore:      newMockSpotStore(),
			locationSource: newMockLocationSource(),
			params: UpdateSpotParams{
				ID:   "1",
				Name: pconv.String("Spot 1"),
				Coordinates: &geo.Coordinates{
					Latitude:  1.23,
					Longitude: -181,
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("longitude"),
		},
		{
			name:      "return error during location source failure",
			spotStore: newMockSpotStore(),
			locationSource: func() geo.LocationSource {
				m := newMockLocationSource()
				m.
					On("Location", geo.Coordinates{Latitude: 1.23, Longitude: 3.21}).
					Return(geo.Location{}, errors.New("something went wrong"))
				return m
			}(),
			params: UpdateSpotParams{
				ID:   "1",
				Name: pconv.String("Spot 1"),
				Coordinates: &geo.Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return error during spot store failure",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("UpdateSpot", UpdateLocalizedSpotParams{
						Location: &geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "Country code 1",
						},
						Name: pconv.String("Spot 1"),
						ID:   "1",
					}).
					Return(Spot{}, errors.New("something went wrong"))
				return m
			}(),
			locationSource: func() geo.LocationSource {
				m := newMockLocationSource()
				m.
					On("Location", geo.Coordinates{Latitude: 1.23, Longitude: 3.21}).
					Return(
						geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "Country code 1",
						},
						nil,
					)
				return m
			}(),
			params: UpdateSpotParams{
				ID:   "1",
				Name: pconv.String("Spot 1"),
				Coordinates: &geo.Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return spot for coordinateless params without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("UpdateSpot", UpdateLocalizedSpotParams{
						Name: pconv.String("Spot 1"),
						ID:   "1",
					}).
					Return(
						Spot{
							Location: geo.Location{
								Coordinates: geo.Coordinates{
									Latitude:  1.23,
									Longitude: 3.21,
								},
								Locality:    "Locality 1",
								CountryCode: "Country code 1",
							},
							Name:      "Spot 1",
							ID:        "1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			locationSource: newMockLocationSource(),
			params: UpdateSpotParams{
				ID:   "1",
				Name: pconv.String("Spot 1"),
			},
			expectedSpot: Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "Country code 1",
				},
				Name:      "Spot 1",
				ID:        "1",
				CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
			},
			expectedErrFn: assert.NoError,
		},
		{
			name: "return spot for nameless params without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("UpdateSpot", UpdateLocalizedSpotParams{
						ID: "1",
						Location: &geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "Country code 1",
						},
					}).
					Return(
						Spot{
							Location: geo.Location{
								Coordinates: geo.Coordinates{
									Latitude:  1.23,
									Longitude: 3.21,
								},
								Locality:    "Locality 1",
								CountryCode: "Country code 1",
							},
							Name:      "Spot 1",
							ID:        "1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			locationSource: func() geo.LocationSource {
				m := newMockLocationSource()
				m.
					On("Location", geo.Coordinates{Latitude: 1.23, Longitude: 3.21}).
					Return(
						geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "Country code 1",
						},
						nil,
					)
				return m
			}(),
			params: UpdateSpotParams{
				ID: "1",
				Coordinates: &geo.Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedSpot: Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "Country code 1",
				},
				Name:      "Spot 1",
				ID:        "1",
				CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
			},
			expectedErrFn: assert.NoError,
		},
		{
			name: "return spot using sanitized params without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("UpdateSpot", UpdateLocalizedSpotParams{
						ID: "1",
						Location: &geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "Country code 1",
						},
						Name: pconv.String("Spot 1"),
					}).
					Return(
						Spot{
							Location: geo.Location{
								Coordinates: geo.Coordinates{
									Latitude:  1.23,
									Longitude: 3.21,
								},
								Locality:    "Locality 1",
								CountryCode: "Country code 1",
							},
							Name:      "Spot 1",
							ID:        "1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			locationSource: func() geo.LocationSource {
				m := newMockLocationSource()
				m.
					On("Location", geo.Coordinates{Latitude: 1.23, Longitude: 3.21}).
					Return(
						geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "Country code 1",
						},
						nil,
					)
				return m
			}(),
			params: UpdateSpotParams{
				ID: " 1 ",
				Coordinates: &geo.Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
				Name: pconv.String(" Spot 1 "),
			},
			expectedSpot: Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "Country code 1",
				},
				Name:      "Spot 1",
				ID:        "1",
				CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
			},
			expectedErrFn: assert.NoError,
		},
		{
			name: "return spot without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("UpdateSpot", UpdateLocalizedSpotParams{
						ID: "1",
						Location: &geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "Country code 1",
						},
						Name: pconv.String("Spot 1"),
					}).
					Return(
						Spot{
							Location: geo.Location{
								Coordinates: geo.Coordinates{
									Latitude:  1.23,
									Longitude: 3.21,
								},
								Locality:    "Locality 1",
								CountryCode: "Country code 1",
							},
							Name:      "Spot 1",
							ID:        "1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			locationSource: func() geo.LocationSource {
				m := newMockLocationSource()
				m.
					On("Location", geo.Coordinates{Latitude: 1.23, Longitude: 3.21}).
					Return(
						geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "Country code 1",
						},
						nil,
					)
				return m
			}(),
			params: UpdateSpotParams{
				ID: "1",
				Coordinates: &geo.Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
				Name: pconv.String("Spot 1"),
			},
			expectedSpot: Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "Country code 1",
				},
				Name:      "Spot 1",
				ID:        "1",
				CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
			},
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewService(test.spotStore, test.locationSource)

			spot, err := s.UpdateSpot(test.params)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpot, spot)
		})
	}
}

func TestService_DeleteSpot(t *testing.T) {
	tests := []struct {
		name           string
		spotStore      SpotStore
		locationSource geo.LocationSource
		id             string
		expectedErrFn  assert.ErrorAssertionFunc
	}{
		{
			name: "return error during spot store failure",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("DeleteSpot", "1").
					Return(errors.New("something went wrong"))
				return m
			}(),
			locationSource: newMockLocationSource(),
			id:             "1",
			expectedErrFn:  assert.Error,
		},
		{
			name: "return spot using sanitized id without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("DeleteSpot", "1").
					Return(nil)
				return m
			}(),
			locationSource: newMockLocationSource(),
			id:             " 1 ",
			expectedErrFn:  assert.NoError,
		},
		{
			name: "return spot without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("DeleteSpot", "1").
					Return(nil)
				return m
			}(),
			locationSource: newMockLocationSource(),
			id:             "1",
			expectedErrFn:  assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewService(test.spotStore, test.locationSource)

			err := s.DeleteSpot(test.id)
			test.expectedErrFn(t, err)
		})
	}
}
