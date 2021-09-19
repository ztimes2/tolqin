package psql

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/ztimes2/tolqin/app/api/internal/geo"
	"github.com/ztimes2/tolqin/app/api/internal/importing"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/psqlutil"
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
			expectedErrFn: assert.Error,
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
					Name: "Spot 1",
					Location: geo.Location{
						Locality:    "Locality 1",
						CountryCode: "Country code 1",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 2",
					Location: geo.Location{
						Locality:    "Locality 2",
						CountryCode: "",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 3",
					Location: geo.Location{
						Locality:    "",
						CountryCode: "Country code 3",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 4",
					Location: geo.Location{
						Locality:    "",
						CountryCode: "",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 5",
					Location: geo.Location{
						Locality:    "Locality 5",
						CountryCode: "Country code 5",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
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
						"INSERT INTO spots (name,latitude,longitude,locality,country_code) "+
							"VALUES ($1,$2,$3,$4,$5),($6,$7,$8,$9,$10)",
					)).
					WithArgs(
						"Spot 1", 1.23, 3.21, "Locality 1", "Country code 1",
						"Spot 2", 1.23, 3.21, "Locality 2", "",
					).
					WillReturnError(errors.New("something went wrong"))

				m.ExpectRollback()
			},
			entries: []importing.SpotEntry{
				{
					Name: "Spot 1",
					Location: geo.Location{
						Locality:    "Locality 1",
						CountryCode: "Country code 1",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 2",
					Location: geo.Location{
						Locality:    "Locality 2",
						CountryCode: "",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 3",
					Location: geo.Location{
						Locality:    "",
						CountryCode: "Country code 3",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 4",
					Location: geo.Location{
						Locality:    "",
						CountryCode: "",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 5",
					Location: geo.Location{
						Locality:    "Locality 5",
						CountryCode: "Country code 5",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
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
						"INSERT INTO spots (name,latitude,longitude,locality,country_code) "+
							"VALUES ($1,$2,$3,$4,$5),($6,$7,$8,$9,$10)",
					)).
					WithArgs(
						"Spot 1", 1.23, 3.21, "Locality 1", "Country code 1",
						"Spot 2", 1.23, 3.21, "Locality 2", "",
					).
					WillReturnResult(sqlmock.NewErrorResult(
						errors.New("something went wrong"),
					))

				m.ExpectRollback()
			},
			entries: []importing.SpotEntry{
				{
					Name: "Spot 1",
					Location: geo.Location{
						Locality:    "Locality 1",
						CountryCode: "Country code 1",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 2",
					Location: geo.Location{
						Locality:    "Locality 2",
						CountryCode: "",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 3",
					Location: geo.Location{
						Locality:    "",
						CountryCode: "Country code 3",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 4",
					Location: geo.Location{
						Locality:    "",
						CountryCode: "",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 5",
					Location: geo.Location{
						Locality:    "Locality 5",
						CountryCode: "Country code 5",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
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
						"INSERT INTO spots (name,latitude,longitude,locality,country_code) "+
							"VALUES ($1,$2,$3,$4,$5),($6,$7,$8,$9,$10)",
					)).
					WithArgs(
						"Spot 1", 1.23, 3.21, "Locality 1", "Country code 1",
						"Spot 2", 1.23, 3.21, "Locality 2", "",
					).
					WillReturnResult(sqlmock.NewResult(0, 0))

				m.ExpectRollback()
			},
			entries: []importing.SpotEntry{
				{
					Name: "Spot 1",
					Location: geo.Location{
						Locality:    "Locality 1",
						CountryCode: "Country code 1",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 2",
					Location: geo.Location{
						Locality:    "Locality 2",
						CountryCode: "",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 3",
					Location: geo.Location{
						Locality:    "",
						CountryCode: "Country code 3",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 4",
					Location: geo.Location{
						Locality:    "",
						CountryCode: "",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 5",
					Location: geo.Location{
						Locality:    "Locality 5",
						CountryCode: "Country code 5",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
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
						"INSERT INTO spots (name,latitude,longitude,locality,country_code) "+
							"VALUES ($1,$2,$3,$4,$5),($6,$7,$8,$9,$10)",
					)).
					WithArgs(
						"Spot 1", 1.23, 3.21, "Locality 1", "Country code 1",
						"Spot 2", 1.23, 3.21, "Locality 2", "",
					).
					WillReturnResult(sqlmock.NewResult(0, 2))

				m.
					ExpectExec(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude,locality,country_code) "+
							"VALUES ($1,$2,$3,$4,$5),($6,$7,$8,$9,$10)",
					)).
					WithArgs(
						"Spot 3", 1.23, 3.21, "", "Country code 3",
						"Spot 4", 1.23, 3.21, "", "",
					).
					WillReturnResult(sqlmock.NewResult(0, 2))

				m.
					ExpectExec(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude,locality,country_code) "+
							"VALUES ($1,$2,$3,$4,$5)",
					)).
					WithArgs(
						"Spot 5", 1.23, 3.21, "Locality 5", "Country code 5",
					).
					WillReturnResult(sqlmock.NewResult(0, 1))

				m.ExpectCommit()
			},
			entries: []importing.SpotEntry{
				{
					Name: "Spot 1",
					Location: geo.Location{
						Locality:    "Locality 1",
						CountryCode: "Country code 1",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 2",
					Location: geo.Location{
						Locality:    "Locality 2",
						CountryCode: "",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 3",
					Location: geo.Location{
						Locality:    "",
						CountryCode: "Country code 3",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 4",
					Location: geo.Location{
						Locality:    "",
						CountryCode: "",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					Name: "Spot 5",
					Location: geo.Location{
						Locality:    "Locality 5",
						CountryCode: "Country code 5",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
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

			importer := NewSpotImporter(sqlx.NewDb(db, psqlutil.DriverNameSQLMock), test.batchSize)
			count, err := importer.ImportSpots(test.entries)
			assert.Equal(t, test.expectedCount, count)
			test.expectedErrFn(t, err)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
