package surfing

import (
	"errors"
	"time"

	"github.com/go-playground/validator"
)

var (
	ErrNotFound        = errors.New("resource not found")
	ErrNothingToUpdate = errors.New("nothing to update")
)

type SpotStore interface {
	Spot(id string) (Spot, error)
	Spots(limit, offset int) ([]Spot, error)
	CreateSpot(CreateSpotParams) (Spot, error)
	UpdateSpot(UpdateSpotParams) (Spot, error)
	DeleteSpot(id string) error
}

type Spot struct {
	ID        string
	Name      string
	Latitude  float64
	Longitude float64
	CreatedAt time.Time
}

type CreateSpotParams struct {
	Name      string
	Latitude  float64
	Longitude float64
}

type UpdateSpotParams struct {
	ID        string
	Name      *string
	Latitude  *float64
	Longitude *float64
}

type Service struct {
	validate  *validator.Validate
	spotStore SpotStore
}

func NewService(v *validator.Validate, spotStore SpotStore) *Service {
	return &Service{
		validate:  v,
		spotStore: spotStore,
	}
}

func (s *Service) Spot(id string) (Spot, error) {
	return s.spotStore.Spot(id)
}

func (s *Service) Spots(limit, offset int) ([]Spot, error) {
	return s.spotStore.Spots(limit, offset)
}

func (s *Service) CreateSpot(p CreateSpotParams) (Spot, error) {
	if err := s.validate.Struct(&p); err != nil {
		return Spot{}, err
	}
	return s.spotStore.CreateSpot(p)
}

func (s *Service) UpdateSpot(p UpdateSpotParams) (Spot, error) {
	if err := s.validate.Struct(&p); err != nil {
		return Spot{}, err
	}
	return s.spotStore.UpdateSpot(p)
}

func (s *Service) DeleteSpot(id string) error {
	return s.spotStore.DeleteSpot(id)
}
