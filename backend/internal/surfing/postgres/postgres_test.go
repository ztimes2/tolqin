package postgres

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/ztimes2/tolqin/backend/internal/surfing"
)

func wrapDB(db *sql.DB) *sqlx.DB {
	return sqlx.NewDb(db, "postgres")
}

func TestSpotStore_Spot(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		dbFn          func() (*sql.DB, error)
		expectedSpot  surfing.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during unexpected db error",
			id:   "1",
			dbFn: func() (*sql.DB, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, created_at " +
							"FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnError(errors.New("unexpected error"))

				return db, nil
			},
			expectedSpot:  surfing.Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return error for unexisting spot",
			id:   "1",
			dbFn: func() (*sql.DB, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, created_at " +
							"FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnError(sql.ErrNoRows)

				return db, nil
			},
			expectedSpot: surfing.Spot{},
			expectedErrFn: func(tt assert.TestingT, e error, i ...interface{}) bool {
				return assert.ErrorIs(tt, e, surfing.ErrSpotNotFound, i...)
			},
		},
		{
			name: "return spot without error",
			id:   "1",
			dbFn: func() (*sql.DB, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, created_at " +
							"FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnRows(
						sqlmock.NewRows([]string{
							"id", "name", "latitude", "longitude", "created_at",
						}).
							AddRow("1", "Test", 1.23, 3.21, time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)),
					)

				return db, nil
			},
			expectedSpot: surfing.Spot{
				ID:        "1",
				Name:      "Test",
				Latitude:  1.23,
				Longitude: 3.21,
				CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, err := test.dbFn()
			if err != nil {
				assert.Fail(t, err.Error())
			}
			defer db.Close()

			store := NewSpotStore(wrapDB(db))

			spot, err := store.Spot(test.id)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpot, spot)
		})
	}
}
