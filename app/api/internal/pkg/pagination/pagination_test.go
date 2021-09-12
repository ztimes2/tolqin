package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimit(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		min           int
		max           int
		dflt          int
		expectedLimit int
	}{
		{
			name:          "return default when limit is less than min",
			limit:         0,
			min:           1,
			max:           100,
			dflt:          10,
			expectedLimit: 10,
		},
		{
			name:          "return max when limit is greater than max",
			limit:         101,
			min:           1,
			max:           100,
			dflt:          10,
			expectedLimit: 100,
		},
		{
			name:          "return limit",
			limit:         20,
			min:           1,
			max:           100,
			dflt:          10,
			expectedLimit: 20,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			limit := Limit(test.limit, test.min, test.max, test.dflt)
			assert.Equal(t, test.expectedLimit, limit)
		})
	}
}

func TestOffset(t *testing.T) {
	tests := []struct {
		name           string
		offset         int
		min            int
		expectedOffset int
	}{
		{
			name:           "return min when offset is less than min",
			offset:         -1,
			min:            0,
			expectedOffset: 0,
		},
		{
			name:           "return offset",
			offset:         1,
			min:            0,
			expectedOffset: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			offset := Offset(test.offset, test.min)
			assert.Equal(t, test.expectedOffset, offset)
		})
	}
}
