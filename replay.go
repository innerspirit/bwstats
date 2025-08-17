package main

import (
	screp "github.com/icza/screp/rep"
	"github.com/icza/screp/rep/repcore"
	"github.com/icza/screp/rep/repcmd"
	"github.com/icza/screp/repparser"
)

// scanReplays scans the user's replay files and gathers race usage statistics
func scanReplays(userNickname string, progressCallback func(float64)) (*RaceStats, error) {
	// Find all .rep files
	repFiles, err := findReplayFiles(progressCallback)
	if err != nil {
		return nil, err
	}

	// Initialize race statistics
	stats := &RaceStats{}

	totalFiles := len(repFiles)
	processedFiles := 0

	// Parse each replay file
	for _, repFile := range repFiles {
		// Parse the replay file with commands enabled
		cfg := repparser.Config{Commands: true}
		rep, err := repparser.ParseFileConfig(repFile, cfg)
		if err != nil {
			// Skip files that can't be parsed
			continue
		}

		// Find the player with the matching nickname
		for _, player := range rep.Header.Players {
			if player.Name == userNickname {
				// Count race usage
				switch player.Race {
				case repcore.RaceTerran:
					stats.Terran++
				case repcore.RaceZerg:
					stats.Zerg++
				case repcore.RaceProtoss:
					stats.Protoss++
				}
				break
			}
		}

		processedFiles++
		// Update progress
		if progressCallback != nil && totalFiles > 0 {
			progressCallback(float64(processedFiles) / float64(totalFiles))
		}
	}

	return stats, nil
}

// scanBuildingStats scans the user's replay files and gathers building construction statistics
func scanBuildingStats(userNickname string, progressCallback func(float64)) (BuildingStatsMap, error) {
	// Find all .rep files
	repFiles, err := findReplayFiles(progressCallback)
	if err != nil {
		return nil, err
	}

	// Initialize building statistics
	buildingStats := make(BuildingStatsMap)

	totalFiles := len(repFiles)
	processedFiles := 0

	// Parse each replay file
	for _, repFile := range repFiles {
		// Parse the replay file with commands enabled
		cfg := repparser.Config{Commands: true}
		rep, err := repparser.ParseFileConfig(repFile, cfg)
		if err != nil {
			// Skip files that can't be parsed
			continue
		}

		// Process commands to gather building construction statistics
		processBuildingCommands(rep, userNickname, buildingStats)

		processedFiles++
		// Update progress
		if progressCallback != nil && totalFiles > 0 {
			progressCallback(float64(processedFiles) / float64(totalFiles))
		}
	}

	return buildingStats, nil
}

// processBuildingCommands processes the commands in a replay to gather building construction statistics
func processBuildingCommands(rep *screp.Replay, userNickname string, buildingStats BuildingStatsMap) {
	// Find the player with the matching nickname
	var targetPlayer *screp.Player
	for _, player := range rep.Header.Players {
		if player.Name == userNickname {
			targetPlayer = player
			break
		}
	}

	// If we didn't find the player, return
	if targetPlayer == nil {
		return
	}

	// Get or create building stats for this player
	playerStats, exists := buildingStats[userNickname]
	if !exists {
		playerStats = &BuildingStats{
			PlayerName:   userNickname,
			SupplyDepots: make([]int, 0),
			Overlords:    make([]int, 0),
			Pylons:       make([]int, 0),
			MaxDuration:  0,
		}
		buildingStats[userNickname] = playerStats
	} else {
		// Extend slices if necessary
		gameDuration := int(rep.Header.Frames.Duration().Seconds()) + 1
		playerStats.SupplyDepots = extendSlice(playerStats.SupplyDepots, gameDuration)
		playerStats.Overlords = extendSlice(playerStats.Overlords, gameDuration)
		playerStats.Pylons = extendSlice(playerStats.Pylons, gameDuration)
	}

	// Process commands to find building construction actions
	for _, cmd := range rep.Commands.Cmds {
		// Check if this command was issued by the target player
		if cmd.BaseCmd().PlayerID == targetPlayer.ID {
			// Check if this is a build command
			if buildCmd, ok := cmd.(*repcmd.BuildCmd); ok {
				// Calculate the second this command was issued
				second := int(buildCmd.Frame.Duration().Seconds())

				// Check what unit was built
				switch buildCmd.Unit.ID {
				case 106: // Supply Depot
					if second < len(playerStats.SupplyDepots) {
						playerStats.SupplyDepots[second]++
					}
				case 109: // Overlord
					if second < len(playerStats.Overlords) {
						playerStats.Overlords[second]++
					}
				case 156: // Pylon
					if second < len(playerStats.Pylons) {
						playerStats.Pylons[second]++
					}
				}
			}
		}
	}
}

// extendSlice extends a slice to the specified size, filling new elements with 0
func extendSlice(slice []int, newSize int) []int {
	if len(slice) >= newSize {
		return slice
	}
	newSlice := make([]int, newSize)
	copy(newSlice, slice)
	return newSlice
}

// sumSlice calculates the sum of all elements in a slice
func sumSlice(slice []int) int {
	sum := 0
	for _, v := range slice {
		sum += v
	}
	return sum
}
