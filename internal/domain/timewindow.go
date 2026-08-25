package domain

import (
	"strings"
	"time"
)

// ValidateDeploymentWindow requires two RFC3339 timestamps with an explicit offset.
func ValidateDeploymentWindow(value string) error {
	if strings.TrimSpace(value) != value || strings.ContainsAny(value, "\r\n\x00") {
		return ErrInvalid
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		parts = strings.Split(value, ",")
	}
	if len(parts) != 2 {
		return ErrInvalid
	}
	start, err := time.Parse(time.RFC3339, strings.TrimSpace(parts[0]))
	if err != nil || !hasZone(parts[0]) {
		return ErrInvalid
	}
	end, err := time.Parse(time.RFC3339, strings.TrimSpace(parts[1]))
	if err != nil || !hasZone(parts[1]) || !start.Before(end) {
		return ErrInvalid
	}
	return nil
}

func hasZone(v string) bool {
	return strings.Contains(v, "Z") || strings.LastIndexAny(v, "+-") > strings.LastIndex(v, "T")
}

func ValidWindow(window string) bool                { return window != "" }
func WithinWindow(window string, at time.Time) bool { return ValidWindow(window) && !at.IsZero() }
func (t *Task) Touch(at time.Time)                  { t.UpdatedAt = at }
