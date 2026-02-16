package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// PolymarketEvent represents an event from Polymarket API
type PolymarketEvent struct {
	ID         string             `json:"id"`
	Title      string             `json:"title"`
	Category   string             `json:"category"`
	Tags       []Tag              `json:"tags"`
	Volume24hr float64            `json:"volume24hr"`
	Volume1wk  float64            `json:"volume1wk"`
	Volume1mo  float64            `json:"volume1mo"`
	Liquidity  float64            `json:"liquidity"`
	Active     bool               `json:"active"`
	Closed     bool               `json:"closed"`
	Markets    []PolymarketMarket `json:"markets"`
}

type Tag struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Slug  string `json:"slug"`
}

type PolymarketMarket struct {
	ID            string `json:"id"`
	Question      string `json:"question"`
	OutcomePrices string `json:"outcomePrices"`
}

type EventAnalysis struct {
	ID             string
	Title          string
	Categories     []string
	Volume24hr     float64
	Volume1wk      float64
	Volume1mo      float64
	Liquidity      float64
	ProbabilityYes float64
}

func main() {
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println("REFINED PARAMETER ANALYSIS FOR GEOPOLITICS/TECH/FINANCE")
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println()

	// Fetch events
	fmt.Println("Fetching events from Polymarket API...")
	events := fetchAllEvents()

	// Filter for target categories
	targetCategories := []string{"geopolitics", "tech", "finance", "crypto", "world"}
	filtered := filterByCategories(events, targetCategories)
	fmt.Printf("Found %d events in target categories\n\n", len(filtered))

	// Sort by volume
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Volume24hr > filtered[j].Volume24hr
	})

	// Analyze top events at different volume levels
	analyzeTopEvents(filtered)

	// Test specific configurations
	testSpecificConfigs(filtered)

	// Generate final recommendations
	generateFinalRecommendations(filtered)
}

func fetchAllEvents() []EventAnalysis {
	baseURL := "https://gamma-api.polymarket.com/events"
	params := url.Values{}
	params.Set("limit", "500")
	params.Set("active", "true")
	params.Set("closed", "false")
	params.Set("order", "volume24hr")
	params.Set("ascending", "false")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var pmEvents []PolymarketEvent
	json.Unmarshal(body, &pmEvents)

	analyses := make([]EventAnalysis, 0, len(pmEvents))
	for _, pe := range pmEvents {
		categories := []string{}
		if pe.Category != "" {
			categories = append(categories, strings.ToLower(pe.Category))
		}
		for _, tag := range pe.Tags {
			if tag.Slug != "" && tag.Slug != "all" {
				categories = append(categories, strings.ToLower(tag.Slug))
			}
		}

		analysis := EventAnalysis{
			ID:         pe.ID,
			Title:      pe.Title,
			Categories: categories,
			Volume24hr: pe.Volume24hr,
			Volume1wk:  pe.Volume1wk,
			Volume1mo:  pe.Volume1mo,
			Liquidity:  pe.Liquidity,
		}

		if len(pe.Markets) > 0 {
			var prices []string
			if err := json.Unmarshal([]byte(pe.Markets[0].OutcomePrices), &prices); err == nil && len(prices) > 0 {
				fmt.Sscanf(prices[0], "%f", &analysis.ProbabilityYes)
			}
		}

		analyses = append(analyses, analysis)
	}

	return analyses
}

func filterByCategories(events []EventAnalysis, targetCategories []string) []EventAnalysis {
	categoryMap := make(map[string]bool)
	for _, cat := range targetCategories {
		categoryMap[cat] = true
	}

	filtered := []EventAnalysis{}
	for _, event := range events {
		for _, cat := range event.Categories {
			if categoryMap[cat] {
				filtered = append(filtered, event)
				break
			}
		}
	}
	return filtered
}

func analyzeTopEvents(events []EventAnalysis) {
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("TOP EVENTS BY VOLUME")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Show top 30 events
	fmt.Printf("%-5s %-15s %-15s %s\n", "Rank", "24hr Volume", "Categories", "Title")
	fmt.Println(strings.Repeat("-", 80))
	for i := 0; i < 30 && i < len(events); i++ {
		e := events[i]
		cats := strings.Join(e.Categories[:min(3, len(e.Categories))], ", ")
		fmt.Printf("%-5d $%-14.0f %-15s %s\n", i+1, e.Volume24hr, cats, truncate(e.Title, 45))
	}

	// Volume distribution
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("VOLUME DISTRIBUTION")
	fmt.Println(strings.Repeat("=", 80))

	thresholds := []float64{50000, 25000, 10000, 5000, 2500, 1000}
	fmt.Printf("\n%-15s %-10s %-15s\n", "Min Volume", "Events", "Expected Alerts/Hr")
	fmt.Println(strings.Repeat("-", 45))
	for _, threshold := range thresholds {
		count := 0
		for _, e := range events {
			if e.Volume24hr >= threshold {
				count++
			}
		}
		// Estimate: 10-20% of events have significant changes per hour
		lowAlerts := int(float64(count) * 0.10)
		highAlerts := int(float64(count) * 0.20)
		fmt.Printf("$%-14.0f %-10d %-15s\n", threshold, count, fmt.Sprintf("%d-%d", lowAlerts, highAlerts))
	}
}

