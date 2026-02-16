package main

import (
	"fmt"
	"sort"
	"strings"
)

// printCategoryStats displays category statistics
func printCategoryStats(category string, stats CategoryStats) {
	fmt.Printf("\n  Category: %s\n", category)
	fmt.Printf("    Events: %d\n", stats.EventCount)
	fmt.Printf("    Avg 24hr Volume: $%.2f\n", stats.AvgVolume24hr)
	fmt.Printf("    Max 24hr Volume: $%.2f\n", stats.MaxVolume24hr)
	fmt.Printf("    Total Volume: $%.2f\n", stats.TotalVolume)
}

// printVolumeDistribution displays volume distribution analysis
func printVolumeDistribution(analysis VolumeAnalysis) {
	fmt.Printf("\nTotal events analyzed: %d\n", analysis.TotalEvents)
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
		pct := float64(count) / float64(analysis.TotalEvents) * 100
		fmt.Printf("  %s: %d events (%.1f%%)\n", threshold, count, pct)
	}
}

// printRecommendations displays configuration recommendations
func printRecommendations(categoryStats map[string]CategoryStats, volumeAnalysis VolumeAnalysis, events []EventAnalysis) {
	fmt.Println("\nRECOMMENDED CONFIGURATION PARAMETERS:")
	fmt.Println(strings.Repeat("-", 80))

	printPollIntervalRecommendation()
	printCategoryRecommendation(categoryStats)
	printVolumeThresholdRecommendation(volumeAnalysis)
	printProbabilityThresholdRecommendation()
	printTopKRecommendation()
	printExpectedBehavior(events)
	printAdditionalRecommendations()
}

func printPollIntervalRecommendation() {
	fmt.Println("\n1. POLL INTERVAL:")
	fmt.Println("   Recommended: 15 minutes")
	fmt.Println("   Rationale:")
	fmt.Println("     - Balances responsiveness with API load")
	fmt.Println("     - Meaningful probability changes happen over hours, not minutes")
	fmt.Println("     - 15 min × 4 = 1 hour window for trend analysis")
	fmt.Println("     - Reduces noise from minor fluctuations")
}

func printCategoryRecommendation(categoryStats map[string]CategoryStats) {
	fmt.Println("\n2. CATEGORIES:")
	fmt.Println("   Recommended: politics, crypto, finance")

	// Sort categories by volume
	type catVol struct {
		name  string
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
}

func printVolumeThresholdRecommendation(volumeAnalysis VolumeAnalysis) {
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
}

func printProbabilityThresholdRecommendation() {
	fmt.Println("\n4. PROBABILITY THRESHOLD:")
	fmt.Println("   Recommended: 0.07 (7%)")
	fmt.Println("   Rationale:")
	fmt.Println("     - 30-second experiments showed most changes are < 1%")
	fmt.Println("     - 5%+ changes are rare and meaningful")
	fmt.Println("     - 7% threshold filters noise while catching signals")
	fmt.Println("     - Too low: floods with minor fluctuations")
	fmt.Println("     - Too high: misses significant movements")
}

func printTopKRecommendation() {
	fmt.Println("\n5. TOP K (events per notification):")
	fmt.Println("   Recommended: 5")
	fmt.Println("   Rationale:")
	fmt.Println("     - Digestible amount of information")
	fmt.Println("     - Enough to show trends without overwhelming")
	fmt.Println("     - Top 5 by volume change ensures quality")
	fmt.Println("     - Users can spot patterns across categories")
}

func printExpectedBehavior(events []EventAnalysis) {
	fmt.Println("\n6. EXPECTED BEHAVIOR:")
	fmt.Println("   With recommended settings:")
	fmt.Printf("     - Monitor ~%d events per poll (filtered by volume)\n",
		countEventsPassingThreshold(events, 50000))
	fmt.Println("     - Alert on ~1-3 events per cycle (7%+ changes)")
	fmt.Println("     - 4 notifications per hour (15 min intervals)")
	fmt.Println("     - 4-12 significant alerts per hour")
	fmt.Println("     - Focus on high-quality, liquid markets")
}

func printAdditionalRecommendations() {
	fmt.Println("\n7. ADDITIONAL RECOMMENDATIONS:")
	fmt.Println("   - Set limit: 200 (fetch enough to find quality events)")
	fmt.Println("   - Set window: 1h (meaningful timeframe for changes)")
	fmt.Println("   - Monitor during active hours (9 AM - 9 PM)")
	fmt.Println("   - Adjust threshold based on alert frequency:")
	fmt.Println("       Too many alerts? Increase threshold to 0.10")
	fmt.Println("       Too few alerts? Decrease threshold to 0.05")
}
