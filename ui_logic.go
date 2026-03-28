package main

import (
	"fmt"
	"math"
)

func formatSummaryLines(summary *MacroSummary) []string {
	if summary == nil {
		return []string{"No results yet."}
	}

	lines := []string{
		fmt.Sprintf("Target: %s", summary.TargetLabel),
		fmt.Sprintf("Matched Replays: %d", summary.MatchedReplays),
	}
	if summary.SkippedReplays > 0 {
		lines = append(lines, fmt.Sprintf("Skipped Replays: %d", summary.SkippedReplays))
	}

	lines = append(lines,
		fmt.Sprintf(
			"Supply Block: %s total, %s avg, rating: %s",
			formatDurationSeconds(summary.TotalSupplyBlockedSeconds),
			formatDurationSeconds(int(math.Round(summary.AvgSupplyBlockedSeconds))),
			summary.SupplyRating,
		),
		fmt.Sprintf(
			"Worker Idle: %s total, %s avg, rating: %s",
			formatDurationSeconds(summary.TotalWorkerIdleSeconds),
			formatDurationSeconds(int(math.Round(summary.AvgWorkerIdleSeconds))),
			summary.WorkerRating,
		),
	)

	return lines
}

func formatDurationSeconds(totalSeconds int) string {
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	if minutes == 0 {
		return fmt.Sprintf("%ds", seconds)
	}
	return fmt.Sprintf("%dm%02ds", minutes, seconds)
}

func formatChartFooter(series []int) string {
	peak := 0
	for _, value := range series {
		if value > peak {
			peak = value
		}
	}
	return fmt.Sprintf("Peak bucket: %ds", peak)
}
