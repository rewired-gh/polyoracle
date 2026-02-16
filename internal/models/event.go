package models

import (
	"errors"
	"time"
)

// Event represents a prediction market event being monitored from Polymarket
type Event struct {
	ID             string    `json:"id"`
	Question       string    `json:"question"`
	Description    string    `json:"description,omitempty"`
	Category       string    `json:"category"`
	YesProbability float64   `json:"yes_probability"`
	NoProbability  float64   `json:"no_probability"`
	Active         bool      `json:"active"`
	LastUpdated    time.Time `json:"last_updated"`
	CreatedAt      time.Time `json:"created_at"`
}

// Validate checks that all event fields are valid
func (e *Event) Validate() error {
	if e.ID == "" {
		return errors.New("event ID must not be empty")
	}
	if e.Question == "" {
		return errors.New("event question must not be empty")
	}
	if e.Category == "" {
		return errors.New("event category must not be empty")
	}
	if e.YesProbability < 0.0 || e.YesProbability > 1.0 {
		return errors.New("yes probability must be between 0.0 and 1.0")
	}
	if e.NoProbability < 0.0 || e.NoProbability > 1.0 {
		return errors.New("no probability must be between 0.0 and 1.0")
	}
	// Allow small tolerance for sum != 1.0 due to floating point precision
	sum := e.YesProbability + e.NoProbability
	if sum < 0.99 || sum > 1.01 {
		return errors.New("yes + no probability should approximately equal 1.0")
	}
	if e.LastUpdated.After(time.Now()) {
		return errors.New("last updated must not be in the future")
	}
	if e.CreatedAt.After(e.LastUpdated) {
		return errors.New("created at must be <= last updated")
	}
	return nil
}
