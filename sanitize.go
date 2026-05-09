package main

import "strings"

// sanitizeProviderKey converts a URL host string into a JSON-safe key
// by replacing non-alphanumeric characters (except - and _) with underscores.
func sanitizeProviderKey(host string) string {
	var b strings.Builder
	b.Grow(len(host))
	for _, r := range host {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}
