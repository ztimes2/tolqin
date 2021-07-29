package importing

import (
	"errors"
	"fmt"

	"github.com/ztimes2/tolqin/internal/surfing"
)

var (
	ErrNothingToImport = errors.New("nothing to import")
)

type SpotImporter interface {
	ImportSpots([]SpotEntry) ([]surfing.Spot, error)
}

type SpotEntry struct {
	Name      string
	Latitude  float64
	Longitude float64
}

type SpotEntrySource interface {
	SpotEntries() ([]SpotEntry, error)
}

func ImportSpots(src SpotEntrySource, importer SpotImporter) ([]surfing.Spot, error) {
	entries, err := src.SpotEntries()
	if err != nil {
		return nil, fmt.Errorf("failed to read spot entries from source: %w", err)
	}

	spots, err := importer.ImportSpots(entries)
	if err != nil {
		return nil, fmt.Errorf("failed to import spots: %w", err)
	}

	return spots, nil
}
