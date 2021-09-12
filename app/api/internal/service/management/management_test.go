package management

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/pconv"
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

func (m *mockSpotStore) CreateSpot(p CreateSpotParams) (Spot, error) {
	args := m.Called(p)
	return args.Get(0).(Spot), args.Error(1)
}

func (m *mockSpotStore) UpdateSpot(p UpdateSpotParams) (Spot, error) {
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
			s := NewService(test.spotStore, newMockLocationSource())

			spot, err := s.Spot(test.id)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpot, spot)
		})
	}
}

func TestService_Spots(t *testing.T) {
	tests := []struct {
		name          string
		spotStore     SpotStore
		params        SpotsParams
		expectedSpots []Spot
		expectedErrFn assert.ErrorAssertionFunc
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
			s := NewService(test.spotStore, newMockLocationSource())

			spots, err := s.Spots(test.params)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpots, spots)
		})
	}
}

func TestService_CreateSpot(t *testing.T) {
	tests := []struct {
		name          string
		spotStore     SpotStore
		params        CreateSpotParams
		expectedSpot  Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name:      "return error for invalid name",
			spotStore: newMockSpotStore(),
			params: CreateSpotParams{
				Name: "",
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "kz",
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("name"),
		},
		{
			name:      "return error for invalid latitude",
			spotStore: newMockSpotStore(),
			params: CreateSpotParams{
				Name: "Spot 1",
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  -91,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "kz",
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("latitude"),
		},
		{
			name:      "return error for invalid longitide",
			spotStore: newMockSpotStore(),
			params: CreateSpotParams{
				Name: "Spot 1",
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 181,
					},
					Locality:    "Locality 1",
					CountryCode: "kz",
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("longitude"),
		},
		{
			name:      "return error for invalid locality",
			spotStore: newMockSpotStore(),
			params: CreateSpotParams{
				Name: "Spot 1",
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 181,
					},
					Locality:    "",
					CountryCode: "kz",
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("locality"),
		},
		{
			name:      "return error for invalid country code",
			spotStore: newMockSpotStore(),
			params: CreateSpotParams{
				Name: "Spot 1",
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 181,
					},
					Locality:    "Locality 1",
					CountryCode: "zz",
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("country code"),
		},
		{
			name: "return error during spot store failure",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("CreateSpot", CreateSpotParams{
						Location: geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "kz",
						},
						Name: "Spot 1",
					}).
					Return(Spot{}, errors.New("something went wrong"))
				return m
			}(),
			params: CreateSpotParams{
				Name: "Spot 1",
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "kz",
				},
			},
			expectedSpot:  Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return spot using sanitized params without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("CreateSpot", CreateSpotParams{
						Location: geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "kz",
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
								CountryCode: "kz",
							},
							Name:      "Spot 1",
							ID:        "1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			params: CreateSpotParams{
				Name: "  Spot 1  ",
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    " Locality 1 ",
					CountryCode: " kz ",
				},
			},
			expectedSpot: Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "kz",
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
					On("CreateSpot", CreateSpotParams{
						Location: geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "kz",
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
								CountryCode: "kz",
							},
							Name:      "Spot 1",
							ID:        "1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			params: CreateSpotParams{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "kz",
				},
				Name: "Spot 1",
			},
			expectedSpot: Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "kz",
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
			s := NewService(test.spotStore, newMockLocationSource())

			spot, err := s.CreateSpot(test.params)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpot, spot)
		})
	}
}

