package geo

import "errors"

var (
	ErrLocationNotFound = errors.New("location not found")
)

type LocationSource interface {
	Location(Coordinates) (Location, error)
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

type Location struct {
	Coordinates
	Locality    string
	CountryCode string
}
