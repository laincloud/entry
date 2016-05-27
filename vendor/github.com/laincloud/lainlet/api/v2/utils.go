package v2

func fixPrefix(s string) string {
	l := len(s)
	if l == 0 {
		return s
	}
	if s[l-1] == '/' {
		return s
	}
	return s + "/"
}
