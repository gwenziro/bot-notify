package utils

import "time"

// FormatTime memformat waktu ke RFC3339 dengan handling nil
func FormatTime(t *time.Time) string {
	if t == nil {
		return time.Now().Format(time.RFC3339)
	}
	return t.Format(time.RFC3339)
}
