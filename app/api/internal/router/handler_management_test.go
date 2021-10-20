package router

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/service/management"
	"github.com/ztimes2/tolqin/app/api/internal/surf"
	"github.com/ztimes2/tolqin/app/api/pkg/pconv"
	"github.com/ztimes2/tolqin/app/api/pkg/valerra"
)

type mockManagementService struct {
	mock.Mock
}

func newMockManagementService() *mockManagementService {
	return &mockManagementService{}
}

func (m *mockManagementService) Spot(ctx context.Context, id string) (surf.Spot, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(surf.Spot), args.Error(1)
}

func (m *mockManagementService) Spots(ctx context.Context, p management.SpotsParams) ([]surf.Spot, error) {
	args := m.Called(ctx, p)
	return args.Get(0).([]surf.Spot), args.Error(1)
}

func (m *mockManagementService) CreateSpot(ctx context.Context, p management.CreateSpotParams) (surf.Spot, error) {
	args := m.Called(ctx, p)
	return args.Get(0).(surf.Spot), args.Error(1)
}

func (m *mockManagementService) UpdateSpot(ctx context.Context, p management.UpdateSpotParams) (surf.Spot, error) {
	args := m.Called(ctx, p)
	return args.Get(0).(surf.Spot), args.Error(1)
}

func (m *mockManagementService) DeleteSpot(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockManagementService) Location(ctx context.Context, c geo.Coordinates) (geo.Location, error) {
	args := m.Called(ctx, c)
	return args.Get(0).(geo.Location), args.Error(1)
}

