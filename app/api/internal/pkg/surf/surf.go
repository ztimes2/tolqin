package surf

import (
	"errors"
	"time"

	"github.com/ztimes2/tolqin/app/api/internal/pkg/geo"
)

var (
	// ErrSpotNotFound is used when a spot is not found.
	ErrSpotNotFound = errors.New("spot not found")

	// ErrEmptySpotUpdateEntry is used when a spot update entry does not contain
	// any fields.
	ErrEmptySpotUpdateEntry = errors.New("empty spot update entry")
)

// Spot represents a surfing spot.
type Spot struct {
	ID        string
	Name      string
	CreatedAt time.Time
	Location  geo.Location
}

// SpotReader is a data storage from which spots can be read.
type SpotReader interface {
	// Spot returns a spot by the given ID. ErrSpotNotFound is returned when spot
	// is not found.
	Spot(id string) (Spot, error)

	// Spots returns multiple spots that match the given parameters.
	Spots(SpotsParams) ([]Spot, error)
}

// SpotsParams holds parameters for reading multiple spots from a data storage.
type SpotsParams struct {
	Limit       int
	Offset      int
	CountryCode string
	SearchQuery SpotSearchQuery
	Bounds      *geo.Bounds
}

// SpotSearchQuery holds a string query for searching for spots. By default, the
// query is compared against names and localities of spots.
type SpotSearchQuery struct {
	Query string

	// WithSpotID can be optionally used to additionally compare the query against
	// spot IDs.
	WithSpotID bool
}

// SpotWriter is a data storage containing spots against which write operations
// can be performed.
type SpotWriter interface {
	// CreateSpot creates a new spot using the given entry and returns it if the
	// creation succeeds.
	CreateSpot(SpotCreationEntry) (Spot, error)

	// UpdateSpot updates an existing spot using the given entry and returns it
	// if the update succeeds. ErrSpotNotFound is returned when spot is not found.
	UpdateSpot(SpotUpdateEntry) (Spot, error)

	// DeleteSpot deletes a spot by the given ID. ErrSpotNotFound is returned when
	// spot is not found.
	DeleteSpot(id string) error
}

// SpotCreationEntry holds parameters for creating a new spot in a data storage.
type SpotCreationEntry struct {
	Location geo.Location
	Name     string
}

// SpotCreationEntrySource is anything that can fetch entries for creating spots.
type SpotCreationEntrySource interface {
	// SpotCreationEntries fetches and returns entries for creating spots.
	SpotCreationEntries() ([]SpotCreationEntry, error)
}

// SpotUpdateEntry holds parameters for updating a spot in a data storage. It can
// be used for both partial and full updates. In order to achieve a full update,
// all pointer-fields must not be nil. For a partial update, only the desired fields
// must not be nil.
type SpotUpdateEntry struct {
	ID          string
	Name        *string
	Latitude    *float64
	Longitude   *float64
	Locality    *string
	CountryCode *string
}

// MultiSpotWriter is a data storage containing spots against which multiple write
// operations can be performed at once.
type MultiSpotWriter interface {
	// CreateSpots creates multiple new spots using the given entries.
	CreateSpots([]SpotCreationEntry) error
}