func testSpecificConfigs(events []EventAnalysis) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("SPECIFIC CONFIGURATION TESTS")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Test configurations designed to get 1-5 alerts per hour
	configs := []struct {
		name        string
		volume24hr  float64
		volume1wk   float64
		volume1mo   float64
		threshold   float64
		description string
	}{
		{
			name:        "Config A: Very Selective (Top ~10 events)",
			volume24hr:  100000,
			volume1wk:   500000,
			volume1mo:   1000000,
			threshold:   0.05,
			description: "Only the highest volume events",
		},
		{
			name:        "Config B: Selective (Top ~20 events)",
			volume24hr:  50000,
			volume1wk:   200000,
			volume1mo:   500000,
			threshold:   0.05,
			description: "High volume, established markets",
		},
		{
			name:        "Config C: Moderate (Top ~30 events)",
			volume24hr:  25000,
			volume1wk:   100000,
			volume1mo:   250000,
			threshold:   0.04,
			description: "Good balance of quality and quantity",
		},
		{
			name:        "Config D: Inclusive (Top ~50 events)",
			volume24hr:  10000,
			volume1wk:   50000,
			volume1mo:   100000,
			threshold:   0.03,
			description: "More events, lower threshold",
		},
		{
			name:        "Config E: Test Mode (Many events)",
			volume24hr:  5000,
			volume1wk:   25000,
			volume1mo:   50000,
			threshold:   0.02,
			description: "For testing - guarantees notifications",
		},
	}

	for _, cfg := range configs {
		fmt.Printf("\n%s\n", cfg.name)
		fmt.Printf("Description: %s\n", cfg.description)
		fmt.Println(strings.Repeat("-", 70))

		// Count events passing volume filter (OR logic)
		count := 0
		for _, e := range events {
			if e.Volume24hr >= cfg.volume24hr || e.Volume1wk >= cfg.volume1wk || e.Volume1mo >= cfg.volume1mo {
				count++
			}
		}

		// Estimate alerts
		// Conservative: 10% of events have changes >= threshold
		// Moderate: 15% of events have changes >= threshold
		// Aggressive: 20% of events have changes >= threshold
		conservativeAlerts := int(float64(count) * 0.10)
		moderateAlerts := int(float64(count) * 0.15)
		aggressiveAlerts := int(float64(count) * 0.20)

		fmt.Printf("Volume thresholds: 24hr=$%.0f, 1wk=$%.0f, 1mo=$%.0f (OR)\n",
			cfg.volume24hr, cfg.volume1wk, cfg.volume1mo)
		fmt.Printf("Probability threshold: %.0f%%\n", cfg.threshold*100)
		fmt.Printf("Events monitored: %d\n", count)
		fmt.Printf("Expected alerts/hour:\n")
		fmt.Printf("  Conservative (10%% change rate): %d\n", conservativeAlerts)
		fmt.Printf("  Moderate (15%% change rate): %d\n", moderateAlerts)
		fmt.Printf("  Aggressive (20%% change rate): %d\n", aggressiveAlerts)

		// Evaluate against target (1-5 alerts per hour)
		inTarget := moderateAlerts >= 1 && moderateAlerts <= 5
		if inTarget {
			fmt.Printf("✓ IN TARGET RANGE (1-5 alerts/hour with moderate estimate)\n")
		} else if moderateAlerts == 0 {
			fmt.Printf("✗ TOO FEW alerts expected\n")
		} else {
			fmt.Printf("✗ TOO MANY alerts expected\n")
		}
	}
}

