package importing

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/surf"
)

type SpotEntrySource interface {
	SpotEntries() ([]SpotEntry, error)
}

type SpotEntry surf.SpotCreationEntry

func (se SpotEntry) sanitize() SpotEntry {
	se.Name = strings.TrimSpace(se.Name)
	se.Location.CountryCode = strings.TrimSpace(se.Location.CountryCode)
	se.Location.Locality = strings.TrimSpace(se.Location.Locality)
	return se
}

func (se SpotEntry) validate() error {
	if se.Name == "" {
		return errors.New("invalid spot name")
	}
	if se.Location.CountryCode == "" || !geo.IsCountry(se.Location.CountryCode) {
		return errors.New("invalid country code")
	}
	if se.Location.Locality == "" {
		return errors.New("invalid locality")
	}
	if !geo.IsLatitude(se.Location.Coordinates.Latitude) {
		return errors.New("invalid latitude")
	}
	if !geo.IsLongitude(se.Location.Coordinates.Longitude) {
		return errors.New("invalid longitude")
	}
	return nil
}

type SpotStore interface {
	surf.MultiSpotWriter
}

func ImportSpots(src SpotEntrySource, store SpotStore) (int, error) {
	ee, err := src.SpotEntries()
	if err != nil {
		return 0, fmt.Errorf("failed to read spot entries from source: %w", err)
	}

	var entries []surf.SpotCreationEntry
	for i, e := range ee {
		e = e.sanitize()

		if err := e.validate(); err != nil {
			return 0, fmt.Errorf("invalid entry #%d: %w", i+1, err)
		}

		entries = append(entries, surf.SpotCreationEntry(e))
	}

	if err := store.CreateSpots(entries); err != nil {
		return 0, fmt.Errorf("failed to import spots: %w", err)
	}

	return len(entries), nil
}