func TestService_UpdateSpot(t *testing.T) {
	tests := []struct {
		name          string
		spotStore     SpotStore
		params        UpdateSpotParams
		expectedSpot  Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name:      "return error for invalid id",
			spotStore: newMockSpotStore(),
			params: UpdateSpotParams{
				ID:   "",
				Name: pconv.String("Spot 1"),
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("id"),
		},
		{
			name:      "return error for invalid name",
			spotStore: newMockSpotStore(),
			params: UpdateSpotParams{
				ID:   "1",
				Name: pconv.String(""),
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("name"),
		},
		{
			name:      "return error for invalid latitude",
			spotStore: newMockSpotStore(),
			params: UpdateSpotParams{
				ID:       "1",
				Name:     pconv.String("Spot 1"),
				Latitude: pconv.Float64(-91),
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("latitude"),
		},
		{
			name:      "return error for invalid longitude",
			spotStore: newMockSpotStore(),
			params: UpdateSpotParams{
				ID:        "1",
				Name:      pconv.String("Spot 1"),
				Latitude:  pconv.Float64(1.23),
				Longitude: pconv.Float64(-181),
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("longitude"),
		},
		{
			name:      "return error for invalid locality",
			spotStore: newMockSpotStore(),
			params: UpdateSpotParams{
				ID:        "1",
				Name:      pconv.String("Spot 1"),
				Latitude:  pconv.Float64(1.23),
				Longitude: pconv.Float64(2.34),
				Locality:  pconv.String(""),
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("locality"),
		},
		{
			name:      "return error for invalid country code",
			spotStore: newMockSpotStore(),
			params: UpdateSpotParams{
				ID:          "1",
				Name:        pconv.String("Spot 1"),
				Latitude:    pconv.Float64(1.23),
				Longitude:   pconv.Float64(2.34),
				Locality:    pconv.String("Locality 1"),
				CountryCode: pconv.String("zz"),
			},
			expectedSpot:  Spot{},
			expectedErrFn: testutil.IsValidationError("country code"),
		},
		{
			name: "return error during spot store failure",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("UpdateSpot", UpdateSpotParams{
						Latitude:    pconv.Float64(1.23),
						Longitude:   pconv.Float64(2.34),
						Locality:    pconv.String("Locality 1"),
						CountryCode: pconv.String("kz"),
						Name:        pconv.String("Spot 1"),
						ID:          "1",
					}).
					Return(Spot{}, errors.New("something went wrong"))
				return m
			}(),
			params: UpdateSpotParams{
				ID:          "1",
				Name:        pconv.String("Spot 1"),
				Latitude:    pconv.Float64(1.23),
				Longitude:   pconv.Float64(2.34),
				Locality:    pconv.String("Locality 1"),
				CountryCode: pconv.String("zz"),
			},
			expectedSpot:  Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return spot for coordinateless params without error",
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("UpdateSpot", UpdateSpotParams{
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
								CountryCode: "kz",
							},
							Name:      "Spot 1",
							ID:        "1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
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
					CountryCode: "kz",
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
					On("UpdateSpot", UpdateSpotParams{
						ID:       "1",
						Latitude: pconv.Float64(1.23),
						Locality: pconv.String("Locality 1"),
					}).
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
							Name:      "Spot 1",
							ID:        "1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			params: UpdateSpotParams{
				ID:       "1",
				Latitude: pconv.Float64(1.23),
				Locality: pconv.String("Locality 1"),
			},
			expectedSpot: Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "kz",
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
					On("UpdateSpot", UpdateSpotParams{
						ID:          "1",
						Latitude:    pconv.Float64(1.23),
						Longitude:   pconv.Float64(2.34),
						Locality:    pconv.String("Locality 1"),
						CountryCode: pconv.String("kz"),
						Name:        pconv.String("Spot 1"),
					}).
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
							Name:      "Spot 1",
							ID:        "1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			params: UpdateSpotParams{
				ID:          " 1 ",
				Latitude:    pconv.Float64(1.23),
				Longitude:   pconv.Float64(2.34),
				Locality:    pconv.String(" Locality 1 "),
				CountryCode: pconv.String(" kz "),
				Name:        pconv.String(" Spot 1 "),
			},
			expectedSpot: Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "kz",
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
					On("UpdateSpot", UpdateSpotParams{
						ID:          "1",
						Latitude:    pconv.Float64(1.23),
						Longitude:   pconv.Float64(2.34),
						Locality:    pconv.String("Locality 1"),
						CountryCode: pconv.String("kz"),
						Name:        pconv.String("Spot 1"),
					}).
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
							Name:      "Spot 1",
							ID:        "1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			params: UpdateSpotParams{
				ID:          "1",
				Latitude:    pconv.Float64(1.23),
				Longitude:   pconv.Float64(2.34),
				Locality:    pconv.String("Locality 1"),
				CountryCode: pconv.String("kz"),
				Name:        pconv.String("Spot 1"),
			},
			expectedSpot: Spot{
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
					Locality:    "Locality 1",
					CountryCode: "kz",
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
			s := NewService(test.spotStore, newMockLocationSource())

			spot, err := s.UpdateSpot(test.params)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpot, spot)
		})
	}
}

