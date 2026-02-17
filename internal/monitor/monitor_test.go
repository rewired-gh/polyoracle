package monitor

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poly-oracle/internal/models"
	"github.com/poly-oracle/internal/storage"
)

func TestDetectChanges(t *testing.T) {
	s := storage.New(100, 50, "/tmp/test-monitor.json", 0644, 0755)
	m := New(s)

	// Create test event
	now := time.Now()
	event := models.Event{
		ID:             "event-1",
		EventID:        "event-1",
		Title:          "Will X happen?",
		Category:       "politics",
		YesProbability: 0.75,
		NoProbability:  0.25,
		Active:         true,
		LastUpdated:    now,
		CreatedAt:      now.Add(-1 * time.Hour),
	}
	if err := s.AddEvent(&event); err != nil {
		t.Fatalf("Failed to add event: %v", err)
	}

	// Add snapshots with changing probabilities
	snapshots := []models.Snapshot{
		{
			ID:             uuid.New().String(),
			EventID:        "event-1",
			YesProbability: 0.60,
			NoProbability:  0.40,
			Timestamp:      now.Add(-1 * time.Hour),
			Source:         "test",
		},
		{
			ID:             uuid.New().String(),
			EventID:        "event-1",
			YesProbability: 0.75,
			NoProbability:  0.25,
			Timestamp:      now,
			Source:         "test",
		},
	}

	for _, snap := range snapshots {
		if err := s.AddSnapshot(&snap); err != nil {
			t.Fatalf("Failed to add snapshot: %v", err)
		}
	}

	// Detect changes
	events := []models.Event{event}
	changes, _, err := m.DetectChanges(events, 0.10, 2*time.Hour)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	if len(changes) == 0 {
		t.Error("Expected at least 1 change, got 0")
		return
	}

	if changes[0].Magnitude < 0.149 || changes[0].Magnitude > 0.151 {
		t.Errorf("Expected magnitude 0.15, got %f", changes[0].Magnitude)
	}

	if changes[0].Direction != "increase" {
		t.Errorf("Expected direction 'increase', got '%s'", changes[0].Direction)
	}
}

func TestRankChanges(t *testing.T) {
	s := storage.New(100, 50, "/tmp/test-monitor-rank.json", 0644, 0755)
	m := New(s)

	changes := []models.Change{
		{
			ID:             "change-1",
			EventID:        "event-1",
			EventQuestion:  "Question 1",
			Magnitude:      0.15,
			Direction:      "increase",
			OldProbability: 0.60,
			NewProbability: 0.75,
			TimeWindow:     time.Hour,
			DetectedAt:     time.Now(),
		},
		{
			ID:             "change-2",
			EventID:        "event-2",
			EventQuestion:  "Question 2",
			Magnitude:      0.25,
			Direction:      "increase",
			OldProbability: 0.50,
			NewProbability: 0.75,
			TimeWindow:     time.Hour,
			DetectedAt:     time.Now(),
		},
		{
			ID:             "change-3",
			EventID:        "event-3",
			EventQuestion:  "Question 3",
			Magnitude:      0.10,
			Direction:      "decrease",
			OldProbability: 0.80,
			NewProbability: 0.70,
			TimeWindow:     time.Hour,
			DetectedAt:     time.Now(),
		},
	}

	// Get top 2
	top := m.RankChanges(changes, 2)

	if len(top) != 2 {
		t.Errorf("Expected 2 changes, got %d", len(top))
	}

	// Verify sorted by magnitude descending
	if top[0].Magnitude < top[1].Magnitude {
		t.Error("Changes not sorted by magnitude descending")
	}

	// Top should be 0.25
	if top[0].Magnitude != 0.25 {
		t.Errorf("Expected top magnitude 0.25, got %f", top[0].Magnitude)
	}

	// Second should be 0.15
	if top[1].Magnitude != 0.15 {
		t.Errorf("Expected second magnitude 0.15, got %f", top[1].Magnitude)
	}
}

