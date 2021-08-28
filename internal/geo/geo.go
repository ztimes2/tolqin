package geo

import (
	"errors"

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
	if c.Latitude < minLatitude || c.Latitude > maxLatitude {
		return validation.NewError("coordinates")
	}
	if c.Longitude < minLongitude || c.Longitude > maxLongitude {
		return validation.NewError("coordinates")
	}
	return nil
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

func (b Bounds) Validate() error {
	if err := b.NorthEast.Validate(); err != nil {
		return validation.NewError("north-east coordinates")
	}
	if err := b.SouthWest.Validate(); err != nil {
		return validation.NewError("south-west coordinates")
	}
	return nil
}
