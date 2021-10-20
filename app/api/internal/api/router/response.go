package router

import "github.com/ztimes2/tolqin/app/api/internal/pkg/surf"

type spotResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Locality    string  `json:"locality"`
	CountryCode string  `json:"country_code"`
}

func toSpotResponse(s surf.Spot) spotResponse {
	return spotResponse{
		ID:          s.ID,
		Name:        s.Name,
		Latitude:    s.Location.Coordinates.Latitude,
		Longitude:   s.Location.Coordinates.Longitude,
		Locality:    s.Location.Locality,
		CountryCode: s.Location.CountryCode,
	}
}

type spotsResponse struct {
	Items []spotResponse `json:"items"`
}

type locationResponse struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Locality    string  `json:"locality"`
	CountryCode string  `json:"country_code"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}
