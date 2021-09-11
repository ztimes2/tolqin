package psql

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/ztimes2/tolqin/internal/geo"
	"github.com/ztimes2/tolqin/internal/pkg/psqlutil"
	"github.com/ztimes2/tolqin/internal/service/surfing"
)

type spot struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Latitude    float64   `db:"latitude"`
	Longitude   float64   `db:"longitude"`
	Locality    string    `db:"locality"`
	CountryCode string    `db:"country_code"`
	CreatedAt   time.Time `db:"created_at"`
}

func toSpot(s spot) surfing.Spot {
	return surfing.Spot{
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
		Where(sq.Eq{psqlutil.CastAsVarchar("id"): id}).
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

	if p.Query != "" {
		builder = builder.Where(sq.Or{
			sq.ILike{"name": psqlutil.Wildcard(p.Query)},
			sq.ILike{"locality": psqlutil.Wildcard(p.Query)},
		})
	}

	if p.Bounds != nil {
		builder = builder.Where(sq.And{
			psqlutil.Between("latitude", p.Bounds.SouthWest.Latitude, p.Bounds.NorthEast.Latitude),
			psqlutil.Between("longitude", p.Bounds.SouthWest.Longitude, p.Bounds.NorthEast.Longitude),
		})
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
