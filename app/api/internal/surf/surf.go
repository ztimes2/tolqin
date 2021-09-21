package surf

import (
	"errors"
	"time"

	"github.com/ztimes2/tolqin/app/api/internal/geo"
)

var (
	ErrSpotNotFound         = errors.New("spot not found")
	ErrEmptySpotUpdateEntry = errors.New("empty spot update entry")
)

type Spot struct {
	ID        string
	Name      string
	CreatedAt time.Time
	Location  geo.Location
}

type SpotReader interface {
	Spot(id string) (Spot, error)
	Spots(SpotsParams) ([]Spot, error)
}

type SpotsParams struct {
	Limit       int
	Offset      int
	CountryCode string
	SearchQuery SpotSearchQuery
	Bounds      *geo.Bounds
}

type SpotSearchQuery struct {
	Query      string
	WithSpotID bool
}

type SpotWriter interface {
	CreateSpot(CreateSpotParams) (Spot, error)
	UpdateSpot(UpdateSpotParams) (Spot, error)
	DeleteSpot(id string) error
}

type CreateSpotParams struct {
	Location geo.Location
	Name     string
}

type UpdateSpotParams struct {
	ID          string
	Name        *string
	Latitude    *float64
	Longitude   *float64
	Locality    *string
	CountryCode *string
}
