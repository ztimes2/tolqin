package psql

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/ztimes2/tolqin/internal/geo"
	"github.com/ztimes2/tolqin/internal/psqlutil"
	"github.com/ztimes2/tolqin/internal/surfing"
	"github.com/ztimes2/tolqin/internal/testutil"
)

func TestSpotStore_Spot(t *testing.T) {
	tests := []struct {
		name          string
		mockFn        func(sqlmock.Sqlmock)
		id            string
		expectedSpot  surfing.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during query execution",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, locality, country_code, created_at " +
							"FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnError(errors.New("something went wrong"))
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
						"SELECT id, name, latitude, longitude, locality, country_code, created_at " +
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
						"SELECT id, name, latitude, longitude, locality, country_code, created_at " +
							"FROM spots WHERE id = $1",
					)).
					WithArgs("1").
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "locality", "country_code", "created_at",
						}).
						AddRow("1", "Spot 1", 1.23, 3.21, "Locality 1", "Country code 1", time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()
			},
			id: "1",
			expectedSpot: surfing.Spot{
				ID:        "1",
				Name:      "Spot 1",
				CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
				Location: geo.Location{
					Locality:    "Locality 1",
					CountryCode: "Country code 1",
					Coordinates: geo.Coordinates{
						Latitude:  1.23,
						Longitude: 3.21,
					},
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

			store := NewSpotStore(psqlutil.WrapDB(db))

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
		params        surfing.SpotsParams
		mockFn        func(sqlmock.Sqlmock)
		expectedSpots []surfing.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during query execution",
			params: surfing.SpotsParams{
				Limit:  10,
				Offset: 0,
			},
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, locality, country_code, created_at " +
							"FROM spots LIMIT 10 OFFSET 0",
					)).
					WillReturnError(errors.New("unexpected error"))
			},
			expectedSpots: nil,
			expectedErrFn: assert.Error,
		},
		{
			name: "return error during row scanning",
			params: surfing.SpotsParams{
				Limit:  10,
				Offset: 0,
			},
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, locality, country_code, created_at " +
							"FROM spots LIMIT 10 OFFSET 0",
					)).
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "locality", "country_code", "created_at",
						}).
						AddRow(1, true, "1.23", "3.21", "Locality 1", "Country code 1", "Not-a-time"),
					).
					RowsWillBeClosed()
			},
			expectedSpots: nil,
			expectedErrFn: assert.Error,
		},
		{
			name: "return 0 spots without error",
			params: surfing.SpotsParams{
				Limit:  10,
				Offset: 0,
			},
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, locality, country_code, created_at " +
							"FROM spots LIMIT 10 OFFSET 0",
					)).
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "locality", "country_code", "created_at",
						}),
					).
					RowsWillBeClosed()
			},
			expectedSpots: nil,
			expectedErrFn: assert.NoError,
		},
		{
			name: "return spots without error",
			params: surfing.SpotsParams{
				Limit:  10,
				Offset: 0,
			},
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, locality, country_code, created_at " +
							"FROM spots LIMIT 10 OFFSET 0",
					)).
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "locality", "country_code", "created_at",
						}).
						AddRow("1", "Spot 1", 1.23, 3.21, "Locality 1", "Country code 1", time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)).
						AddRow("2", "Spot 2", 2.34, 4.32, "Locality 2", "Country code 2", time.Date(2021, 3, 2, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()
			},
			expectedSpots: []surfing.Spot{
				{
					ID:        "1",
					Name:      "Spot 1",
					CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
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
					ID:        "2",
					Name:      "Spot 2",
					CreatedAt: time.Date(2021, 3, 2, 0, 0, 0, 0, time.UTC),
					Location: geo.Location{
						Locality:    "Locality 2",
						CountryCode: "Country code 2",
						Coordinates: geo.Coordinates{
							Latitude:  2.34,
							Longitude: 4.32,
						},
					},
				},
			},
			expectedErrFn: assert.NoError,
		},
		{
			name: "return spots by country without error",
			params: surfing.SpotsParams{
				Limit:       10,
				Offset:      0,
				CountryCode: "kz",
			},
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, locality, country_code, created_at " +
							"FROM spots WHERE country_code = $1 LIMIT 10 OFFSET 0",
					)).
					WithArgs("kz").
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "locality", "country_code", "created_at",
						}).
						AddRow("1", "Spot 1", 1.23, 3.21, "Locality 1", "kz", time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)).
						AddRow("2", "Spot 2", 2.34, 4.32, "Locality 2", "kz", time.Date(2021, 3, 2, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()
			},
			expectedSpots: []surfing.Spot{
				{
					ID:        "1",
					Name:      "Spot 1",
					CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
					Location: geo.Location{
						Locality:    "Locality 1",
						CountryCode: "kz",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					ID:        "2",
					Name:      "Spot 2",
					CreatedAt: time.Date(2021, 3, 2, 0, 0, 0, 0, time.UTC),
					Location: geo.Location{
						Locality:    "Locality 2",
						CountryCode: "kz",
						Coordinates: geo.Coordinates{
							Latitude:  2.34,
							Longitude: 4.32,
						},
					},
				},
			},
			expectedErrFn: assert.NoError,
		},
		{
			name: "return spots by query without error",
			params: surfing.SpotsParams{
				Limit:  10,
				Offset: 0,
				Query:  "query",
			},
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, locality, country_code, created_at "+
							"FROM spots WHERE (name ILIKE $1 OR locality ILIKE $2) LIMIT 10 OFFSET 0",
					)).
					WithArgs("%query%", "%query%").
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "locality", "country_code", "created_at",
						}).
						AddRow("1", "Spot 1", 1.23, 3.21, "Locality 1", "kz", time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)).
						AddRow("2", "Spot 2", 2.34, 4.32, "Locality 2", "kz", time.Date(2021, 3, 2, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()
			},
			expectedSpots: []surfing.Spot{
				{
					ID:        "1",
					Name:      "Spot 1",
					CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
					Location: geo.Location{
						Locality:    "Locality 1",
						CountryCode: "kz",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					ID:        "2",
					Name:      "Spot 2",
					CreatedAt: time.Date(2021, 3, 2, 0, 0, 0, 0, time.UTC),
					Location: geo.Location{
						Locality:    "Locality 2",
						CountryCode: "kz",
						Coordinates: geo.Coordinates{
							Latitude:  2.34,
							Longitude: 4.32,
						},
					},
				},
			},
			expectedErrFn: assert.NoError,
		},
		{
			name: "return spots by bounds without error",
			params: surfing.SpotsParams{
				Limit:  10,
				Offset: 0,
				Bounds: &geo.Bounds{
					NorthEast: geo.Coordinates{
						Latitude:  90,
						Longitude: 180,
					},
					SouthWest: geo.Coordinates{
						Latitude:  -90,
						Longitude: -180,
					},
				},
			},
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, locality, country_code, created_at "+
							"FROM spots WHERE (latitude BETWEEN $1 AND $2 AND longitude BETWEEN $3 AND $4) "+
							"LIMIT 10 OFFSET 0",
					)).
					WithArgs(-90.0, 90.0, -180.0, 180.0).
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "locality", "country_code", "created_at",
						}).
						AddRow("1", "Spot 1", 1.23, 3.21, "Locality 1", "kz", time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)).
						AddRow("2", "Spot 2", 2.34, 4.32, "Locality 2", "kz", time.Date(2021, 3, 2, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()
			},
			expectedSpots: []surfing.Spot{
				{
					ID:        "1",
					Name:      "Spot 1",
					CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
					Location: geo.Location{
						Locality:    "Locality 1",
						CountryCode: "kz",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					ID:        "2",
					Name:      "Spot 2",
					CreatedAt: time.Date(2021, 3, 2, 0, 0, 0, 0, time.UTC),
					Location: geo.Location{
						Locality:    "Locality 2",
						CountryCode: "kz",
						Coordinates: geo.Coordinates{
							Latitude:  2.34,
							Longitude: 4.32,
						},
					},
				},
			},
			expectedErrFn: assert.NoError,
		},
		{
			name: "return spots by country code and query without error",
			params: surfing.SpotsParams{
				Limit:       10,
				Offset:      0,
				CountryCode: "kz",
				Query:       "query",
			},
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, locality, country_code, created_at "+
							"FROM spots WHERE country_code = $1 AND (name ILIKE $2 OR locality ILIKE $3) LIMIT 10 OFFSET 0",
					)).
					WithArgs("kz", "%query%", "%query%").
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "locality", "country_code", "created_at",
						}).
						AddRow("1", "Spot 1", 1.23, 3.21, "Locality 1", "kz", time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)).
						AddRow("2", "Spot 2", 2.34, 4.32, "Locality 2", "kz", time.Date(2021, 3, 2, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()
			},
			expectedSpots: []surfing.Spot{
				{
					ID:        "1",
					Name:      "Spot 1",
					CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
					Location: geo.Location{
						Locality:    "Locality 1",
						CountryCode: "kz",
						Coordinates: geo.Coordinates{
							Latitude:  1.23,
							Longitude: 3.21,
						},
					},
				},
				{
					ID:        "2",
					Name:      "Spot 2",
					CreatedAt: time.Date(2021, 3, 2, 0, 0, 0, 0, time.UTC),
					Location: geo.Location{
						Locality:    "Locality 2",
						CountryCode: "kz",
						Coordinates: geo.Coordinates{
							Latitude:  2.34,
							Longitude: 4.32,
						},
					},
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

			store := NewSpotStore(psqlutil.WrapDB(db))

			spots, err := store.Spots(test.params)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedSpots, spots)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
