package valerra

// StringNotEmpty returns a condition function that checks if the given string is
// empty or not.
func StringNotEmpty(s string) func() bool {
	return func() bool {
		return s != ""
	}
}

// StringLessOrEqual return a condition function that checks if the character length
// of the given string is less or equal to the given size.
func StringLessOrEqual(s string, size int) func() bool {
	return func() bool {
		return len(s) <= size
	}
}
