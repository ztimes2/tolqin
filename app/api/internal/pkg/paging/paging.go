package paging

func Limit(limit, min, max, fallback int) int {
	if limit < min {
		return fallback
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
