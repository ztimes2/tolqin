package importing

import (
	"errors"
	"fmt"

	"github.com/ztimes2/tolqin/internal/geo"
	"github.com/ztimes2/tolqin/internal/validation"
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

func (se SpotEntry) validate() error {
	if se.Location.Locality == "" {
		return validation.NewError("locality")
	}
	if !geo.IsCountry(se.Location.CountryCode) {
		return validation.NewError("country code")
	}
	if err := se.Location.Coordinates.Validate(); err != nil {
		return err
	}
	if se.Name == "" {
		return validation.NewError("name")
	}
	return nil
}

func ImportSpots(src SpotEntrySource, importer SpotImporter) (int, error) {
	entries, err := src.SpotEntries()
	if err != nil {
		return 0, fmt.Errorf("failed to read spot entries from source: %w", err)
	}

	for i, e := range entries {
		if err := e.validate(); err != nil {
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
