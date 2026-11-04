package crawler

import (
	"errors"
	"time"
)

const (
	ModeManual   = "MANUAL"
	ModeInterval = "INTERVAL"

	// MinIntervalMinutes keeps a pesantren from hammering its own (or Google's) servers.
	MinIntervalMinutes = 360
	MaxIntervalMinutes = 30 * 24 * 60
)

var (
	ErrScheduleInvalid     = errors.New("INVALID_REQUEST")
	ErrScheduleTooFrequent = errors.New("CRAWL_SCHEDULE_TOO_FREQUENT")
)

// ValidateSchedule checks a mode/interval pair as sent by the dashboard.
func ValidateSchedule(mode string, intervalMinutes int) error {
	switch mode {
	case ModeManual:
		return nil
	case ModeInterval:
		if intervalMinutes < MinIntervalMinutes {
			return ErrScheduleTooFrequent
		}
		if intervalMinutes > MaxIntervalMinutes {
			return ErrScheduleInvalid
		}
		return nil
	}
	return ErrScheduleInvalid
}

// NextRun is when a source should next be fetched after `from`; nil for manual sources.
func NextRun(mode string, intervalMinutes int, from time.Time) *time.Time {
	if mode != ModeInterval || intervalMinutes <= 0 {
		return nil
	}
	next := from.Add(time.Duration(intervalMinutes) * time.Minute)
	return &next
}
