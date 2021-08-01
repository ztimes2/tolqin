package importing

import (
	"errors"
	"fmt"

	"github.com/ztimes2/tolqin/internal/geo"
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
	geo.Location
	Name string
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
