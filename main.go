package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// RaceStats holds statistics about race usage
type RaceStats struct {
	Terran  int
	Zerg    int
	Protoss int
}

// BuildingStats holds statistics about building construction per player
type BuildingStats struct {
	PlayerName    string
	SupplyDepots  []int // Count per second
	Overlords     []int // Count per second
	Pylons        []int // Count per second
	MaxDuration   int   // Maximum game duration in seconds
}

// BuildingStatsMap holds building statistics for all players
type BuildingStatsMap map[string]*BuildingStats

func main() {
	myApp := app.NewWithID("com.innerspirit.bwstats")
	myApp.Settings().SetTheme(&FuturisticTheme{})

	myWindow := myApp.NewWindow("BW Stats - Replay Analyzer")
	myWindow.Resize(fyne.NewSize(500, 400))

	// Load user settings
	userNickname, err := loadUserNickname()
	if err != nil {
		userNickname = "Error loading settings: " + err.Error()
	}

	// Create UI via helpers
	content, terrLabel, zergLabel, protossLabel, totalLabel, progress, statusLabel, scanButton, scanBuildingButton := CreateUI(userNickname)
	scanButton.OnTapped = func() {
		// Reset and show progress
		ResetStatsUI(terrLabel, zergLabel, protossLabel, totalLabel)
		ShowProgress(progress, statusLabel, "Scanning replay files...")
		scanButton.Disable()

		// Start scanning in a goroutine
		go func() {
			stats, err := scanReplays(userNickname, func(p float64) {
				// Update progress bar in UI thread
				fyne.Do(func() {
					progress.SetValue(p)
				})
			})
			if err != nil {
				fyne.Do(func() {
					statusLabel.SetText("Error: " + err.Error())
					progress.Hide()
					scanButton.Enable()
				})
				return
			}

			// Update UI with results
			fyne.Do(func() {
				UpdateRaceStatsUI(terrLabel, zergLabel, protossLabel, totalLabel, stats)
				HideProgress(progress, statusLabel, "Scan completed successfully!")
				scanButton.Enable()

				// Send notification
				fyne.CurrentApp().SendNotification(&fyne.Notification{
					Title:   "Scan Complete",
					Content: fmt.Sprintf("Found %d Terran, %d Zerg, %d Protoss games", stats.Terran, stats.Zerg, stats.Protoss),
				})
			})
		}()
	}

	scanBuildingButton.OnTapped = func() {
		// Reset stats and show progress
		ResetStatsUI(terrLabel, zergLabel, protossLabel, totalLabel)
		ShowProgress(progress, statusLabel, "Scanning building stats...")
		scanBuildingButton.Disable()
		scanButton.Disable()

		// Start scanning in a goroutine
		go func() {
			buildingStats, err := scanBuildingStats(userNickname, func(p float64) {
				fyne.Do(func() {
					progress.SetValue(p)
				})
			})
			if err != nil {
				fyne.Do(func() {
					statusLabel.SetText("Error: " + err.Error())
					progress.Hide()
					scanBuildingButton.Enable()
					scanButton.Enable()
				})
				return
			}

			// Update UI with results
			fyne.Do(func() {
				if playerStats, exists := buildingStats[userNickname]; exists {
					UpdateBuildingStatsUI(terrLabel, zergLabel, protossLabel, totalLabel, playerStats)
				} else {
					statusLabel.SetText("No building stats found for user")
				}
				HideProgress(progress, statusLabel, "Building stats scan completed!")
				scanBuildingButton.Enable()
				scanButton.Enable()

				// Send notification
				fyne.CurrentApp().SendNotification(&fyne.Notification{
					Title:   "Building Stats Scan Complete",
					Content: fmt.Sprintf("Building stats collected for %s", userNickname),
				})
			})
		}()
	}

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

