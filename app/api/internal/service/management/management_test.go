package management

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ztimes2/tolqin/app/api/internal/auth"
	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/jwt"
	"github.com/ztimes2/tolqin/app/api/internal/surf"
	"github.com/ztimes2/tolqin/app/api/pkg/pconv"
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

func (m *mockSpotStore) CreateSpot(p surf.SpotCreationEntry) (surf.Spot, error) {
	args := m.Called(p)
	return args.Get(0).(surf.Spot), args.Error(1)
}

func (m *mockSpotStore) UpdateSpot(p surf.SpotUpdateEntry) (surf.Spot, error) {
	args := m.Called(p)
	return args.Get(0).(surf.Spot), args.Error(1)
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
		ctxFn         func() context.Context
		spotStore     SpotStore
		id            string
		expectedSpot  surf.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error for unauthenticated request",
			ctxFn: func() context.Context {
				return context.Background()
			},
			spotStore:     newMockSpotStore(),
			id:            "",
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.IsError(jwt.ErrClaimsNotFound),
		},
		{
			name: "return error for unauthorized request",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: "",
				})
			},
			spotStore:     newMockSpotStore(),
			id:            "",
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.IsError(jwt.ErrMismatchedRole),
		},
		{
			name: "return error for invalid spot id",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore:     newMockSpotStore(),
			id:            "",
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidSpotID),
		},
		{
			name: "return error during spot store failure",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			s := NewService(test.spotStore, newMockLocationSource())

			spot, err := s.Spot(test.ctxFn(), test.id)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpot, spot)
		})
	}
}

func TestService_Spots(t *testing.T) {
	tests := []struct {
		name          string
		ctxFn         func() context.Context
		spotStore     SpotStore
		params        SpotsParams
		expectedSpots []surf.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error for unauthenticated request",
			ctxFn: func() context.Context {
				return context.Background()
			},
			spotStore: newMockSpotStore(),
			params: SpotsParams{
				Limit:       20,
				Offset:      0,
				CountryCode: "invalid",
			},
			expectedSpots: nil,
			expectedErrFn: testutil.IsError(jwt.ErrClaimsNotFound),
		},
		{
			name: "return error for unauthorized request",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: "",
				})
			},
			spotStore: newMockSpotStore(),
			params: SpotsParams{
				Limit:       20,
				Offset:      0,
				CountryCode: "invalid",
			},
			expectedSpots: nil,
			expectedErrFn: testutil.IsError(jwt.ErrMismatchedRole),
		},
		{
			name: "return error for invalid country code",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			name: "return error for invalid query",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			name: "return error for invalid north-east latitude",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			name: "return error for invalid north-east longitude",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			name: "return error for invalid south-west latitude",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			name: "return error for invalid south-west longitude",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			name: "return error during spot store failure",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("Spots", surf.SpotsParams{
						Limit:       10,
						Offset:      0,
						CountryCode: "kz",
						SearchQuery: surf.SpotSearchQuery{
							Query:      "query",
							WithSpotID: true,
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
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			s := NewService(test.spotStore, newMockLocationSource())

			spots, err := s.Spots(test.ctxFn(), test.params)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpots, spots)
		})
	}
}

