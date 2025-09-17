package internal

func DefaultIfEmpty(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
