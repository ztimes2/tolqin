package surfing

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator"
	"github.com/ztimes2/tolqin/internal/geo"
	"github.com/ztimes2/tolqin/internal/p8n"
)

const (
	minLimit     = 1
	maxLimit     = 100
	defaultLimit = 10

	minOffset = 0
)

var (
	ErrNotFound        = errors.New("resource not found")
	ErrNothingToUpdate = errors.New("nothing to update")
)

type SpotStore interface {
	Spot(id string) (Spot, error)
	Spots(limit, offset int) ([]Spot, error)
	CreateSpot(CreateLocalizedSpotParams) (Spot, error)
	UpdateSpot(UpdateLocalizedSpotParams) (Spot, error)
	DeleteSpot(id string) error
}

type Spot struct {
	geo.Location
	ID        string
	Name      string
	CreatedAt time.Time
}

type CreateLocalizedSpotParams struct {
	geo.Location
	Name string
}

type UpdateLocalizedSpotParams struct {
	*geo.Location
	ID   string
	Name *string
}

type Service struct {
	validate       *validator.Validate
	spotStore      SpotStore
	locationSource geo.LocationSource
}

func NewService(
	v *validator.Validate,
	s SpotStore,
	l geo.LocationSource,
) *Service {
	return &Service{
		validate:       v,
		spotStore:      s,
		locationSource: l,
	}
}

func (s *Service) Spot(id string) (Spot, error) {
	return s.spotStore.Spot(id)
}

func (s *Service) Spots(limit, offset int) ([]Spot, error) {
	return s.spotStore.Spots(
		p8n.Limit(limit, minLimit, maxLimit, defaultLimit),
		p8n.Offset(offset, minOffset),
	)
}

type CreateSpotParams struct {
	geo.Coordinates
	Name string
}

func (s *Service) CreateSpot(p CreateSpotParams) (Spot, error) {
	if err := s.validate.Struct(&p); err != nil {
		return Spot{}, fmt.Errorf("invalid params: %w", err)
	}

	l, err := s.locationSource.Location(p.Coordinates)
	if err != nil {
		if !errors.Is(err, geo.ErrLocationNotFound) {
			return Spot{}, fmt.Errorf("failed to fetch location: %w", err)
		}
		l = geo.Location{
			Coordinates: p.Coordinates,
		}
	}

	return s.spotStore.CreateSpot(CreateLocalizedSpotParams{
		Name:     p.Name,
		Location: l,
	})
}

type UpdateSpotParams struct {
	*geo.Coordinates
	ID   string
	Name *string
}

func (s *Service) UpdateSpot(p UpdateSpotParams) (Spot, error) {
	if err := s.validate.Struct(&p); err != nil {
		return Spot{}, err
	}

	localized := UpdateLocalizedSpotParams{
		ID:   p.ID,
		Name: p.Name,
	}
	if p.Coordinates != nil {
		l, err := s.locationSource.Location(*p.Coordinates)
		if err != nil {
			if !errors.Is(err, geo.ErrLocationNotFound) {
				return Spot{}, fmt.Errorf("failed to fetch location: %w", err)
			}
			l = geo.Location{
				Coordinates: *p.Coordinates,
			}
		}

		localized.Location = &l
	}

	return s.spotStore.UpdateSpot(localized)
}

func (s *Service) DeleteSpot(id string) error {
	return s.spotStore.DeleteSpot(id)
}
