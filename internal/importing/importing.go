package importing

import (
	"errors"
	"fmt"
)

var (
	ErrNothingToImport = errors.New("nothing to import")
)

type SpotImporter interface {
	ImportSpots([]SpotEntry) (int, error)
}

type SpotEntry struct {
	Name      string
	Latitude  float64
	Longitude float64
}

type SpotEntrySource interface {
	SpotEntries() ([]SpotEntry, error)
}

func ImportSpots(src SpotEntrySource, importer SpotImporter) (int, error) {
	entries, err := src.SpotEntries()
	if err != nil {
		return 0, fmt.Errorf("failed to read spot entries from source: %w", err)
	}

	count, err := importer.ImportSpots(entries)
	if err != nil {
		return 0, fmt.Errorf("failed to import spots: %w", err)
	}

	return count, nil
}
