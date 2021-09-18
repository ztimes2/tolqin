package management

import (
	"errors"
	"strings"
	"time"

	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/paging"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/pconv"
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
	ErrInvalidLocality           = errors.New("invalid locality")
	ErrInvalidCountryCode        = errors.New("invalid country code")
	ErrInvalidLatitude           = errors.New("invalid latitude")
	ErrInvalidLongitude          = errors.New("invalid longitude")
	ErrInvalidNorthEastLatitude  = errors.New("invalid north-east latitude")
	ErrInvalidNorthEastLongitude = errors.New("invalid north-east longitude")
	ErrInvalidSouthWestLatitude  = errors.New("invalid south-west latitude")
	ErrInvalidSouthWestLongitude = errors.New("invalid south-west longitude")
	ErrInvalidSpotName           = errors.New("invalid spot name")
	ErrInvalidSpotID             = errors.New("invalid spot id")
)

type SpotStore interface {
	Spot(id string) (Spot, error)
	Spots(SpotsParams) ([]Spot, error)
	CreateSpot(CreateSpotParams) (Spot, error)
	UpdateSpot(UpdateSpotParams) (Spot, error)
	DeleteSpot(id string) error
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

type CreateSpotParams struct {
	Location geo.Location
	Name     string
}

func (p CreateSpotParams) sanitize() CreateSpotParams {
	p.Name = strings.TrimSpace(p.Name)
	p.Location.CountryCode = strings.TrimSpace(p.Location.CountryCode)
	p.Location.Locality = strings.TrimSpace(p.Location.Locality)
	return p
}

func (p CreateSpotParams) validate() error {
	v := valerra.New()

	v.IfFalse(valerra.StringNotEmpty(p.Name), ErrInvalidSpotName)
	v.IfFalse(valerrautil.IsCountry(p.Location.CountryCode), ErrInvalidCountryCode)
	v.IfFalse(valerra.StringNotEmpty(p.Location.Locality), ErrInvalidLocality)
	v.IfFalse(valerrautil.IsLatitude(p.Location.Coordinates.Latitude), ErrInvalidLatitude)
	v.IfFalse(valerrautil.IsLongitude(p.Location.Coordinates.Longitude), ErrInvalidLongitude)

	return v.Validate()
}

type UpdateSpotParams struct {
	ID          string
	Name        *string
	Latitude    *float64
	Longitude   *float64
	Locality    *string
	CountryCode *string
}

func (p UpdateSpotParams) sanitize() UpdateSpotParams {
	sanitized := UpdateSpotParams{
		ID:        strings.TrimSpace(p.ID),
		Latitude:  p.Latitude,
		Longitude: p.Longitude,
	}
	if p.Name != nil {
		sanitized.Name = pconv.String(strings.TrimSpace(*p.Name))
	}
	if p.Locality != nil {
		sanitized.Locality = pconv.String(strings.TrimSpace(*p.Locality))
	}
	if p.CountryCode != nil {
		sanitized.CountryCode = pconv.String(strings.TrimSpace(*p.CountryCode))
	}
	return sanitized
}

func (p UpdateSpotParams) validate() error {
	v := valerra.New()

	v.IfFalse(valerra.StringNotEmpty(p.ID), ErrInvalidSpotID)
	if p.Name != nil {
		v.IfFalse(valerra.StringNotEmpty(*p.Name), ErrInvalidSpotName)
	}
	if p.Latitude != nil {
		v.IfFalse(valerrautil.IsLatitude(*p.Latitude), ErrInvalidLatitude)
	}
	if p.Longitude != nil {
		v.IfFalse(valerrautil.IsLongitude(*p.Longitude), ErrInvalidLongitude)
	}
	if p.Locality != nil {
		v.IfFalse(valerra.StringNotEmpty(*p.Locality), ErrInvalidLocality)
	}
	if p.CountryCode != nil {
		v.IfFalse(valerrautil.IsCountry(*p.CountryCode), ErrInvalidCountryCode)
	}

	return v.Validate()
}

type Service struct {
	spotStore      SpotStore
	locationSource geo.LocationSource
}

func NewService(s SpotStore, l geo.LocationSource) *Service {
	return &Service{
		spotStore:      s,
		locationSource: l,
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

func (s *Service) CreateSpot(p CreateSpotParams) (Spot, error) {
	p = p.sanitize()

	if err := p.validate(); err != nil {
		return Spot{}, err
	}

	return s.spotStore.CreateSpot(CreateSpotParams{
		Name:     p.Name,
		Location: p.Location,
	})
}

func (s *Service) UpdateSpot(p UpdateSpotParams) (Spot, error) {
	p = p.sanitize()

	if err := p.validate(); err != nil {
		return Spot{}, err
	}

	return s.spotStore.UpdateSpot(p)
}

func (s *Service) DeleteSpot(id string) error {
	id = strings.TrimSpace(id)

	if err := valerra.IfFalse(valerra.StringNotEmpty(id), ErrInvalidSpotID); err != nil {
		return err
	}

	return s.spotStore.DeleteSpot(id)
}

func (s *Service) Location(c geo.Coordinates) (geo.Location, error) {
	v := valerra.New()
	v.IfFalse(valerrautil.IsLatitude(c.Latitude), ErrInvalidLatitude)
	v.IfFalse(valerrautil.IsLongitude(c.Longitude), ErrInvalidLongitude)
	if err := v.Validate(); err != nil {
		return geo.Location{}, err
	}

	l, err := s.locationSource.Location(c)
	if err != nil {
		if errors.Is(err, geo.ErrLocationNotFound) {
			return geo.Location{}, ErrNotFound
		}
		return geo.Location{}, err
	}

	return l, nil
}
