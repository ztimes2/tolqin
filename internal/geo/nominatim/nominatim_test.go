package nominatim

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ztimes2/tolqin/internal/geo"
	"github.com/ztimes2/tolqin/internal/testutil"
)

func TestNominatim_Location(t *testing.T) {
	tests := []struct {
		name             string
		handlerFn        func(t *testing.T) http.HandlerFunc
		coord            geo.Coordinates
		expectedLocation geo.Location
		expectedErrFn    assert.ErrorAssertionFunc
	}{
		{
			name: "return error for response with non-200 http status code",
			handlerFn: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "1.23", r.URL.Query().Get(queryParamLatitude))
					assert.Equal(t, "3.21", r.URL.Query().Get(queryParamLongitude))
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":"Something went wrong."}`))
				}
			},
			coord: geo.Coordinates{
				Latitude:  1.23,
				Longitude: 3.21,
			},
			expectedLocation: geo.Location{},
			expectedErrFn:    assert.Error,
		},
		{
			name: "return error for response with unexpected body",
			handlerFn: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "1.23", r.URL.Query().Get(queryParamLatitude))
					assert.Equal(t, "3.21", r.URL.Query().Get(queryParamLongitude))
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(nil)
				}
			},
			coord: geo.Coordinates{
				Latitude:  1.23,
				Longitude: 3.21,
			},
			expectedLocation: geo.Location{},
			expectedErrFn:    assert.Error,
		},
		{
			name: "return error for response with error body",
			handlerFn: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "1.23", r.URL.Query().Get(queryParamLatitude))
					assert.Equal(t, "3.21", r.URL.Query().Get(queryParamLongitude))
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{"error":"Location not found."}`))
				}
			},
			coord: geo.Coordinates{
				Latitude:  1.23,
				Longitude: 3.21,
			},
			expectedLocation: geo.Location{},
			expectedErrFn:    testutil.IsError(geo.ErrLocationNotFound),
		},
		{
			name: "return location for response with address body",
			handlerFn: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "1.23", r.URL.Query().Get(queryParamLatitude))
					assert.Equal(t, "3.21", r.URL.Query().Get(queryParamLongitude))
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(
						`{
							"address": {
								"country_code": "Country code",
								"region": "Region",
								"territory": "Territory",
								"state": "State",
								"county": "County",
								"municipality": "Municipality",
								"city_district": "City district",
								"city": "City",
								"town": "Town",
								"village": "Village",
								"hamlet": "Hamlet"
							}
						}`,
					))
				}
			},
			coord: geo.Coordinates{
				Latitude:  1.23,
				Longitude: 3.21,
			},
			expectedLocation: geo.Location{
				Coordinates: geo.Coordinates{
					Latitude:  1.23,
					Longitude: 3.21,
				},
				CountryCode: "Country code",
				Locality:    "Hamlet",
			},
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, endpointReverseGeocoding, r.URL.Path)
				assert.Equal(t, languageCodeEnglish, r.Header.Get(headerAcceptLanguage))
				assert.Equal(t, formatJSON, r.URL.Query().Get(queryParamFormat))
				test.handlerFn(t)(w, r)
			}))
			defer server.Close()

			n := New(Config{
				BaseURL: server.URL,
			})

			location, err := n.Location(test.coord)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedLocation, location)
		})
	}
}

func TestReverseGeocodingAddressResponse_Locality(t *testing.T) {
	tests := []struct {
		name             string
		resp             reverseGeocodingAddressResponse
		expectedLocality string
	}{
		{
			name: "return hamlet",
			resp: reverseGeocodingAddressResponse{
				Region:       "Region",
				Territory:    "Territory",
				State:        "State",
				County:       "County",
				Municipality: "Municipality",
				CityDistrict: "City district",
				City:         "City",
				Town:         "Town",
				Village:      "Village",
				Hamlet:       "Hamlet",
			},
			expectedLocality: "Hamlet",
		},
		{
			name: "return village",
			resp: reverseGeocodingAddressResponse{
				Region:       "Region",
				Territory:    "Territory",
				State:        "State",
				County:       "County",
				Municipality: "Municipality",
				CityDistrict: "City district",
				City:         "City",
				Town:         "Town",
				Village:      "Village",
			},
			expectedLocality: "Village",
		},
		{
			name: "return town",
			resp: reverseGeocodingAddressResponse{
				Region:       "Region",
				Territory:    "Territory",
				State:        "State",
				County:       "County",
				Municipality: "Municipality",
				CityDistrict: "City district",
				City:         "City",
				Town:         "Town",
			},
			expectedLocality: "Town",
		},
		{
			name: "return city",
			resp: reverseGeocodingAddressResponse{
				Region:       "Region",
				Territory:    "Territory",
				State:        "State",
				County:       "County",
				Municipality: "Municipality",
				CityDistrict: "City district",
				City:         "City",
			},
			expectedLocality: "City",
		},
		{
			name: "return city district",
			resp: reverseGeocodingAddressResponse{
				Region:       "Region",
				Territory:    "Territory",
				State:        "State",
				County:       "County",
				Municipality: "Municipality",
				CityDistrict: "City district",
			},
			expectedLocality: "City district",
		},
		{
			name: "return municipality",
			resp: reverseGeocodingAddressResponse{
				Region:       "Region",
				Territory:    "Territory",
				State:        "State",
				County:       "County",
				Municipality: "Municipality",
			},
			expectedLocality: "Municipality",
		},
		{
			name: "return county",
			resp: reverseGeocodingAddressResponse{
				Region:    "Region",
				Territory: "Territory",
				State:     "State",
				County:    "County",
			},
			expectedLocality: "County",
		},
		{
			name: "return state",
			resp: reverseGeocodingAddressResponse{
				Region:    "Region",
				Territory: "Territory",
				State:     "State",
			},
			expectedLocality: "State",
		},
		{
			name: "return territory",
			resp: reverseGeocodingAddressResponse{
				Region:    "Region",
				Territory: "Territory",
			},
			expectedLocality: "Territory",
		},
		{
			name: "return region",
			resp: reverseGeocodingAddressResponse{
				Region: "Region",
			},
			expectedLocality: "Region",
		},
		{
			name:             "return empty string",
			resp:             reverseGeocodingAddressResponse{},
			expectedLocality: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			locality := test.resp.locality()
			assert.Equal(t, test.expectedLocality, locality)
		})
	}
}
