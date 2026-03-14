package services

import (
	"fmt"
	"time"
)

// HumanizeAge returns a human-readable string representing how long ago t was.
func HumanizeAge(t time.Time) string {
	if t.IsZero() {
		return "—"
	}

	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	case d < 30*24*time.Hour:
		weeks := int(d.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week"
		}
		return fmt.Sprintf("%d weeks", weeks)
	case d < 365*24*time.Hour:
		months := int(d.Hours() / 24 / 30)
		if months == 1 {
			return "1 month"
		}
		return fmt.Sprintf("%d months", months)
	default:
		years := int(d.Hours() / 24 / 365)
		if years == 1 {
			return "1 year"
		}
		return fmt.Sprintf("%d years", years)
	}
}
