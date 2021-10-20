package nominatim

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ztimes2/tolqin/app/api/internal/pkg/geo"
)

const (
	endpointReverseGeocoding = "/reverse"

	headerAcceptLanguage = "Accept-Language"

	queryParamLatitude  = "lat"
	queryParamLongitude = "lon"
	queryParamFormat    = "format"

	formatJSON          = "json"
	languageCodeEnglish = "en"
)

// Nominatim is an adapter for communicating with the Notimatim API.
type Nominatim struct {
	client  *http.Client
	baseURL string
}

// Config holds configuration for connecting to the Nominatim API.
type Config struct {
	BaseURL string
	Timeout time.Duration
}

// New returns a new *Nominatim.
func New(cfg Config) *Nominatim {
	return &Nominatim{
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		baseURL: cfg.BaseURL,
	}
}

// Location implements geo.LocationSource interface and fetches a location by the
// given coordinates. ErrLocationNotFound is returned when location is not found.
func (n *Nominatim) Location(c geo.Coordinates) (geo.Location, error) {
	req, err := http.NewRequest(http.MethodGet, n.baseURL+endpointReverseGeocoding, nil)
	if err != nil {
		return geo.Location{}, fmt.Errorf("failed to prepare request: %w", err)
	}

	q := url.Values{
		queryParamLatitude:  []string{floatToString(c.Latitude)},
		queryParamLongitude: []string{floatToString(c.Longitude)},
		queryParamFormat:    []string{formatJSON},
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set(headerAcceptLanguage, languageCodeEnglish)

	resp, err := n.client.Do(req)
	if err != nil {
		return geo.Location{}, fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return geo.Location{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != 200 {
		return geo.Location{}, fmt.Errorf("unsuccessful response: %s %s", resp.Status, string(body))
	}

	var r reverseGeocodingResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return geo.Location{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if r.hasError() {
		return geo.Location{}, geo.ErrLocationNotFound
	}

	return r.toLocation(c), nil
}

func floatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

type reverseGeocodingResponse struct {
	Error   string                          `json:"error"`
	Address reverseGeocodingAddressResponse `json:"address"`
}

func (r reverseGeocodingResponse) hasError() bool {
	return r.Error != ""
}

func (r reverseGeocodingResponse) toLocation(c geo.Coordinates) geo.Location {
	return geo.Location{
		CountryCode: r.Address.CountryCode,
		Locality:    r.Address.locality(),
		Coordinates: c,
	}
}

type reverseGeocodingAddressResponse struct {
	CountryCode  string `json:"country_code"`
	Region       string `json:"region"`
	Territory    string `json:"territory"`
	State        string `json:"state"`
	County       string `json:"county"`
	Municipality string `json:"municipality"`
	CityDistrict string `json:"city_district"`
	City         string `json:"city"`
	Town         string `json:"town"`
	Village      string `json:"village"`
	Hamlet       string `json:"hamlet"`
}

func (r reverseGeocodingAddressResponse) locality() string {
	if r.Hamlet != "" {
		return r.Hamlet
	}
	if r.Village != "" {
		return r.Village
	}
	if r.Town != "" {
		return r.Town
	}
	if r.City != "" {
		return r.City
	}
	if r.CityDistrict != "" {
		return r.CityDistrict
	}
	if r.Municipality != "" {
		return r.Municipality
	}
	if r.County != "" {
		return r.County
	}
	if r.State != "" {
		return r.State
	}
	if r.Territory != "" {
		return r.Territory
	}
	if r.Region != "" {
		return r.Region
	}
	return ""
}
