package conversion

import "time"

// FormatTime renders a timestamp in RFC3339 format. Zero values become an
// empty string.
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.Format(time.RFC3339)
}
