package inmemory

import (
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ztimes2/tolqin/backend/internal/surfing"
)

type SpotStore struct {
	m     *sync.Mutex
	spots map[string]surfing.Spot
}

func NewSpotStore() *SpotStore {
	return &SpotStore{
		m:     &sync.Mutex{},
		spots: make(map[string]surfing.Spot),
	}
}

func (s *SpotStore) Spot(id string) (surfing.Spot, error) {
	s.m.Lock()
	defer s.m.Unlock()

	spot, ok := s.spots[id]
	if !ok {
		return surfing.Spot{}, surfing.ErrNotFound
	}

	return spot, nil
}

func (s *SpotStore) Spots() ([]surfing.Spot, error) {
	s.m.Lock()
	defer s.m.Unlock()

	var spots []surfing.Spot
	for _, spot := range s.spots {
		spots = append(spots, spot)
	}

	sort.SliceStable(spots, func(i, j int) bool {
		return spots[i].CreatedAt.Before(spots[j].CreatedAt)
	})

	return spots, nil
}

func (s *SpotStore) CreateSpot(p surfing.CreateSpotParams) (surfing.Spot, error) {
	s.m.Lock()
	defer s.m.Unlock()

	spot := surfing.Spot{
		ID:        uuid.New().String(),
		Name:      p.Name,
		Latitude:  p.Latitude,
		Longitude: p.Longitude,
		CreatedAt: time.Now(),
	}

	s.spots[spot.ID] = spot

	return spot, nil
}

func (s *SpotStore) UpdateSpot(p surfing.UpdateSpotParams) (surfing.Spot, error) {
	s.m.Lock()
	defer s.m.Unlock()

	spot, ok := s.spots[p.ID]
	if !ok {
		return surfing.Spot{}, surfing.ErrNotFound
	}

	if p.Name != nil {
		spot.Name = *p.Name
	}
	if p.Latitude != nil {
		spot.Latitude = *p.Latitude
	}
	if p.Longitude != nil {
		spot.Longitude = *p.Longitude
	}

	s.spots[p.ID] = spot

	return spot, nil
}

func (s *SpotStore) DeleteSpot(id string) error {
	s.m.Lock()
	defer s.m.Unlock()

	spot, ok := s.spots[id]
	if !ok {
		return surfing.ErrNotFound
	}

	delete(s.spots, spot.ID)

	return nil
}
