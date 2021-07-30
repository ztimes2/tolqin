package batch

type Batch struct {
	I int
	J int
}

type Coord struct {
	length  int
	size    int
	i       int
	j       int
	hasNext bool
}

func New(length, size int) *Coord {
	var hasNext bool
	if size > 0 && length > 0 {
		hasNext = true
	}
	return &Coord{
		length:  length,
		size:    size,
		i:       0,
		j:       clampIntMax(size-1, length-1),
		hasNext: hasNext,
	}
}

func (c *Coord) HasNext() bool {
	return c.hasNext
}

func (c *Coord) Batch() Batch {
	if !c.hasNext {
		return Batch{}
	}

	b := Batch{
		I: c.i,
		J: c.j,
	}

	c.i = c.j + 1
	c.j = clampIntMax(c.j+c.size, c.length-1)

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
