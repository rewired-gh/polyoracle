// Package storage provides thread-safe in-memory storage with file-based persistence.
// It manages events, probability snapshots, and detected changes with automatic
// data rotation to prevent unbounded memory growth.
//
// Storage is designed for reliability with atomic file writes and graceful
// handling of persistence failures. Data is persisted to JSON files and can
// be restored on application restart.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/poly-oracle/internal/models"
)

// Storage provides thread-safe in-memory storage with file-based persistence
type Storage struct {
	events    map[string]*models.Event
	snapshots map[string][]models.Snapshot
	changes   []models.Change
	mu        sync.RWMutex

	// Configuration
	maxEvents            int
	maxSnapshotsPerEvent int
	filePath             string
}

// PersistenceFile represents the file structure for JSON persistence
type PersistenceFile struct {
	Version   string                       `json:"version"`
	SavedAt   time.Time                    `json:"saved_at"`
	Events    map[string]*models.Event     `json:"events"`
	Snapshots map[string][]models.Snapshot `json:"snapshots"`
}

// New creates a new Storage instance with persistence to tmp directory
// If filePath is empty, uses OS-appropriate tmp directory
func New(maxEvents, maxSnapshotsPerEvent int, filePath string) *Storage {
	// Use OS-appropriate tmp directory if no path provided
	if filePath == "" {
		filePath = filepath.Join(os.TempDir(), "poly-oracle", "data.json")
	}

	return &Storage{
		events:               make(map[string]*models.Event),
		snapshots:            make(map[string][]models.Snapshot),
		changes:              make([]models.Change, 0),
		maxEvents:            maxEvents,
		maxSnapshotsPerEvent: maxSnapshotsPerEvent,
		filePath:             filePath,
	}
}

// AddEvent adds or updates an event in storage
func (s *Storage) AddEvent(event *models.Event) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.events[event.ID] = event
	return nil
}

// GetEvent retrieves an event by ID
func (s *Storage) GetEvent(id string) (*models.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, exists := s.events[id]
	if !exists {
		return nil, fmt.Errorf("event not found: %s", id)
	}
	return event, nil
}

// GetAllEvents returns all events
func (s *Storage) GetAllEvents() ([]*models.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]*models.Event, 0, len(s.events))
	for _, event := range s.events {
		events = append(events, event)
	}
	return events, nil
}

// UpdateEvent updates an existing event
func (s *Storage) UpdateEvent(event *models.Event) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[event.ID]; !exists {
		return fmt.Errorf("event not found: %s", event.ID)
	}

	s.events[event.ID] = event
	return nil
}

// AddSnapshot adds a new snapshot for an event
func (s *Storage) AddSnapshot(snapshot *models.Snapshot) error {
	if err := snapshot.Validate(); err != nil {
		return fmt.Errorf("invalid snapshot: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify event exists
	if _, exists := s.events[snapshot.EventID]; !exists {
		return fmt.Errorf("event not found: %s", snapshot.EventID)
	}

	s.snapshots[snapshot.EventID] = append(s.snapshots[snapshot.EventID], *snapshot)
	return nil
}

// GetSnapshots retrieves all snapshots for an event
func (s *Storage) GetSnapshots(eventID string) ([]models.Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshots, exists := s.snapshots[eventID]
	if !exists {
		return []models.Snapshot{}, nil
	}

	return snapshots, nil
}

// GetSnapshotsInWindow retrieves snapshots within a time window for an event
func (s *Storage) GetSnapshotsInWindow(eventID string, window time.Duration) ([]models.Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshots, exists := s.snapshots[eventID]
	if !exists {
		return []models.Snapshot{}, nil
	}

	now := time.Now()
	var filtered []models.Snapshot
	for _, snapshot := range snapshots {
		if now.Sub(snapshot.Timestamp) <= window {
			filtered = append(filtered, snapshot)
		}
	}

	return filtered, nil
}

// AddChange adds a detected change
func (s *Storage) AddChange(change *models.Change) error {
	if err := change.Validate(); err != nil {
		return fmt.Errorf("invalid change: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.changes = append(s.changes, *change)
	return nil
}

// GetTopChanges returns the top K changes sorted by magnitude
func (s *Storage) GetTopChanges(k int) ([]models.Change, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Sort changes by magnitude descending
	sorted := make([]models.Change, len(s.changes))
	copy(sorted, s.changes)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Magnitude > sorted[j].Magnitude
	})

	// Return top K
	if k > len(sorted) {
		k = len(sorted)
	}
	return sorted[:k], nil
}

// ClearChanges removes all stored changes
func (s *Storage) ClearChanges() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.changes = make([]models.Change, 0)
	return nil
}

// Save persists storage state to file
func (s *Storage) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create data directory if needed
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Prepare persistence file
	data := PersistenceFile{
		Version:   "1.0",
		SavedAt:   time.Now(),
		Events:    s.events,
		Snapshots: s.snapshots,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Write to temporary file first (atomic write)
	tempPath := s.filePath + ".tmp"
	if err := os.WriteFile(tempPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Rename temp file to actual file
	if err := os.Rename(tempPath, s.filePath); err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// Load restores storage state from file
func (s *Storage) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if file exists
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		// No file to load, start fresh
		return nil
	}

	// Read file
	jsonData, err := os.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal
	var data PersistenceFile
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// Restore state
	s.events = data.Events
	s.snapshots = data.Snapshots

	return nil
}

// RotateSnapshots removes old snapshots exceeding max limit
func (s *Storage) RotateSnapshots() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for eventID, snapshots := range s.snapshots {
		if len(snapshots) > s.maxSnapshotsPerEvent {
			// Keep only the most recent snapshots
			start := len(snapshots) - s.maxSnapshotsPerEvent
			s.snapshots[eventID] = snapshots[start:]
		}
	}

	return nil
}

// RotateEvents removes events when exceeding max limit
func (s *Storage) RotateEvents() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.events) <= s.maxEvents {
		return nil
	}

	// Find oldest events to remove
	type eventWithTime struct {
		id          string
		lastUpdated time.Time
	}

	var eventList []eventWithTime
	for id, event := range s.events {
		eventList = append(eventList, eventWithTime{id: id, lastUpdated: event.LastUpdated})
	}

	// Sort by last updated (oldest first)
	sort.Slice(eventList, func(i, j int) bool {
		return eventList[i].lastUpdated.Before(eventList[j].lastUpdated)
	})

	// Remove oldest events
	toRemove := len(s.events) - s.maxEvents
	for i := 0; i < toRemove; i++ {
		eventID := eventList[i].id
		delete(s.events, eventID)
		delete(s.snapshots, eventID)
	}

	return nil
}
