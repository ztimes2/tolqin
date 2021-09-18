package batch

type Batch struct {
	I int
	J int
}

type Batcher struct {
	length    int
	batchSize int
	i         int
	j         int
	hasNext   bool
}

func New(length, batchSize int) *Batcher {
	var hasNext bool
	if batchSize > 0 && length > 0 {
		hasNext = true
	}
	return &Batcher{
		length:    length,
		batchSize: batchSize,
		i:         0,
		j:         clampIntMax(batchSize-1, length-1),
		hasNext:   hasNext,
	}
}

func (b *Batcher) HasNext() bool {
	return b.hasNext
}

func (b *Batcher) Batch() Batch {
	if !b.hasNext {
		return Batch{}
	}

	batch := Batch{
		I: b.i,
		J: b.j,
	}

	b.i = b.j + 1
	b.j = clampIntMax(b.j+b.batchSize, b.length-1)

	if b.i > b.length-1 {
		b.hasNext = false
	}

	return batch
}

func clampIntMax(i, max int) int {
	if i > max {
		return max
	}
	return i
}
