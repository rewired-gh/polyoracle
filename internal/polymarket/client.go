package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/poly-oracle/internal/models"
)

// Client provides access to Polymarket API
type Client struct {
	apiBaseURL string
	httpClient *http.Client
	timeout    time.Duration
}

// PolymarketEvent represents an event from Polymarket API
type PolymarketEvent struct {
	ID          string   `json:"id"`
	Question    string   `json:"question"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Active      bool     `json:"active"`
	Markets     []Market `json:"markets"`
}

// Market represents a market from Polymarket API
type Market struct {
	ID            string   `json:"id"`
	EventID       string   `json:"event_id"`
	Outcome       string   `json:"outcome"`
	OutcomePrices []string `json:"outcome_prices"`
}

// NewClient creates a new Polymarket client
func NewClient(apiBaseURL string, timeout time.Duration) *Client {
	return &Client{
		apiBaseURL: apiBaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// FetchEvents retrieves events from Polymarket Gamma API
func (c *Client) FetchEvents(ctx context.Context, categories []string) ([]models.Event, error) {
	url := fmt.Sprintf("%s/events?active=true&limit=100", c.apiBaseURL)

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		Events []PolymarketEvent `json:"events"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode events: %w", err)
	}

	// Filter by categories and convert to internal models
	var events []models.Event
	categoryMap := make(map[string]bool)
	for _, cat := range categories {
		categoryMap[cat] = true
	}

	for _, pe := range response.Events {
		// Filter by category if specified
		if len(categories) > 0 && !categoryMap[pe.Category] {
			continue
		}

		// Extract probabilities from markets
		var yesProb, noProb float64
		for _, market := range pe.Markets {
			if market.Outcome == "Yes" && len(market.OutcomePrices) > 0 {
				fmt.Sscanf(market.OutcomePrices[0], "%f", &yesProb)
			} else if market.Outcome == "No" && len(market.OutcomePrices) > 0 {
				fmt.Sscanf(market.OutcomePrices[0], "%f", &noProb)
			}
		}

		event := models.Event{
			ID:             pe.ID,
			Question:       pe.Question,
			Description:    pe.Description,
			Category:       pe.Category,
			YesProbability: yesProb,
			NoProbability:  noProb,
			Active:         pe.Active,
			LastUpdated:    time.Now(),
			CreatedAt:      time.Now(),
		}

		events = append(events, event)
	}

	return events, nil
}

// FetchMarketData retrieves market data for an event
func (c *Client) FetchMarketData(ctx context.Context, eventID string) ([]Market, error) {
	url := fmt.Sprintf("%s/events/%s", c.apiBaseURL, eventID)

	resp, err := c.doRequest(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch market data: %w", err)
	}
	defer resp.Body.Close()

	var pe PolymarketEvent
	if err := json.NewDecoder(resp.Body).Decode(&pe); err != nil {
		return nil, fmt.Errorf("failed to decode market data: %w", err)
	}

	return pe.Markets, nil
}

// doRequest performs HTTP request with retry logic
func (c *Client) doRequest(ctx context.Context, url string) (*http.Response, error) {
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		if resp.StatusCode >= 500 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
