package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	screp "github.com/icza/screp/rep"
	"github.com/icza/screp/rep/repcore"
	"github.com/icza/screp/rep/repcmd"
	"github.com/icza/screp/repparser"
)

// Account represents a user account from CSettings.json
type Account struct {
	Account string `json:"account"`
}

// Settings mirrors akafinder's structure for CSettings.json
// See akafinder `main.go` for reference.
type Settings struct {
	GatewayHistory []Account `json:"Gateway History"`
}

// FuturisticTheme defines a custom dark theme
type FuturisticTheme struct{}

func (f FuturisticTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{10, 15, 25, 255}
	case theme.ColorNameButton:
		return color.RGBA{20, 30, 50, 255}
	case theme.ColorNameDisabledButton:
		return color.RGBA{15, 20, 30, 255}
	case theme.ColorNameForeground:
		return color.RGBA{0, 255, 200, 255}
	case theme.ColorNameDisabled:
		return color.RGBA{100, 100, 100, 255}
	case theme.ColorNamePlaceHolder:
		return color.RGBA{150, 150, 150, 255}
	case theme.ColorNamePressed:
		return color.RGBA{0, 200, 255, 255}
	case theme.ColorNameSelection:
		return color.RGBA{0, 100, 150, 80}
	case theme.ColorNameSeparator:
		return color.RGBA{0, 0, 0, 0}
	case theme.ColorNameShadow:
		return color.RGBA{0, 0, 0, 100}
	case theme.ColorNameInputBackground:
		return color.RGBA{15, 25, 40, 255}
	case theme.ColorNameMenuBackground:
		return color.RGBA{20, 30, 50, 255}
	case theme.ColorNameOverlayBackground:
		return color.RGBA{0, 0, 0, 180}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (f FuturisticTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (f FuturisticTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (f FuturisticTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 16
	case theme.SizeNameCaptionText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 20
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInlineIcon:
		return 24
	case theme.SizeNameScrollBar:
		return 16
	case theme.SizeNameScrollBarSmall:
		return 8
	default:
		return theme.DefaultTheme().Size(name)
	}
}

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

	// UI elements
	welcomeLabel := widget.NewLabel("BW Stats - Replay Analyzer")
	welcomeLabel.TextStyle = fyne.TextStyle{Bold: true}

	userLabel := widget.NewLabel("Current User: " + userNickname)
	userLabel.Wrapping = fyne.TextWrapWord

	// Race statistics labels
	terrLabel := widget.NewLabel("Terran: 0")
	zergLabel := widget.NewLabel("Zerg: 0")
	protossLabel := widget.NewLabel("Protoss: 0")
	totalLabel := widget.NewLabel("Total Games: 0")

	// Progress bar for scanning
	progress := widget.NewProgressBar()
	progress.Hide()

	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignCenter

	// Button to start scanning
	scanButton := widget.NewButton("Scan Replay Files", nil)
	scanButton.OnTapped = func() {
		// Reset stats
		terrLabel.SetText("Terran: 0")
		zergLabel.SetText("Zerg: 0")
		protossLabel.SetText("Protoss: 0")
		totalLabel.SetText("Total Games: 0")

		// Show progress bar
		progress.Show()
		statusLabel.SetText("Scanning replay files...")
		scanButton.Disable()

		// Start scanning in a goroutine
		go func(btn *widget.Button, prog *widget.ProgressBar) {
			stats, err := scanReplays(userNickname, func(p float64) {
				// Update progress bar in UI thread
				fyne.Do(func() {
					prog.SetValue(p)
				})
			})
			if err != nil {
				fyne.Do(func() {
					statusLabel.SetText("Error: " + err.Error())
					progress.Hide()
					btn.Enable()
				})
				return
			}

			// Update UI with results
			fyne.Do(func() {
				terrLabel.SetText(fmt.Sprintf("Terran: %d", stats.Terran))
				zergLabel.SetText(fmt.Sprintf("Zerg: %d", stats.Zerg))
				protossLabel.SetText(fmt.Sprintf("Protoss: %d", stats.Protoss))
				totalLabel.SetText(fmt.Sprintf("Total Games: %d", stats.Terran+stats.Zerg+stats.Protoss))
				progress.Hide()
				statusLabel.SetText("Scan completed successfully!")
				btn.Enable()

				// Send notification
				fyne.CurrentApp().SendNotification(&fyne.Notification{
					Title:   "Scan Complete",
					Content: fmt.Sprintf("Found %d Terran, %d Zerg, %d Protoss games", stats.Terran, stats.Zerg, stats.Protoss),
				})
			})
		}(scanButton, progress)
	}

	// Button to scan building stats
	scanBuildingButton := widget.NewButton("Scan Building Stats", nil)
	scanBuildingButton.OnTapped = func() {
		// Reset stats
		terrLabel.SetText("Terran: 0")
		zergLabel.SetText("Zerg: 0")
		protossLabel.SetText("Protoss: 0")
		totalLabel.SetText("Total Games: 0")

		// Show progress bar
		progress.Show()
		statusLabel.SetText("Scanning building stats...")
		scanBuildingButton.Disable()
		scanButton.Disable()

		// Start scanning in a goroutine
		go func(btn *widget.Button, prog *widget.ProgressBar) {
			buildingStats, err := scanBuildingStats(userNickname, func(p float64) {
				// Update progress bar in UI thread
				fyne.Do(func() {
					prog.SetValue(p)
				})
			})
			if err != nil {
				fyne.Do(func() {
					statusLabel.SetText("Error: " + err.Error())
					progress.Hide()
					btn.Enable()
					scanButton.Enable()
				})
				return
			}

			// Update UI with results
			fyne.Do(func() {
				// Display building stats
				if playerStats, exists := buildingStats[userNickname]; exists {
					terrLabel.SetText(fmt.Sprintf("Supply Depots: %d (avg: %.2f/sec)", 
						sumSlice(playerStats.SupplyDepots), 
						float64(sumSlice(playerStats.SupplyDepots))/float64(len(playerStats.SupplyDepots))))
					zergLabel.SetText(fmt.Sprintf("Overlords: %d (avg: %.2f/sec)", 
						sumSlice(playerStats.Overlords), 
						float64(sumSlice(playerStats.Overlords))/float64(len(playerStats.Overlords))))
					protossLabel.SetText(fmt.Sprintf("Pylons: %d (avg: %.2f/sec)", 
						sumSlice(playerStats.Pylons), 
						float64(sumSlice(playerStats.Pylons))/float64(len(playerStats.Pylons))))
					totalLabel.SetText(fmt.Sprintf("Total Buildings: %d", 
						sumSlice(playerStats.SupplyDepots) + sumSlice(playerStats.Overlords) + sumSlice(playerStats.Pylons)))
				} else {
					statusLabel.SetText("No building stats found for user")
				}
				progress.Hide()
				statusLabel.SetText("Building stats scan completed!")
				btn.Enable()
				scanButton.Enable()

				// Send notification
				fyne.CurrentApp().SendNotification(&fyne.Notification{
					Title:   "Building Stats Scan Complete",
					Content: fmt.Sprintf("Building stats collected for %s", userNickname),
				})
			})
		}(scanBuildingButton, progress)
	}

	content := container.NewVBox(
		welcomeLabel,
		userLabel,
		widget.NewSeparator(),
		scanButton,
		scanBuildingButton,
		progress,
		statusLabel,
		widget.NewSeparator(),
		terrLabel,
		zergLabel,
		protossLabel,
		totalLabel,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

// loadUserNickname loads the user's nickname from CSettings.json
func loadUserNickname() (string, error) {
	// Get user home directory
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %v", err)
	}

	// Construct path to CSettings.json similarly to akafinder
	settingsPath := filepath.Join(userHome, "Documents", "StarCraft", "CSettings.json")

	// Check if file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return "Settings file not found", nil
	}

	// Open and read the file
	f, err := os.Open(settingsPath)
	if err != nil {
		return "", fmt.Errorf("failed to open settings file: %v", err)
	}
	defer f.Close()

	// Read file content
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed to read settings file: %v", err)
	}

	// Parse JSON using the same structure as akafinder
	var settings Settings
	if err := json.Unmarshal(b, &settings); err != nil {
		return "", fmt.Errorf("failed to parse settings: %v", err)
	}

	// Use the first account as the current user
	if len(settings.GatewayHistory) > 0 {
		return settings.GatewayHistory[0].Account, nil
	}

	return "No accounts found", nil
}

