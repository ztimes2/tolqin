package importing

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/validation"
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
	se.Location = se.Location.Sanitize()
	return se
}

func (se SpotEntry) validate() error {
	if err := se.Location.Validate(); err != nil {
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
