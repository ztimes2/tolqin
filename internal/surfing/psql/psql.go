package psql

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/ztimes2/tolqin/internal/surfing"
)

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

func (ss *SpotStore) Spots(limit, offset int) ([]surfing.Spot, error) {
	query, args, err := ss.builder.
		Select("id", "name", "latitude", "longitude", "created_at").
		From("spots").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
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
