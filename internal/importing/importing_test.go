package importing

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ztimes2/tolqin/internal/geo"
)

type mockSpotEntrySource struct {
	mock.Mock
}

func newMockSpotEntrySource() *mockSpotEntrySource {
	return &mockSpotEntrySource{}
}

func (m *mockSpotEntrySource) SpotEntries() ([]SpotEntry, error) {
	args := m.Called()
	return args.Get(0).([]SpotEntry), args.Error(1)
}

type mockSpotImporter struct {
	mock.Mock
}

func newMockSpotImporter() *mockSpotImporter {
	return &mockSpotImporter{}
}

func (m *mockSpotImporter) ImportSpots(entries []SpotEntry) (int, error) {
	args := m.Called(entries)
	return args.Int(0), args.Error(1)
}

func TestImportSpots(t *testing.T) {
	tests := []struct {
		name          string
		source        SpotEntrySource
		importer      SpotImporter
		expectedCount int
		expectedErrFn assert.ErrorAssertionFunc
	}{
		{
			name: "return error during spot entry source failure",
			source: func() SpotEntrySource {
				m := newMockSpotEntrySource()
				m.
					On("SpotEntries").
					Return(([]SpotEntry)(nil), errors.New("something went wrong"))
				return m
			}(),
			importer:      newMockSpotImporter(),
			expectedCount: 0,
			expectedErrFn: assert.Error,
		},
		{
			name: "return error for entry with invalid country code",
			source: func() SpotEntrySource {
				m := newMockSpotEntrySource()
				m.
					On("SpotEntries").
					Return(
						[]SpotEntry{
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  1.23,
										Longitude: 3.21,
									},
									CountryCode: "zz",
									Locality:    "Locality 1",
								},
								Name: "Spot 1",
							},
						},
						nil,
					)
				return m
			}(),
			importer:      newMockSpotImporter(),
			expectedCount: 0,
			expectedErrFn: assert.Error,
		},
		{
			name: "return error for entry with empty locality",
			source: func() SpotEntrySource {
				m := newMockSpotEntrySource()
				m.
					On("SpotEntries").
					Return(
						[]SpotEntry{
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  1.23,
										Longitude: 3.21,
									},
									CountryCode: "kz",
									Locality:    "",
								},
								Name: "Spot 1",
							},
						},
						nil,
					)
				return m
			}(),
			importer:      newMockSpotImporter(),
			expectedCount: 0,
			expectedErrFn: assert.Error,
		},
		{
			name: "return error for entry with empty name",
			source: func() SpotEntrySource {
				m := newMockSpotEntrySource()
				m.
					On("SpotEntries").
					Return(
						[]SpotEntry{
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  1.23,
										Longitude: 3.21,
									},
									CountryCode: "kz",
									Locality:    "Locality 1",
								},
								Name: "",
							},
						},
						nil,
					)
				return m
			}(),
			importer:      newMockSpotImporter(),
			expectedCount: 0,
			expectedErrFn: assert.Error,
		},
		{
			name: "return error for entry with invalid latitude",
			source: func() SpotEntrySource {
				m := newMockSpotEntrySource()
				m.
					On("SpotEntries").
					Return(
						[]SpotEntry{
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  -91,
										Longitude: 3.21,
									},
									CountryCode: "kz",
									Locality:    "Locality 1",
								},
								Name: "Spot 1",
							},
						},
						nil,
					)
				return m
			}(),
			importer:      newMockSpotImporter(),
			expectedCount: 0,
			expectedErrFn: assert.Error,
		},
		{
			name: "return error for entry with invalid longitude",
			source: func() SpotEntrySource {
				m := newMockSpotEntrySource()
				m.
					On("SpotEntries").
					Return(
						[]SpotEntry{
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  -90,
										Longitude: 181,
									},
									CountryCode: "kz",
									Locality:    "Locality 1",
								},
								Name: "Spot 1",
							},
						},
						nil,
					)
				return m
			}(),
			importer:      newMockSpotImporter(),
			expectedCount: 0,
			expectedErrFn: assert.Error,
		},
		{
			name: "return error during spot importer failure",
			source: func() SpotEntrySource {
				m := newMockSpotEntrySource()
				m.
					On("SpotEntries").
					Return(
						[]SpotEntry{
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  1.23,
										Longitude: 3.21,
									},
									CountryCode: "kz",
									Locality:    "Locality 1",
								},
								Name: "Spot 1",
							},
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  1.23,
										Longitude: 3.21,
									},
									CountryCode: "kz",
									Locality:    "Locality 2",
								},
								Name: "Spot 2",
							},
						},
						nil,
					)
				return m
			}(),
			importer: func() SpotImporter {
				m := newMockSpotImporter()
				m.
					On(
						"ImportSpots",
						[]SpotEntry{
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  1.23,
										Longitude: 3.21,
									},
									CountryCode: "kz",
									Locality:    "Locality 1",
								},
								Name: "Spot 1",
							},
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  1.23,
										Longitude: 3.21,
									},
									CountryCode: "kz",
									Locality:    "Locality 2",
								},
								Name: "Spot 2",
							},
						},
					).
					Return(0, errors.New("something went wrong"))
				return m
			}(),
			expectedCount: 0,
			expectedErrFn: assert.Error,
		},
		{
			name: "return 0 count without error when there is nothing to import",
			source: func() SpotEntrySource {
				m := newMockSpotEntrySource()
				m.
					On("SpotEntries").
					Return(([]SpotEntry)(nil), nil)
				return m
			}(),
			importer: func() SpotImporter {
				m := newMockSpotImporter()
				m.
					On("ImportSpots", ([]SpotEntry)(nil)).
					Return(0, ErrNothingToImport)
				return m
			}(),
			expectedCount: 0,
			expectedErrFn: assert.NoError,
		},
		{
			name: "return non-0 count without error",
			source: func() SpotEntrySource {
				m := newMockSpotEntrySource()
				m.
					On("SpotEntries").
					Return(
						[]SpotEntry{
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  1.23,
										Longitude: 3.21,
									},
									CountryCode: "kz",
									Locality:    "Locality 1",
								},
								Name: "Spot 1",
							},
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  1.23,
										Longitude: 3.21,
									},
									CountryCode: "kz",
									Locality:    "Locality 2",
								},
								Name: "Spot 2",
							},
						},
						nil,
					)
				return m
			}(),
			importer: func() SpotImporter {
				m := newMockSpotImporter()
				m.
					On(
						"ImportSpots",
						[]SpotEntry{
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  1.23,
										Longitude: 3.21,
									},
									CountryCode: "kz",
									Locality:    "Locality 1",
								},
								Name: "Spot 1",
							},
							{
								Location: geo.Location{
									Coordinates: geo.Coordinates{
										Latitude:  1.23,
										Longitude: 3.21,
									},
									CountryCode: "kz",
									Locality:    "Locality 2",
								},
								Name: "Spot 2",
							},
						},
					).
					Return(2, nil)
				return m
			}(),
			expectedCount: 2,
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			count, err := ImportSpots(test.source, test.importer)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedCount, count)
		})
	}
}
