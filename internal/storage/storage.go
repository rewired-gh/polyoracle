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
	"strings"
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
	filePermissions      os.FileMode
	dirPermissions       os.FileMode
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
func New(maxEvents, maxSnapshotsPerEvent int, filePath string, filePermissions, dirPermissions os.FileMode) *Storage {
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
		filePermissions:      filePermissions,
		dirPermissions:       dirPermissions,
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

	// Sort by timestamp ascending (oldest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.Before(filtered[j].Timestamp)
	})

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
	if err := os.MkdirAll(dir, s.dirPermissions); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Prepare persistence file
	data := PersistenceFile{
		Version:   "2.0",
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
	if err := os.WriteFile(tempPath, jsonData, s.filePermissions); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Rename temp file to actual file
	if err := os.Rename(tempPath, s.filePath); err != nil {
		_ = os.Remove(tempPath) // Clean up temp file on rename failure
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// Load restores storage state from file
func (s *Storage) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clean up any stale temp files from previous crashes
	tempPath := s.filePath + ".tmp"
	if _, err := os.Stat(tempPath); err == nil {
		_ = os.Remove(tempPath)
	}

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

	// Clear and restore state
	s.events = data.Events
	if s.events == nil {
		s.events = make(map[string]*models.Event)
	}

	s.snapshots = data.Snapshots
	if s.snapshots == nil {
		s.snapshots = make(map[string][]models.Snapshot)
	}

	// Changes are transient, clear them
	s.changes = make([]models.Change, 0)

	// Migrate if needed (version < 2.0)
	if data.Version == "" || data.Version == "1.0" {
		s.migrateToCompositeIDs()
	}

	return nil
}

// migrateToCompositeIDs migrates v1.0 data format to v2.0 composite ID format
// In v1.0, single-market events used just EventID, while multi-market used EventID:MarketID
// In v2.0, all events use EventID:MarketID format for consistency
func (s *Storage) migrateToCompositeIDs() {
	newEvents := make(map[string]*models.Event)
	newSnapshots := make(map[string][]models.Snapshot)

	for id, event := range s.events {
		// Check if ID needs migration (doesn't contain ":" and has MarketID)
		if !strings.Contains(id, ":") && event.MarketID != "" {
			// Migrate to composite ID format
			newID := event.EventID + ":" + event.MarketID
			event.ID = newID
			newEvents[newID] = event

			// Migrate associated snapshots
			if snaps, exists := s.snapshots[id]; exists {
				for i := range snaps {
					snaps[i].EventID = newID
				}
				newSnapshots[newID] = snaps
			}
		} else {
			// Already in correct format or no MarketID
			newEvents[id] = event
			if snaps, exists := s.snapshots[id]; exists {
				newSnapshots[id] = snaps
			}
		}
	}

	s.events = newEvents
	s.snapshots = newSnapshots
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
