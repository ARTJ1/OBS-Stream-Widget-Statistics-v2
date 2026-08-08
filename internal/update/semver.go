package update

import (
	"strconv"
	"strings"
)

func normalizeTag(tag string) string {
	tag = strings.TrimSpace(tag)
	tag = strings.TrimPrefix(strings.ToLower(tag), "v")
	if i := strings.IndexAny(tag, "-+"); i >= 0 {
		tag = tag[:i]
	}
	return tag
}

// Newer reports whether latest is a higher semver than current.
// Non-semver / empty / equal → false (bonus tags like streamdock-icons-* are ignored).
func Newer(latest, current string) bool {
	l := parseSemver(normalizeTag(latest))
	c := parseSemver(normalizeTag(current))
	if l == nil || c == nil {
		return false
	}
	for i := 0; i < 3; i++ {
		if l[i] > c[i] {
			return true
		}
		if l[i] < c[i] {
			return false
		}
	}
	return false
}

func parseSemver(s string) []int {
	if s == "" || s == "dev" {
		return nil
	}
	parts := strings.Split(s, ".")
	if len(parts) < 1 {
		return nil
	}
	out := make([]int, 3)
	for i := 0; i < 3 && i < len(parts); i++ {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			return nil
		}
		out[i] = n
	}
	return out
}
