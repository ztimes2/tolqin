package valerra

func StringNotEmpty(s string) func() bool {
	return func() bool {
		return s != ""
	}
}

func StringLessOrEqual(s string, size int) func() bool {
	return func() bool {
		return len(s) <= size
	}
}
