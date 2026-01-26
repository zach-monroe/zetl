package database

import "strings"

// ParsePostgresTags converts PostgreSQL array format {a,b,c} to []string
func ParsePostgresTags(tagsBytes []byte) []string {
	tagsStr := string(tagsBytes)
	clean := strings.Trim(tagsStr, "{}")
	if len(clean) == 0 {
		return []string{}
	}
	return strings.Split(clean, ",")
}

// FormatPostgresTags converts []string to PostgreSQL array format {a,b,c}
func FormatPostgresTags(tags []string) string {
	return "{" + strings.Join(tags, ",") + "}"
}
