package geo

import (
	"errors"
	"strings"

	"github.com/ztimes2/tolqin/internal/validation"
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

func (c Coordinates) Validate() error {
	if !IsLatitude(c.Latitude) {
		return validation.NewError("latitude")
	}
	if !IsLongitude(c.Longitude) {
		return validation.NewError("longitude")
	}
	return nil
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

func (l Location) Sanitize() Location {
	l.CountryCode = strings.TrimSpace(l.CountryCode)
	l.Locality = strings.TrimSpace(l.Locality)
	return l
}

func (l Location) Validate() error {
	if l.Locality == "" {
		return validation.NewError("locality")
	}
	if !IsCountry(l.CountryCode) {
		return validation.NewError("country code")
	}
	if err := l.Coordinates.Validate(); err != nil {
		return err
	}
	return nil
}

type Bounds struct {
	NorthEast Coordinates
	SouthWest Coordinates
}

func (b Bounds) Validate() error {
	if err := b.NorthEast.Validate(); err != nil {
		return validation.NewError("north-east coordinates")
	}
	if err := b.SouthWest.Validate(); err != nil {
		return validation.NewError("south-west coordinates")
	}
	return nil
}
