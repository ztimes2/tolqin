package batch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoordinator(t *testing.T) {
	tests := []struct {
		name            string
		size            int
		entries         []int
		expectedBatches []Batch
		expectedEntries []int
	}{
		{
			name:            "do not produce batches when size is 0",
			size:            0,
			entries:         []int{1, 2, 3, 4, 5},
			expectedBatches: nil,
			expectedEntries: nil,
		},
		{
			name:            "do not produce batches when slice is empty",
			size:            2,
			entries:         nil,
			expectedBatches: nil,
			expectedEntries: nil,
		},
		{
			name:    "produce batches",
			size:    2,
			entries: []int{1, 2, 3, 4, 5},
			expectedBatches: []Batch{
				{I: 0, J: 1},
				{I: 2, J: 3},
				{I: 4, J: 4},
			},
			expectedEntries: []int{1, 2, 3, 4, 5},
		},
		{
			name:    "produce batches",
			size:    2,
			entries: []int{1, 2, 3, 4},
			expectedBatches: []Batch{
				{I: 0, J: 1},
				{I: 2, J: 3},
			},
			expectedEntries: []int{1, 2, 3, 4},
		},
		{
			name:    "produce 1 batch when size equals to slice length",
			size:    4,
			entries: []int{1, 2, 3, 4},
			expectedBatches: []Batch{
				{I: 0, J: 3},
			},
			expectedEntries: []int{1, 2, 3, 4},
		},
		{
			name:    "produce 1 batch when size is greater than slice length",
			size:    5,
			entries: []int{1, 2, 3, 4},
			expectedBatches: []Batch{
				{I: 0, J: 3},
			},
			expectedEntries: []int{1, 2, 3, 4},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				batches []Batch
				entries []int
			)

			coord := New(len(test.entries), test.size)
			for coord.HasNext() {
				b := coord.Batch()

				batches = append(batches, b)
				entries = append(entries, test.entries[b.I:b.J+1]...)
			}

			assert.Equal(t, test.expectedBatches, batches)
			assert.Equal(t, test.expectedEntries, entries)
		})
	}
}