func TestService_DeleteSpot(t *testing.T) {
	tests := []struct {
		name          string
		spotStore     SpotStore
		id            string
		expectedErrFn assert.ErrorAssertionFunc
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
			id:            "1",
			expectedErrFn: assert.Error,
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
			id:            " 1 ",
			expectedErrFn: assert.NoError,
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
			id:            "1",
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewService(test.spotStore, newMockLocationSource())

			err := s.DeleteSpot(test.id)
			test.expectedErrFn(t, err)
		})
	}
}

func TestService_Location(t *testing.T) {
	tests := []struct {
		name             string
		locationSource   geo.LocationSource
		coord            geo.Coordinates
		expectedLocation geo.Location
		expectedErrFn    assert.ErrorAssertionFunc
	}{
		{
			name:           "return error for invalid latitude",
			locationSource: newMockLocationSource(),
			coord: geo.Coordinates{
				Latitude:  -91,
				Longitude: 180,
			},
			expectedLocation: geo.Location{},
			expectedErrFn:    testutil.IsValidationError("latitude"),
		},
		{
			name:           "return error for invalid longitude",
			locationSource: newMockLocationSource(),
			coord: geo.Coordinates{
				Latitude:  -90,
				Longitude: 181,
			},
			expectedLocation: geo.Location{},
			expectedErrFn:    testutil.IsValidationError("longitude"),
		},
		{
			name: "return error during unexpected location source failure",
			locationSource: func() geo.LocationSource {
				m := newMockLocationSource()
				m.
					On("Location", geo.Coordinates{
						Latitude:  -90,
						Longitude: 180,
					}).
					Return(geo.Location{}, errors.New("something went wrong"))
				return m
			}(),
			coord: geo.Coordinates{
				Latitude:  -90,
				Longitude: 180,
			},
			expectedLocation: geo.Location{},
			expectedErrFn:    assert.Error,
		},
		{
			name: "return error when location is not found",
			locationSource: func() geo.LocationSource {
				m := newMockLocationSource()
				m.
					On("Location", geo.Coordinates{
						Latitude:  -90,
						Longitude: 180,
					}).
					Return(geo.Location{}, geo.ErrLocationNotFound)
				return m
			}(),
			coord: geo.Coordinates{
				Latitude:  -90,
				Longitude: 180,
			},
			expectedLocation: geo.Location{},
			expectedErrFn:    testutil.IsError(ErrNotFound),
		},
		{
			name: "return location without error",
			locationSource: func() geo.LocationSource {
				m := newMockLocationSource()
				m.
					On("Location", geo.Coordinates{
						Latitude:  -90,
						Longitude: 180,
					}).
					Return(
						geo.Location{
							Locality:    "Locality 1",
							CountryCode: "kz",
							Coordinates: geo.Coordinates{
								Latitude:  -90,
								Longitude: 180,
							},
						},
						nil,
					)
				return m
			}(),
			coord: geo.Coordinates{
				Latitude:  -90,
				Longitude: 180,
			},
			expectedLocation: geo.Location{
				Locality:    "Locality 1",
				CountryCode: "kz",
				Coordinates: geo.Coordinates{
					Latitude:  -90,
					Longitude: 180,
				},
			},
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewService(newMockSpotStore(), test.locationSource)

			l, err := s.Location(test.coord)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedLocation, l)
		})
	}
}
