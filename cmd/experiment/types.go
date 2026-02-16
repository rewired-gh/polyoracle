package main

import "time"

// EventAnalysis holds analyzed event data
type EventAnalysis struct {
	ID             string
	Title          string
	Categories     []string
	Volume24hr     float64
	Volume1wk      float64
	Volume1mo      float64
	ProbabilityYes float64
	Active         bool
}

// ProbabilitySnapshot captures a moment in time
type ProbabilitySnapshot struct {
	Timestamp  time.Time
	EventID    string
	Title      string
	ProbYes    float64
	Volume24hr float64
}

// CategoryStats holds statistical data for a category
type CategoryStats struct {
	Category      string
	EventCount    int
	AvgVolume24hr float64
	MaxVolume24hr float64
	TotalVolume   float64
}

// VolumeAnalysis holds volume distribution analysis
type VolumeAnalysis struct {
	TotalEvents     int
	Top10Percent    []EventAnalysis
	Top20Percent    []EventAnalysis
	Top50Percent    []EventAnalysis
	ThresholdCounts map[string]int
}

// NewVolumeAnalysis creates a new VolumeAnalysis with initialized map
func NewVolumeAnalysis() VolumeAnalysis {
	return VolumeAnalysis{
		ThresholdCounts: make(map[string]int),
	}
}
