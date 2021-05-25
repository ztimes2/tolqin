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
	"github.com/ztimes2/tolqin/backend/internal/testutil"
	"github.com/ztimes2/tolqin/backend/internal/typeutil"
)

func wrapDB(db *sql.DB) *sqlx.DB {
	return sqlx.NewDb(db, "postgres")
}

func TestSpotStore_Spot(t *testing.T) {
	tests := []struct {
		name          string
		mockDBFn      func() (*sql.DB, sqlmock.Sqlmock, error)
		id            string
		expectedSpot  surfing.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during unexpected db error",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, created_at " +
							"FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnError(errors.New("unexpected error"))

				return db, mock, nil
			},
			id:            "1",
			expectedSpot:  surfing.Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return error for unexisting resource",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, created_at " +
							"FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnError(sql.ErrNoRows)

				return db, mock, nil
			},
			id:            "1",
			expectedSpot:  surfing.Spot{},
			expectedErrFn: testutil.IsError(surfing.ErrNotFound),
		},
		{
			name: "return spot without error",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, created_at " +
							"FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "created_at",
						}).
						AddRow("1", "Test", 1.23, 3.21, time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()

				return db, mock, nil
			},
			id: "1",
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
			db, mock, err := test.mockDBFn()
			if err != nil {
				assert.Fail(t, err.Error())
			}
			defer db.Close()

			store := NewSpotStore(wrapDB(db))

			spot, err := store.Spot(test.id)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpot, spot)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSpotStore_Spots(t *testing.T) {
	tests := []struct {
		name          string
		mockDBFn      func() (*sql.DB, sqlmock.Sqlmock, error)
		expectedSpots []surfing.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during unexpected db error",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, created_at " +
							"FROM spots",
					)).
					WillReturnError(errors.New("unexpected error"))

				return db, mock, nil
			},
			expectedSpots: nil,
			expectedErrFn: assert.Error,
		},
		{
			name: "return error during row scanning",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, created_at " +
							"FROM spots",
					)).
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "created_at",
						}).
						AddRow(1, true, "1.23", "3.21", "Not-a-time"),
					).
					RowsWillBeClosed()

				return db, mock, nil
			},
			expectedSpots: nil,
			expectedErrFn: assert.Error,
		},
		{
			name: "return 0 spots without error",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, created_at " +
							"FROM spots",
					)).
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "created_at",
						}),
					).
					RowsWillBeClosed()

				return db, mock, nil
			},
			expectedSpots: nil,
			expectedErrFn: assert.NoError,
		},
		{
			name: "return multiple spots without error",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, created_at " +
							"FROM spots",
					)).
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "created_at",
						}).
						AddRow("1", "Test", 1.23, 3.21, time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)).
						AddRow("2", "Test", 2.34, 4.32, time.Date(2021, 3, 2, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()

				return db, mock, nil
			},
			expectedSpots: []surfing.Spot{
				{
					ID:        "1",
					Name:      "Test",
					Latitude:  1.23,
					Longitude: 3.21,
					CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:        "2",
					Name:      "Test",
					Latitude:  2.34,
					Longitude: 4.32,
					CreatedAt: time.Date(2021, 3, 2, 0, 0, 0, 0, time.UTC),
				},
			},
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := test.mockDBFn()
			if err != nil {
				assert.Fail(t, err.Error())
			}
			defer db.Close()

			store := NewSpotStore(wrapDB(db))

			spots, err := store.Spots()
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpots, spots)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSpotStore_CreateSpot(t *testing.T) {
	tests := []struct {
		name          string
		mockDBFn      func() (*sql.DB, sqlmock.Sqlmock, error)
		params        surfing.CreateSpotParams
		expectedSpot  surfing.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during unexpected db error",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude) "+
							"VALUES ($1,$2,$3) "+
							"RETURNING id, created_at",
					)).
					WithArgs("Test", 1.23, 3.21).
					WillReturnError(errors.New("unexpected error"))

				return db, mock, nil
			},
			params: surfing.CreateSpotParams{
				Name:      "Test",
				Latitude:  1.23,
				Longitude: 3.21,
			},
			expectedSpot:  surfing.Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return spot without error",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude) "+
							"VALUES ($1,$2,$3) "+
							"RETURNING id, created_at",
					)).
					WithArgs("Test", 1.23, 3.21).
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "created_at",
						}).
						AddRow("1", time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()

				return db, mock, nil
			},
			params: surfing.CreateSpotParams{
				Name:      "Test",
				Latitude:  1.23,
				Longitude: 3.21,
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
			db, mock, err := test.mockDBFn()
			if err != nil {
				assert.Fail(t, err.Error())
			}
			defer db.Close()

			store := NewSpotStore(wrapDB(db))
			spot, err := store.CreateSpot(test.params)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpot, spot)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSpotStore_UpdateSpot(t *testing.T) {
	tests := []struct {
		name          string
		mockDBFn      func() (*sql.DB, sqlmock.Sqlmock, error)
		params        surfing.UpdateSpotParams
		expectedSpot  surfing.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during unexpected db error",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"UPDATE spots "+
							"SET latitude = $1, longitude = $2, name = $3 "+
							"WHERE id = $4 "+
							"RETURNING id, name, latitude, longitude, created_at",
					)).
					WithArgs(2.34, 4.32, "Test updated", "1").
					WillReturnError(errors.New("unexpected error"))

				return db, mock, nil
			},
			params: surfing.UpdateSpotParams{
				ID:        "1",
				Name:      typeutil.String("Test updated"),
				Latitude:  typeutil.Float64(2.34),
				Longitude: typeutil.Float64(4.32),
			},
			expectedSpot:  surfing.Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return error for unexisting resource",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"UPDATE spots "+
							"SET latitude = $1, longitude = $2, name = $3 "+
							"WHERE id = $4 "+
							"RETURNING id, name, latitude, longitude, created_at",
					)).
					WithArgs(2.34, 4.32, "Test updated", "1").
					WillReturnError(sql.ErrNoRows)

				return db, mock, nil
			},
			params: surfing.UpdateSpotParams{
				ID:        "1",
				Name:      typeutil.String("Test updated"),
				Latitude:  typeutil.Float64(2.34),
				Longitude: typeutil.Float64(4.32),
			},
			expectedSpot:  surfing.Spot{},
			expectedErrFn: testutil.IsError(surfing.ErrNotFound),
		},
		{
			name: "return error when nothing to update",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}
				return db, mock, nil
			},
			params: surfing.UpdateSpotParams{
				ID: "1",
			},
			expectedSpot:  surfing.Spot{},
			expectedErrFn: testutil.IsError(surfing.ErrNothingToUpdate),
		},
		{
			name: "return spot without error for full update",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"UPDATE spots "+
							"SET latitude = $1, longitude = $2, name = $3 "+
							"WHERE id = $4 "+
							"RETURNING id, name, latitude, longitude, created_at",
					)).
					WithArgs(2.34, 4.32, "Test updated", "1").
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "created_at",
						}).
						AddRow("1", "Test updated", 2.34, 4.32, time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()

				return db, mock, nil
			},
			params: surfing.UpdateSpotParams{
				ID:        "1",
				Name:      typeutil.String("Test updated"),
				Latitude:  typeutil.Float64(2.34),
				Longitude: typeutil.Float64(4.32),
			},
			expectedSpot: surfing.Spot{
				ID:        "1",
				Name:      "Test updated",
				Latitude:  2.34,
				Longitude: 4.32,
				CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedErrFn: assert.NoError,
		},
		{
			name: "return spot without error for partial update",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectQuery(regexp.QuoteMeta(
						"UPDATE spots "+
							"SET name = $1 "+
							"WHERE id = $2 "+
							"RETURNING id, name, latitude, longitude, created_at",
					)).
					WithArgs("Test updated", "1").
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "created_at",
						}).
						AddRow("1", "Test updated", 2.34, 4.32, time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()

				return db, mock, nil
			},
			params: surfing.UpdateSpotParams{
				ID:   "1",
				Name: typeutil.String("Test updated"),
			},
			expectedSpot: surfing.Spot{
				ID:        "1",
				Name:      "Test updated",
				Latitude:  2.34,
				Longitude: 4.32,
				CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := test.mockDBFn()
			if err != nil {
				assert.Fail(t, err.Error())
			}
			defer db.Close()

			store := NewSpotStore(wrapDB(db))
			spot, err := store.UpdateSpot(test.params)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpot, spot)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSpotStore_DeleteSpot(t *testing.T) {
	tests := []struct {
		name          string
		mockDBFn      func() (*sql.DB, sqlmock.Sqlmock, error)
		id            string
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during unexpected db error",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectExec(regexp.QuoteMeta(
						"DELETE FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnError(errors.New("unexpected error"))

				return db, mock, nil
			},
			id:            "1",
			expectedErrFn: assert.Error,
		},
		{
			name: "return error when reading affected rows",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectExec(regexp.QuoteMeta(
						"DELETE FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnResult(sqlmock.NewErrorResult(
						errors.New("unexpected error"),
					))

				return db, mock, nil
			},
			id:            "1",
			expectedErrFn: assert.Error,
		},
		{
			name: "return error for unexisting resource",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectExec(regexp.QuoteMeta(
						"DELETE FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnResult(sqlmock.NewResult(0, 0))

				return db, mock, nil
			},
			id:            "1",
			expectedErrFn: assert.Error,
		},
		{
			name: "return no error",
			mockDBFn: func() (*sql.DB, sqlmock.Sqlmock, error) {
				db, mock, err := sqlmock.New()
				if err != nil {
					return nil, nil, err
				}

				mock.
					ExpectExec(regexp.QuoteMeta(
						"DELETE FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnResult(sqlmock.NewResult(
						0, // Postgres driver does not support LastInsertId
						1,
					))

				return db, mock, nil
			},
			id:            "1",
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := test.mockDBFn()
			if err != nil {
				assert.Fail(t, err.Error())
			}
			defer db.Close()

			store := NewSpotStore(wrapDB(db))
			err = store.DeleteSpot(test.id)
			test.expectedErrFn(t, err)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