func TestService_CreateSpot(t *testing.T) {
	tests := []struct {
		name          string
		ctxFn         func() context.Context
		spotStore     SpotStore
		params        CreateSpotParams
		expectedSpot  surf.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error for unauthenticated request",
			ctxFn: func() context.Context {
				return context.Background()
			},
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
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.IsError(jwt.ErrClaimsNotFound),
		},
		{
			name: "return error for unauthorized request",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: "",
				})
			},
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
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.IsError(jwt.ErrMismatchedRole),
		},
		{
			name: "return error for invalid name",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidSpotName),
		},
		{
			name: "return error for invalid latitude",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidLatitude),
		},
		{
			name: "return error for invalid longitide",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidLongitude),
		},
		{
			name: "return error for invalid locality",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: newMockSpotStore(),
			params: CreateSpotParams{
				Name: "Spot 1",
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 180,
					},
					Locality:    "",
					CountryCode: "kz",
				},
			},
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidLocality),
		},
		{
			name: "return error for invalid country code",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: newMockSpotStore(),
			params: CreateSpotParams{
				Name: "Spot 1",
				Location: geo.Location{
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 180,
					},
					Locality:    "Locality 1",
					CountryCode: "zz",
				},
			},
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidCountryCode),
		},
		{
			name: "return error during spot store failure",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("CreateSpot", surf.SpotCreationEntry{
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
					Return(surf.Spot{}, errors.New("something went wrong"))
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
			expectedSpot:  surf.Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return spot using sanitized params without error",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("CreateSpot", surf.SpotCreationEntry{
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
						surf.Spot{
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
			expectedSpot: surf.Spot{
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
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("CreateSpot", surf.SpotCreationEntry{
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
						surf.Spot{
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
			expectedSpot: surf.Spot{
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

			spot, err := s.CreateSpot(test.ctxFn(), test.params)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpot, spot)
		})
	}
}

func TestService_UpdateSpot(t *testing.T) {
	tests := []struct {
		name          string
		ctxFn         func() context.Context
		spotStore     SpotStore
		params        UpdateSpotParams
		expectedSpot  surf.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error for unauthenticated request",
			ctxFn: func() context.Context {
				return context.Background()
			},
			spotStore: newMockSpotStore(),
			params: UpdateSpotParams{
				ID:   "",
				Name: pconv.String("Spot 1"),
			},
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.IsError(jwt.ErrClaimsNotFound),
		},
		{
			name: "return error for unauthorized request",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: "",
				})
			},
			spotStore: newMockSpotStore(),
			params: UpdateSpotParams{
				ID:   "",
				Name: pconv.String("Spot 1"),
			},
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.IsError(jwt.ErrMismatchedRole),
		},
		{
			name: "return error for invalid id",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: newMockSpotStore(),
			params: UpdateSpotParams{
				ID:   "",
				Name: pconv.String("Spot 1"),
			},
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidSpotID),
		},
		{
			name: "return error for invalid name",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: newMockSpotStore(),
			params: UpdateSpotParams{
				ID:   "1",
				Name: pconv.String(""),
			},
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidSpotName),
		},
		{
			name: "return error for invalid latitude",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: newMockSpotStore(),
			params: UpdateSpotParams{
				ID:       "1",
				Name:     pconv.String("Spot 1"),
				Latitude: pconv.Float64(-91),
			},
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidLatitude),
		},
		{
			name: "return error for invalid longitude",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: newMockSpotStore(),
			params: UpdateSpotParams{
				ID:        "1",
				Name:      pconv.String("Spot 1"),
				Latitude:  pconv.Float64(1.23),
				Longitude: pconv.Float64(-181),
			},
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidLongitude),
		},
		{
			name: "return error for invalid locality",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: newMockSpotStore(),
			params: UpdateSpotParams{
				ID:        "1",
				Name:      pconv.String("Spot 1"),
				Latitude:  pconv.Float64(1.23),
				Longitude: pconv.Float64(2.34),
				Locality:  pconv.String(""),
			},
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidLocality),
		},
		{
			name: "return error for invalid country code",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: newMockSpotStore(),
			params: UpdateSpotParams{
				ID:          "1",
				Name:        pconv.String("Spot 1"),
				Latitude:    pconv.Float64(1.23),
				Longitude:   pconv.Float64(2.34),
				Locality:    pconv.String("Locality 1"),
				CountryCode: pconv.String("zz"),
			},
			expectedSpot:  surf.Spot{},
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidCountryCode),
		},
		{
			name: "return error during spot store failure",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("UpdateSpot", surf.SpotUpdateEntry{
						Latitude:    pconv.Float64(1.23),
						Longitude:   pconv.Float64(2.34),
						Locality:    pconv.String("Locality 1"),
						CountryCode: pconv.String("kz"),
						Name:        pconv.String("Spot 1"),
						ID:          "1",
					}).
					Return(surf.Spot{}, errors.New("something went wrong"))
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
			expectedSpot:  surf.Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return spot for coordinateless params without error",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("UpdateSpot", surf.SpotUpdateEntry{
						Name: pconv.String("Spot 1"),
						ID:   "1",
					}).
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
			expectedSpot: surf.Spot{
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
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("UpdateSpot", surf.SpotUpdateEntry{
						ID:       "1",
						Latitude: pconv.Float64(1.23),
						Locality: pconv.String("Locality 1"),
					}).
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
			expectedSpot: surf.Spot{
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
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("UpdateSpot", surf.SpotUpdateEntry{
						ID:          "1",
						Latitude:    pconv.Float64(1.23),
						Longitude:   pconv.Float64(2.34),
						Locality:    pconv.String("Locality 1"),
						CountryCode: pconv.String("kz"),
						Name:        pconv.String("Spot 1"),
					}).
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
			expectedSpot: surf.Spot{
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
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore: func() SpotStore {
				m := newMockSpotStore()
				m.
					On("UpdateSpot", surf.SpotUpdateEntry{
						ID:          "1",
						Latitude:    pconv.Float64(1.23),
						Longitude:   pconv.Float64(2.34),
						Locality:    pconv.String("Locality 1"),
						CountryCode: pconv.String("kz"),
						Name:        pconv.String("Spot 1"),
					}).
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
			expectedSpot: surf.Spot{
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

			spot, err := s.UpdateSpot(test.ctxFn(), test.params)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpot, spot)
		})
	}
}

