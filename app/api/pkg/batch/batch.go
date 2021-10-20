/*
Package batch helps with creation and coordination of batches for a slice/array.

For example, the following piece of code splits a slice that contains 10 elements
into batches with a size of 3 and prints them one by one:

	list := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	coord := batch.New(len(list), 3)
	for coord.HasNext() {
		b := coord.Batch()
		fmt.Println(list[b.I:b.J+1])
	}

	// Output:
	// [1, 2, 3]
	// [4, 5, 6]
	// [7, 8, 9]
	// [10]
*/
package batch

// Coordinator creates and coordinates batches for a slice/array.
type Coordinator struct {
	length    int
	batchSize int
	i         int
	j         int
	hasNext   bool
}

// New returns a new *Coordinator for a slice/array with the given length and the
// desired batch size.
func New(length, batchSize int) *Coordinator {
	var hasNext bool
	if batchSize > 0 && length > 0 {
		hasNext = true
	}
	return &Coordinator{
		length:    length,
		batchSize: batchSize,
		i:         0,
		j:         clampIntMax(batchSize-1, length-1),
		hasNext:   hasNext,
	}
}

// HasNext checks whether there is a batch that is waiting to be consumed.
//
// When true gets returned, the caller is expected to envoke Batch() method for
// consuming that awaiting batch.
func (c *Coordinator) HasNext() bool {
	return c.hasNext
}

// Batch returns a batch that is waiting to be consumed.
//
// If there are no more batches left, then an empty Batch gets returned. The caller
// is expected to check the availability of a batch using HasNext() method prior
// to its consumption.
//
// Once a batch is consumed, the coordinator immediately queues the next batch if
// available.
func (c *Coordinator) Batch() Batch {
	if !c.hasNext {
		return Batch{}
	}

	b := Batch{
		I: c.i,
		J: c.j,
	}

	c.i = c.j + 1
	c.j = clampIntMax(c.j+c.batchSize, c.length-1)

	if c.i > c.length-1 {
		c.hasNext = false
	}

	return b
}

// Batch holds indices of a batch.
type Batch struct {
	// I is the index of the first element of a batch.
	I int
	// J is the index of the last element of a batch.
	J int
}

func clampIntMax(i, max int) int {
	if i > max {
		return max
	}
	return i
}