// scanReplays scans the user's replay files and gathers race usage statistics
func scanReplays(userNickname string, progressCallback func(float64)) (*RaceStats, error) {
	// Get the replay directory path
	replayDir := filepath.Join(os.Getenv("USERPROFILE"), "Documents", "StarCraft", "Maps", "Replays", "AutoSave")
	// Check if the replay directory exists
	if _, err := os.Stat(replayDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("autoreplays directory not found: %s", replayDir)
	}

	// Get subdirectories (date folders)
	subdirs, err := os.ReadDir(replayDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read autoreplays directory: %v", err)
	}

	if len(subdirs) == 0 {
		return nil, fmt.Errorf("no date folders found in autoreplays directory")
	}

	// Find all .rep files in the directory and subdirectories
	var repFiles []string

	// Walk through each subdirectory to find all .rep files
	for i, subdir := range subdirs {
		if !subdir.IsDir() {
			continue
		}

		subDirPath := filepath.Join(replayDir, subdir.Name())
		err := filepath.Walk(subDirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Check if it's a .rep file
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".rep") {
				repFiles = append(repFiles, path)
			}

			return nil
		})

		// Update progress after each subdirectory is processed
		if progressCallback != nil {
			progressCallback(float64(i+1) / float64(len(subdirs)))
		}

		if err != nil {
			return nil, fmt.Errorf("failed to walk subdirectory %s: %v", subdir.Name(), err)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to walk replay directory: %v", err)
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
	// Get the replay directory path
	replayDir := filepath.Join(os.Getenv("USERPROFILE"), "Documents", "StarCraft", "Maps", "Replays", "AutoSave")
	// Check if the replay directory exists
	if _, err := os.Stat(replayDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("autoreplays directory not found: %s", replayDir)
	}

	// Get subdirectories (date folders)
	subdirs, err := os.ReadDir(replayDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read autoreplays directory: %v", err)
	}

	if len(subdirs) == 0 {
		return nil, fmt.Errorf("no date folders found in autoreplays directory")
	}

	// Find all .rep files in the directory and subdirectories
	var repFiles []string

	// Walk through each subdirectory to find all .rep files
	for i, subdir := range subdirs {
		if !subdir.IsDir() {
			continue
		}

		subDirPath := filepath.Join(replayDir, subdir.Name())
		err := filepath.Walk(subDirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Check if it's a .rep file
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".rep") {
				repFiles = append(repFiles, path)
			}

			return nil
		})

		// Update progress after each subdirectory is processed
		if progressCallback != nil {
			progressCallback(float64(i+1) / float64(len(subdirs)))
		}

		if err != nil {
			return nil, fmt.Errorf("failed to walk subdirectory %s: %v", subdir.Name(), err)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to walk replay directory: %v", err)
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
