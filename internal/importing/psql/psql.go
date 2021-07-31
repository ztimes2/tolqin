package psql

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/ztimes2/tolqin/internal/batch"
	"github.com/ztimes2/tolqin/internal/importing"
)

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

func (si *SpotImporter) ImportSpots(entries []importing.SpotEntry) (int, error) {
	if len(entries) == 0 {
		return 0, importing.ErrNothingToImport
	}

	tx, err := si.db.Beginx()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	var count int

	coord := batch.New(len(entries), si.batchSize)
	for coord.HasNext() {
		b := coord.Batch()

		n, err := si.importSpots(tx, entries[b.I:b.J+1])
		if err != nil {
			_ = tx.Rollback()
			return 0, fmt.Errorf("failed to import spots: %w", err)
		}

		count += n
	}

	_ = tx.Commit()
	return count, nil
}

func (si *SpotImporter) importSpots(tx *sqlx.Tx, entries []importing.SpotEntry) (int, error) {
	builder := si.builder.
		Insert("spots").
		Columns("name", "latitude", "longitude")

	for _, e := range entries {
		builder = builder.Values(e.Name, e.Latitude, e.Longitude)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	res, err := tx.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to read affected rows: %w", err)
	}

	if count == 0 {
		return 0, fmt.Errorf("no rows affected")
	}

	return int(count), nil
}
