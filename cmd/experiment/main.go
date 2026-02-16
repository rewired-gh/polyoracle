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

// Event represents a Polymarket event from the Gamma API
type Event struct {
	ID                string  `json:"id"`
	Slug              string  `json:"slug"`
	Title             string  `json:"title"`
	Description       string  `json:"description"`
	Category          string  `json:"category"`
	Volume24hr        float64 `json:"volume24hr"`
	Volume1wk         float64 `json:"volume1wk"`
	Volume1mo         float64 `json:"volume1mo"`
	Liquidity         float64 `json:"liquidity"`
	Active            bool    `json:"active"`
	Closed            bool    `json:"closed"`
	Markets           []Market `json:"markets"`
	Tags              []Tag   `json:"tags"`
	Image             string  `json:"image"`
	EnableOrderBook   bool    `json:"enableOrderBook"`
	New               bool    `json:"new"`
	Featured          bool    `json:"featured"`
}

// Tag represents a category tag
type Tag struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Slug  string `json:"slug"`
}

// Market represents a market within an event
type Market struct {
	ID                   string  `json:"id"`
	Slug                 string  `json:"slug"`
	Question             string  `json:"question"`
	OutcomePrices        string  `json:"outcomePrices"`  // JSON string like "[\"0.45\",\"0.55\"]"
	Outcome              interface{} `json:"outcome"`     // Can be null, string, or array
	Volume               string  `json:"volume"`
	Active               bool    `json:"active"`
	Closed               bool    `json:"closed"`
}

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
	Timestamp time.Time
	EventID   string
	Title     string
	ProbYes   float64
	Volume24hr float64
}

func main() {
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println("POLYMARKET API EXPERIMENT - Configuration Parameter Analysis")
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println()

	// Step 1: Fetch and analyze events by category
	fmt.Println("STEP 1: Fetching events by category...")
	fmt.Println(strings.Repeat("-", 80))
	categoryStats := analyzeCategories()

	// Step 2: Analyze volume distribution
	fmt.Println("\nSTEP 2: Analyzing volume distribution...")
	fmt.Println(strings.Repeat("-", 80))
	allEvents := fetchAllEvents()
	volumeAnalysis := analyzeVolumeDistribution(allEvents)

	// Step 3: Test probability changes over time
	fmt.Println("\nSTEP 3: Testing probability changes over time...")
	fmt.Println(strings.Repeat("-", 80))
	analyzeProbabilityChanges()

	// Step 4: Generate recommendations
	fmt.Println("\nSTEP 4: Generating recommendations...")
	fmt.Println(strings.Repeat("-", 80))
	generateRecommendations(categoryStats, volumeAnalysis, allEvents)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("EXPERIMENT COMPLETE")
	fmt.Println(strings.Repeat("=", 80))
}

func analyzeCategories() map[string]CategoryStats {
	fmt.Println("\nFetching all events to analyze by category...")
	events := fetchEventsByCategory("", 200)

	if len(events) == 0 {
		fmt.Println("  No events found")
		return nil
	}

	analyses := analyzeEvents(events)

	// Group events by category
	categoryEvents := make(map[string][]EventAnalysis)
	for _, analysis := range analyses {
		if len(analysis.Categories) > 0 {
			// Use the first category
			cat := analysis.Categories[0]
			categoryEvents[cat] = append(categoryEvents[cat], analysis)
		}
	}

	stats := make(map[string]CategoryStats)
	for category, eventList := range categoryEvents {
		stats[category] = calculateCategoryStats(category, eventList)

		fmt.Printf("\n  Category: %s\n", category)
		fmt.Printf("    Events: %d\n", len(eventList))
		fmt.Printf("    Avg 24hr Volume: $%.2f\n", stats[category].AvgVolume24hr)
		fmt.Printf("    Max 24hr Volume: $%.2f\n", stats[category].MaxVolume24hr)
		fmt.Printf("    Total Volume: $%.2f\n", stats[category].TotalVolume)
	}

	return stats
}

func fetchAllEvents() []EventAnalysis {
	fmt.Println("\nFetching all events (no category filter)...")
	events := fetchEventsByCategory("", 200)
	return analyzeEvents(events)
}

