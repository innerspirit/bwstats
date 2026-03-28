package main

import "fmt"

func formatRaceStats(stats *RaceStats) (terr, zerg, protoss, total string) {
	return fmt.Sprintf("Terran: %d", stats.Terran),
		fmt.Sprintf("Zerg: %d", stats.Zerg),
		fmt.Sprintf("Protoss: %d", stats.Protoss),
		fmt.Sprintf("Total Games: %d", stats.Terran+stats.Zerg+stats.Protoss)
}

func formatBuildingStats(playerStats *BuildingStats) (terr, zerg, protoss, total string) {
	return fmt.Sprintf("Supply Depots: %d (avg: %.2f/sec)",
			sumSlice(playerStats.SupplyDepots),
			float64(sumSlice(playerStats.SupplyDepots))/float64(len(playerStats.SupplyDepots))),
		fmt.Sprintf("Overlords: %d (avg: %.2f/sec)",
			sumSlice(playerStats.Overlords),
			float64(sumSlice(playerStats.Overlords))/float64(len(playerStats.Overlords))),
		fmt.Sprintf("Pylons: %d (avg: %.2f/sec)",
			sumSlice(playerStats.Pylons),
			float64(sumSlice(playerStats.Pylons))/float64(len(playerStats.Pylons))),
		fmt.Sprintf("Total Buildings: %d",
			sumSlice(playerStats.SupplyDepots)+sumSlice(playerStats.Overlords)+sumSlice(playerStats.Pylons))
}
