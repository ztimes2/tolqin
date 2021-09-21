package psql

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/psqlutil"
	"github.com/ztimes2/tolqin/app/api/internal/surf"
)

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

func (ss *SpotStore) Spot(id string) (surf.Spot, error) {
	query, args, err := ss.builder.
		Select("id", "name", "latitude", "longitude", "locality", "country_code", "created_at").
		From("spots").
		Where(sq.Eq{psqlutil.CastAsVarchar("id"): id}).
		ToSql()
	if err != nil {
		return surf.Spot{}, fmt.Errorf("failed to build query: %w", err)
	}

	var s spot
	if err := ss.db.QueryRowx(query, args...).StructScan(&s); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return surf.Spot{}, surf.ErrSpotNotFound
		}
		return surf.Spot{}, fmt.Errorf("failed to execute query: %w", err)
	}

	return toSpot(s), nil
}

func (ss *SpotStore) Spots(p surf.SpotsParams) ([]surf.Spot, error) {
	builder := buildSpotsSQL(ss.builder, p)

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := ss.db.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	var spots []surf.Spot
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

func buildSpotsSQL(b sq.StatementBuilderType, p surf.SpotsParams) sq.SelectBuilder {
	builder := b.
		Select("id", "name", "latitude", "longitude", "locality", "country_code", "created_at").
		From("spots").
		Limit(uint64(p.Limit)).
		Offset(uint64(p.Offset))

	if p.CountryCode != "" {
		builder = builder.Where(sq.Eq{"country_code": p.CountryCode})
	}

	if p.SearchQuery.Query != "" {
		or := sq.Or{
			sq.ILike{"name": psqlutil.Wildcard(p.SearchQuery.Query)},
			sq.ILike{"locality": psqlutil.Wildcard(p.SearchQuery.Query)},
		}
		if p.SearchQuery.WithSpotID {
			or = append(or, sq.ILike{psqlutil.CastAsVarchar("id"): psqlutil.Wildcard(p.SearchQuery.Query)})
		}
		builder = builder.Where(or)
	}

	if p.Bounds != nil {
		builder = builder.Where(sq.And{
			psqlutil.Between("latitude", p.Bounds.SouthWest.Latitude, p.Bounds.NorthEast.Latitude),
			psqlutil.Between("longitude", p.Bounds.SouthWest.Longitude, p.Bounds.NorthEast.Longitude),
		})
	}

	return builder
}

func (ss *SpotStore) CreateSpot(p surf.CreateSpotParams) (surf.Spot, error) {
	query, args, err := ss.builder.
		Insert("spots").
		Columns("name", "latitude", "longitude", "locality", "country_code").
		Values(
			p.Name,
			p.Location.Coordinates.Latitude,
			p.Location.Coordinates.Longitude,
			p.Location.Locality,
			p.Location.CountryCode,
		).
		Suffix("RETURNING id, name, latitude, longitude, locality, country_code, created_at").
		ToSql()
	if err != nil {
		return surf.Spot{}, fmt.Errorf("failed to build query: %w", err)
	}

	var s spot
	if err := ss.db.QueryRowx(query, args...).StructScan(&s); err != nil {
		return surf.Spot{}, fmt.Errorf("failed to execute query: %w", err)
	}

	return toSpot(s), nil
}

func (ss *SpotStore) UpdateSpot(p surf.UpdateSpotParams) (surf.Spot, error) {
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
	if p.Locality != nil {
		values["locality"] = *p.Locality
	}
	if p.CountryCode != nil {
		values["country_code"] = *p.CountryCode
	}
	if len(values) == 0 {
		return surf.Spot{}, surf.ErrEmptySpotUpdateEntry
	}

	query, args, err := ss.builder.
		Update("spots").
		SetMap(values).
		Where(sq.Eq{psqlutil.CastAsVarchar("id"): p.ID}).
		Suffix("RETURNING id, name, latitude, longitude, locality, country_code, created_at").
		ToSql()
	if err != nil {
		return surf.Spot{}, fmt.Errorf("failed to build query: %w", err)
	}

	var s spot
	if err := ss.db.QueryRowx(query, args...).StructScan(&s); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return surf.Spot{}, surf.ErrSpotNotFound
		}
		return surf.Spot{}, fmt.Errorf("failed to execute query: %w", err)
	}

	return toSpot(s), nil
}

func (ss *SpotStore) DeleteSpot(id string) error {
	query, args, err := ss.builder.
		Delete("spots").
		Where(sq.Eq{psqlutil.CastAsVarchar("id"): id}).
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
		return surf.ErrSpotNotFound
	}

	return nil
}

type spot struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Latitude    float64   `db:"latitude"`
	Longitude   float64   `db:"longitude"`
	Locality    string    `db:"locality"`
	CountryCode string    `db:"country_code"`
	CreatedAt   time.Time `db:"created_at"`
}

func toSpot(s spot) surf.Spot {
	return surf.Spot{
		ID:        s.ID,
		Name:      s.Name,
		CreatedAt: s.CreatedAt,
		Location: geo.Location{
			Locality:    s.Locality,
			CountryCode: s.CountryCode,
			Coordinates: geo.Coordinates{
				Latitude:  s.Latitude,
				Longitude: s.Longitude,
			},
		},
	}
}
