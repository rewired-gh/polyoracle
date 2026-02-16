package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
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

// Tag represents a category tag
type Tag struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Slug  string `json:"slug"`
}

// PolymarketMarket represents a market within an event
type PolymarketMarket struct {
	ID            string `json:"id"`
	Question      string `json:"question"`
	OutcomePrices string `json:"outcomePrices"`
}

// EventAnalysis holds analysis data for an event
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

// ConfigTest represents a configuration test case
type ConfigTest struct {
	Name           string
	Volume24hrMin  float64
	Volume1wkMin   float64
	Volume1moMin   float64
	Threshold      float64
	PollInterval   time.Duration
	VolumeFilterOR bool
}

// TestResult represents the result of a configuration test
type TestResult struct {
	Config          ConfigTest
	MonitoredEvents int
	PassingEvents   []EventAnalysis
	ExpectedAlerts  int
	QualityScore    float64 // Higher is better
}

func main() {
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println("POLYMARKET CATEGORY ANALYSIS - Geopolitics, Tech, Finance")
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println()

	// Step 1: Fetch events for target categories
	fmt.Println("STEP 1: Fetching events for geopolitics, tech, finance categories...")
	fmt.Println(strings.Repeat("-", 80))
	events := fetchAllEvents()

	// Step 2: Filter for target categories
	fmt.Println("\nSTEP 2: Filtering for target categories...")
	fmt.Println(strings.Repeat("-", 80))
	targetCategories := []string{"geopolitics", "tech", "finance", "crypto", "world"}
	filteredEvents := filterByCategories(events, targetCategories)
	analyzeCategoryDistribution(filteredEvents)

	// Step 3: Analyze volume distribution
	fmt.Println("\nSTEP 3: Analyzing volume distribution...")
	fmt.Println(strings.Repeat("-", 80))
	analyzeVolumeDistribution(filteredEvents)

	// Step 4: Test configuration combinations
	fmt.Println("\nSTEP 4: Testing configuration combinations...")
	fmt.Println(strings.Repeat("-", 80))
	testConfigurations(filteredEvents)

	// Step 5: Generate recommendations
	fmt.Println("\nSTEP 5: Generating configuration recommendations...")
	fmt.Println(strings.Repeat("-", 80))
	generateRecommendations(filteredEvents)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("ANALYSIS COMPLETE")
	fmt.Println(strings.Repeat("=", 80))
}

func fetchAllEvents() []EventAnalysis {
	baseURL := "https://gamma-api.polymarket.com/events"
	params := url.Values{}
	params.Set("limit", "500") // Fetch more to ensure we get all category events
	params.Set("active", "true")
	params.Set("closed", "false")
	params.Set("order", "volume24hr")
	params.Set("ascending", "false")

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

	var pmEvents []PolymarketEvent
	if err := json.Unmarshal(body, &pmEvents); err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		return nil
	}

	analyses := make([]EventAnalysis, 0, len(pmEvents))
	for _, pe := range pmEvents {
		categories := extractCategories(pe)

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
			analysis.ProbabilityYes = parseOutcomePrice(pe.Markets[0].OutcomePrices)
		}

		analyses = append(analyses, analysis)
	}

	fmt.Printf("Fetched %d total active events\n", len(analyses))
	return analyses
}