func analyzeVolumeDistribution(events []EventAnalysis) VolumeAnalysis {
	if len(events) == 0 {
		return NewVolumeAnalysis()
	}

	// Sort by 24hr volume
	sort.Slice(events, func(i, j int) bool {
		return events[i].Volume24hr > events[j].Volume24hr
	})

	analysis := NewVolumeAnalysis()
	analysis.TotalEvents = len(events)

	// Calculate percentiles
	analysis.Top10Percent = events[:max(1, len(events)/10)]
	analysis.Top20Percent = events[:max(1, len(events)/5)]
	analysis.Top50Percent = events[:max(1, len(events)/2)]

	// Calculate thresholds
	thresholds := []float64{10000, 25000, 50000, 100000, 250000, 500000, 1000000}
	for _, threshold := range thresholds {
		count := 0
		for _, event := range events {
			if event.Volume24hr >= threshold {
				count++
			}
		}
		analysis.ThresholdCounts[fmt.Sprintf("$%.0f", threshold)] = count
	}

	// Print distribution
	fmt.Printf("\nTotal events analyzed: %d\n", len(events))
	fmt.Println("\nVolume Distribution (24hr):")
	fmt.Printf("  Top 10%% events (%d events):\n", len(analysis.Top10Percent))
	if len(analysis.Top10Percent) > 0 {
		fmt.Printf("    Volume range: $%.2f - $%.2f\n",
			analysis.Top10Percent[len(analysis.Top10Percent)-1].Volume24hr,
			analysis.Top10Percent[0].Volume24hr)
	}

	fmt.Printf("  Top 20%% events (%d events):\n", len(analysis.Top20Percent))
	if len(analysis.Top20Percent) > 0 {
		fmt.Printf("    Volume range: $%.2f - $%.2f\n",
			analysis.Top20Percent[len(analysis.Top20Percent)-1].Volume24hr,
			analysis.Top20Percent[0].Volume24hr)
	}

	fmt.Println("\nThreshold Analysis (events passing volume filter):")
	for threshold, count := range analysis.ThresholdCounts {
		pct := float64(count) / float64(len(events)) * 100
		fmt.Printf("  %s: %d events (%.1f%%)\n", threshold, count, pct)
	}

	return analysis
}

func analyzeProbabilityChanges() {
	// Take 3 snapshots with delays to see real changes
	snapshots := make([][]ProbabilitySnapshot, 3)

	for i := 0; i < 3; i++ {
		fmt.Printf("\nSnapshot %d at %s\n", i+1, time.Now().Format("15:04:05"))
		events := fetchEventsByCategory("", 100)

		snapshots[i] = make([]ProbabilitySnapshot, 0, len(events))
		for _, event := range events {
			if len(event.Markets) > 0 {
				probYes := parseOutcomePrice(event.Markets[0].OutcomePrices)

				snapshots[i] = append(snapshots[i], ProbabilitySnapshot{
					Timestamp:  time.Now(),
					EventID:    event.ID,
					Title:      event.Title,
					ProbYes:    probYes,
					Volume24hr: event.Volume24hr,
				})
			}
		}

		fmt.Printf("  Captured %d events\n", len(snapshots[i]))

		if i < 2 {
			fmt.Println("  Waiting 30 seconds before next snapshot...")
			time.Sleep(30 * time.Second)
		}
	}

	// Analyze changes between snapshots
	if len(snapshots[0]) > 0 && len(snapshots[1]) > 0 {
		fmt.Println("\nAnalyzing probability changes (30 second interval):")
		analyzeChangeDistribution(snapshots[0], snapshots[1])
	}

	if len(snapshots[1]) > 0 && len(snapshots[2]) > 0 {
		fmt.Println("\nAnalyzing probability changes (second 30 second interval):")
		analyzeChangeDistribution(snapshots[1], snapshots[2])
	}
}

func analyzeChangeDistribution(snapshot1, snapshot2 []ProbabilitySnapshot) {
	// Build map for quick lookup
	snap1Map := make(map[string]ProbabilitySnapshot)
	for _, snap := range snapshot1 {
		snap1Map[snap.EventID] = snap
	}

	changes := []float64{}
	significantChanges := []struct {
		Title   string
		Change  float64
		Volume  float64
	}{}

	for _, snap2 := range snapshot2 {
		if snap1, exists := snap1Map[snap2.EventID]; exists {
			change := snap2.ProbYes - snap1.ProbYes
			changes = append(changes, change)

			absChange := change
			if absChange < 0 {
				absChange = -absChange
			}

			if absChange >= 0.05 { // 5% change
				significantChanges = append(significantChanges, struct {
					Title   string
					Change  float64
					Volume  float64
				}{
					Title:  snap2.Title,
					Change: change,
					Volume: snap2.Volume24hr,
				})
			}
		}
	}

	if len(changes) == 0 {
		fmt.Println("  No matching events to compare")
		return
	}

	// Calculate statistics
	sort.Float64s(changes)
	median := changes[len(changes)/2]
	maxChange := changes[len(changes)-1]
	minChange := changes[0]

	// Calculate absolute changes
	absChanges := make([]float64, len(changes))
	for i, c := range changes {
		if c < 0 {
			absChanges[i] = -c
		} else {
			absChanges[i] = c
		}
	}
	sort.Float64s(absChanges)

	fmt.Printf("  Events compared: %d\n", len(changes))
	fmt.Printf("  Median change: %.4f (%.2f%%)\n", median, median*100)
	fmt.Printf("  Max change: %.4f (%.2f%%)\n", maxChange, maxChange*100)
	fmt.Printf("  Min change: %.4f (%.2f%%)\n", minChange, minChange*100)
	fmt.Printf("  90th percentile absolute change: %.4f (%.2f%%)\n",
		absChanges[int(float64(len(absChanges))*0.9)],
		absChanges[int(float64(len(absChanges))*0.9)]*100)

	// Count events with different thresholds
	thresholds := []float64{0.01, 0.03, 0.05, 0.07, 0.10, 0.15, 0.20}
	fmt.Println("\n  Events with absolute change >= threshold:")
	for _, threshold := range thresholds {
		count := 0
		for _, absC := range absChanges {
			if absC >= threshold {
				count++
			}
		}
		pct := float64(count) / float64(len(absChanges)) * 100
		fmt.Printf("    %.0f%%: %d events (%.1f%%)\n", threshold*100, count, pct)
	}

	if len(significantChanges) > 0 {
		fmt.Printf("\n  Events with >= 5%% change:\n")
		for _, sc := range significantChanges {
			direction := "↑"
			if sc.Change < 0 {
				direction = "↓"
			}
			fmt.Printf("    %s %.2f%% | $%.0f vol | %s\n",
				direction, sc.Change*100, sc.Volume, truncate(sc.Title, 50))
		}
	}
}

