package surfer

import (
	"errors"
	"strings"
	"time"

	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/paging"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/valerra"
	"github.com/ztimes2/tolqin/app/api/internal/valerrautil"
)

const (
	minLimit     = 1
	maxLimit     = 100
	defaultLimit = 10

	minOffset = 0

	maxQueryChars = 100
)

var (
	ErrNotFound        = errors.New("resource not found")
	ErrNothingToUpdate = errors.New("nothing to update")

	ErrInvalidSearchQuery        = errors.New("invalid search query")
	ErrInvalidCountryCode        = errors.New("invalid country code")
	ErrInvalidNorthEastLatitude  = errors.New("invalid north-east latitude")
	ErrInvalidNorthEastLongitude = errors.New("invalid north-east longitude")
	ErrInvalidSouthWestLatitude  = errors.New("invalid south-west latitude")
	ErrInvalidSouthWestLongitude = errors.New("invalid south-west longitude")
	ErrInvalidSpotID             = errors.New("invalid spot id")
)

type SpotStore interface {
	Spot(id string) (Spot, error)
	Spots(SpotsParams) ([]Spot, error)
}

type Spot struct {
	ID        string
	Name      string
	CreatedAt time.Time
	Location  geo.Location
}

type SpotsParams struct {
	Limit       int
	Offset      int
	CountryCode string
	Query       string
	Bounds      *geo.Bounds
}

func (p SpotsParams) sanitize() SpotsParams {
	p.Limit = paging.Limit(p.Limit, minLimit, maxLimit, defaultLimit)
	p.Offset = paging.Offset(p.Offset, minOffset)
	p.CountryCode = strings.ToLower(strings.TrimSpace(p.CountryCode))
	p.Query = strings.TrimSpace(p.Query)
	return p
}

func (p SpotsParams) validate() error {
	v := valerra.New()

	v.IfFalse(valerra.StringLessOrEqual(p.Query, maxQueryChars), ErrInvalidSearchQuery)
	if p.CountryCode != "" {
		v.IfFalse(valerrautil.IsCountry(p.CountryCode), ErrInvalidCountryCode)
	}
	if p.Bounds != nil {
		v.IfFalse(valerrautil.IsLatitude(p.Bounds.NorthEast.Latitude), ErrInvalidNorthEastLatitude)
		v.IfFalse(valerrautil.IsLongitude(p.Bounds.NorthEast.Longitude), ErrInvalidNorthEastLongitude)
		v.IfFalse(valerrautil.IsLatitude(p.Bounds.SouthWest.Latitude), ErrInvalidSouthWestLatitude)
		v.IfFalse(valerrautil.IsLongitude(p.Bounds.SouthWest.Longitude), ErrInvalidSouthWestLongitude)
	}

	return v.Validate()
}

type Service struct {
	spotStore SpotStore
}

func NewService(s SpotStore) *Service {
	return &Service{
		spotStore: s,
	}
}

func (s *Service) Spot(id string) (Spot, error) {
	id = strings.TrimSpace(id)

	if err := valerra.IfFalse(valerra.StringNotEmpty(id), ErrInvalidSpotID); err != nil {
		return Spot{}, err
	}

	return s.spotStore.Spot(id)
}

func (s *Service) Spots(p SpotsParams) ([]Spot, error) {
	p = p.sanitize()

	if err := p.validate(); err != nil {
		return nil, err
	}

	return s.spotStore.Spots(p)
}
