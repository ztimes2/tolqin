package management

import (
	"errors"
	"strings"
	"time"

	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/pagination"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/pconv"
	"github.com/ztimes2/tolqin/app/api/internal/validation"
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
	p.Limit = pagination.Limit(p.Limit, minLimit, maxLimit, defaultLimit)
	p.Offset = pagination.Offset(p.Offset, minOffset)
	p.CountryCode = strings.ToLower(strings.TrimSpace(p.CountryCode))
	p.Query = strings.TrimSpace(p.Query)
	return p
}

func (p SpotsParams) validate() error {
	if p.CountryCode != "" && !geo.IsCountry(p.CountryCode) {
		return validation.NewError("country code")
	}
	if len(p.Query) > maxQueryChars {
		return validation.NewError("query")
	}
	if p.Bounds != nil {
		if err := p.Bounds.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type CreateSpotParams struct {
	Location geo.Location
	Name     string
}

func (p CreateSpotParams) sanitize() CreateSpotParams {
	p.Name = strings.TrimSpace(p.Name)
	p.Location = p.Location.Sanitize()
	return p
}

func (p CreateSpotParams) validate() error {
	if p.Name == "" {
		return validation.NewError("name")
	}
	return p.Location.Validate()
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
	if p.ID == "" {
		return validation.NewError("id")
	}
	if p.Name != nil && *p.Name == "" {
		return validation.NewError("name")
	}
	if p.Latitude != nil && !geo.IsLatitude(*p.Latitude) {
		return validation.NewError("latitude")
	}
	if p.Longitude != nil && !geo.IsLongitude(*p.Longitude) {
		return validation.NewError("longitude")
	}
	if p.Locality != nil && *p.Locality == "" {
		return validation.NewError("locality")
	}
	if p.CountryCode != nil && !geo.IsCountry(*p.CountryCode) {
		return validation.NewError("country code")
	}
	return nil
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
	return s.spotStore.Spot(strings.TrimSpace(id))
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
	return s.spotStore.DeleteSpot(strings.TrimSpace(id))
}

func (s *Service) Location(c geo.Coordinates) (geo.Location, error) {
	if err := c.Validate(); err != nil {
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