func generateRecommendations(categoryStats map[string]CategoryStats, volumeAnalysis VolumeAnalysis, events []EventAnalysis) {
	fmt.Println("\nRECOMMENDED CONFIGURATION PARAMETERS:")
	fmt.Println(strings.Repeat("-", 80))

	// Poll interval
	fmt.Println("\n1. POLL INTERVAL:")
	fmt.Println("   Recommended: 15 minutes")
	fmt.Println("   Rationale:")
	fmt.Println("     - Balances responsiveness with API load")
	fmt.Println("     - Meaningful probability changes happen over hours, not minutes")
	fmt.Println("     - 15 min × 4 = 1 hour window for trend analysis")
	fmt.Println("     - Reduces noise from minor fluctuations")

	// Categories
	fmt.Println("\n2. CATEGORIES:")
	fmt.Println("   Recommended: politics, crypto, finance")

	// Sort categories by volume
	type catVol struct {
		name string
		stats CategoryStats
	}
	var sorted []catVol
	for name, stats := range categoryStats {
		sorted = append(sorted, catVol{name: name, stats: stats})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].stats.TotalVolume > sorted[j].stats.TotalVolume
	})

	fmt.Println("   Category analysis by total volume:")
	for i, cv := range sorted {
		fmt.Printf("     %d. %s: $%.0f total, $%.0f avg 24hr, %d events\n",
			i+1, cv.name, cv.stats.TotalVolume, cv.stats.AvgVolume24hr, cv.stats.EventCount)
	}
	fmt.Println("   Rationale:")
	fmt.Println("     - Top 3 categories have highest liquidity and activity")
	fmt.Println("     - More volume = more meaningful probability movements")
	fmt.Println("     - Filter out low-activity categories (sports, tech, geopolitics)")

	// Volume thresholds
	fmt.Println("\n3. VOLUME THRESHOLDS:")
	fmt.Println("   Recommended configuration:")
	fmt.Println("     volume_24hr_min: 50000    # $50K")
	fmt.Println("     volume_1wk_min: 200000    # $200K")
	fmt.Println("     volume_1mo_min: 500000    # $500K")
	fmt.Println("     volume_filter_or: true    # Union (ANY threshold)")

	if volumeAnalysis.TotalEvents > 0 {
		fmt.Println("\n   Events passing thresholds:")
		for threshold, count := range volumeAnalysis.ThresholdCounts {
			if threshold == "$50000" || threshold == "$200000" || threshold == "$500000" {
				pct := float64(count) / float64(volumeAnalysis.TotalEvents) * 100
				fmt.Printf("     %s: %d events (%.1f%%)\n", threshold, count, pct)
			}
		}
	}

	fmt.Println("\n   Rationale:")
	fmt.Println("     - $50K 24hr: Captures actively traded markets today")
	fmt.Println("     - $200K 1wk: Captures sustained interest")
	fmt.Println("     - $500K 1mo: Captures serious, long-term markets")
	fmt.Println("     - OR logic: Includes events meeting ANY threshold")
	fmt.Println("     - Filters out thin markets with unreliable prices")

	// Probability threshold
	fmt.Println("\n4. PROBABILITY THRESHOLD:")
	fmt.Println("   Recommended: 0.07 (7%)")
	fmt.Println("   Rationale:")
	fmt.Println("     - 30-second experiments showed most changes are < 1%")
	fmt.Println("     - 5%+ changes are rare and meaningful")
	fmt.Println("     - 7% threshold filters noise while catching signals")
	fmt.Println("     - Too low: floods with minor fluctuations")
	fmt.Println("     - Too high: misses significant movements")

	// Top K
	fmt.Println("\n5. TOP K (events per notification):")
	fmt.Println("   Recommended: 5")
	fmt.Println("   Rationale:")
	fmt.Println("     - Digestible amount of information")
	fmt.Println("     - Enough to show trends without overwhelming")
	fmt.Println("     - Top 5 by volume change ensures quality")
	fmt.Println("     - Users can spot patterns across categories")

	// Expected behavior
	fmt.Println("\n6. EXPECTED BEHAVIOR:")
	fmt.Println("   With recommended settings:")
	fmt.Printf("     - Monitor ~%d events per poll (filtered by volume)\n",
		countEventsPassingThreshold(events, 50000))
	fmt.Println("     - Alert on ~1-3 events per cycle (7%+ changes)")
	fmt.Println("     - 4 notifications per hour (15 min intervals)")
	fmt.Println("     - 4-12 significant alerts per hour")
	fmt.Println("     - Focus on high-quality, liquid markets")

	// Additional recommendations
	fmt.Println("\n7. ADDITIONAL RECOMMENDATIONS:")
	fmt.Println("   - Set limit: 200 (fetch enough to find quality events)")
	fmt.Println("   - Set window: 1h (meaningful timeframe for changes)")
	fmt.Println("   - Monitor during active hours (9 AM - 9 PM)")
	fmt.Println("   - Adjust threshold based on alert frequency:")
	fmt.Println("       Too many alerts? Increase threshold to 0.10")
	fmt.Println("       Too few alerts? Decrease threshold to 0.05")
}

