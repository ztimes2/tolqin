package router

import (
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
	"github.com/ztimes2/tolqin/app/api/internal/pkg/pconv"
	"github.com/ztimes2/tolqin/app/api/internal/service/management"
	"github.com/ztimes2/tolqin/app/api/internal/validation"
)

type mockManagementService struct {
	mock.Mock
}

func newMockManagementService() *mockManagementService {
	return &mockManagementService{}
}

func (m *mockManagementService) Spot(id string) (management.Spot, error) {
	args := m.Called(id)
	return args.Get(0).(management.Spot), args.Error(1)
}

func (m *mockManagementService) Spots(p management.SpotsParams) ([]management.Spot, error) {
	args := m.Called(p)
	return args.Get(0).([]management.Spot), args.Error(1)
}

func (m *mockManagementService) CreateSpot(p management.CreateSpotParams) (management.Spot, error) {
	args := m.Called(p)
	return args.Get(0).(management.Spot), args.Error(1)
}

func (m *mockManagementService) UpdateSpot(p management.UpdateSpotParams) (management.Spot, error) {
	args := m.Called(p)
	return args.Get(0).(management.Spot), args.Error(1)
}

func (m *mockManagementService) DeleteSpot(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockManagementService) Location(c geo.Coordinates) (geo.Location, error) {
	args := m.Called(c)
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
					On("Spot", "1").
					Return(management.Spot{}, errors.New("something went wrong"))
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
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Spot", "1").
					Return(management.Spot{}, management.ErrNotFound)
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
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Spot", "1").
					Return(
						management.Spot{
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
			server := httptest.NewServer(newRouter(newMockSurferService(), test.service, test.logger))
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
					`{"error_description":"Invalid limit."}`,
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
					`{"error_description":"Invalid offset."}`,
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
					`{"error_description":"Invalid coordinates."}`,
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
					`{"error_description":"Invalid coordinates."}`,
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
					`{"error_description":"Invalid coordinates."}`,
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
					`{"error_description":"Invalid coordinates."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 400 status code and error body for validation error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Spots", management.SpotsParams{
						Limit:       10,
						Offset:      0,
						CountryCode: "zz",
					}).
					Return(([]management.Spot)(nil), validation.NewError("country"))
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
			name: "respond with 500 status code and error body for unexpected error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Spots", management.SpotsParams{
						Limit:  10,
						Offset: 0,
					}).
					Return(([]management.Spot)(nil), errors.New("something went wrong"))
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
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Spots", management.SpotsParams{
						Limit:  0,
						Offset: 0,
					}).
					Return(([]management.Spot)(nil), nil)
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
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Spots", management.SpotsParams{
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
						[]management.Spot{
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
			server := httptest.NewServer(newRouter(newMockSurferService(), test.service, test.logger))
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
					`{"error_description":"Invalid input."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 400 status code and error body for validation error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("CreateSpot", management.CreateSpotParams{
						Location: geo.Location{
							Coordinates: geo.Coordinates{
								Latitude:  1.23,
								Longitude: 3.21,
							},
							Locality:    "Locality 1",
							CountryCode: "kz",
						},
					}).
					Return(management.Spot{}, validation.NewError("name"))
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
					`{"error_description":"Invalid name."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 500 status code and error body for unexpected error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("CreateSpot", management.CreateSpotParams{
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
					Return(management.Spot{}, errors.New("something went wrong"))
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
					`{"error_description":"Something went wrong..."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 201 status code and spot body",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("CreateSpot", management.CreateSpotParams{
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
						management.Spot{
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
			server := httptest.NewServer(newRouter(newMockSurferService(), test.service, test.logger))
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
					`{"error_description":"Invalid input."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 400 status code and error body for validation error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("UpdateSpot", management.UpdateSpotParams{
						Latitude:  pconv.Float64(1.23),
						Longitude: pconv.Float64(3.21),
						Name:      pconv.String(""),
						ID:        "1",
					}).
					Return(management.Spot{}, validation.NewError("name"))
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
					`{"error_description":"Invalid name."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 400 status code and error body for empty input",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("UpdateSpot", management.UpdateSpotParams{
						ID: "1",
					}).
					Return(management.Spot{}, management.ErrNothingToUpdate)
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
					`{"error_description":"Nothing to update."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 404 status code and error body for unexisting spot",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("UpdateSpot", management.UpdateSpotParams{
						Latitude:  pconv.Float64(1.23),
						Longitude: pconv.Float64(3.21),
						Name:      pconv.String("Spot 1"),
						ID:        "1",
					}).
					Return(management.Spot{}, management.ErrNotFound)
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
					`{"error_description":"Such spot doesn't exist."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 500 status code and error body for unexpected error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("UpdateSpot", management.UpdateSpotParams{
						Latitude:  pconv.Float64(1.23),
						Longitude: pconv.Float64(3.21),
						Name:      pconv.String("Spot 1"),
						ID:        "1",
					}).
					Return(management.Spot{}, errors.New("something went wrong"))
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
					`{"error_description":"Something went wrong..."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 200 status code and spot body for partial input",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("UpdateSpot", management.UpdateSpotParams{
						Name:      pconv.String("Spot 1"),
						Latitude:  pconv.Float64(1.23),
						Longitude: pconv.Float64(3.21),
						ID:        "1",
					}).
					Return(
						management.Spot{
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
						"id": "1",
						"name": "Spot 1",
						"latitude": 1.23,
						"longitude": 3.21,
						"locality": "Locality 1",
						"country_code": "kz"
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
					On("UpdateSpot", management.UpdateSpotParams{
						Locality:    pconv.String("Locality 1"),
						CountryCode: pconv.String("kz"),
						ID:          "1",
					}).
					Return(
						management.Spot{
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
						"id": "1",
						"name": "Spot 1",
						"latitude": 1.23,
						"longitude": 3.21,
						"locality": "Locality 1",
						"country_code": "kz"
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
					On("UpdateSpot", management.UpdateSpotParams{
						Name:        pconv.String("Spot 1"),
						Latitude:    pconv.Float64(1.23),
						Longitude:   pconv.Float64(3.21),
						Locality:    pconv.String("Locality 1"),
						CountryCode: pconv.String("kz"),
						ID:          "1",
					}).
					Return(
						management.Spot{
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
						"id": "1",
						"name": "Spot 1",
						"latitude": 1.23,
						"longitude": 3.21,
						"locality": "Locality 1",
						"country_code": "kz"
					}`,
					string(body),
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(newRouter(newMockSurferService(), test.service, test.logger))
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
					On("DeleteSpot", "1").
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
					`{"error_description":"Something went wrong..."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 404 status code and error body for unexisting spot",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("DeleteSpot", "1").
					Return(management.ErrNotFound)
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
			name: "respond with 204 status code",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("DeleteSpot", "1").
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
			server := httptest.NewServer(newRouter(newMockSurferService(), test.service, test.logger))
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
					`{"error_description":"Invalid latitude."}`,
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
					`{"error_description":"Invalid latitude."}`,
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
					`{"error_description":"Invalid longitude."}`,
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
					`{"error_description":"Invalid longitude."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 500 status code and error body for unexpected error",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Location", geo.Coordinates{
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
					`{"error_description":"Something went wrong..."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 404 status code and error body when location is not found",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Location", geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					}).
					Return(geo.Location{}, management.ErrNotFound)
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
					`{"error_description":"Location was not found."}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 200 status code and location body",
			service: func() managementService {
				m := newMockManagementService()
				m.
					On("Location", geo.Coordinates{
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
						"latitude": 1.23,
						"longitude": 3.21,
						"locality": "Locality 1",
						"country_code": "kz"	
					}`,
					string(body),
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(newRouter(newMockSurferService(), test.service, test.logger))
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