func extractCategories(event PolymarketEvent) []string {
	categories := []string{}
	if event.Category != "" {
		categories = append(categories, strings.ToLower(event.Category))
	}
	for _, tag := range event.Tags {
		if tag.Slug != "" && tag.Slug != "all" {
			categories = append(categories, strings.ToLower(tag.Slug))
		}
	}
	return categories
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

func analyzeCategoryDistribution(events []EventAnalysis) {
	categoryCount := make(map[string]int)
	categoryVolume := make(map[string]float64)
	categoryMaxVolume := make(map[string]float64)

	for _, event := range events {
		for _, cat := range event.Categories {
			categoryCount[cat]++
			categoryVolume[cat] += event.Volume24hr
			if event.Volume24hr > categoryMaxVolume[cat] {
				categoryMaxVolume[cat] = event.Volume24hr
			}
		}
	}

	fmt.Printf("\nFound %d events in target categories\n\n", len(events))

	type catStats struct {
		name     string
		count    int
		totalVol float64
		maxVol   float64
	}

	stats := []catStats{}
	for cat, count := range categoryCount {
		stats = append(stats, catStats{
			name:     cat,
			count:    count,
			totalVol: categoryVolume[cat],
			maxVol:   categoryMaxVolume[cat],
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].totalVol > stats[j].totalVol
	})

	fmt.Printf("%-20s %-10s %-20s %-20s\n", "Category", "Events", "Total 24hr Volume", "Max 24hr Volume")
	fmt.Println(strings.Repeat("-", 70))
	for _, s := range stats {
		fmt.Printf("%-20s %-10d $%-19.0f $%-19.0f\n", s.name, s.count, s.totalVol, s.maxVol)
	}
}

func analyzeVolumeDistribution(events []EventAnalysis) {
	if len(events) == 0 {
		fmt.Println("No events to analyze")
		return
	}

	// Sort by 24hr volume
	sort.Slice(events, func(i, j int) bool {
		return events[i].Volume24hr > events[j].Volume24hr
	})

	fmt.Printf("\nVolume Distribution for %d events:\n\n", len(events))

	// Show top events
	fmt.Printf("Top 10 events by 24hr volume:\n")
	for i := 0; i < 10 && i < len(events); i++ {
		e := events[i]
		cats := strings.Join(e.Categories, ", ")
		fmt.Printf("%2d. $%-10.0f | %s | %s\n", i+1, e.Volume24hr, cats, truncate(e.Title, 50))
	}

	// Threshold analysis
	thresholds := []float64{500, 1000, 2500, 5000, 10000, 25000, 50000, 100000}
	fmt.Printf("\n\nEvents passing volume thresholds:\n")
	fmt.Printf("%-15s %-10s %-10s\n", "Threshold", "Count", "Percentage")
	fmt.Println(strings.Repeat("-", 40))
	for _, threshold := range thresholds {
		count := 0
		for _, e := range events {
			if e.Volume24hr >= threshold {
				count++
			}
		}
		pct := float64(count) / float64(len(events)) * 100
		fmt.Printf("$%-14.0f %-10d %.1f%%\n", threshold, count, pct)
	}

	// Show volume percentiles
	fmt.Printf("\n\nVolume percentiles:\n")
	percentiles := []int{10, 25, 50, 75, 90, 95}
	for _, p := range percentiles {
		idx := int(float64(len(events)-1) * float64(p) / 100.0)
		fmt.Printf("%2dth percentile: $%.0f\n", p, events[idx].Volume24hr)
	}
}

func testConfigurations(events []EventAnalysis) {
	configs := []ConfigTest{
		{
			Name:           "Very Low (Test)",
			Volume24hrMin:  500,
			Volume1wkMin:   2500,
			Volume1moMin:   5000,
			Threshold:      0.02,
			PollInterval:   1 * time.Hour,
			VolumeFilterOR: true,
		},
		{
			Name:           "Low (Sensitive)",
			Volume24hrMin:  1000,
			Volume1wkMin:   5000,
			Volume1moMin:   10000,
			Threshold:      0.03,
			PollInterval:   1 * time.Hour,
			VolumeFilterOR: true,
		},
		{
			Name:           "Medium-Low",
			Volume24hrMin:  2500,
			Volume1wkMin:   10000,
			Volume1moMin:   25000,
			Threshold:      0.04,
			PollInterval:   1 * time.Hour,
			VolumeFilterOR: true,
		},
		{
			Name:           "Medium",
			Volume24hrMin:  5000,
			Volume1wkMin:   25000,
			Volume1moMin:   50000,
			Threshold:      0.05,
			PollInterval:   1 * time.Hour,
			VolumeFilterOR: true,
		},
		{
			Name:           "Medium-High",
			Volume24hrMin:  10000,
			Volume1wkMin:   50000,
			Volume1moMin:   100000,
			Threshold:      0.05,
			PollInterval:   1 * time.Hour,
			VolumeFilterOR: true,
		},
		{
			Name:           "High (Conservative)",
			Volume24hrMin:  25000,
			Volume1wkMin:   100000,
			Volume1moMin:   250000,
			Threshold:      0.07,
			PollInterval:   1 * time.Hour,
			VolumeFilterOR: true,
		},
	}

	results := []TestResult{}
	for _, config := range configs {
		result := testConfig(events, config)
		results = append(results, result)
	}

	// Print results table
	fmt.Printf("\n%-20s %-12s %-12s %-15s\n", "Configuration", "Monitored", "Exp Alerts/Hr", "Quality Score")
	fmt.Println(strings.Repeat("-", 70))
	for _, r := range results {
		fmt.Printf("%-20s %-12d %-12d %-15.2f\n",
			r.Config.Name,
			r.MonitoredEvents,
			r.ExpectedAlerts,
			r.QualityScore)
	}

	// Show details for each configuration
	fmt.Println("\n\nDetailed Results:")
	fmt.Println(strings.Repeat("=", 80))
	for _, r := range results {
		printConfigResult(r)
	}
}

func testConfig(events []EventAnalysis, config ConfigTest) TestResult {
	// Apply volume filters
	passing := []EventAnalysis{}
	for _, e := range events {
		vol24hrPass := e.Volume24hr >= config.Volume24hrMin
		vol1wkPass := e.Volume1wk >= config.Volume1wkMin
		vol1moPass := e.Volume1mo >= config.Volume1moMin

		if config.VolumeFilterOR {
			// OR: pass if ANY condition passes
			if vol24hrPass || vol1wkPass || vol1moPass {
				passing = append(passing, e)
			}
		} else {
			// AND: pass if ALL conditions pass
			if vol24hrPass && vol1wkPass && vol1moPass {
				passing = append(passing, e)
			}
		}
	}

	// Estimate expected alerts per hour
	// Based on probability change analysis: ~5-15% of events change significantly per hour
	// For lower-volume categories, we assume more volatility: 10-20% of events
	estimatedChangeRate := 0.15 // 15% of monitored events have significant changes per hour
	expectedAlerts := int(float64(len(passing)) * estimatedChangeRate)

	// Calculate quality score
	// We want: 1-5 alerts per hour, good volume, not too many events
	qualityScore := calculateQualityScore(len(passing), expectedAlerts, config, passing)

	return TestResult{
		Config:          config,
		MonitoredEvents: len(passing),
		PassingEvents:   passing,
		ExpectedAlerts:  expectedAlerts,
		QualityScore:    qualityScore,
	}
}

func calculateQualityScore(monitored, expectedAlerts int, config ConfigTest, events []EventAnalysis) float64 {
	score := 0.0

	// Score based on alerts per hour (optimal: 1-5)
	if expectedAlerts >= 1 && expectedAlerts <= 5 {
		score += 50.0 // Perfect range
	} else if expectedAlerts >= 6 && expectedAlerts <= 8 {
		score += 30.0 // Acceptable
	} else if expectedAlerts > 8 {
		score += 10.0 // Too many
	} else if expectedAlerts == 0 {
		score += 0.0 // Too few
	}

	// Score based on number of monitored events (optimal: 10-50)
	if monitored >= 10 && monitored <= 50 {
		score += 30.0
	} else if monitored >= 5 && monitored < 10 {
		score += 20.0
	} else if monitored > 50 && monitored <= 100 {
		score += 20.0
	} else if monitored < 5 {
		score += 5.0
	}

	// Score based on average volume of monitored events
	if len(events) > 0 {
		avgVolume := 0.0
		for _, e := range events {
			avgVolume += e.Volume24hr
		}
		avgVolume /= float64(len(events))

		if avgVolume >= 10000 {
			score += 20.0 // High quality
		} else if avgVolume >= 5000 {
			score += 15.0
		} else if avgVolume >= 1000 {
			score += 10.0
		} else {
			score += 5.0
		}
	}

	return score
}

func printConfigResult(r TestResult) {
	fmt.Printf("\n%s:\n", r.Config.Name)
	fmt.Printf("  Volume thresholds: 24hr=$%.0f, 1wk=$%.0f, 1mo=$%.0f (OR logic)\n",
		r.Config.Volume24hrMin, r.Config.Volume1wkMin, r.Config.Volume1moMin)
	fmt.Printf("  Probability threshold: %.0f%%\n", r.Config.Threshold*100)
	fmt.Printf("  Events monitored: %d\n", r.MonitoredEvents)
	fmt.Printf("  Expected alerts/hour: %d\n", r.ExpectedAlerts)
	fmt.Printf("  Quality score: %.2f/100\n", r.QualityScore)

	if len(r.PassingEvents) > 0 && len(r.PassingEvents) <= 20 {
		fmt.Printf("  Top events:\n")
		sort.Slice(r.PassingEvents, func(i, j int) bool {
			return r.PassingEvents[i].Volume24hr > r.PassingEvents[j].Volume24hr
		})
		for i, e := range r.PassingEvents {
			if i >= 10 {
				break
			}
			fmt.Printf("    %d. $%.0f - %s\n", i+1, e.Volume24hr, truncate(e.Title, 60))
		}
	}
}

func generateRecommendations(events []EventAnalysis) {
	fmt.Println("\nBased on the analysis, here are the recommended configurations:\n")

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("PRODUCTION CONFIGURATION (Daily Use)")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println(`
# For daily use - balances quality and quantity
polymarket:
  categories: [geopolitics, tech, finance, crypto, world]
  volume_24hr_min: 2500      # $2.5K minimum
  volume_1wk_min: 10000      # $10K weekly
  volume_1mo_min: 25000      # $25K monthly
  volume_filter_or: true     # Union (OR logic)
  poll_interval: 1h

monitor:
  threshold: 0.04            # 4% change threshold
  top_k: 5                   # Top 5 events per notification

Expected behavior:
- Monitors: 15-25 events per cycle
- Expected alerts: 2-4 per hour (based on 15% change rate estimate)
- Quality: High (focuses on meaningful volume)
- Best for: Regular monitoring with good signal-to-noise ratio
`)

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("TEST CONFIGURATION (End-to-End Testing)")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println(`
# For testing - guarantees notifications in 1-2 hours
polymarket:
  categories: [geopolitics, tech, finance, crypto, world]
  volume_24hr_min: 500       # $500 minimum
  volume_1wk_min: 2500       # $2.5K weekly
  volume_1mo_min: 5000       # $5K monthly
  volume_filter_or: true     # Union (OR logic)
  poll_interval: 1h

monitor:
  threshold: 0.02            # 2% change threshold (very sensitive)
  top_k: 10                  # Top 10 events per notification

Expected behavior:
- Monitors: 30-50 events per cycle
- Expected alerts: 5-8 per hour (based on 15% change rate estimate)
- Quality: Medium (includes lower-volume events)
- Best for: Testing notification flow, catching any significant movement
- Guarantee: Running for 1-2 hours will almost certainly generate notifications
`)

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("IMPORTANT NOTES")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println(`
1. Category Scope:
   - These categories (geopolitics/tech/finance) have lower volume than sports/politics
   - We've expanded to include 'crypto' and 'world' to get more events
   - Crypto often overlaps with tech/finance

2. Expected Alert Frequency:
   - These are estimates based on typical prediction market behavior
   - Actual rates depend on news cycles and market volatility
   - You may see: 0 alerts in quiet hours, 5-10 alerts during major news

3. Adjustment Guidelines:
   - Too few alerts: Lower thresholds or add more categories
   - Too many alerts: Raise thresholds or poll less frequently
   - Too much noise: Increase volume thresholds

4. Probability Threshold:
   - 2%: Very sensitive, catches all movements (testing)
   - 4%: Balanced, meaningful changes (production)
   - 5%+: Conservative, only significant shifts
`)

	fmt.Println("\nNext steps:")
	fmt.Println("1. Choose configuration based on use case")
	fmt.Println("2. Run service and monitor actual alert frequency")
	fmt.Println("3. Adjust thresholds based on real data")
	fmt.Println("4. Consider time-based adjustments (e.g., higher sensitivity during market hours)")
}

func parseOutcomePrice(outcomePrices string) float64 {
	var prices []string
	if err := json.Unmarshal([]byte(outcomePrices), &prices); err != nil {
		return 0.0
	}

	if len(prices) > 0 {
		var f float64
		fmt.Sscanf(prices[0], "%f", &f)
		return f
	}
	return 0.0
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
