package core

import "unicode"

// IsAlphanumericOrUnderscore checks the given string contains only letters and digits or underscore
// and does not have any other characters.
func IsAlphanumericOrUnderscore(v string) bool {
	if v == "" {
		return false
	}
	for _, c := range v {
		if unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' {
			continue
		}
		return false
	}
	return true
}
