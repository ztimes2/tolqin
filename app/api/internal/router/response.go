package router

type spotResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Locality    string  `json:"locality"`
	CountryCode string  `json:"country_code"`
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
