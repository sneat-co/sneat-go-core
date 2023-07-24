package core

// IsAlphanumericOrUnderscore checks the given string contains only letters and digits or underscore
// and does not have any other characters.
func IsAlphanumericOrUnderscore(v string) bool {
	if v == "" {
		return false
	}
	for _, c := range v {
		if !(c == '_' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9') {
			return false
		}
	}
	return true
}
