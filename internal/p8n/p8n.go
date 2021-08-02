package p8n

func Limit(limit, min, max, dflt int) int {
	if limit < min {
		return dflt
	}
	if limit > max {
		return max
	}
	return limit
}

func Offset(offset, min int) int {
	if offset < min {
		return min
	}
	return offset
}
