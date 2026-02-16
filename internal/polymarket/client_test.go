package polymarket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchEvents(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events" {
			t.Errorf("Expected /events, got %s", r.URL.Path)
		}

		response := struct {
			Events []PolymarketEvent `json:"events"`
		}{
			Events: []PolymarketEvent{
				{
					ID:          "event-1",
					Question:    "Will X happen?",
					Category:    "politics",
					Active:      true,
					Description: "Test event",
					Markets: []Market{
						{
							ID:            "market-1",
							EventID:       "event-1",
							Outcome:       "Yes",
							OutcomePrices: []string{"0.75"},
						},
						{
							ID:            "market-2",
							EventID:       "event-1",
							Outcome:       "No",
							OutcomePrices: []string{"0.25"},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, 30*time.Second)

	// Fetch events
	events, err := client.FetchEvents(context.Background(), []string{"politics"})
	if err != nil {
		t.Fatalf("FetchEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].ID != "event-1" {
		t.Errorf("Expected ID event-1, got %s", events[0].ID)
	}

	if events[0].YesProbability != 0.75 {
		t.Errorf("Expected yes probability 0.75, got %f", events[0].YesProbability)
	}
}

func TestFetchEventsCategoryFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := struct {
			Events []PolymarketEvent `json:"events"`
		}{
			Events: []PolymarketEvent{
				{
					ID:       "event-1",
					Question: "Politics question?",
					Category: "politics",
					Active:   true,
					Markets:  []Market{},
				},
				{
					ID:       "event-2",
					Question: "Sports question?",
					Category: "sports",
					Active:   true,
					Markets:  []Market{},
				},
			},
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, 30*time.Second)

	// Fetch only politics
	events, err := client.FetchEvents(context.Background(), []string{"politics"})
	if err != nil {
		t.Fatalf("FetchEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].Category != "politics" {
		t.Errorf("Expected category politics, got %s", events[0].Category)
	}
}