func TestManagementHandler_Spot(t *testing.T) {
	tests := []struct {
		name               string
		service            managementService
		logger             *logrus.Logger
		id                 string
		expectedResponseFn func(t *testing.T, r *http.Response)
	}{
		{
			name: "respond with 500 status code and error body for unexpected error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Spot", mock.Anything, "1").
					Return(surf.Spot{}, errors.New("something went wrong"))
				return m
			}(),
			logger: nil, // FIXME catch error logs
			id:     "1",
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusInternalServerError, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "unexpected",
							"description": "Something went wrong..."
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 404 status code and error body for unexisting spot",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Spot", mock.Anything, "1").
					Return(surf.Spot{}, surf.ErrSpotNotFound)
				return m
			}(),
			logger: nil, // FIXME catch error logs
			id:     "1",
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusNotFound, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "not_found",
							"description": "Such spot doesn't exist."
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 400 status code and error body for invalid spot id",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Spot", mock.Anything, "1").
					Return(surf.Spot{}, valerra.NewErrors(management.ErrInvalidSpotID))
				return m
			}(),
			logger: nil, // FIXME catch error logs
			id:     "1",
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "spot_id",
									"reason": "Must be a non empty string."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 200 status code and spot body",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Spot", mock.Anything, "1").
					Return(
						surf.Spot{
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
			logger: nil, // FIXME catch error logs
			id:     "1",
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusOK, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"data": {
							"id": "1",
							"name": "Spot 1",
							"latitude": 1.23,
							"longitude": 3.21,
							"locality": "Locality 1",
							"country_code": "Country code 1"
						}
					}`,
					string(body),
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(newRouter(nil, newMockSurferService(), test.service, nil, test.logger))
			defer server.Close()

			req, err := http.NewRequest(http.MethodGet, server.URL+"/management/v1/spots/"+test.id, nil)
			assert.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			test.expectedResponseFn(t, resp)
		})
	}
}

func TestManagementHandler_Spots(t *testing.T) {
	tests := []struct {
		name               string
		service            managementService
		logger             *logrus.Logger
		requestFn          func(r *http.Request)
		expectedResponseFn func(t *testing.T, r *http.Response)
	}{
		{
			name:    "respond with 400 status code and error body for invalid limit",
			service: newMockManagementService(),
			logger:  nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"limit":  []string{"a"},
					"offset": []string{"0"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "limit",
									"reason": "Must be a valid integer."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name:    "respond with 400 status code and error body for invalid offset",
			service: newMockManagementService(),
			logger:  nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"limit":  []string{"10"},
					"offset": []string{"a"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "offset",
									"reason": "Must be a valid integer."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name:    "respond with 400 status code and error body for invalid north-east latitude",
			service: newMockManagementService(),
			logger:  nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"limit":  []string{"10"},
					"offset": []string{"0"},
					"ne_lat": []string{"a"},
					"ne_lon": []string{"180"},
					"sw_lat": []string{"-90"},
					"sw_lon": []string{"-180"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "ne_lat",
									"reason": "Must be a valid latitude."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name:    "respond with 400 status code and error body for invalid north-east longitude",
			service: newMockManagementService(),
			logger:  nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"limit":  []string{"10"},
					"offset": []string{"0"},
					"ne_lat": []string{"90"},
					"ne_lon": []string{"a"},
					"sw_lat": []string{"-90"},
					"sw_lon": []string{"-180"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "ne_lon",
									"reason": "Must be a valid longitude."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name:    "respond with 400 status code and error body for invalid south-west latitude",
			service: newMockManagementService(),
			logger:  nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"limit":  []string{"10"},
					"offset": []string{"0"},
					"ne_lat": []string{"90"},
					"ne_lon": []string{"180"},
					"sw_lat": []string{"a"},
					"sw_lon": []string{"-180"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "sw_lat",
									"reason": "Must be a valid latitude."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name:    "respond with 400 status code and error body for invalid south-west longitude",
			service: newMockManagementService(),
			logger:  nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"limit":  []string{"10"},
					"offset": []string{"0"},
					"ne_lat": []string{"90"},
					"ne_lon": []string{"180"},
					"sw_lat": []string{"-90"},
					"sw_lon": []string{"a"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "sw_lon",
									"reason": "Must be a valid longitude."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 400 status code and error body for validation error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Spots", mock.Anything, management.SpotsParams{
						Limit:       10,
						Offset:      0,
						CountryCode: "zz",
					}).
					Return(([]surf.Spot)(nil), valerra.NewErrors(
						management.ErrInvalidSearchQuery,
						management.ErrInvalidCountryCode,
						management.ErrInvalidNorthEastLatitude,
						management.ErrInvalidNorthEastLongitude,
						management.ErrInvalidSouthWestLatitude,
						management.ErrInvalidSouthWestLongitude,
					))
				return m
			}(),
			logger: nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"limit":   []string{"10"},
					"offset":  []string{"0"},
					"country": []string{"zz"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "query",
									"reason": "Must not exceed character limit."
								},
								{
									"key": "country",
									"reason": "Must be a valid ISO-2 country code."
								},
								{
									"key": "ne_lat",
									"reason": "Must be a valid latitude."
								},
								{
									"key": "ne_lon",
									"reason": "Must be a valid longitude."
								},
								{
									"key": "sw_lat",
									"reason": "Must be a valid latitude."
								},
								{
									"key": "sw_lon",
									"reason": "Must be a valid longitude."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 500 status code and error body for unexpected error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Spots", mock.Anything, management.SpotsParams{
						Limit:  10,
						Offset: 0,
					}).
					Return(([]surf.Spot)(nil), errors.New("something went wrong"))
				return m
			}(),
			logger: nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"limit":  []string{"10"},
					"offset": []string{"0"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusInternalServerError, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "unexpected",
							"description": "Something went wrong..."
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 200 status code and empty spot list body",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Spots", mock.Anything, management.SpotsParams{
						Limit:  0,
						Offset: 0,
					}).
					Return(([]surf.Spot)(nil), nil)
				return m
			}(),
			logger: nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				// Omit query parameters
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusOK, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"data": {
							"items":[]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 200 status code and spot list body",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Spots", mock.Anything, management.SpotsParams{
						Limit:       10,
						Offset:      0,
						CountryCode: "kz",
						SearchQuery: "query",
						Bounds: &geo.Bounds{
							NorthEast: geo.Coordinates{
								Latitude:  90,
								Longitude: 180,
							},
							SouthWest: geo.Coordinates{
								Latitude:  -90,
								Longitude: -180,
							},
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
			logger: nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"limit":   []string{"10"},
					"offset":  []string{"0"},
					"country": []string{"kz"},
					"query":   []string{"query"},
					"ne_lat":  []string{"90"},
					"ne_lon":  []string{"180"},
					"sw_lat":  []string{"-90"},
					"sw_lon":  []string{"-180"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusOK, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"data": {
							"items": [
								{
									"id": "1",
									"name": "Spot 1",
									"latitude": 1.23,
									"longitude": 3.21,
									"locality": "Locality 1",
									"country_code": "kz"
								},
								{
									"id": "2",
									"name": "Spot 2",
									"latitude": 1.23,
									"longitude": 3.21,
									"locality": "Locality 2",
									"country_code": "kz"
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(newRouter(nil, newMockSurferService(), test.service, nil, test.logger))
			defer server.Close()

			req, err := http.NewRequest(http.MethodGet, server.URL+"/management/v1/spots", nil)
			assert.NoError(t, err)

			test.requestFn(req)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			test.expectedResponseFn(t, resp)
		})
	}
}

func TestManagementHandler_CreateSpot(t *testing.T) {
	tests := []struct {
		name               string
		service            managementService
		logger             *logrus.Logger
		requestFn          func(r *http.Request)
		expectedResponseFn func(t *testing.T, r *http.Response)
	}{
		{
			name:    "respond with 400 status code and error body for invalid request body format",
			service: newMockManagementService(),
			logger:  nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				// Omit request body
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid payload.",
							"fields": []
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 400 status code and error body for validation error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("CreateSpot", mock.Anything, management.CreateSpotParams{
						Location: geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "kz",
						},
					}).
					Return(surf.Spot{}, valerra.NewErrors(
						management.ErrInvalidSpotName,
						management.ErrInvalidCountryCode,
						management.ErrInvalidLocality,
						management.ErrInvalidLatitude,
						management.ErrInvalidLongitude,
					))
				return m
			}(),
			logger: nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				r.Body = ioutil.NopCloser(strings.NewReader(
					`{
						"latitude": 1.23,
						"longitude": 3.21,
						"locality": "Locality 1",
						"country_code": "kz"
					}`,
				))
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "name",
									"reason": "Must be a non empty string."
								},
								{
									"key": "country_code",
									"reason": "Must be a valid ISO-2 country code."
								},
								{
									"key": "locality",
									"reason": "Must be a non empty string."
								},
								{
									"key": "latitude",
									"reason": "Must be a valid latitude."
								},
								{
									"key": "longitude",
									"reason": "Must be a valid longitude."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 500 status code and error body for unexpected error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("CreateSpot", mock.Anything, management.CreateSpotParams{
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
			logger: nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				r.Body = ioutil.NopCloser(strings.NewReader(
					`{
						"name": "Spot 1",
						"latitude": 1.23,
						"longitude": 3.21,
						"locality": "Locality 1",
						"country_code": "kz"
					}`,
				))
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusInternalServerError, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "unexpected",
							"description": "Something went wrong..."
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 201 status code and spot body",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("CreateSpot", mock.Anything, management.CreateSpotParams{
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
			logger: nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				r.Body = ioutil.NopCloser(strings.NewReader(
					`{
						"name": "Spot 1",
						"latitude": 1.23,
						"longitude": 3.21,
						"locality": "Locality 1",
						"country_code": "kz"
					}`,
				))
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusCreated, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"data": {
							"id": "1",
							"name": "Spot 1",
							"latitude": 1.23,
							"longitude": 3.21,
							"locality": "Locality 1",
							"country_code": "Country code 1"
						}
					}`,
					string(body),
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(newRouter(nil, newMockSurferService(), test.service, nil, test.logger))
			defer server.Close()

			req, err := http.NewRequest(http.MethodPost, server.URL+"/management/v1/spots", nil)
			assert.NoError(t, err)

			test.requestFn(req)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			test.expectedResponseFn(t, resp)
		})
	}
}

