package util

import "strings"

// In returns true if s is within vs or if vs contains any string that contain s.
func In(vs []string, s string) bool {
	for _, v := range vs {
		if s == v || strings.Contains(s, v) {
			return true
		}
	}

	return false
}
