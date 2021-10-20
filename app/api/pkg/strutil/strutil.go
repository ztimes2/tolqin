package strutil

// RepeatRune repeats the given rune n times and returns the result as string.
func RepeatRune(r rune, n int) string {
	var s string
	for i := 0; i < n; i++ {
		s += string(r)
	}
	return s
}
