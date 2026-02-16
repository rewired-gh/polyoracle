package main

import (
	"fmt"
	"strings"
	"time"
)

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
	printVolumeDistribution(volumeAnalysis)

	// Step 3: Test probability changes over time
	fmt.Println("\nSTEP 3: Testing probability changes over time...")
	fmt.Println(strings.Repeat("-", 80))
	analyzeProbabilityChanges()

	// Step 4: Generate recommendations
	fmt.Println("\nSTEP 4: Generating recommendations...")
	fmt.Println(strings.Repeat("-", 80))
	printRecommendations(categoryStats, volumeAnalysis, allEvents)

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
		printCategoryStats(category, stats[category])
	}

	return stats
}

func fetchAllEvents() []EventAnalysis {
	fmt.Println("\nFetching all events (no category filter)...")
	events := fetchEventsByCategory("", 200)
	return analyzeEvents(events)
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

// Helper functions

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
