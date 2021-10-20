package paging

// Limit clamps the given limit value to the given min and max boundaries. If the
// value is less than min, then the given fallback is used.
func Limit(limit, min, max, fallback int) int {
	if limit < min {
		return fallback
	}
	if limit > max {
		return max
	}
	return limit
}

// Offset clamps the given offset value to the given min boundary.
func Offset(offset, min int) int {
	if offset < min {
		return min
	}
	return offset
}
