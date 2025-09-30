package composit

// Mostly a copy of Go source before we can upgrade to 1.21

func delete[S ~[]E, E any](s S, i, j int) S {
	_ = s[i:j:len(s)] // bounds check

	if i == j {
		return s
	}

	s = append(s[:i], s[j:]...)
	return s
}

func indexFunc[S ~[]E, E any](s S, f func(E) bool) int {
	for i := range s {
		if f(s[i]) {
			return i
		}
	}
	return -1
}

func insert[S ~[]E, E any](s S, i int, v E) S {
	if len(s) == i { // nil or empty slice or after last element
		return append(s, v)
	}
	s = append(s[:i+1], s[i:]...) // index < len(a)
	s[i] = v
	return s
}
