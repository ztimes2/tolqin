package surfing

import (
	"errors"
	"strings"
	"time"

	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/pagination"
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

type Service struct {
	spotStore SpotStore
}

func NewService(s SpotStore) *Service {
	return &Service{
		spotStore: s,
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
