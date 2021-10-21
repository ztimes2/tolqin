package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/ztimes2/tolqin/app/api/internal/pkg/geo"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/surf"
)

type SpotCreationEntrySource struct {
	reader io.Reader
}

func NewSpotCreationEntrySource(r io.Reader) *SpotCreationEntrySource {
	return &SpotCreationEntrySource{
		reader: r,
	}
}

func (s *SpotCreationEntrySource) SpotCreationEntries() ([]surf.SpotCreationEntry, error) {
	records, err := csv.NewReader(s.reader).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not read csv: %w", err)
	}

	if len(records) <= 1 {
		return nil, nil
	}

	var entries []surf.SpotCreationEntry
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

		entries = append(entries, surf.SpotCreationEntry{
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
