package importing

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/geo"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/surf"
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

type mockSpotStore struct {
	mock.Mock
}

func newMockSpotStore() *mockSpotStore {
	return &mockSpotStore{}
}

func (m *mockSpotStore) CreateSpots(entries []surf.SpotCreationEntry) error {
	args := m.Called(entries)
	return args.Error(0)
}

func TestImportSpots(t *testing.T) {
	tests := []struct {
		name          string
		source        SpotEntrySource
		store         SpotStore
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
			store:         newMockSpotStore(),
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
			store:         newMockSpotStore(),
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
			store:         newMockSpotStore(),
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
			store:         newMockSpotStore(),
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
			store:         newMockSpotStore(),
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
			store:         newMockSpotStore(),
			expectedCount: 0,
			expectedErrFn: assert.Error,
		},
		{
			name: "return error during spot store failure",
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
			store: func() SpotStore {
				m := newMockSpotStore()
				m.
					On(
						"CreateSpots",
						[]surf.SpotCreationEntry{
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
					Return(errors.New("something went wrong"))
				return m
			}(),
			expectedCount: 0,
			expectedErrFn: assert.Error,
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
									CountryCode: " kz ",
									Locality:    " Locality 1 ",
								},
								Name: " Spot 1 ",
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
			store: func() SpotStore {
				m := newMockSpotStore()
				m.
					On(
						"CreateSpots",
						[]surf.SpotCreationEntry{
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
					Return(nil)
				return m
			}(),
			expectedCount: 2,
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			count, err := ImportSpots(test.source, test.store)
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedCount, count)
		})
	}
}
