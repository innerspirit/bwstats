package main

import (
	"testing"
)

func TestFormatRaceStats(t *testing.T) {
	terr, zerg, protoss, total := formatRaceStats(&RaceStats{Terran: 4, Zerg: 5, Protoss: 6})

	if terr != "Terran: 4" {
		t.Fatalf("unexpected terran label: %q", terr)
	}
	if zerg != "Zerg: 5" {
		t.Fatalf("unexpected zerg label: %q", zerg)
	}
	if protoss != "Protoss: 6" {
		t.Fatalf("unexpected protoss label: %q", protoss)
	}
	if total != "Total Games: 15" {
		t.Fatalf("unexpected total label: %q", total)
	}
}

func TestFormatBuildingStats(t *testing.T) {
	playerStats := &BuildingStats{
		SupplyDepots: []int{1, 1},
		Overlords:    []int{2, 0},
		Pylons:       []int{0, 3},
	}

	terr, zerg, protoss, total := formatBuildingStats(playerStats)

	if terr != "Supply Depots: 2 (avg: 1.00/sec)" {
		t.Fatalf("unexpected terran building label: %q", terr)
	}
	if zerg != "Overlords: 2 (avg: 1.00/sec)" {
		t.Fatalf("unexpected zerg building label: %q", zerg)
	}
	if protoss != "Pylons: 3 (avg: 1.50/sec)" {
		t.Fatalf("unexpected protoss building label: %q", protoss)
	}
	if total != "Total Buildings: 7" {
		t.Fatalf("unexpected total building label: %q", total)
	}
}
