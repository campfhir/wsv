package internal

func Map[T any, V any](s []T, fn func(e T, i int, a []T) V) []V {
	u := make([]V, len(s))
	c := make([]T, len(s))
	copy(c, s)
	for i, e := range s {
		x := fn(e, i, c)
		u[i] = x
	}
	return u
}
