package geo

import (
	"errors"
)

const (
	minLatitude float64 = -90
	maxLatitude float64 = 90

	minLongitude float64 = -180
	maxLongitude float64 = 180
)

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

func IsLatitude(lat float64) bool {
	return minLatitude <= lat && lat <= maxLatitude
}

func IsLongitude(lon float64) bool {
	return minLongitude <= lon && lon <= maxLongitude
}

type Location struct {
	Locality    string
	CountryCode string
	Coordinates Coordinates
}

type Bounds struct {
	NorthEast Coordinates
	SouthWest Coordinates
}
