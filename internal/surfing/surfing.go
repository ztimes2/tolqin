package surfing

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ztimes2/tolqin/internal/geo"
	"github.com/ztimes2/tolqin/internal/pagination"
	"github.com/ztimes2/tolqin/internal/pconv"
	"github.com/ztimes2/tolqin/internal/validation"
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
	CreateSpot(CreateLocalizedSpotParams) (Spot, error)
	UpdateSpot(UpdateLocalizedSpotParams) (Spot, error)
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

type CreateLocalizedSpotParams struct {
	Location geo.Location
	Name     string
}

type UpdateLocalizedSpotParams struct {
	Location *geo.Location
	ID       string
	Name     *string
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

type CreateSpotParams struct {
	Coordinates geo.Coordinates
	Name        string
}

func (p CreateSpotParams) sanitize() CreateSpotParams {
	p.Name = strings.TrimSpace(p.Name)
	return p
}

func (p CreateSpotParams) validate() error {
	if p.Name == "" {
		return validation.NewError("name")
	}
	return p.Coordinates.Validate()
}

func (s *Service) CreateSpot(p CreateSpotParams) (Spot, error) {
	p = p.sanitize()

	if err := p.validate(); err != nil {
		return Spot{}, err
	}

	l, err := localize(s.locationSource, p.Coordinates)
	if err != nil {
		return Spot{}, err
	}

	return s.spotStore.CreateSpot(CreateLocalizedSpotParams{
		Name:     p.Name,
		Location: l,
	})
}

type UpdateSpotParams struct {
	Coordinates *geo.Coordinates
	ID          string
	Name        *string
}

func (p UpdateSpotParams) sanitize() UpdateSpotParams {
	sanitized := UpdateSpotParams{
		Coordinates: p.Coordinates,
		ID:          strings.TrimSpace(p.ID),
	}
	if p.Name != nil {
		sanitized.Name = pconv.String(strings.TrimSpace(*p.Name))
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
	if p.Coordinates != nil {
		if err := p.Coordinates.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) UpdateSpot(p UpdateSpotParams) (Spot, error) {
	p = p.sanitize()

	if err := p.validate(); err != nil {
		return Spot{}, err
	}

	localized := UpdateLocalizedSpotParams{
		ID:   p.ID,
		Name: p.Name,
	}
	if p.Coordinates != nil {
		l, err := localize(s.locationSource, *p.Coordinates)
		if err != nil {
			return Spot{}, err
		}
		localized.Location = &l
	}

	return s.spotStore.UpdateSpot(localized)
}

func (s *Service) DeleteSpot(id string) error {
	return s.spotStore.DeleteSpot(strings.TrimSpace(id))
}

func localize(src geo.LocationSource, c geo.Coordinates) (geo.Location, error) {
	l, err := src.Location(c)
	if err != nil {
		if !errors.Is(err, geo.ErrLocationNotFound) {
			return geo.Location{}, fmt.Errorf("failed to fetch location: %w", err)
		}
		return geo.Location{
			Coordinates: c,
		}, nil
	}
	return l, nil
}
