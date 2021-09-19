package pconv

// String returns a pointer to the given string.
func String(s string) *string {
	return &s
}

// Float64 returns a pointer to the given float64.
func Float64(f float64) *float64 {
	return &f
}
