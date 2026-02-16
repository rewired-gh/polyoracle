package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// PolymarketAPI types for fetching data from Gamma API
type PolymarketEvent struct {
	ID              string             `json:"id"`
	Slug            string             `json:"slug"`
	Title           string             `json:"title"`
	Description     string             `json:"description"`
	Category        string             `json:"category"`
	Volume24hr      float64            `json:"volume24hr"`
	Volume1wk       float64            `json:"volume1wk"`
	Volume1mo       float64            `json:"volume1mo"`
	Liquidity       float64            `json:"liquidity"`
	Active          bool               `json:"active"`
	Closed          bool               `json:"closed"`
	Markets         []PolymarketMarket `json:"markets"`
	Tags            []Tag              `json:"tags"`
	Image           string             `json:"image"`
	EnableOrderBook bool               `json:"enableOrderBook"`
	New             bool               `json:"new"`
	Featured        bool               `json:"featured"`
}

// Tag represents a category tag
type Tag struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Slug  string `json:"slug"`
}

// PolymarketMarket represents a market within an event
type PolymarketMarket struct {
	ID            string      `json:"id"`
	Slug          string      `json:"slug"`
	Question      string      `json:"question"`
	OutcomePrices string      `json:"outcomePrices"` // JSON string like "[\"0.45\",\"0.55\"]"
	Outcome       interface{} `json:"outcome"`       // Can be null, string, or array
	Volume        string      `json:"volume"`
	Active        bool        `json:"active"`
	Closed        bool        `json:"closed"`
}

// fetchEventsByCategory fetches events from Polymarket Gamma API
func fetchEventsByCategory(category string, limit int) []PolymarketEvent {
	baseURL := "https://gamma-api.polymarket.com/events"
	params := url.Values{}

	if category != "" {
		params.Set("tag", category)
	}
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("active", "true")
	params.Set("closed", "false")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Printf("Error fetching events: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return nil
	}

	var events []PolymarketEvent
	if err := json.Unmarshal(body, &events); err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		return nil
	}

	return events
}
