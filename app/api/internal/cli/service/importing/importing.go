package importing

import (
	"fmt"

	"github.com/ztimes2/tolqin/app/api/internal/pkg/surf"
)

func ImportSpots(src surf.SpotCreationEntrySource, dest surf.MultiSpotWriter) (int, error) {
	entries, err := src.SpotCreationEntries()
	if err != nil {
		return 0, fmt.Errorf("could not read spot entries from source: %w", err)
	}

	// TODO sanitize each entry
	// TODO validate each entry

	if err := dest.CreateSpots(entries); err != nil {
		return 0, fmt.Errorf("could not create spots in the destination: %w", err)
	}

	return len(entries), nil
}
