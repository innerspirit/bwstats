//go:build windows

package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	myApp := app.NewWithID("com.innerspirit.bwstats")
	myApp.Settings().SetTheme(&FuturisticTheme{})

	myWindow := myApp.NewWindow("BW Stats - Replay Analyzer")
	myWindow.Resize(fyne.NewSize(720, 760))

	// Load user settings
	identity, err := loadPlayerIdentity()
	if err != nil {
		identity = PlayerIdentity{
			DisplayName: "Error loading settings: " + err.Error(),
		}
	}

	ui := CreateUI(identity)
	ui.ScanButton.OnTapped = func() {
		target := resolveScanTarget(identity, ui.ManualEntry.Text)
		UpdateSummaryUI(ui.SummaryLabel, nil)
		ui.SupplyChart.SetSeries(make([]int, chartBucketCount))
		ui.WorkerChart.SetSeries(make([]int, chartBucketCount))
		ShowProgress(ui.Progress, ui.StatusLabel, "Scanning replay files...")
		ui.ScanButton.Disable()

		go func() {
			summary, err := scanMacroStats(target, func(p float64) {
				fyne.Do(func() {
					ui.Progress.SetValue(p)
				})
			})
			if err != nil {
				fyne.Do(func() {
					ui.StatusLabel.SetText("Error: " + err.Error())
					ui.Progress.Hide()
					ui.ScanButton.Enable()
				})
				return
			}

			fyne.Do(func() {
				UpdateSummaryUI(ui.SummaryLabel, summary)
				ui.SupplyChart.SetSeries(summary.SupplyChart)
				ui.WorkerChart.SetSeries(summary.WorkerChart)
				HideProgress(ui.Progress, ui.StatusLabel, "Scan completed successfully!")
				ui.ScanButton.Enable()

				fyne.CurrentApp().SendNotification(&fyne.Notification{
					Title: "Scan Complete",
					Content: fmt.Sprintf(
						"%s: %d matched replays, %s avg supply block, %s avg worker idle",
						target.DisplayLabel,
						summary.MatchedReplays,
						formatDurationSeconds(int(summary.AvgSupplyBlockedSeconds)),
						formatDurationSeconds(int(summary.AvgWorkerIdleSeconds)),
					),
				})
			})
		}()
	}

	myWindow.SetContent(ui.Content)
	myWindow.ShowAndRun()
}
