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
		return validation.NewError("latitude")
	}
	if c.Longitude < minLongitude || c.Longitude > maxLongitude {
		return validation.NewError("longitude")
	}
	return nil
}

type CoordinateRange struct {
	NorthWest Coordinates
	SouthEast Coordinates
}

func (cr CoordinateRange) Validate() error {
	if err := cr.NorthWest.Validate(); err != nil {
		vErr := (err).(*validation.Error)
		return validation.NewError("north west " + vErr.Field())
	}
	if err := cr.SouthEast.Validate(); err != nil {
		vErr := (err).(*validation.Error)
		return validation.NewError("south east " + vErr.Field())
	}
	return nil
}

type Location struct {
	Coordinates
	Locality    string
	CountryCode string
}
