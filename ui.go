//go:build windows

package main

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const chartHeight float32 = 72

// FuturisticTheme defines a custom dark theme.
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

type AppUI struct {
	Content      fyne.CanvasObject
	ManualEntry  *widget.Entry
	SummaryLabel *widget.Label
	Progress     *widget.ProgressBar
	StatusLabel  *widget.Label
	ScanButton   *widget.Button
	SupplyChart  *MiniBarChart
	WorkerChart  *MiniBarChart
}

type MiniBarChart struct {
	title  *widget.Label
	footer *widget.Label
	bars   *fyne.Container
	root   *fyne.Container
	color  color.Color
}

// CreateUI builds the macro-analysis UI.
func CreateUI(identity PlayerIdentity) *AppUI {
	welcomeLabel := widget.NewLabel("BW Stats - Brood War Macro Analyzer")
	welcomeLabel.TextStyle = fyne.TextStyle{Bold: true}

	autoTarget := widget.NewLabel(fmt.Sprintf("Auto Target: %s", formatIdentityLabel(identity)))
	autoTarget.Wrapping = fyne.TextWrapWord

	manualEntry := widget.NewEntry()
	manualEntry.SetPlaceHolder("Manual player name override (optional)")

	summaryLabel := widget.NewLabel(strings.Join(formatSummaryLines(nil), "\n"))
	summaryLabel.Wrapping = fyne.TextWrapWord

	progress := widget.NewProgressBar()
	progress.Hide()

	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignCenter

	scanButton := widget.NewButton("Scan Macro Stats", nil)

	supplyChart := NewMiniBarChart("Supply Block Chart (0:00-15:00)", color.RGBA{0, 255, 200, 255})
	workerChart := NewMiniBarChart("Worker Idle Chart (0:00-15:00)", color.RGBA{255, 190, 64, 255})

	content := container.NewVBox(
		welcomeLabel,
		autoTarget,
		manualEntry,
		widget.NewSeparator(),
		scanButton,
		progress,
		statusLabel,
		widget.NewSeparator(),
		summaryLabel,
		widget.NewSeparator(),
		supplyChart.CanvasObject(),
		workerChart.CanvasObject(),
	)

	return &AppUI{
		Content:      container.NewVScroll(content),
		ManualEntry:  manualEntry,
		SummaryLabel: summaryLabel,
		Progress:     progress,
		StatusLabel:  statusLabel,
		ScanButton:   scanButton,
		SupplyChart:  supplyChart,
		WorkerChart:  workerChart,
	}
}

func NewMiniBarChart(title string, barColor color.Color) *MiniBarChart {
	titleLabel := widget.NewLabel(title)
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	footerLabel := widget.NewLabel("Peak bucket: 0s")
	bars := container.NewGridWithColumns(chartBucketCount)

	chart := &MiniBarChart{
		title:  titleLabel,
		footer: footerLabel,
		bars:   bars,
		color:  barColor,
	}
	chart.root = container.NewVBox(titleLabel, bars, footerLabel)
	chart.SetSeries(make([]int, chartBucketCount))

	return chart
}

func (c *MiniBarChart) CanvasObject() fyne.CanvasObject {
	return c.root
}

func (c *MiniBarChart) SetSeries(series []int) {
	c.bars.Objects = nil
	maxValue := 1
	for _, value := range series {
		if value > maxValue {
			maxValue = value
		}
	}

	for _, value := range series {
		height := float32(value) / float32(maxValue) * chartHeight
		if height < 2 {
			height = 2
		}
		rect := canvas.NewRectangle(c.color)
		rect.SetMinSize(fyne.NewSize(6, height))
		c.bars.Add(container.NewVBox(layout.NewSpacer(), rect))
	}

	c.footer.SetText(formatChartFooter(series))
	c.bars.Refresh()
}

func UpdateSummaryUI(label *widget.Label, summary *MacroSummary) {
	label.SetText(strings.Join(formatSummaryLines(summary), "\n"))
}

func formatIdentityLabel(identity PlayerIdentity) string {
	if len(identity.Aliases) == 0 {
		return identity.DisplayName
	}
	return fmt.Sprintf("%s (%s)", identity.DisplayName, strings.Join(identity.Aliases, ", "))
}

func ShowProgress(progress *widget.ProgressBar, statusLabel *widget.Label, message string) {
	progress.Show()
	statusLabel.SetText(message)
}

func HideProgress(progress *widget.ProgressBar, statusLabel *widget.Label, message string) {
	progress.Hide()
	statusLabel.SetText(message)
}
