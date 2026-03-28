package main

import (
	"testing"

	screp "github.com/icza/screp/rep"
	"github.com/icza/screp/rep/repcmd"
	"github.com/icza/screp/rep/repcore"
)

func TestAnalyzeReplaySupplyBlockLateDepot(t *testing.T) {
	rep := terranReplayWithCommands([]timedCmd{
		buildWorker(0),
		buildWorker(13),
		buildWorker(26),
		buildWorker(39),
		buildWorker(52),
		buildWorker(65),
		buildWorker(78),
		buildBuilding(60, repcmd.UnitIDSupplyDepot),
	}, 120)

	player := rep.Header.Players[0]
	result := analyzeMatchedReplay(rep, player)

	if result.SupplyBlockedSeconds == 0 {
		t.Fatalf("expected supply block time, got 0")
	}
	if result.SupplyChart[2] == 0 {
		t.Fatalf("expected blocked seconds in 60-90s bucket, got %v", result.SupplyChart)
	}
}

func TestAnalyzeReplaySupplyBlockTimelyDepot(t *testing.T) {
	rep := terranReplayWithCommands([]timedCmd{
		buildWorker(0),
		buildWorker(13),
		buildWorker(26),
		buildWorker(39),
		buildWorker(52),
		buildWorker(65),
		buildWorker(78),
		buildBuilding(40, repcmd.UnitIDSupplyDepot),
	}, 120)

	player := rep.Header.Players[0]
	result := analyzeMatchedReplay(rep, player)

	if result.SupplyBlockedSeconds != 0 {
		t.Fatalf("expected no supply block time, got %d", result.SupplyBlockedSeconds)
	}
}

func TestAnalyzeReplayWorkerIdleAndChart(t *testing.T) {
	rep := terranReplayWithCommands([]timedCmd{
		buildWorker(30),
	}, 120)

	player := rep.Header.Players[0]
	result := analyzeMatchedReplay(rep, player)

	if result.WorkerIdleSeconds < 30 {
		t.Fatalf("expected worker idle time before first worker, got %d", result.WorkerIdleSeconds)
	}
	if result.WorkerChart[0] == 0 {
		t.Fatalf("expected worker idle time in first chart bucket, got %v", result.WorkerChart)
	}
}

func TestAnalyzeReplayStopsWorkerIdleAfterSixtyWorkers(t *testing.T) {
	var cmds []timedCmd
	for second := 0; second < 13*56; second += 13 {
		cmds = append(cmds, buildWorker(second))
	}
	for _, second := range []int{40, 150, 256, 360, 464, 568, 672} {
		cmds = append(cmds, buildBuilding(second, repcmd.UnitIDSupplyDepot))
	}

	rep := terranReplayWithCommands(cmds, 1000)
	player := rep.Header.Players[0]
	result := analyzeMatchedReplay(rep, player)

	if result.WorkerIdleSeconds != 0 {
		t.Fatalf("expected no worker idle after continuous worker production to 60, got %d", result.WorkerIdleSeconds)
	}
}

func TestAggregateMacroResults(t *testing.T) {
	first := ReplayMacroResult{
		Matched:              true,
		SupplyBlockedSeconds: 12,
		WorkerIdleSeconds:    24,
		SupplyChart:          chartSeriesWithValue(2, 12),
		WorkerChart:          chartSeriesWithValue(1, 24),
	}
	second := ReplayMacroResult{
		Matched:              true,
		SupplyBlockedSeconds: 18,
		WorkerIdleSeconds:    36,
		SupplyChart:          chartSeriesWithValue(2, 18),
		WorkerChart:          chartSeriesWithValue(1, 36),
	}

	summary := aggregateMacroResults(ScanTarget{DisplayLabel: "alpha"}, []ReplayMacroResult{first, second}, 1)

	if summary.MatchedReplays != 2 {
		t.Fatalf("expected 2 matched replays, got %d", summary.MatchedReplays)
	}
	if summary.TotalSupplyBlockedSeconds != 30 {
		t.Fatalf("expected 30 total blocked seconds, got %d", summary.TotalSupplyBlockedSeconds)
	}
	if summary.AvgWorkerIdleSeconds != 30 {
		t.Fatalf("expected 30 average worker idle seconds, got %.2f", summary.AvgWorkerIdleSeconds)
	}
	if summary.SupplyChart[2] != 30 {
		t.Fatalf("expected aggregated supply chart bucket, got %v", summary.SupplyChart)
	}
	if summary.WorkerRating == "" {
		t.Fatalf("expected worker rating to be populated")
	}
}

func terranReplayWithCommands(cmds []timedCmd, durationSeconds int) *screp.Replay {
	player := &screp.Player{
		ID:   1,
		Name: "alpha",
		Race: repcore.RaceTerran,
		Type: repcore.PlayerTypeHuman,
	}

	replayCmds := make([]repcmd.Cmd, 0, len(cmds))
	for _, cmd := range cmds {
		replayCmds = append(replayCmds, cmd.toCmd(player.ID))
	}

	return &screp.Replay{
		Header: &screp.Header{
			Players: []*screp.Player{player},
			Frames:  secondFrame(durationSeconds),
			Type:    repcore.GameTypeMelee,
		},
		Commands: &screp.Commands{Cmds: replayCmds},
	}
}

type timedCmd struct {
	second int
	kind   byte
	unitID uint16
}

func buildWorker(second int) timedCmd {
	return timedCmd{second: second, kind: repcmd.TypeIDTrain, unitID: unitIDSCV}
}

func buildBuilding(second int, unitID uint16) timedCmd {
	return timedCmd{second: second, kind: repcmd.TypeIDBuild, unitID: unitID}
}

func (c timedCmd) toCmd(playerID byte) repcmd.Cmd {
	base := &repcmd.Base{
		Frame:    secondFrame(c.second),
		PlayerID: playerID,
	}

	switch c.kind {
	case repcmd.TypeIDBuild:
		base.Type = repcmd.TypeBuild
		return &repcmd.BuildCmd{
			Base: base,
			Unit: repcmd.UnitByID(c.unitID),
		}
	default:
		base.Type = repcmd.TypeTrain
		return &repcmd.TrainCmd{
			Base: base,
			Unit: repcmd.UnitByID(c.unitID),
		}
	}
}

func chartSeriesWithValue(index, value int) []int {
	series := make([]int, chartBucketCount)
	series[index] = value
	return series
}

func secondFrame(second int) repcore.Frame {
	return repcore.Frame((second*1000 + 41) / 42)
}
