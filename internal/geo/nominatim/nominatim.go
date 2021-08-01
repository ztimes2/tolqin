package nominatim

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ztimes2/tolqin/internal/geo"
)

const (
	endpointReverseGeocoding = "/reverse"

	formatJSON = "json"

	headerAcceptLanguage = "Accept-Language"
	languageCodeEnglish  = "en"
)

type Nominatim struct {
	client  *http.Client
	baseURL string
}

type Config struct {
	BaseURL string
	Timeout time.Duration
}

func New(cfg Config) *Nominatim {
	return &Nominatim{
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		baseURL: cfg.BaseURL,
	}
}

func (n *Nominatim) Location(c geo.Coordinates) (geo.Location, error) {
	req, err := http.NewRequest(http.MethodGet, n.baseURL+endpointReverseGeocoding, nil)
	if err != nil {
		return geo.Location{}, fmt.Errorf("failed to prepare request: %w", err)
	}

	q := url.Values{
		"lat":    []string{floatToString(c.Latitude)},
		"lon":    []string{floatToString(c.Longitude)},
		"format": []string{formatJSON},
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set(headerAcceptLanguage, languageCodeEnglish)

	resp, err := n.client.Do(req)
	if err != nil {
		return geo.Location{}, fmt.Errorf("failed to send request: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return geo.Location{}, fmt.Errorf("failed to read response body: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return geo.Location{}, fmt.Errorf("unsuccessful response: %s %s", resp.Status, string(body))
	}

	var r reverseGeocodingResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return geo.Location{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	l := r.locality()
	if l == "" {
		fmt.Println(string(body))
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
	Error   string `json:"error"`
	Address struct {
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
	} `json:"address"`
}

func (r reverseGeocodingResponse) hasError() bool {
	return r.Error != ""
}

func (r reverseGeocodingResponse) toLocation(c geo.Coordinates) geo.Location {
	return geo.Location{
		CountryCode: r.Address.CountryCode,
		Locality:    r.locality(),
		Coordinates: c,
	}
}

func (r reverseGeocodingResponse) locality() string {
	if r.Address.Hamlet != "" {
		return r.Address.Hamlet
	}
	if r.Address.Village != "" {
		return r.Address.Village
	}
	if r.Address.Town != "" {
		return r.Address.Town
	}
	if r.Address.City != "" {
		return r.Address.City
	}
	if r.Address.CityDistrict != "" {
		return r.Address.CityDistrict
	}
	if r.Address.Municipality != "" {
		return r.Address.Municipality
	}
	if r.Address.County != "" {
		return r.Address.County
	}
	if r.Address.State != "" {
		return r.Address.State
	}
	if r.Address.Territory != "" {
		return r.Address.Territory
	}
	if r.Address.Region != "" {
		return r.Address.Region
	}
	return ""
}