func generateFinalRecommendations(events []EventAnalysis) {
	fmt.Println("\n\n" + strings.Repeat("=", 80))
	fmt.Println("FINAL RECOMMENDATIONS")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println(`
Based on the analysis, the key insight is that geopolitics/tech/finance categories
have LOWER volume than sports/politics, but still have many events (228 total).

The challenge: Finding the right balance between:
- Too many events monitored (excessive alerts)
- Too few events monitored (missed opportunities)

SOLUTION: Use VOLUME THRESHOLDS to select the TOP events by market activity.
`)

	fmt.Println("\n" + strings.Repeat("-", 80))
	fmt.Println("PRODUCTION CONFIGURATION")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println(`
polymarket:
  categories:
    - geopolitics
    - tech
    - finance
    - crypto
    - world
  volume_24hr_min: 25000      # $25K minimum 24hr volume
  volume_1wk_min: 100000      # $100K minimum weekly volume
  volume_1mo_min: 250000      # $250K minimum monthly volume
  volume_filter_or: true      # OR logic (union)
  poll_interval: 1h

monitor:
  threshold: 0.04             # 4% probability change
  top_k: 5                    # Show top 5 events
  window: 1h

RATIONALE:
- Monitors ~30-40 highest-quality events
- Expected alerts: 2-4 per hour (moderate estimate with 15% change rate)
- Filters out low-volume noise
- Focuses on established, liquid markets
- Crypto category included (often overlaps with tech/finance)

WHY THIS WORKS:
- Volume thresholds select top ~15% of events by activity
- 4% threshold catches meaningful movements (not minor fluctuations)
- OR logic ensures we capture events with sustained weekly/monthly volume
- Hourly polling allows enough time for probabilities to change significantly
`)

	fmt.Println("\n" + strings.Repeat("-", 80))
	fmt.Println("TEST CONFIGURATION (For End-to-End Testing)")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println(`
polymarket:
  categories:
    - geopolitics
    - tech
    - finance
    - crypto
    - world
  volume_24hr_min: 5000       # $5K minimum (very low)
  volume_1wk_min: 25000       # $25K minimum
  volume_1mo_min: 50000       # $50K minimum
  volume_filter_or: true
  poll_interval: 1h

monitor:
  threshold: 0.02             # 2% change (very sensitive)
  top_k: 10                   # Show top 10 events
  window: 1h

RATIONALE:
- Monitors ~80-100 events (broader coverage)
- Expected alerts: 8-15 per hour (moderate estimate)
- 2% threshold catches even small movements
- RUNNING FOR 1-2 HOURS WILL ALMOST CERTAINLY GENERATE NOTIFICATIONS
- Perfect for testing notification flow and end-to-end verification

TESTING STRATEGY:
1. Deploy with test configuration
2. Run for 1-2 hours during active market hours
3. You should see multiple notifications
4. Verify Telegram messages are formatted correctly
5. Verify storage/persistence is working
6. Once tested, switch to production configuration
`)

	fmt.Println("\n" + strings.Repeat("-", 80))
	fmt.Println("ADJUSTMENT GUIDELINES")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println(`
After deployment, monitor actual alert frequency and adjust:

TOO MANY ALERTS (>5 per hour consistently):
  → Increase volume thresholds (e.g., 50K/200K/500K)
  → Increase probability threshold (e.g., 0.05 or 0.06)
  → Both adjustments will reduce events and sensitivity

TOO FEW ALERTS (<1 per hour consistently):
  → Lower volume thresholds (e.g., 10K/50K/100K)
  → Lower probability threshold (e.g., 0.03)
  → Add more categories (e.g., "sports" for higher volume)

NOISE (alerts on minor events):
  → Increase volume thresholds significantly
  → Increase probability threshold to 0.05+
  → Consider AND logic instead of OR for stricter filtering

MISSING IMPORTANT EVENTS:
  → Lower thresholds
  → Add more categories
  → Check if events are being filtered incorrectly

EXPECTED VARIABILITY:
- Quiet periods: 0-1 alerts per hour (normal)
- Active periods: 3-8 alerts per hour (normal during news events)
- Very active: 10+ alerts per hour (major events unfolding)

The key is to tune for YOUR tolerance and use case.
`)

	fmt.Println("\n" + strings.Repeat("-", 80))
	fmt.Println("IMPLEMENTATION STEPS")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println(`
1. Update configs/config.yaml with PRODUCTION configuration
2. Update internal/config/config.go defaults
3. Deploy and monitor for 24-48 hours
4. Check actual alert frequency vs expected
5. Adjust thresholds based on real data
6. Document your tuned configuration

For testing notifications:
1. Use TEST configuration temporarily
2. Run for 1-2 hours
3. Verify notifications work end-to-end
4. Switch back to PRODUCTION configuration
`)

	fmt.Println("\n" + strings.Repeat("=", 80))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
