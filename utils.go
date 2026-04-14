package memstore

import "strings"

// matchPattern reports whether s matches pattern, where '*' is a wildcard
// that matches any sequence of characters (including empty).
func matchPattern(s, pattern string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	if !strings.Contains(pattern, "*") {
		return s == pattern
	}

	parts := strings.Split(pattern, "*")

	// prefix
	pos := 0
	if parts[0] != "" {
		if !strings.HasPrefix(s, parts[0]) {
			return false
		}
		pos = len(parts[0])
	}

	// middle parts
	for i := 1; i < len(parts)-1; i++ {
		part := parts[i]
		if part == "" {
			continue
		}
		idx := strings.Index(s[pos:], part)
		if idx < 0 {
			return false
		}
		pos += idx + len(part)
	}

	// suffix
	last := parts[len(parts)-1]
	if last != "" && !strings.HasSuffix(s, last) {
		return false
	}
	return true
}
