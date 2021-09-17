package importing

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ztimes2/tolqin/app/api/internal/geo"
)

var (
	ErrNothingToImport = errors.New("nothing to import")
)

type SpotEntrySource interface {
	SpotEntries() ([]SpotEntry, error)
}

type SpotImporter interface {
	ImportSpots([]SpotEntry) (int, error)
}

type SpotEntry struct {
	Location geo.Location
	Name     string
}

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

func ImportSpots(src SpotEntrySource, importer SpotImporter) (int, error) {
	entries, err := src.SpotEntries()
	if err != nil {
		return 0, fmt.Errorf("failed to read spot entries from source: %w", err)
	}

	for i := range entries {
		entries[i] = entries[i].sanitize()

		if err := entries[i].validate(); err != nil {
			return 0, fmt.Errorf("invalid entry #%d: %w", i+1, err)
		}
	}

	count, err := importer.ImportSpots(entries)
	if err != nil {
		if errors.Is(err, ErrNothingToImport) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to import spots: %w", err)
	}

	return count, nil
}
