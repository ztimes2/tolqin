package valerra

// StringNotEmpty returns a condition that checks if the given string is not empty.
func StringNotEmpty(s string) Condition {
	return func() bool {
		return s != ""
	}
}

// StringLessOrEqual returns a condition that checks if the character length of
// the given string is less or equal to the given size.
func StringLessOrEqual(s string, size int) Condition {
	return func() bool {
		return len(s) <= size
	}
}
