package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/ztimes2/tolqin/app/api/internal/importer/importing"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/geo"
)

type SpotEntrySource struct {
	reader io.Reader
}

func NewSpotEntrySource(r io.Reader) SpotEntrySource {
	return SpotEntrySource{
		reader: r,
	}
}

func (ss SpotEntrySource) SpotEntries() ([]importing.SpotEntry, error) {
	records, err := csv.NewReader(ss.reader).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read csv: %w", err)
	}

	if len(records) <= 1 {
		return nil, nil
	}

	var entries []importing.SpotEntry
	for _, r := range records[1:] {
		if len(r) != 5 {
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
			Name: r[0],
			Location: geo.Location{
				Locality:    r[3],
				CountryCode: r[4],
				Coordinates: geo.Coordinates{
					Latitude:  lat,
					Longitude: long,
				},
			},
		})
	}

	return entries, nil
}
