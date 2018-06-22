package go_contentline

import "strings"

//ValidID checks if a string is a valid identifier (iana-token or x-token).
// If not, returns the conflicting rune.
func ValidID(in string) *rune {
	for _, r := range []rune(in) {
		if !strings.ContainsRune(parName, r) {
			return &r
		}
	}
	return nil
}
