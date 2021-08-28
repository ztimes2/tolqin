package psql

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/ztimes2/tolqin/internal/geo"
	"github.com/ztimes2/tolqin/internal/psqlutil"
	"github.com/ztimes2/tolqin/internal/surfing"
)

type spot struct {
	ID          string         `db:"id"`
	Name        string         `db:"name"`
	Latitude    float64        `db:"latitude"`
	Longitude   float64        `db:"longitude"`
	Locality    sql.NullString `db:"locality"`
	CountryCode sql.NullString `db:"country_code"`
	CreatedAt   time.Time      `db:"created_at"`
}

func toSpot(s spot) surfing.Spot {
	return surfing.Spot{
		ID:        s.ID,
		Name:      s.Name,
		CreatedAt: s.CreatedAt,
		Location: geo.Location{
			Locality:    s.Locality.String,
			CountryCode: s.CountryCode.String,
			Coordinates: geo.Coordinates{
				Latitude:  s.Latitude,
				Longitude: s.Longitude,
			},
		},
	}
}

type SpotStore struct {
	db      *sqlx.DB
	builder sq.StatementBuilderType
}

func NewSpotStore(db *sqlx.DB) *SpotStore {
	return &SpotStore{
		db:      db,
		builder: psqlutil.NewQueryBuilder(),
	}
}

func (ss *SpotStore) Spot(id string) (surfing.Spot, error) {
	query, args, err := ss.builder.
		Select("id", "name", "latitude", "longitude", "locality", "country_code", "created_at").
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

	return toSpot(s), nil
}

func (ss *SpotStore) Spots(p surfing.SpotsParams) ([]surfing.Spot, error) {
	builder := ss.builder.
		Select("id", "name", "latitude", "longitude", "locality", "country_code", "created_at").
		From("spots").
		Limit(uint64(p.Limit)).
		Offset(uint64(p.Offset))

	if p.CountryCode != "" {
		builder = builder.Where(sq.Eq{"country_code": p.CountryCode})
	}

	query, args, err := builder.ToSql()
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
		spots = append(spots, toSpot(s))
	}

	return spots, nil
}

func (ss *SpotStore) CreateSpot(p surfing.CreateLocalizedSpotParams) (surfing.Spot, error) {
	query, args, err := ss.builder.
		Insert("spots").
		Columns("name", "latitude", "longitude", "locality", "country_code").
		Values(
			p.Name,
			p.Location.Coordinates.Latitude,
			p.Location.Coordinates.Longitude,
			psqlutil.String(p.Location.Locality),
			psqlutil.String(p.Location.CountryCode),
		).
		Suffix("RETURNING id, name, latitude, longitude, locality, country_code, created_at").
		ToSql()
	if err != nil {
		return surfing.Spot{}, fmt.Errorf("failed to build query: %w", err)
	}

	var s spot
	if err := ss.db.QueryRowx(query, args...).StructScan(&s); err != nil {
		return surfing.Spot{}, fmt.Errorf("failed to execute query: %w", err)
	}

	return toSpot(s), nil
}

func (ss *SpotStore) UpdateSpot(p surfing.UpdateLocalizedSpotParams) (surfing.Spot, error) {
	values := make(map[string]interface{})
	if p.Name != nil {
		values["name"] = *p.Name
	}
	if p.Location != nil {
		values["latitude"] = p.Location.Coordinates.Latitude
		values["longitude"] = p.Location.Coordinates.Longitude
		values["locality"] = psqlutil.String(p.Location.Locality)
		values["country_code"] = psqlutil.String(p.Location.CountryCode)
	}
	if len(values) == 0 {
		return surfing.Spot{}, surfing.ErrNothingToUpdate
	}

	query, args, err := ss.builder.
		Update("spots").
		SetMap(values).
		Where(sq.Eq{"id": p.ID}).
		Suffix("RETURNING id, name, latitude, longitude, locality, country_code, created_at").
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

	return toSpot(s), nil
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
