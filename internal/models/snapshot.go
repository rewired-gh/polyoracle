package models

import (
	"errors"
	"time"
)

// Snapshot represents a point-in-time probability reading for an event.
// Snapshots are recorded periodically to track probability changes over time
// and enable change detection within configurable time windows.
//
// Each snapshot captures the Yes/No probabilities at a specific moment,
// along with metadata about when it was recorded and the data source.
type Snapshot struct {
	ID             string    `json:"id"`
	EventID        string    `json:"event_id"`
	YesProbability float64   `json:"yes_probability"`
	NoProbability  float64   `json:"no_probability"`
	Timestamp      time.Time `json:"timestamp"`
	Source         string    `json:"source"` // Data source identifier (e.g., "polymarket-gamma-api")
}

// Validate checks that all snapshot fields are valid
func (s *Snapshot) Validate() error {
	if s.ID == "" {
		return errors.New("snapshot ID must not be empty")
	}
	if s.EventID == "" {
		return errors.New("event ID must not be empty")
	}
	if s.YesProbability < 0.0 || s.YesProbability > 1.0 {
		return errors.New("yes probability must be between 0.0 and 1.0")
	}
	if s.NoProbability < 0.0 || s.NoProbability > 1.0 {
		return errors.New("no probability must be between 0.0 and 1.0")
	}
	if s.Timestamp.After(time.Now()) {
		return errors.New("timestamp must not be in the future")
	}
	if s.Source == "" {
		return errors.New("source must not be empty")
	}
	return nil
}
