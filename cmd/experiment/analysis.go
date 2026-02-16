package main

import (
	"encoding/json"
	"fmt"
	"sort"
)

// analyzeEvents converts PolymarketEvent slice to EventAnalysis slice
func analyzeEvents(events []PolymarketEvent) []EventAnalysis {
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

// calculateCategoryStats computes statistics for events in a category
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

// analyzeVolumeDistribution analyzes the volume distribution of events
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

	return analysis
}

// analyzeChangeDistribution analyzes probability changes between snapshots
func analyzeChangeDistribution(snapshot1, snapshot2 []ProbabilitySnapshot) {
	// Build map for quick lookup
	snap1Map := make(map[string]ProbabilitySnapshot)
	for _, snap := range snapshot1 {
		snap1Map[snap.EventID] = snap
	}

	changes := []float64{}
	significantChanges := []struct {
		Title  string
		Change float64
		Volume float64
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
					Title  string
					Change float64
					Volume float64
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

// parseOutcomePrice parses outcome prices from JSON string
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

// parseFloat parses a string to float64
func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
