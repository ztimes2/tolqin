package router

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ztimes2/tolqin/app/api/internal/api/service/surfing"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/geo"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/surf"
	"github.com/ztimes2/tolqin/app/api/pkg/valerra"
)

type mockSurfingService struct {
	mock.Mock
}

func newMockSurfingService() *mockSurfingService {
	return &mockSurfingService{}
}

func (m *mockSurfingService) Spot(id string) (surf.Spot, error) {
	args := m.Called(id)
	return args.Get(0).(surf.Spot), args.Error(1)
}

func (m *mockSurfingService) Spots(p surfing.SpotsParams) ([]surf.Spot, error) {
	args := m.Called(p)
	return args.Get(0).([]surf.Spot), args.Error(1)
}

func TestSurfingHandler_Spot(t *testing.T) {
	tests := []struct {
		name               string
		service            surfingService
		logger             *logrus.Logger
		id                 string
		expectedResponseFn func(t *testing.T, r *http.Response)
	}{
		{
			name: "respond with 500 status code and error body for unexpected error",
			service: func() surfingService {
				m := newMockSurfingService()
				m.
					On("Spot", "1").
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
			service: func() surfingService {
				m := newMockSurfingService()
				m.
					On("Spot", "1").
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
			service: func() surfingService {
				m := newMockSurfingService()
				m.
					On("Spot", "invalid").
					Return(surf.Spot{}, valerra.NewErrors(surfing.ErrInvalidSpotID))
				return m
			}(),
			logger: nil, // FIXME catch error logs
			id:     "invalid",
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
			service: func() surfingService {
				m := newMockSurfingService()
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
			server := httptest.NewServer(newRouter(nil, test.service, nil, nil, test.logger)) // TODO replace nil
			defer server.Close()

			req, err := http.NewRequest(http.MethodGet, server.URL+"/surfing/v1/spots/"+test.id, nil)
			assert.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			test.expectedResponseFn(t, resp)
		})
	}
}

func TestSurfingHandler_Spots(t *testing.T) {
	tests := []struct {
		name               string
		service            surfingService
		logger             *logrus.Logger
		requestFn          func(r *http.Request)
		expectedResponseFn func(t *testing.T, r *http.Response)
	}{
		{
			name:    "respond with 400 status code and error body for invalid limit",
			service: newMockSurfingService(),
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
			service: newMockSurfingService(),
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
			service: newMockSurfingService(),
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
			service: newMockSurfingService(),
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
			service: newMockSurfingService(),
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
			service: newMockSurfingService(),
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
			service: func() surfingService {
				m := newMockSurfingService()
				m.
					On("Spots", surfing.SpotsParams{
						Limit:       10,
						Offset:      0,
						CountryCode: "zz",
					}).
					Return(([]surf.Spot)(nil), valerra.NewErrors(
						surfing.ErrInvalidSearchQuery,
						surfing.ErrInvalidCountryCode,
						surfing.ErrInvalidNorthEastLatitude,
						surfing.ErrInvalidNorthEastLongitude,
						surfing.ErrInvalidSouthWestLatitude,
						surfing.ErrInvalidSouthWestLongitude,
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
			name: "respond with 500 status code and error body for expected error",
			service: func() surfingService {
				m := newMockSurfingService()
				m.
					On("Spots", surfing.SpotsParams{
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
			service: func() surfingService {
				m := newMockSurfingService()
				m.
					On("Spots", surfing.SpotsParams{
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
			service: func() surfingService {
				m := newMockSurfingService()
				m.
					On("Spots", surfing.SpotsParams{
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
			server := httptest.NewServer(newRouter(nil, test.service, nil, nil, test.logger)) // TODO replace nil
			defer server.Close()

			req, err := http.NewRequest(http.MethodGet, server.URL+"/surfing/v1/spots", nil)
			assert.NoError(t, err)

			test.requestFn(req)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			test.expectedResponseFn(t, resp)
		})
	}
}
