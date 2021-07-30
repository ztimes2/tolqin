package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ztimes2/tolqin/internal/importing"
	"github.com/ztimes2/tolqin/internal/surfing"
)

const (
	driverName = "postgres"

	sslModeNameDisable = "disable"
)

type Config struct {
	Host         string
	Port         string
	Username     string
	Password     string
	DatabaseName string
	SSLMode      SSLMode
}

func (c Config) String() string {
	entries := []string{
		"host=" + c.Host,
		"port=" + c.Port,
		"dbname=" + c.DatabaseName,
	}
	if c.SSLMode != SSLModeUndefined {
		entries = append(entries, "sslmode="+c.SSLMode.String())
	}
	if c.Username != "" {
		entries = append(entries, "user="+c.Username)
	}
	if c.Password != "" {
		entries = append(entries, "password="+c.Password)
	}
	return strings.Join(entries, " ")
}

type SSLMode int

const (
	SSLModeUndefined SSLMode = iota
	SSLModeDisabled
)

func NewSSLMode(s string) SSLMode {
	switch s {
	case sslModeNameDisable:
		return SSLModeDisabled
	default:
		return SSLModeUndefined
	}
}

func (s SSLMode) String() string {
	switch s {
	case SSLModeDisabled:
		return sslModeNameDisable
	default:
		return ""
	}
}

func NewDB(cfg Config) (*sqlx.DB, error) {
	return sqlx.Open(driverName, cfg.String())
}

type spot struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Latitude  float64   `db:"latitude"`
	Longitude float64   `db:"longitude"`
	CreatedAt time.Time `db:"created_at"`
}

type SpotStore struct {
	db      *sqlx.DB
	builder sq.StatementBuilderType
}

func NewSpotStore(db *sqlx.DB) *SpotStore {
	return &SpotStore{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (ss *SpotStore) Spot(id string) (surfing.Spot, error) {
	query, args, err := ss.builder.
		Select("id", "name", "latitude", "longitude", "created_at").
		From("spots").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return surfing.Spot{}, fmt.Errorf("failed to build query: %w", err)
	}

	var s spot
	if err := ss.db.QueryRowx(query, args...).StructScan(&s); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return surfing.Spot{}, surfing.ErrNotFound
		}
		return surfing.Spot{}, fmt.Errorf("failed to execute query: %w", err)
	}

	return surfing.Spot(s), nil
}

func (ss *SpotStore) Spots() ([]surfing.Spot, error) {
	query, args, err := ss.builder.
		Select("id", "name", "latitude", "longitude", "created_at").
		From("spots").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := ss.db.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	var spots []surfing.Spot
	defer rows.Close()
	for rows.Next() {
		var s spot
		if err := rows.StructScan(&s); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		spots = append(spots, surfing.Spot(s))
	}

	return spots, nil
}

func (ss *SpotStore) CreateSpot(p surfing.CreateSpotParams) (surfing.Spot, error) {
	query, args, err := ss.builder.
		Insert("spots").
		Columns("name", "latitude", "longitude").
		Values(p.Name, p.Latitude, p.Longitude).
		Suffix("RETURNING id, created_at").
		ToSql()
	if err != nil {
		return surfing.Spot{}, fmt.Errorf("failed to build query: %w", err)
	}

	var s spot
	if err := ss.db.QueryRowx(query, args...).StructScan(&s); err != nil {
		return surfing.Spot{}, fmt.Errorf("failed to execute query: %w", err)
	}

	return surfing.Spot{
		Name:      p.Name,
		Latitude:  p.Latitude,
		Longitude: p.Longitude,
		ID:        s.ID,
		CreatedAt: s.CreatedAt,
	}, nil
}

func (ss *SpotStore) UpdateSpot(p surfing.UpdateSpotParams) (surfing.Spot, error) {
	values := make(map[string]interface{})
	if p.Name != nil {
		values["name"] = *p.Name
	}
	if p.Latitude != nil {
		values["latitude"] = *p.Latitude
	}
	if p.Longitude != nil {
		values["longitude"] = *p.Longitude
	}
	if len(values) == 0 {
		return surfing.Spot{}, surfing.ErrNothingToUpdate
	}

	query, args, err := ss.builder.
		Update("spots").
		SetMap(values).
		Where(sq.Eq{"id": p.ID}).
		Suffix("RETURNING id, name, latitude, longitude, created_at").
		ToSql()
	if err != nil {
		return surfing.Spot{}, fmt.Errorf("failed to build query: %w", err)
	}

	var s spot
	if err := ss.db.QueryRowx(query, args...).StructScan(&s); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return surfing.Spot{}, surfing.ErrNotFound
		}
		return surfing.Spot{}, fmt.Errorf("failed to execute query: %w", err)
	}

	return surfing.Spot(s), nil
}

func (ss *SpotStore) DeleteSpot(id string) error {
	query, args, err := ss.builder.
		Delete("spots").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	res, err := ss.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read affected rows: %w", err)
	}

	if count == 0 {
		return surfing.ErrNotFound
	}

	return nil
}

type SpotImporter struct {
	db        *sqlx.DB
	builder   sq.StatementBuilderType
	batchSize int
}

func NewSpotImporter(db *sqlx.DB, batchSize int) *SpotImporter {
	return &SpotImporter{
		db:        db,
		builder:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		batchSize: batchSize,
	}
}

func (si *SpotImporter) ImportSpots(entries []importing.SpotEntry,
) ([]surfing.Spot, error) {

	if len(entries) == 0 {
		return nil, importing.ErrNothingToImport
	}

	tx, err := si.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	var spots []surfing.Spot

	i := 0
	j := clampIntMax(si.batchSize - 1, len(entries) - 1)
	for {
		batch := entries[i : j+1]
		if len(batch) == 0 {
			break
		}

		ss, err := si.importSpots(tx, batch)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to import spots: %w", err)
		}

		spots = append(spots, ss...)

		i = j + 1
		j = clampIntMax(j+si.batchSize, len(entries)-1)
	}

	tx.Commit()
	return spots, nil
}

func clampIntMax(i, max int) int {
	if i > max {
		return max
	}
	return i
}

func (si *SpotImporter) importSpots(tx *sqlx.Tx, entries []importing.SpotEntry,
) ([]surfing.Spot, error) {

	builder := si.builder.
		Insert("spots").
		Columns("name", "latitude", "longitude")

	for _, e := range entries {
		builder = builder.Values(e.Name, e.Latitude, e.Longitude)
	}

	query, args, err := builder.
		Suffix("RETURNING id, name, latitude, longitude, created_at").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := tx.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	var spots []surfing.Spot
	defer rows.Close()
	for rows.Next() {
		var s spot
		if err := rows.StructScan(&s); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		spots = append(spots, surfing.Spot(s))
	}

	return spots, nil
}
