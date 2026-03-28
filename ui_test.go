package main

import (
	"strings"
	"testing"
)

func TestFormatSummaryLines(t *testing.T) {
	summary := &MacroSummary{
		TargetLabel:               "Current Player (alpha, bravo)",
		MatchedReplays:            4,
		SkippedReplays:            1,
		TotalSupplyBlockedSeconds: 80,
		TotalWorkerIdleSeconds:    160,
		AvgSupplyBlockedSeconds:   20,
		AvgWorkerIdleSeconds:      40,
		SupplyRating:              "Solid",
		WorkerRating:              "Needs Work",
	}

	lines := formatSummaryLines(summary)
	joined := strings.Join(lines, "\n")

	if !strings.Contains(joined, "Target: Current Player (alpha, bravo)") {
		t.Fatalf("missing target line: %q", joined)
	}
	if !strings.Contains(joined, "Matched Replays: 4") {
		t.Fatalf("missing matched replay count: %q", joined)
	}
	if !strings.Contains(joined, "Skipped Replays: 1") {
		t.Fatalf("missing skipped replay count: %q", joined)
	}
	if !strings.Contains(joined, "Supply Block: 1m20s total, 20s avg, rating: Solid") {
		t.Fatalf("missing supply summary: %q", joined)
	}
	if !strings.Contains(joined, "Worker Idle: 2m40s total, 40s avg, rating: Needs Work") {
		t.Fatalf("missing worker summary: %q", joined)
	}
}

func TestFormatChartFooter(t *testing.T) {
	if got := formatChartFooter([]int{0, 4, 9}); got != "Peak bucket: 9s" {
		t.Fatalf("unexpected chart footer: %q", got)
	}
}
