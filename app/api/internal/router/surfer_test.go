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
	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/service/surfer"
	"github.com/ztimes2/tolqin/app/api/internal/validation"
)

type mockSurferService struct {
	mock.Mock
}

func newMockSurferService() *mockSurferService {
	return &mockSurferService{}
}

func (m *mockSurferService) Spot(id string) (surfer.Spot, error) {
	args := m.Called(id)
	return args.Get(0).(surfer.Spot), args.Error(1)
}

func (m *mockSurferService) Spots(p surfer.SpotsParams) ([]surfer.Spot, error) {
	args := m.Called(p)
	return args.Get(0).([]surfer.Spot), args.Error(1)
}

func TestSurferHandler_Spot(t *testing.T) {
	tests := []struct {
		name               string
		service            surferService
		logger             *logrus.Logger
		id                 string
		expectedResponseFn func(t *testing.T, r *http.Response)
	}{
		{
			name: "respond with 500 status code and error body for unexpected error",
			service: func() surferService {
				m := newMockSurferService()
				m.
					On("Spot", "1").
					Return(surfer.Spot{}, errors.New("something went wrong"))
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
					`{"error_description":"Something went wrong..."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 404 status code and error body for unexisting spot",
			service: func() surferService {
				m := newMockSurferService()
				m.
					On("Spot", "1").
					Return(surfer.Spot{}, surfer.ErrNotFound)
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
					`{"error_description":"Such spot doesn't exist."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 200 status code and spot body",
			service: func() surferService {
				m := newMockSurferService()
				m.
					On("Spot", "1").
					Return(
						surfer.Spot{
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
						"id": "1",
						"name": "Spot 1",
						"latitude": 1.23,
						"longitude": 3.21,
						"locality": "Locality 1",
						"country_code": "Country code 1"
					}`,
					string(body),
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(newRouter(test.service, nil, test.logger)) // TODO replace nil
			defer server.Close()

			req, err := http.NewRequest(http.MethodGet, server.URL+"/v1/spots/"+test.id, nil)
			assert.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			test.expectedResponseFn(t, resp)
		})
	}
}

func TestSurferHandler_Spots(t *testing.T) {
	tests := []struct {
		name               string
		service            surferService
		logger             *logrus.Logger
		requestFn          func(r *http.Request)
		expectedResponseFn func(t *testing.T, r *http.Response)
	}{
		{
			name:    "respond with 400 status code and error body for invalid limit",
			service: newMockSurferService(),
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
					`{"error_description":"Invalid limit."}`,
					string(body),
				)
			},
		},
		{
			name:    "respond with 400 status code and error body for invalid offset",
			service: newMockSurferService(),
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
					`{"error_description":"Invalid offset."}`,
					string(body),
				)
			},
		},
		{
			name:    "respond with 400 status code and error body for invalid north-east latitude",
			service: newMockSurferService(),
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
					`{"error_description":"Invalid coordinates."}`,
					string(body),
				)
			},
		},
		{
			name:    "respond with 400 status code and error body for invalid north-east longitude",
			service: newMockSurferService(),
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
					`{"error_description":"Invalid coordinates."}`,
					string(body),
				)
			},
		},
		{
			name:    "respond with 400 status code and error body for invalid south-west latitude",
			service: newMockSurferService(),
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
					`{"error_description":"Invalid coordinates."}`,
					string(body),
				)
			},
		},
		{
			name:    "respond with 400 status code and error body for invalid south-west longitude",
			service: newMockSurferService(),
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
					`{"error_description":"Invalid coordinates."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 400 status code and error body for validation error",
			service: func() surferService {
				m := newMockSurferService()
				m.
					On("Spots", surfer.SpotsParams{
						Limit:       10,
						Offset:      0,
						CountryCode: "zz",
					}).
					Return(([]surfer.Spot)(nil), validation.NewError("country"))
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
					`{"error_description":"Invalid country."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 500 status code and error body for expected error",
			service: func() surferService {
				m := newMockSurferService()
				m.
					On("Spots", surfer.SpotsParams{
						Limit:  10,
						Offset: 0,
					}).
					Return(([]surfer.Spot)(nil), errors.New("something went wrong"))
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
					`{"error_description":"Something went wrong..."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 200 status code and empty spot list body",
			service: func() surferService {
				m := newMockSurferService()
				m.
					On("Spots", surfer.SpotsParams{
						Limit:  0,
						Offset: 0,
					}).
					Return(([]surfer.Spot)(nil), nil)
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
					`{"items":[]}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 200 status code and spot list body",
			service: func() surferService {
				m := newMockSurferService()
				m.
					On("Spots", surfer.SpotsParams{
						Limit:       10,
						Offset:      0,
						CountryCode: "kz",
						Query:       "query",
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
						[]surfer.Spot{
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
					"q":       []string{"query"},
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
					}`,
					string(body),
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(newRouter(test.service, nil, test.logger)) // TODO replace nil
			defer server.Close()

			req, err := http.NewRequest(http.MethodGet, server.URL+"/v1/spots", nil)
			assert.NoError(t, err)

			test.requestFn(req)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			test.expectedResponseFn(t, resp)
		})
	}
}
