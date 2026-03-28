package main

// PlayerIdentity represents the current player's display name and alias set.
type PlayerIdentity struct {
	DisplayName string
	Aliases     []string
}

// ScanTarget represents the selected replay scan target.
type ScanTarget struct {
	DisplayLabel string
	Names        []string
	ManualName   string
}

// ReplayMacroResult holds estimated macro metrics for one replay.
type ReplayMacroResult struct {
	Matched              bool
	SupplyBlockedSeconds int
	WorkerIdleSeconds    int
	SupplyChart          []int
	WorkerChart          []int
}

// MacroSummary holds the aggregated results shown in the UI.
type MacroSummary struct {
	TargetLabel               string
	ScannedReplays            int
	MatchedReplays            int
	SkippedReplays            int
	TotalSupplyBlockedSeconds int
	TotalWorkerIdleSeconds    int
	AvgSupplyBlockedSeconds   float64
	AvgWorkerIdleSeconds      float64
	SupplyRating              string
	WorkerRating              string
	SupplyChart               []int
	WorkerChart               []int
}
