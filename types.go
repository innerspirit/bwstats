package main

// RaceStats holds statistics about race usage.
type RaceStats struct {
	Terran  int
	Zerg    int
	Protoss int
}

// BuildingStats holds building construction statistics per player.
type BuildingStats struct {
	PlayerName   string
	SupplyDepots []int // Count per second
	Overlords    []int // Count per second
	Pylons       []int // Count per second
	MaxDuration  int   // Maximum game duration in seconds
}

// BuildingStatsMap holds building statistics for all players.
type BuildingStatsMap map[string]*BuildingStats
