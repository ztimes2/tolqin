package csv

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/geo"
	"github.com/ztimes2/tolqin/app/api/internal/pkg/surf"
)

type mockReader struct {
	mock.Mock
}

func newMockReader() *mockReader {
	return &mockReader{}
}

func (m *mockReader) Read(b []byte) (int, error) {
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func TestSpotCreationEntrySource_SpotCreationEntries(t *testing.T) {
	tests := []struct {
		name            string
		readerFn        func(t *testing.T) io.Reader
		expectedEntries []surf.SpotCreationEntry
		expectedErrFn   assert.ErrorAssertionFunc
	}{
		{
			name: "return reader error",
			readerFn: func(t *testing.T) io.Reader {
				m := newMockReader()
				m.
					On("Read", mock.Anything).
					Return(0, errors.New("something went wrong"))
				return m
			},
			expectedEntries: nil,
			expectedErrFn:   assert.Error,
		},
		{
			name: "return 0 entries for empty csv",
			readerFn: func(t *testing.T) io.Reader {
				return strings.NewReader("")
			},
			expectedEntries: nil,
			expectedErrFn:   assert.NoError,
		},
		{
			name: "return 0 entries for csv with 0 rows",
			readerFn: func(t *testing.T) io.Reader {
				b, err := ioutil.ReadFile("testdata/rowless.csv")
				assert.NoError(t, err)
				return bytes.NewReader(b)
			},
			expectedEntries: nil,
			expectedErrFn:   assert.NoError,
		},
		{
			name: "return error for csv with invalid columns",
			readerFn: func(t *testing.T) io.Reader {
				b, err := ioutil.ReadFile("testdata/invalid_columns.csv")
				assert.NoError(t, err)
				return bytes.NewReader(b)
			},
			expectedEntries: nil,
			expectedErrFn:   assert.Error,
		},
		{
			name: "return error for csv with invalid latitude",
			readerFn: func(t *testing.T) io.Reader {
				b, err := ioutil.ReadFile("testdata/invalid_latitude.csv")
				assert.NoError(t, err)
				return bytes.NewReader(b)
			},
			expectedEntries: nil,
			expectedErrFn:   assert.Error,
		},
		{
			name: "return error for csv with invalid longitude",
			readerFn: func(t *testing.T) io.Reader {
				b, err := ioutil.ReadFile("testdata/invalid_longitude.csv")
				assert.NoError(t, err)
				return bytes.NewReader(b)
			},
			expectedEntries: nil,
			expectedErrFn:   assert.Error,
		},
		{
			name: "return entries without error",
			readerFn: func(t *testing.T) io.Reader {
				b, err := ioutil.ReadFile("testdata/valid.csv")
				assert.NoError(t, err)
				return bytes.NewReader(b)
			},
			expectedEntries: []surf.SpotCreationEntry{
				{
					Name: "Abrolhos Islands",
					Location: geo.Location{
						CountryCode: "au",
						Locality:    "City Of Greater Geraldton",
						Coordinates: geo.Coordinates{
							Latitude:  -28.92683,
							Longitude: 113.97929,
						},
					},
				},
				{
					Name: "Cables",
					Location: geo.Location{
						CountryCode: "au",
						Locality:    "Town of Mosman Park",
						Coordinates: geo.Coordinates{
							Latitude:  -32.01783,
							Longitude: 115.7512,
						},
					},
				},
			},
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := NewSpotCreationEntrySource(test.readerFn(t))
			entries, err := s.SpotCreationEntries()
			test.expectedErrFn(t, err)
			assert.Equal(t, test.expectedEntries, entries)
		})
	}
}
