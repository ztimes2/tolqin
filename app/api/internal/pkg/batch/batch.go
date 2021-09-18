package batch

type Batch struct {
	I int
	J int
}

type Coordinator struct {
	length    int
	batchSize int
	i         int
	j         int
	hasNext   bool
}

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

func (c *Coordinator) HasNext() bool {
	return c.hasNext
}

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

func clampIntMax(i, max int) int {
	if i > max {
		return max
	}
	return i
}
