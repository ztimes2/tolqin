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
	"github.com/ztimes2/tolqin/internal/geo"
	"github.com/ztimes2/tolqin/internal/pconv"
	"github.com/ztimes2/tolqin/internal/surfing"
	"github.com/ztimes2/tolqin/internal/validation"
)

type mockService struct {
	mock.Mock
}

func newMockService() *mockService {
	return &mockService{}
}

func (m *mockService) Spot(id string) (surfing.Spot, error) {
	args := m.Called(id)
	return args.Get(0).(surfing.Spot), args.Error(1)
}

func (m *mockService) Spots(p surfing.SpotsParams) ([]surfing.Spot, error) {
	args := m.Called(p)
	return args.Get(0).([]surfing.Spot), args.Error(1)
}

func (m *mockService) CreateSpot(p surfing.CreateSpotParams) (surfing.Spot, error) {
	args := m.Called(p)
	return args.Get(0).(surfing.Spot), args.Error(1)
}

func (m *mockService) UpdateSpot(p surfing.UpdateSpotParams) (surfing.Spot, error) {
	args := m.Called(p)
	return args.Get(0).(surfing.Spot), args.Error(1)
}

func (m *mockService) DeleteSpot(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestHandler_Spot(t *testing.T) {
	tests := []struct {
		name               string
		service            service
		logger             *logrus.Logger
		id                 string
		expectedResponseFn func(t *testing.T, r *http.Response)
	}{
		{
			name: "respond with 500 status code and error body for unexpected error",
			service: func() service {
				m := newMockService()
				m.
					On("Spot", "1").
					Return(surfing.Spot{}, errors.New("something went wrong"))
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
			service: func() service {
				m := newMockService()
				m.
					On("Spot", "1").
					Return(surfing.Spot{}, surfing.ErrNotFound)
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
			service: func() service {
				m := newMockService()
				m.
					On("Spot", "1").
					Return(
						surfing.Spot{
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
			server := httptest.NewServer(newRouter(test.service, test.logger))
			defer server.Close()

			req, err := http.NewRequest(http.MethodGet, server.URL+"/spots/"+test.id, nil)
			assert.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			test.expectedResponseFn(t, resp)
		})
	}
}

func TestHandler_Spots(t *testing.T) {
	tests := []struct {
		name               string
		service            service
		logger             *logrus.Logger
		requestFn          func(r *http.Request)
		expectedResponseFn func(t *testing.T, r *http.Response)
	}{
		{
			name:    "respond with 400 status code and error body for invalid limit",
			service: newMockService(),
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
			service: newMockService(),
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
			name: "respond with 400 status code and error body for validation error",
			service: func() service {
				m := newMockService()
				m.
					On("Spots", surfing.SpotsParams{
						Limit:       10,
						Offset:      0,
						CountryCode: "zz",
					}).
					Return(([]surfing.Spot)(nil), validation.NewError("country"))
				return m
			}(),
			logger: nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				vals := url.Values{
					"limit":  []string{"10"},
					"offset": []string{"0"},
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
			service: func() service {
				m := newMockService()
				m.
					On("Spots", surfing.SpotsParams{
						Limit:  10,
						Offset: 0,
					}).
					Return(([]surfing.Spot)(nil), errors.New("something went wrong"))
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
			service: func() service {
				m := newMockService()
				m.
					On("Spots", surfing.SpotsParams{
						Limit:  0,
						Offset: 0,
					}).
					Return(([]surfing.Spot)(nil), nil)
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
			service: func() service {
				m := newMockService()
				m.
					On("Spots", surfing.SpotsParams{
						Limit:       10,
						Offset:      0,
						CountryCode: "kz",
						Query:       "query",
					}).
					Return(
						[]surfing.Spot{
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
			server := httptest.NewServer(newRouter(test.service, test.logger))
			defer server.Close()

			req, err := http.NewRequest(http.MethodGet, server.URL+"/spots", nil)
			assert.NoError(t, err)

			test.requestFn(req)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			test.expectedResponseFn(t, resp)
		})
	}
}

func TestHandler_CreateSpot(t *testing.T) {
	tests := []struct {
		name               string
		service            service
		logger             *logrus.Logger
		requestFn          func(r *http.Request)
		expectedResponseFn func(t *testing.T, r *http.Response)
	}{
		{
			name:    "respond with 400 status code and error body for invalid request body format",
			service: newMockService(),
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
			service: func() service {
				m := newMockService()
				m.
					On("CreateSpot", surfing.CreateSpotParams{
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					}).
					Return(surfing.Spot{}, validation.NewError("name"))
				return m
			}(),
			logger: nil, // FIXME catch error logs
			requestFn: func(r *http.Request) {
				r.Body = ioutil.NopCloser(strings.NewReader(
					`{
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
			name: "respond with 500 status code and error body for unexpected error",
			service: func() service {
				m := newMockService()
				m.
					On("CreateSpot", surfing.CreateSpotParams{
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
						Name: "Spot 1",
					}).
					Return(surfing.Spot{}, errors.New("something went wrong"))
				return m
			}(),
			logger: nil, // FIXME catch error logs
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
			name: "respond with 201 status code and spot body",
			service: func() service {
				m := newMockService()
				m.
					On("CreateSpot", surfing.CreateSpotParams{
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
						Name: "Spot 1",
					}).
					Return(
						surfing.Spot{
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
						"longitude": 3.21
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
			server := httptest.NewServer(newRouter(test.service, test.logger))
			defer server.Close()

			req, err := http.NewRequest(http.MethodPost, server.URL+"/spots", nil)
			assert.NoError(t, err)

			test.requestFn(req)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			test.expectedResponseFn(t, resp)
		})
	}
}

func TestHandler_UpdateSpot(t *testing.T) {
	tests := []struct {
		name               string
		service            service
		logger             *logrus.Logger
		id                 string
		requestFn          func(r *http.Request)
		expectedResponseFn func(t *testing.T, r *http.Response)
	}{
		{
			name:    "respond with 400 status code and error body for invalid request body format",
			service: newMockService(),
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
			service: func() service {
				m := newMockService()
				m.
					On("UpdateSpot", surfing.UpdateSpotParams{
						Coordinates: &geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
						Name: pconv.String(""),
						ID:   "1",
					}).
					Return(surfing.Spot{}, validation.NewError("name"))
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
			service: func() service {
				m := newMockService()
				m.
					On("UpdateSpot", surfing.UpdateSpotParams{
						ID: "1",
					}).
					Return(surfing.Spot{}, surfing.ErrNothingToUpdate)
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
			service: func() service {
				m := newMockService()
				m.
					On("UpdateSpot", surfing.UpdateSpotParams{
						Coordinates: &geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
						Name: pconv.String("Spot 1"),
						ID:   "1",
					}).
					Return(surfing.Spot{}, surfing.ErrNotFound)
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
			service: func() service {
				m := newMockService()
				m.
					On("UpdateSpot", surfing.UpdateSpotParams{
						Coordinates: &geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
						Name: pconv.String("Spot 1"),
						ID:   "1",
					}).
					Return(surfing.Spot{}, errors.New("something went wrong"))
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
			service: func() service {
				m := newMockService()
				m.
					On("UpdateSpot", surfing.UpdateSpotParams{
						Name: pconv.String("Spot 1"),
						ID:   "1",
					}).
					Return(
						surfing.Spot{
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
			requestFn: func(r *http.Request) {
				r.Body = ioutil.NopCloser(strings.NewReader(
					`{
						"name": "Spot 1"
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
						"country_code": "Country code 1"
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 200 status code and spot body for partial input",
			service: func() service {
				m := newMockService()
				m.
					On("UpdateSpot", surfing.UpdateSpotParams{
						Coordinates: &geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
						ID: "1",
					}).
					Return(
						surfing.Spot{
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
			requestFn: func(r *http.Request) {
				r.Body = ioutil.NopCloser(strings.NewReader(
					`{
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
						"country_code": "Country code 1"
					}`,
					string(body),
				)
			},
		},
		{
			name: "respond with 200 status code and spot body for full input",
			service: func() service {
				m := newMockService()
				m.
					On("UpdateSpot", surfing.UpdateSpotParams{
						Coordinates: &geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
						Name: pconv.String("Spot 1"),
						ID:   "1",
					}).
					Return(
						surfing.Spot{
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
			requestFn: func(r *http.Request) {
				r.Body = ioutil.NopCloser(strings.NewReader(
					`{
						"latitude": 1.23,
						"longitude": 3.21,
						"name": "Spot 1"
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
						"country_code": "Country code 1"
					}`,
					string(body),
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(newRouter(test.service, test.logger))
			defer server.Close()

			req, err := http.NewRequest(http.MethodPatch, server.URL+"/spots/"+test.id, nil)
			assert.NoError(t, err)

			test.requestFn(req)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			test.expectedResponseFn(t, resp)
		})
	}
}

func TestHandler_DeleteSpot(t *testing.T) {
	tests := []struct {
		name               string
		service            service
		logger             *logrus.Logger
		id                 string
		expectedResponseFn func(t *testing.T, r *http.Response)
	}{
		{
			name: "respond with 500 status code and error body for unexpected error",
			service: func() service {
				m := newMockService()
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
			service: func() service {
				m := newMockService()
				m.
					On("DeleteSpot", "1").
					Return(surfing.ErrNotFound)
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
			service: func() service {
				m := newMockService()
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
			server := httptest.NewServer(newRouter(test.service, test.logger))
			defer server.Close()

			req, err := http.NewRequest(http.MethodDelete, server.URL+"/spots/"+test.id, nil)
			assert.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			test.expectedResponseFn(t, resp)
		})
	}
}
