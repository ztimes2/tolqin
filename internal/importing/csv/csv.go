package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/ztimes2/tolqin/internal/importing"
)

type SpotEntries struct {
	reader io.Reader
}

func NewSpotEntries(r io.Reader) SpotEntries {
	return SpotEntries{
		reader: r,
	}
}

func (se SpotEntries) SpotEntries() ([]importing.SpotEntry, error) {
	records, err := csv.NewReader(se.reader).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv: %w", err)
	}

	var entries []importing.SpotEntry
	for _, r := range records {
		if len(r) != 3 {
			return nil, errors.New("invalid csv record: must contain exactly 3 fields")
		}

		lat, err := strconv.ParseFloat(r[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid latitide: %w", err)
		}

		long, err := strconv.ParseFloat(r[2], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid longitude: %w", err)
		}

		entries = append(entries, importing.SpotEntry{
			Name:      r[0],
			Latitude:  lat,
			Longitude: long,
		})
	}

	return entries, nil
}