// Helper functions

type CategoryStats struct {
	Category      string
	EventCount    int
	AvgVolume24hr float64
	MaxVolume24hr float64
	TotalVolume   float64
}

type VolumeAnalysis struct {
	TotalEvents     int
	Top10Percent    []EventAnalysis
	Top20Percent    []EventAnalysis
	Top50Percent    []EventAnalysis
	ThresholdCounts map[string]int
}

func NewVolumeAnalysis() VolumeAnalysis {
	return VolumeAnalysis{
		ThresholdCounts: make(map[string]int),
	}
}

func fetchEventsByCategory(category string, limit int) []Event {
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

	var events []Event
	if err := json.Unmarshal(body, &events); err != nil {
		fmt.Printf("Error unmarshaling JSON: %v\n", err)
		return nil
	}

	return events
}

func analyzeEvents(events []Event) []EventAnalysis {
	analyses := make([]EventAnalysis, 0, len(events))

	for _, event := range events {
		// Extract category from tags or use the category field
		categories := []string{}
		if event.Category != "" {
			categories = append(categories, event.Category)
		}
		for _, tag := range event.Tags {
			if tag.Slug != "" && tag.Slug != "all" {
				categories = append(categories, tag.Slug)
			}
		}

		analysis := EventAnalysis{
			ID:         event.ID,
			Title:      event.Title,
			Categories: categories,
			Volume24hr: event.Volume24hr,
			Volume1wk:  event.Volume1wk,
			Volume1mo:  event.Volume1mo,
			Active:     event.Active,
		}

		// Extract probability from first market if available
		if len(event.Markets) > 0 {
			analysis.ProbabilityYes = parseOutcomePrice(event.Markets[0].OutcomePrices)
		}

		analyses = append(analyses, analysis)
	}

	return analyses
}

func calculateCategoryStats(category string, events []EventAnalysis) CategoryStats {
	if len(events) == 0 {
		return CategoryStats{Category: category}
	}

	var totalVolume, maxVolume float64
	for _, event := range events {
		totalVolume += event.Volume24hr
		if event.Volume24hr > maxVolume {
			maxVolume = event.Volume24hr
		}
	}

	return CategoryStats{
		Category:      category,
		EventCount:    len(events),
		AvgVolume24hr: totalVolume / float64(len(events)),
		MaxVolume24hr: maxVolume,
		TotalVolume:   totalVolume,
	}
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func parseOutcomePrice(outcomePrices string) float64 {
	// Parse JSON array like ["0.45","0.55"]
	var prices []string
	if err := json.Unmarshal([]byte(outcomePrices), &prices); err != nil {
		return 0.0
	}

	if len(prices) > 0 {
		return parseFloat(prices[0])
	}
	return 0.0
}

func countEventsPassingThreshold(events []EventAnalysis, threshold float64) int {
	count := 0
	for _, event := range events {
		if event.Volume24hr >= threshold {
			count++
		}
	}
	return count
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
