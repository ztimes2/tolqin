package psql

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/ztimes2/tolqin/internal/importing"
	"github.com/ztimes2/tolqin/internal/psqlutil"
	"github.com/ztimes2/tolqin/internal/testutil"
)

func TestSpotImporter_ImportSpots(t *testing.T) {
	tests := []struct {
		name          string
		batchSize     int
		mockFn        func(sqlmock.Sqlmock)
		entries       []importing.SpotEntry
		expectedCount int
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name:          "return error when nothing to import",
			batchSize:     2,
			mockFn:        func(m sqlmock.Sqlmock) {},
			entries:       []importing.SpotEntry{},
			expectedCount: 0,
			expectedErrFn: testutil.IsError(importing.ErrNothingToImport),
		},
		{
			name:      "return error during tx init",
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
			expectedCount: 0,
			expectedErrFn: assert.Error,
		},
		{
			name:      "return error during query execution",
			batchSize: 2,
			mockFn: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()

				m.
					ExpectExec(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude) "+
							"VALUES ($1,$2,$3),($4,$5,$6)",
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
			expectedCount: 0,
			expectedErrFn: assert.Error,
		},
		{
			name:      "return error when reading affected rows",
			batchSize: 2,
			mockFn: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()

				m.
					ExpectExec(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude) "+
							"VALUES ($1,$2,$3),($4,$5,$6)",
					)).
					WithArgs(
						"Test 1", 1.23, 3.21,
						"Test 2", 1.23, 3.21,
					).
					WillReturnResult(sqlmock.NewErrorResult(
						errors.New("something went wrong"),
					))

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
			expectedCount: 0,
			expectedErrFn: assert.Error,
		},
		{
			name:      "return error when no rows affected",
			batchSize: 2,
			mockFn: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()

				m.
					ExpectExec(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude) "+
							"VALUES ($1,$2,$3),($4,$5,$6)",
					)).
					WithArgs(
						"Test 1", 1.23, 3.21,
						"Test 2", 1.23, 3.21,
					).
					WillReturnResult(sqlmock.NewResult(0, 0))

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
			expectedCount: 0,
			expectedErrFn: assert.Error,
		},
		{
			name:      "return spots without error",
			batchSize: 2,
			mockFn: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()

				m.
					ExpectExec(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude) "+
							"VALUES ($1,$2,$3),($4,$5,$6)",
					)).
					WithArgs(
						"Test 1", 1.23, 3.21,
						"Test 2", 1.23, 3.21,
					).
					WillReturnResult(sqlmock.NewResult(0, 2))

				m.
					ExpectExec(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude) "+
							"VALUES ($1,$2,$3),($4,$5,$6)",
					)).
					WithArgs(
						"Test 3", 1.23, 3.21,
						"Test 4", 1.23, 3.21,
					).
					WillReturnResult(sqlmock.NewResult(0, 2))

				m.
					ExpectExec(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude) "+
							"VALUES ($1,$2,$3)",
					)).
					WithArgs("Test 5", 1.23, 3.21).
					WillReturnResult(sqlmock.NewResult(0, 1))

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
			expectedCount: 5,
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

			importer := NewSpotImporter(psqlutil.WrapDB(db), test.batchSize)
			count, err := importer.ImportSpots(test.entries)
			assert.Equal(t, test.expectedCount, count)
			test.expectedErrFn(t, err)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
