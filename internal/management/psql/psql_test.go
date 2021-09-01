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
	"github.com/ztimes2/tolqin/internal/management"
	"github.com/ztimes2/tolqin/internal/pconv"
	"github.com/ztimes2/tolqin/internal/psqlutil"
	"github.com/ztimes2/tolqin/internal/testutil"
)

func TestSpotStore_Spot(t *testing.T) {
	tests := []struct {
		name          string
		mockFn        func(sqlmock.Sqlmock)
		id            string
		expectedSpot  management.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during query execution",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, locality, country_code, created_at " +
							"FROM spots WHERE CAST(id AS VARCHAR) = $1",
					)).
					WithArgs("1").
					WillReturnError(errors.New("something went wrong"))
			},
			id:            "1",
			expectedSpot:  management.Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return error for unexisting resource",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, locality, country_code, created_at " +
							"FROM spots WHERE CAST(id AS VARCHAR) = $1",
					)).
					WithArgs("1").
					WillReturnError(sql.ErrNoRows)
			},
			id:            "1",
			expectedSpot:  management.Spot{},
			expectedErrFn: testutil.IsError(management.ErrNotFound),
		},
		{
			name: "return spot without error",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, locality, country_code, created_at " +
							"FROM spots WHERE CAST(id AS VARCHAR) = $1",
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
			expectedSpot: management.Spot{
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
		params        management.SpotsParams
		mockFn        func(sqlmock.Sqlmock)
		expectedSpots []management.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during query execution",
			params: management.SpotsParams{
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
			params: management.SpotsParams{
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
			params: management.SpotsParams{
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
			params: management.SpotsParams{
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
			expectedSpots: []management.Spot{
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
			params: management.SpotsParams{
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
			expectedSpots: []management.Spot{
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
			params: management.SpotsParams{
				Limit:  10,
				Offset: 0,
				Query:  "query",
			},
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, locality, country_code, created_at "+
							"FROM spots "+
							"WHERE (name ILIKE $1 OR locality ILIKE $2 OR CAST(id AS VARCHAR) ILIKE $3) "+
							"LIMIT 10 OFFSET 0",
					)).
					WithArgs("%query%", "%query%", "%query%").
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "locality", "country_code", "created_at",
						}).
						AddRow("1", "Spot 1", 1.23, 3.21, "Locality 1", "kz", time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)).
						AddRow("2", "Spot 2", 2.34, 4.32, "Locality 2", "kz", time.Date(2021, 3, 2, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()
			},
			expectedSpots: []management.Spot{
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
			params: management.SpotsParams{
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
			expectedSpots: []management.Spot{
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
			params: management.SpotsParams{
				Limit:       10,
				Offset:      0,
				CountryCode: "kz",
				Query:       "query",
			},
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"SELECT id, name, latitude, longitude, locality, country_code, created_at "+
							"FROM spots WHERE country_code = $1 "+
							"AND (name ILIKE $2 OR locality ILIKE $3 OR CAST(id AS VARCHAR) ILIKE $4) "+
							"LIMIT 10 OFFSET 0",
					)).
					WithArgs("kz", "%query%", "%query%", "%query%").
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "locality", "country_code", "created_at",
						}).
						AddRow("1", "Spot 1", 1.23, 3.21, "Locality 1", "kz", time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)).
						AddRow("2", "Spot 2", 2.34, 4.32, "Locality 2", "kz", time.Date(2021, 3, 2, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()
			},
			expectedSpots: []management.Spot{
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

func TestSpotStore_CreateSpot(t *testing.T) {
	tests := []struct {
		name          string
		mockFn        func(sqlmock.Sqlmock)
		params        management.CreateSpotParams
		expectedSpot  management.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during query execution",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude,locality,country_code) "+
							"VALUES ($1,$2,$3,$4,$5) "+
							"RETURNING id, name, latitude, longitude, locality, country_code, created_at",
					)).
					WithArgs("Spot 1", 1.23, 3.21, "Locality 1", "Country code 1").
					WillReturnError(errors.New("unexpected error"))
			},
			params: management.CreateSpotParams{
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
			expectedSpot:  management.Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return spot without error",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"INSERT INTO spots (name,latitude,longitude,locality,country_code) "+
							"VALUES ($1,$2,$3,$4,$5) "+
							"RETURNING id, name, latitude, longitude, locality, country_code, created_at",
					)).
					WithArgs("Spot 1", 1.23, 3.21, "Locality 1", "Country code 1").
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "locality", "country_code", "created_at",
						}).
						AddRow("1", "Spot 1", 1.23, 3.21, "Locality 1", "Country code 1", time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()
			},
			params: management.CreateSpotParams{
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
			expectedSpot: management.Spot{
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
		params        management.UpdateSpotParams
		expectedSpot  management.Spot
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during query execution",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"UPDATE spots "+
							"SET country_code = $1, latitude = $2, locality = $3, longitude = $4, name = $5 "+
							"WHERE CAST(id AS VARCHAR) = $6 "+
							"RETURNING id, name, latitude, longitude, locality, country_code, created_at",
					)).
					WithArgs("Country code 1", 2.34, "Locality 1", 4.32, "Updated spot 1", "1").
					WillReturnError(errors.New("unexpected error"))
			},
			params: management.UpdateSpotParams{
				ID:          "1",
				Name:        pconv.String("Updated spot 1"),
				Locality:    pconv.String("Locality 1"),
				CountryCode: pconv.String("Country code 1"),
				Latitude:    pconv.Float64(2.34),
				Longitude:   pconv.Float64(4.32),
			},
			expectedSpot:  management.Spot{},
			expectedErrFn: assert.Error,
		},
		{
			name: "return error for unexisting resource",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"UPDATE spots "+
							"SET country_code = $1, latitude = $2, locality = $3, longitude = $4, name = $5 "+
							"WHERE CAST(id AS VARCHAR) = $6 "+
							"RETURNING id, name, latitude, longitude, locality, country_code, created_at",
					)).
					WithArgs("Country code 1", 2.34, "Locality 1", 4.32, "Updated spot 1", "1").
					WillReturnError(sql.ErrNoRows)
			},
			params: management.UpdateSpotParams{
				ID:          "1",
				Name:        pconv.String("Updated spot 1"),
				Locality:    pconv.String("Locality 1"),
				CountryCode: pconv.String("Country code 1"),
				Latitude:    pconv.Float64(2.34),
				Longitude:   pconv.Float64(4.32),
			},
			expectedSpot:  management.Spot{},
			expectedErrFn: testutil.IsError(management.ErrNotFound),
		},
		{
			name:   "return error when nothing to update",
			mockFn: func(m sqlmock.Sqlmock) {},
			params: management.UpdateSpotParams{
				ID: "1",
			},
			expectedSpot:  management.Spot{},
			expectedErrFn: testutil.IsError(management.ErrNothingToUpdate),
		},
		{
			name: "return spot without error for full update",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"UPDATE spots "+
							"SET country_code = $1, latitude = $2, locality = $3, longitude = $4, name = $5 "+
							"WHERE CAST(id AS VARCHAR) = $6 "+
							"RETURNING id, name, latitude, longitude, locality, country_code, created_at",
					)).
					WithArgs("Country code 1", 2.34, "Locality 1", 4.32, "Updated spot 1", "1").
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "locality", "country_code", "created_at",
						}).
						AddRow("1", "Updated spot 1", 2.34, 4.32, "Locality 1", "Country code 1", time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()
			},
			params: management.UpdateSpotParams{
				ID:          "1",
				Name:        pconv.String("Updated spot 1"),
				Locality:    pconv.String("Locality 1"),
				CountryCode: pconv.String("Country code 1"),
				Latitude:    pconv.Float64(2.34),
				Longitude:   pconv.Float64(4.32),
			},
			expectedSpot: management.Spot{
				ID:        "1",
				Name:      "Updated spot 1",
				CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
				Location: geo.Location{
					Locality:    "Locality 1",
					CountryCode: "Country code 1",
					Coordinates: geo.Coordinates{
						Latitude:  2.34,
						Longitude: 4.32,
					},
				},
			},
			expectedErrFn: assert.NoError,
		},
		{
			name: "return spot without error for partial update",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectQuery(regexp.QuoteMeta(
						"UPDATE spots "+
							"SET latitude = $1, name = $2 "+
							"WHERE CAST(id AS VARCHAR) = $3 "+
							"RETURNING id, name, latitude, longitude, locality, country_code, created_at",
					)).
					WithArgs(2.34, "Updated spot 1", "1").
					WillReturnRows(sqlmock.
						NewRows([]string{
							"id", "name", "latitude", "longitude", "locality", "country_code", "created_at",
						}).
						AddRow("1", "Updated spot 1", 2.34, 4.32, "Locality 1", "Country code 1", time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)),
					).
					RowsWillBeClosed()
			},
			params: management.UpdateSpotParams{
				ID:       "1",
				Name:     pconv.String("Updated spot 1"),
				Latitude: pconv.Float64(2.34),
			},
			expectedSpot: management.Spot{
				ID:        "1",
				Name:      "Updated spot 1",
				CreatedAt: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC),
				Location: geo.Location{
					Locality:    "Locality 1",
					CountryCode: "Country code 1",
					Coordinates: geo.Coordinates{
						Latitude:  2.34,
						Longitude: 4.32,
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
			name: "return error during query execution",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectExec(regexp.QuoteMeta(
						"DELETE FROM spots WHERE CAST(id AS VARCHAR) = $1",
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
						"DELETE FROM spots WHERE CAST(id AS VARCHAR) = $1",
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
						"DELETE FROM spots WHERE CAST(id AS VARCHAR) = $1",
					)).
					WithArgs("1").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			id:            "1",
			expectedErrFn: testutil.IsError(management.ErrNotFound),
		},
		{
			name: "return no error",
			mockFn: func(m sqlmock.Sqlmock) {
				m.
					ExpectExec(regexp.QuoteMeta(
						"DELETE FROM spots WHERE CAST(id AS VARCHAR) = $1",
					)).
					WithArgs("1").
					WillReturnResult(sqlmock.NewResult(0, 1))
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

			store := NewSpotStore(psqlutil.WrapDB(db))
			err = store.DeleteSpot(test.id)
			test.expectedErrFn(t, err)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
