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

	"github.com/ztimes2/tolqin/internal/importing"
	"github.com/ztimes2/tolqin/internal/pconv"
	"github.com/ztimes2/tolqin/internal/surfing"
	"github.com/ztimes2/tolqin/internal/testutil"
)

func wrapDB(db *sql.DB) *sqlx.DB {
	return sqlx.NewDb(db, driverName)
}

func TestSpotStore_Spot(t *testing.T) {
	tests := []struct {
		name          string
		mockFn        func(sqlmock.Sqlmock)
		id            string
		expectedSpot  surfing.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during unexpected db error",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, created_at " +
							"FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnError(errors.New("unexpected error"))
			},
			id:            "1",
			expectedSpot:  surfing.Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return error for unexisting resource",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, created_at " +
							"FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnError(sql.ErrNoRows)
			},
			id:            "1",
			expectedSpot:  surfing.Spot{},
			expectedErrFn: testutil.IsError(surfing.ErrNotFound),
		},
		{
			name: "return spot without error",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
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
			db, mock, err := sqlmock.New()
			if err != nil {
				assert.Fail(t, err.Error())
			}
			defer db.Close()

			test.mockFn(mock)

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
		mockFn        func(sqlmock.Sqlmock)
		expectedSpots []surfing.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during unexpected db error",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, created_at " +
							"FROM spots",
					)).
					WillReturnError(errors.New("unexpected error"))
			},
			expectedSpots: nil,
			expectedErrFn: assert.Error,
		},
		{
			name: "return error during row scanning",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
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
			},
			expectedSpots: nil,
			expectedErrFn: assert.Error,
		},
		{
			name: "return 0 spots without error",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
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
			},
			expectedSpots: nil,
			expectedErrFn: assert.NoError,
		},
		{
			name: "return multiple spots without error",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
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
			db, mock, err := sqlmock.New()
			if err != nil {
				assert.Fail(t, err.Error())
			}
			defer db.Close()

			test.mockFn(mock)

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
		mockFn        func(sqlmock.Sqlmock)
		params        surfing.CreateSpotParams
		expectedSpot  surfing.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during unexpected db error",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude) "+
							"VALUES ($1,$2,$3) "+
							"RETURNING id, created_at",
					)).
					WithArgs("Test", 1.23, 3.21).
					WillReturnError(errors.New("unexpected error"))
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
			mockFn: func(m sqlmock.Sqlmock) {
				m.
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
			db, mock, err := sqlmock.New()
			if err != nil {
				assert.Fail(t, err.Error())
			}
			defer db.Close()

			test.mockFn(mock)

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
		mockFn        func(sqlmock.Sqlmock)
		params        surfing.UpdateSpotParams
		expectedSpot  surfing.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during unexpected db error",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"UPDATE spots "+
							"SET latitude = $1, longitude = $2, name = $3 "+
							"WHERE id = $4 "+
							"RETURNING id, name, latitude, longitude, created_at",
					)).
					WithArgs(2.34, 4.32, "Test updated", "1").
					WillReturnError(errors.New("unexpected error"))
			},
			params: surfing.UpdateSpotParams{
				ID:        "1",
				Name:      pconv.String("Test updated"),
				Latitude:  pconv.Float64(2.34),
				Longitude: pconv.Float64(4.32),
			},
			expectedSpot:  surfing.Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return error for unexisting resource",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"UPDATE spots "+
							"SET latitude = $1, longitude = $2, name = $3 "+
							"WHERE id = $4 "+
							"RETURNING id, name, latitude, longitude, created_at",
					)).
					WithArgs(2.34, 4.32, "Test updated", "1").
					WillReturnError(sql.ErrNoRows)
			},
			params: surfing.UpdateSpotParams{
				ID:        "1",
				Name:      pconv.String("Test updated"),
				Latitude:  pconv.Float64(2.34),
				Longitude: pconv.Float64(4.32),
			},
			expectedSpot:  surfing.Spot{},
			expectedErrFn: testutil.IsError(surfing.ErrNotFound),
		},
		{
			name:   "return error when nothing to update",
			mockFn: func(m sqlmock.Sqlmock) {},
			params: surfing.UpdateSpotParams{
				ID: "1",
			},
			expectedSpot:  surfing.Spot{},
			expectedErrFn: testutil.IsError(surfing.ErrNothingToUpdate),
		},
		{
			name: "return spot without error for full update",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
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
			},
			params: surfing.UpdateSpotParams{
				ID:        "1",
				Name:      pconv.String("Test updated"),
				Latitude:  pconv.Float64(2.34),
				Longitude: pconv.Float64(4.32),
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
			mockFn: func(m sqlmock.Sqlmock) {
				m.
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
			},
			params: surfing.UpdateSpotParams{
				ID:   "1",
				Name: pconv.String("Test updated"),
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
			db, mock, err := sqlmock.New()
			if err != nil {
				assert.Fail(t, err.Error())
			}
			defer db.Close()

			test.mockFn(mock)

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
		mockFn        func(sqlmock.Sqlmock)
		id            string
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during unexpected db error",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectExec(regexp.QuoteMeta(
						"DELETE FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnError(errors.New("unexpected error"))
			},
			id:            "1",
			expectedErrFn: assert.Error,
		},
		{
			name: "return error when reading affected rows",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectExec(regexp.QuoteMeta(
						"DELETE FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnResult(sqlmock.NewErrorResult(
						errors.New("unexpected error"),
					))
			},
			id:            "1",
			expectedErrFn: assert.Error,
		},
		{
			name: "return error for unexisting resource",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectExec(regexp.QuoteMeta(
						"DELETE FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			id:            "1",
			expectedErrFn: assert.Error,
		},
		{
			name: "return no error",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectExec(regexp.QuoteMeta(
						"DELETE FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnResult(sqlmock.NewResult(
						0, // Postgres driver does not support LastInsertId
						1,
					))
			},
			id:            "1",
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				assert.Fail(t, err.Error())
			}
			defer db.Close()

			test.mockFn(mock)

			store := NewSpotStore(wrapDB(db))
			err = store.DeleteSpot(test.id)
			test.expectedErrFn(t, err)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSpotImporter_ImportSpots(t *testing.T) {
	tests := []struct {
		name          string
		batchSize     int
		mockFn        func(sqlmock.Sqlmock)
		entries       []importing.SpotEntry
		expectedSpots []surfing.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name:          "return error when nothing to import",
			batchSize:     2,
			mockFn:        func(m sqlmock.Sqlmock) {},
			entries:       []importing.SpotEntry{},
			expectedSpots: nil,
			expectedErrFn: testutil.IsError(importing.ErrNothingToImport),
		},
		{
			name:      "return error during tx init error",
			batchSize: 2,
			mockFn: func(m sqlmock.Sqlmock) {
				m.ExpectBegin().
					WillReturnError(errors.New("something went wrong"))
			},
			entries: []importing.SpotEntry{
				{
					Name:      "Test 1",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 2",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 3",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 4",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 5",
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedSpots: nil,
			expectedErrFn: assert.Error,
		},
		{
			name:      "return error during unexpected db failure",
			batchSize: 2,
			mockFn: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()

				m.
					ExpectQuery(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude) "+
							"VALUES ($1,$2,$3),($4,$5,$6) "+
							"RETURNING id, name, latitude, longitude, created_at",
					)).
					WithArgs(
						"Test 1", 1.23, 3.21,
						"Test 2", 1.23, 3.21,
					).
					WillReturnError(errors.New("something went wrong"))

				m.ExpectRollback()
			},
			entries: []importing.SpotEntry{
				{
					Name:      "Test 1",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 2",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 3",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 4",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 5",
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedSpots: nil,
			expectedErrFn: assert.Error,
		},
		{
			name:      "return error during row scanning",
			batchSize: 2,
			mockFn: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()

				m.
					ExpectQuery(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude) "+
							"VALUES ($1,$2,$3),($4,$5,$6) "+
							"RETURNING id, name, latitude, longitude, created_at",
					)).
					WithArgs(
						"Test 1", 1.23, 3.21,
						"Test 2", 1.23, 3.21,
					).
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "created_at",
						}).
						AddRow("1", "Test 1", 1.23, 3.21, false).
						AddRow("2", "Test 2", 1.23, 3.21, false),
					)

				m.ExpectRollback()
			},
			entries: []importing.SpotEntry{
				{
					Name:      "Test 1",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 2",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 3",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 4",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 5",
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedSpots: nil,
			expectedErrFn: assert.Error,
		},
		{
			name:      "return spots without error",
			batchSize: 2,
			mockFn: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()

				m.
					ExpectQuery(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude) "+
							"VALUES ($1,$2,$3),($4,$5,$6) "+
							"RETURNING id, name, latitude, longitude, created_at",
					)).
					WithArgs(
						"Test 1", 1.23, 3.21,
						"Test 2", 1.23, 3.21,
					).
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "created_at",
						}).
						AddRow("1", "Test 1", 1.23, 3.21, time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)).
						AddRow("2", "Test 2", 1.23, 3.21, time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)),
					)

				m.
					ExpectQuery(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude) "+
							"VALUES ($1,$2,$3),($4,$5,$6) "+
							"RETURNING id, name, latitude, longitude, created_at",
					)).
					WithArgs(
						"Test 3", 1.23, 3.21,
						"Test 4", 1.23, 3.21,
					).
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "created_at",
						}).
						AddRow("3", "Test 3", 1.23, 3.21, time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)).
						AddRow("4", "Test 4", 1.23, 3.21, time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)),
					)

				m.
					ExpectQuery(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude) "+
							"VALUES ($1,$2,$3) "+
							"RETURNING id, name, latitude, longitude, created_at",
					)).
					WithArgs("Test 5", 1.23, 3.21).
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "created_at",
						}).
						AddRow("5", "Test 5", 1.23, 3.21, time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)),
					)

				m.ExpectCommit()
			},
			entries: []importing.SpotEntry{
				{
					Name:      "Test 1",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 2",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 3",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 4",
					Latitude:  1.23,
					Longitude: 3.21,
				},
				{
					Name:      "Test 5",
					Latitude:  1.23,
					Longitude: 3.21,
				},
			},
			expectedSpots: []surfing.Spot{
				{
					ID:        "1",
					Name:      "Test 1",
					Latitude:  1.23,
					Longitude: 3.21,
					CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:        "2",
					Name:      "Test 2",
					Latitude:  1.23,
					Longitude: 3.21,
					CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:        "3",
					Name:      "Test 3",
					Latitude:  1.23,
					Longitude: 3.21,
					CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:        "4",
					Name:      "Test 4",
					Latitude:  1.23,
					Longitude: 3.21,
					CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:        "5",
					Name:      "Test 5",
					Latitude:  1.23,
					Longitude: 3.21,
					CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				assert.Fail(t, err.Error())
			}
			defer db.Close()

			test.mockFn(mock)

			importer := NewSpotImporter(wrapDB(db), test.batchSize)
			spots, err := importer.ImportSpots(test.entries)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpots, spots)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