func TestDetectChanges_BelowThreshold(t *testing.T) {
	s := storage.New(100, 50, "/tmp/test-threshold.json", 0644, 0755)
	m := New(s)

	now := time.Now()
	event := models.Event{
		ID:             "event-1",
		EventID:        "event-1",
		Title:          "Test?",
		Category:       "politics",
		YesProbability: 0.65,
		NoProbability:  0.35,
		Active:         true,
		LastUpdated:    now,
		CreatedAt:      now.Add(-1 * time.Hour),
	}
	if err := s.AddEvent(&event); err != nil {
		t.Fatalf("Failed to add event: %v", err)
	}

	// Add snapshots with small change (below threshold)
	snapshots := []models.Snapshot{
		{
			ID:             uuid.New().String(),
			EventID:        "event-1",
			YesProbability: 0.60,
			NoProbability:  0.40,
			Timestamp:      now.Add(-1 * time.Hour),
			Source:         "test",
		},
		{
			ID:             uuid.New().String(),
			EventID:        "event-1",
			YesProbability: 0.65,
			NoProbability:  0.35,
			Timestamp:      now,
			Source:         "test",
		},
	}

	for _, snap := range snapshots {
		if err := s.AddSnapshot(&snap); err != nil {
			t.Fatalf("Failed to add snapshot: %v", err)
		}
	}

	// Detect with 0.10 threshold (change is 0.05)
	events := []models.Event{event}
	changes, _, err := m.DetectChanges(events, 0.10, 2*time.Hour)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	// Should not detect change below threshold
	if len(changes) != 0 {
		t.Errorf("Expected 0 changes (below threshold), got %d", len(changes))
	}
}

func TestDetectChanges_OutOfOrderSnapshots(t *testing.T) {
	s := storage.New(100, 50, "/tmp/test-out-of-order.json", 0644, 0755)
	m := New(s)

	now := time.Now()
	event := models.Event{
		ID:             "event-1",
		EventID:        "event-1",
		Title:          "Test?",
		Category:       "politics",
		YesProbability: 0.85,
		NoProbability:  0.15,
		Active:         true,
		LastUpdated:    now,
		CreatedAt:      now.Add(-2 * time.Hour),
	}
	if err := s.AddEvent(&event); err != nil {
		t.Fatalf("Failed to add event: %v", err)
	}

	// Add snapshots OUT OF ORDER to test sorting
	// Oldest: 0.50 (2 hours ago)
	// Middle: 0.70 (1 hour ago)
	// Newest: 0.85 (now)
	// Expected change: 0.85 - 0.50 = 0.35
	snapshots := []models.Snapshot{
		{
			ID:             uuid.New().String(),
			EventID:        "event-1",
			YesProbability: 0.70, // Middle (added first, but not oldest)
			NoProbability:  0.30,
			Timestamp:      now.Add(-1 * time.Hour),
			Source:         "test",
		},
		{
			ID:             uuid.New().String(),
			EventID:        "event-1",
			YesProbability: 0.50, // Oldest (added second)
			NoProbability:  0.50,
			Timestamp:      now.Add(-2 * time.Hour),
			Source:         "test",
		},
		{
			ID:             uuid.New().String(),
			EventID:        "event-1",
			YesProbability: 0.85, // Newest (added last)
			NoProbability:  0.15,
			Timestamp:      now,
			Source:         "test",
		},
	}

	for _, snap := range snapshots {
		if err := s.AddSnapshot(&snap); err != nil {
			t.Fatalf("Failed to add snapshot: %v", err)
		}
	}

	// Detect changes with 0.10 threshold
	events := []models.Event{event}
	changes, _, err := m.DetectChanges(events, 0.10, 3*time.Hour)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	if len(changes) == 0 {
		t.Fatal("Expected at least 1 change, got 0")
	}

	// Verify magnitude is calculated from oldest (0.50) to newest (0.85)
	// NOT from first added (0.70) to last added (0.85)
	expectedMagnitude := 0.35
	if changes[0].Magnitude < expectedMagnitude-0.01 || changes[0].Magnitude > expectedMagnitude+0.01 {
		t.Errorf("Expected magnitude %.2f (0.85 - 0.50), got %.2f", expectedMagnitude, changes[0].Magnitude)
	}

	if changes[0].Direction != "increase" {
		t.Errorf("Expected direction 'increase', got '%s'", changes[0].Direction)
	}

	// Verify old and new probabilities
	if changes[0].OldProbability != 0.50 {
		t.Errorf("Expected old probability 0.50, got %.2f", changes[0].OldProbability)
	}
	if changes[0].NewProbability != 0.85 {
		t.Errorf("Expected new probability 0.85, got %.2f", changes[0].NewProbability)
	}
}