func TestManagementHandler_UpdateSpot(t *testing.T) {
	tests := []struct {
		name               string
		service            managementService
		logger             *logrus.Logger
		id                 string
		requestFn          func(r *http.Request)
		expectedResponseFn func(t *testing.T, r *http.Response)
	}{
		{
			name:    "respond with 400 status code and error body for invalid request body format",
			service: newMockManagementService(),
			logger:  nil, // FIXME catch error logs
			id:      "1",
			requestFn: func(r *http.Request) {
				// Omit request body
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid payload.",
							"fields": []
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 400 status code and error body for validation error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("UpdateSpot", mock.Anything, management.UpdateSpotParams{
						Latitude:  pconv.Float64(1.23),
						Longitude: pconv.Float64(3.21),
						Name:      pconv.String(""),
						ID:        "1",
					}).
					Return(surf.Spot{}, valerra.NewErrors(
						management.ErrInvalidSpotID,
						management.ErrInvalidSpotName,
						management.ErrInvalidCountryCode,
						management.ErrInvalidLocality,
						management.ErrInvalidLatitude,
						management.ErrInvalidLongitude,
					))
				return m
			}(),
			logger: nil, // FIXME catch error logs
			id:     "1",
			requestFn: func(r *http.Request) {
				r.Body = ioutil.NopCloser(strings.NewReader(
					`{
						"name": "",
						"latitude": 1.23,
						"longitude": 3.21
					}`,
				))
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "spot_id",
									"reason": "Must be a non empty string."
								},
								{
									"key": "name",
									"reason": "Must be a non empty string."
								},
								{
									"key": "country_code",
									"reason": "Must be a valid ISO-2 country code."
								},
								{
									"key": "locality",
									"reason": "Must be a non empty string."
								},
								{
									"key": "latitude",
									"reason": "Must be a valid latitude."
								},
								{
									"key": "longitude",
									"reason": "Must be a valid longitude."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 400 status code and error body for empty input",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("UpdateSpot", mock.Anything, management.UpdateSpotParams{
						ID: "1",
					}).
					Return(surf.Spot{}, surf.ErrEmptySpotUpdateEntry)
				return m
			}(),
			logger: nil, // FIXME catch error logs
			id:     "1",
			requestFn: func(r *http.Request) {
				r.Body = ioutil.NopCloser(strings.NewReader(`{}`))
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Nothing to update.",
							"fields": []
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 404 status code and error body for unexisting spot",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("UpdateSpot", mock.Anything, management.UpdateSpotParams{
						Latitude:  pconv.Float64(1.23),
						Longitude: pconv.Float64(3.21),
						Name:      pconv.String("Spot 1"),
						ID:        "1",
					}).
					Return(surf.Spot{}, surf.ErrSpotNotFound)
				return m
			}(),
			logger: nil, // FIXME catch error logs
			id:     "1",
			requestFn: func(r *http.Request) {
				r.Body = ioutil.NopCloser(strings.NewReader(
					`{
						"name": "Spot 1",
						"latitude": 1.23,
						"longitude": 3.21
					}`,
				))
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusNotFound, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "not_found",
							"description": "Such spot doesn't exist."
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 500 status code and error body for unexpected error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("UpdateSpot", mock.Anything, management.UpdateSpotParams{
						Latitude:  pconv.Float64(1.23),
						Longitude: pconv.Float64(3.21),
						Name:      pconv.String("Spot 1"),
						ID:        "1",
					}).
					Return(surf.Spot{}, errors.New("something went wrong"))
				return m
			}(),
			logger: nil, // FIXME catch error logs
			id:     "1",
			requestFn: func(r *http.Request) {
				r.Body = ioutil.NopCloser(strings.NewReader(
					`{
						"name": "Spot 1",
						"latitude": 1.23,
						"longitude": 3.21
					}`,
				))
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusInternalServerError, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "unexpected",
							"description": "Something went wrong..."
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 200 status code and spot body for partial input",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("UpdateSpot", mock.Anything, management.UpdateSpotParams{
						Name:      pconv.String("Spot 1"),
						Latitude:  pconv.Float64(1.23),
						Longitude: pconv.Float64(3.21),
						ID:        "1",
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
							ID:        "1",
							Name:      "Spot 1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			logger: nil, // FIXME catch error logs
			id:     "1",
			requestFn: func(r *http.Request) {
				r.Body = ioutil.NopCloser(strings.NewReader(
					`{
						"name": "Spot 1",
						"latitude": 1.23,
						"longitude": 3.21
					}`,
				))
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusOK, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"data": {
							"id": "1",
							"name": "Spot 1",
							"latitude": 1.23,
							"longitude": 3.21,
							"locality": "Locality 1",
							"country_code": "kz"
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 200 status code and spot body for partial input",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("UpdateSpot", mock.Anything, management.UpdateSpotParams{
						Locality:    pconv.String("Locality 1"),
						CountryCode: pconv.String("kz"),
						ID:          "1",
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
							ID:        "1",
							Name:      "Spot 1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			logger: nil, // FIXME catch error logs
			id:     "1",
			requestFn: func(r *http.Request) {
				r.Body = ioutil.NopCloser(strings.NewReader(
					`{
						"locality": "Locality 1",
						"country_code": "kz"
					}`,
				))
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusOK, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"data": {
							"id": "1",
							"name": "Spot 1",
							"latitude": 1.23,
							"longitude": 3.21,
							"locality": "Locality 1",
							"country_code": "kz"
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 200 status code and spot body for full input",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("UpdateSpot", mock.Anything, management.UpdateSpotParams{
						Name:        pconv.String("Spot 1"),
						Latitude:    pconv.Float64(1.23),
						Longitude:   pconv.Float64(3.21),
						Locality:    pconv.String("Locality 1"),
						CountryCode: pconv.String("kz"),
						ID:          "1",
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
							ID:        "1",
							Name:      "Spot 1",
							CreatedAt: time.Date(2021, 1, 1, 1, 1, 1, 1, time.UTC),
						},
						nil,
					)
				return m
			}(),
			logger: nil, // FIXME catch error logs
			id:     "1",
			requestFn: func(r *http.Request) {
				r.Body = ioutil.NopCloser(strings.NewReader(
					`{
						"name": "Spot 1",
						"latitude": 1.23,
						"longitude": 3.21,
						"locality": "Locality 1",
						"country_code": "kz"
					}`,
				))
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusOK, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"data": {
							"id": "1",
							"name": "Spot 1",
							"latitude": 1.23,
							"longitude": 3.21,
							"locality": "Locality 1",
							"country_code": "kz"
						}
					}`,
					string(body),
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(newRouter(nil, newMockSurferService(), test.service, nil, test.logger))
			defer server.Close()

			req, err := http.NewRequest(http.MethodPatch, server.URL+"/management/v1/spots/"+test.id, nil)
			assert.NoError(t, err)

			test.requestFn(req)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			test.expectedResponseFn(t, resp)
		})
	}
}

func TestManagementHandler_DeleteSpot(t *testing.T) {
	tests := []struct {
		name               string
		service            managementService
		logger             *logrus.Logger
		id                 string
		expectedResponseFn func(t *testing.T, r *http.Response)
	}{
		{
			name: "respond with 500 status code and error body for unexpected error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("DeleteSpot", mock.Anything, "1").
					Return(errors.New("something went wrong"))
				return m
			}(),
			logger: nil, // FIXME catch error logs
			id:     "1",
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusInternalServerError, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "unexpected",
							"description": "Something went wrong..."
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 404 status code and error body for unexisting spot",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("DeleteSpot", mock.Anything, "1").
					Return(surf.ErrSpotNotFound)
				return m
			}(),
			logger: nil, // FIXME catch error logs
			id:     "1",
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusNotFound, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "not_found",
							"description": "Such spot doesn't exist."
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 400 status code and error body for invalid spot id",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("DeleteSpot", mock.Anything, "1").
					Return(valerra.NewErrors(management.ErrInvalidSpotID))
				return m
			}(),
			logger: nil, // FIXME catch error logs
			id:     "1",
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "spot_id",
									"reason": "Must be a non empty string."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 204 status code",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("DeleteSpot", mock.Anything, "1").
					Return(nil)
				return m
			}(),
			logger: nil, // FIXME catch error logs
			id:     "1",
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusNoContent, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.Equal(t, "", string(body))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(newRouter(nil, newMockSurferService(), test.service, nil, test.logger))
			defer server.Close()

			req, err := http.NewRequest(http.MethodDelete, server.URL+"/management/v1/spots/"+test.id, nil)
			assert.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			test.expectedResponseFn(t, resp)
		})
	}
}

func TestManagementHandler_Location(t *testing.T) {
	tests := []struct {
		name               string
		service            managementService
		logger             *logrus.Logger
		requestFn          func(r *http.Request)
		expectedResponseFn func(t *testing.T, r *http.Response)
	}{
		{
			name:    "respond with 400 status code and error body for invalid latitude",
			service: newMockManagementService(),
			logger:  nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"lat": []string{"a"},
					"lon": []string{"3.21"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "lat",
									"reason": "Must be a valid latitude."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name:    "respond with 400 status code and error body for empty latitude",
			service: newMockManagementService(),
			logger:  nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"lon": []string{"3.21"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "lat",
									"reason": "Must be a valid latitude."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name:    "respond with 400 status code and error body for invalid longitude",
			service: newMockManagementService(),
			logger:  nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"lat": []string{"1.23"},
					"lon": []string{"a"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "lon",
									"reason": "Must be a valid longitude."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name:    "respond with 400 status code and error body for empty longitude",
			service: newMockManagementService(),
			logger:  nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"lat": []string{"1.23"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "lon",
									"reason": "Must be a valid longitude."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 400 status code and error body for validation error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Location", mock.Anything, geo.Coordinates{
						Latitude:  -91,
						Longitude: -181,
					}).
					Return(geo.Location{}, valerra.NewErrors(
						management.ErrInvalidLatitude,
						management.ErrInvalidLongitude,
					))
				return m
			}(),
			logger: nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"lat": []string{"-91"},
					"lon": []string{"-181"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusBadRequest, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "invalid_input",
							"description": "Invalid input parameters.",
							"fields": [
								{
									"key": "lat",
									"reason": "Must be a valid latitude."
								},
								{
									"key": "lon",
									"reason": "Must be a valid longitude."
								}
							]
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 500 status code and error body for unexpected error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Location", mock.Anything, geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					}).
					Return(geo.Location{}, errors.New("something went wrong"))
				return m
			}(),
			logger: nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"lat": []string{"1.23"},
					"lon": []string{"3.21"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusInternalServerError, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "unexpected",
							"description": "Something went wrong..."
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 404 status code and error body when location is not found",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Location", mock.Anything, geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					}).
					Return(geo.Location{}, geo.ErrLocationNotFound)
				return m
			}(),
			logger: nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"lat": []string{"1.23"},
					"lon": []string{"3.21"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusNotFound, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"error": {
							"code": "not_found",
							"description": "Location was not found."
						}
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 200 status code and location body",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Location", mock.Anything, geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					}).
					Return(
						geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "kz",
						},
						nil,
					)
				return m
			}(),
			logger: nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"lat": []string{"1.23"},
					"lon": []string{"3.21"},
				}
				r.URL.RawQuery = vals.Encode()
			},
			expectedResponseFn: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusOK, r.StatusCode)

				body, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				assert.NoError(t, err)

				assert.JSONEq(
					t,
					`{
						"data": {
							"latitude": 1.23,
							"longitude": 3.21,
							"locality": "Locality 1",
							"country_code": "kz"
						}
					}`,
					string(body),
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(newRouter(nil, newMockSurferService(), test.service, nil, test.logger))
			defer server.Close()

			req, err := http.NewRequest(http.MethodGet, server.URL+"/management/v1/geo/location", nil)
			assert.NoError(t, err)

			test.requestFn(req)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			test.expectedResponseFn(t, resp)
		})
	}
}