func TestService_DeleteSpot(t *testing.T) {
	tests := []struct {
		name          string
		ctxFn         func() context.Context
		spotStore     SpotStore
		id            string
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error for unauthenticated request",
			ctxFn: func() context.Context {
				return context.Background()
			},
			spotStore:     newMockSpotStore(),
			id:            "",
			expectedErrFn: testutil.IsError(jwt.ErrClaimsNotFound),
		},
		{
			name: "return error for unauthorized request",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: "",
				})
			},
			spotStore:     newMockSpotStore(),
			id:            "",
			expectedErrFn: testutil.IsError(jwt.ErrMismatchedRole),
		},
		{
			name: "return error for invalid spot id",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			spotStore:     newMockSpotStore(),
			id:            "",
			expectedErrFn: testutil.AreValidationErrors(ErrInvalidSpotID),
		},
		{
			name: "return error during spot store failure",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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

			err := s.DeleteSpot(test.ctxFn(), test.id)
			test.expectedErrFn(t, err)
		})
	}
}

func TestService_Location(t *testing.T) {
	tests := []struct {
		name             string
		ctxFn            func() context.Context
		locationSource   geo.LocationSource
		coord            geo.Coordinates
		expectedLocation geo.Location
		expectedErrFn    assert.ErrorAssertionFunc
	}{
		{
			name: "return error for unauthenticated request",
			ctxFn: func() context.Context {
				return context.Background()
			},
			locationSource: newMockLocationSource(),
			coord: geo.Coordinates{
				Latitude:  -91,
				Longitude: 180,
			},
			expectedLocation: geo.Location{},
			expectedErrFn:    testutil.IsError(jwt.ErrClaimsNotFound),
		},
		{
			name: "return error for unauthorized request",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: "",
				})
			},
			locationSource: newMockLocationSource(),
			coord: geo.Coordinates{
				Latitude:  -91,
				Longitude: 180,
			},
			expectedLocation: geo.Location{},
			expectedErrFn:    testutil.IsError(jwt.ErrMismatchedRole),
		},
		{
			name: "return error for invalid latitude",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			locationSource: newMockLocationSource(),
			coord: geo.Coordinates{
				Latitude:  -91,
				Longitude: 180,
			},
			expectedLocation: geo.Location{},
			expectedErrFn:    testutil.AreValidationErrors(ErrInvalidLatitude),
		},
		{
			name: "return error for invalid longitude",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
			locationSource: newMockLocationSource(),
			coord: geo.Coordinates{
				Latitude:  -90,
				Longitude: 181,
			},
			expectedLocation: geo.Location{},
			expectedErrFn:    testutil.AreValidationErrors(ErrInvalidLongitude),
		},
		{
			name: "return error during unexpected location source failure",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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
			expectedErrFn:    testutil.IsError(geo.ErrLocationNotFound),
		},
		{
			name: "return location without error",
			ctxFn: func() context.Context {
				return jwt.ContextWith(context.Background(), jwt.Claims{
					Role: jwt.RoleName(auth.RoleAdmin),
				})
			},
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

			l, err := s.Location(test.ctxFn(), test.coord)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedLocation, l)
		})
	}
}
