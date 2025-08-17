package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

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

// CreateUI creates and returns all UI elements for the main window
func CreateUI(userNickname string) (fyne.CanvasObject, *widget.Label, *widget.Label, *widget.Label, *widget.Label, *widget.ProgressBar, *widget.Label, *widget.Button, *widget.Button) {
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

	// Buttons
	scanButton := widget.NewButton("Scan Replay Files", nil)
	scanBuildingButton := widget.NewButton("Scan Building Stats", nil)

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

	return content, terrLabel, zergLabel, protossLabel, totalLabel, progress, statusLabel, scanButton, scanBuildingButton
}

// UpdateRaceStatsUI updates the UI elements with race statistics
func UpdateRaceStatsUI(terrLabel, zergLabel, protossLabel, totalLabel *widget.Label, stats *RaceStats) {
	terrLabel.SetText(fmt.Sprintf("Terran: %d", stats.Terran))
	zergLabel.SetText(fmt.Sprintf("Zerg: %d", stats.Zerg))
	protossLabel.SetText(fmt.Sprintf("Protoss: %d", stats.Protoss))
	totalLabel.SetText(fmt.Sprintf("Total Games: %d", stats.Terran+stats.Zerg+stats.Protoss))
}

// UpdateBuildingStatsUI updates the UI elements with building statistics
func UpdateBuildingStatsUI(terrLabel, zergLabel, protossLabel, totalLabel *widget.Label, playerStats *BuildingStats) {
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
		sumSlice(playerStats.SupplyDepots)+sumSlice(playerStats.Overlords)+sumSlice(playerStats.Pylons)))
}

// ResetStatsUI resets the UI elements to their initial state
func ResetStatsUI(terrLabel, zergLabel, protossLabel, totalLabel *widget.Label) {
	terrLabel.SetText("Terran: 0")
	zergLabel.SetText("Zerg: 0")
	protossLabel.SetText("Protoss: 0")
	totalLabel.SetText("Total Games: 0")
}

// ShowProgress shows the progress bar and updates the status label
func ShowProgress(progress *widget.ProgressBar, statusLabel *widget.Label, message string) {
	progress.Show()
	statusLabel.SetText(message)
}

// HideProgress hides the progress bar and updates the status label
func HideProgress(progress *widget.ProgressBar, statusLabel *widget.Label, message string) {
	progress.Hide()
	statusLabel.SetText(message)
}
