package main

import (
	"fmt"
	"strings"
	"time"

	screp "github.com/icza/screp/rep"
	"github.com/icza/screp/rep/repcmd"
	"github.com/icza/screp/rep/repcore"
	"github.com/icza/screp/repparser"
)

const (
	chartWindowSeconds = 15 * 60
	chartBucketSeconds = 30
	chartBucketCount   = chartWindowSeconds / chartBucketSeconds

	supplyGreatThreshold = 15.0
	supplySolidThreshold = 45.0
	workerGreatThreshold = 45.0
	workerSolidThreshold = 120.0
)

const (
	unitIDMarine        = 0x00
	unitIDGhost         = 0x01
	unitIDVulture       = 0x02
	unitIDGoliath       = 0x03
	unitIDSiegeTankTM   = 0x05
	unitIDSCV           = 0x07
	unitIDWraith        = 0x08
	unitIDScienceVessel = 0x09
	unitIDDropship      = 0x0B
	unitIDBattlecruiser = 0x0C
	unitIDFirebat       = 0x20
	unitIDMedic         = 0x22
	unitIDZergling      = 0x25
	unitIDHydralisk     = 0x26
	unitIDUltralisk     = 0x27
	unitIDDrone         = 0x29
	unitIDOverlord      = 0x2A
	unitIDMutalisk      = 0x2B
	unitIDGuardian      = 0x2C
	unitIDQueen         = 0x2D
	unitIDDefiler       = 0x2E
	unitIDScourge       = 0x2F
	unitIDValkyrie      = 0x3A
	unitIDCorsair       = 0x3C
	unitIDDarkTemplar   = 0x3D
	unitIDDevourer      = 0x3E
	unitIDDarkArchon    = 0x3F
	unitIDProbe         = 0x40
	unitIDZealot        = 0x41
	unitIDDragoon       = 0x42
	unitIDHighTemplar   = 0x43
	unitIDArchon        = 0x44
	unitIDShuttle       = 0x45
	unitIDScout         = 0x46
	unitIDArbiter       = 0x47
	unitIDCarrier       = 0x48
	unitIDReaver        = 0x53
	unitIDObserver      = 0x54
	unitIDLurker        = 0x67
)

type raceConfig struct {
	startingSupplyHalf int
	startingWorkers    int
	workerUnitID       uint16
	workerTrainSeconds int
	baseProducerCount  int
}

type commandEvent struct {
	second  int
	command repcmd.Cmd
}

type scheduledEvent struct {
	second int
	kind   string
	unitID uint16
}

var raceConfigs = map[byte]raceConfig{
	repcore.RaceTerran.ID: {
		startingSupplyHalf: 20,
		startingWorkers:    4,
		workerUnitID:       unitIDSCV,
		workerTrainSeconds: 13,
		baseProducerCount:  1,
	},
	repcore.RaceProtoss.ID: {
		startingSupplyHalf: 18,
		startingWorkers:    4,
		workerUnitID:       unitIDProbe,
		workerTrainSeconds: 13,
		baseProducerCount:  1,
	},
	repcore.RaceZerg.ID: {
		startingSupplyHalf: 18,
		startingWorkers:    4,
		workerUnitID:       unitIDDrone,
		workerTrainSeconds: 13,
		baseProducerCount:  1,
	},
}

var supplyProvidedHalf = map[uint16]int{
	repcmd.UnitIDCommandCenter: 20,
	repcmd.UnitIDSupplyDepot:   16,
	repcmd.UnitIDNexus:         18,
	repcmd.UnitIDPylon:         16,
	repcmd.UnitIDHatchery:      2,
	repcmd.UnitIDLair:          2,
	repcmd.UnitIDHive:          2,
	unitIDOverlord:             16,
}

var unitSupplyCostHalf = map[uint16]int{
	unitIDSCV:           2,
	unitIDMarine:        2,
	unitIDGhost:         2,
	unitIDVulture:       4,
	unitIDGoliath:       4,
	unitIDSiegeTankTM:   4,
	unitIDWraith:        4,
	unitIDScienceVessel: 4,
	unitIDDropship:      4,
	unitIDBattlecruiser: 12,
	unitIDFirebat:       2,
	unitIDMedic:         2,
	unitIDValkyrie:      6,
	unitIDDrone:         2,
	unitIDZergling:      1,
	unitIDHydralisk:     2,
	unitIDUltralisk:     8,
	unitIDMutalisk:      4,
	unitIDGuardian:      4,
	unitIDQueen:         4,
	unitIDDefiler:       4,
	unitIDScourge:       1,
	unitIDLurker:        4,
	unitIDDevourer:      4,
	unitIDProbe:         2,
	unitIDZealot:        4,
	unitIDDragoon:       4,
	unitIDHighTemplar:   4,
	unitIDDarkTemplar:   4,
	unitIDArchon:        8,
	unitIDShuttle:       4,
	unitIDScout:         6,
	unitIDArbiter:       8,
	unitIDCarrier:       12,
	unitIDReaver:        8,
	unitIDObserver:      2,
	unitIDCorsair:       4,
	unitIDDarkArchon:    8,
}

var unitBuildSeconds = map[uint16]int{
	unitIDSCV:                  13,
	unitIDProbe:                13,
	unitIDDrone:                13,
	repcmd.UnitIDSupplyDepot:   30,
	repcmd.UnitIDPylon:         25,
	unitIDOverlord:             25,
	repcmd.UnitIDCommandCenter: 100,
	repcmd.UnitIDNexus:         100,
	repcmd.UnitIDHatchery:      120,
	repcmd.UnitIDLair:          80,
	repcmd.UnitIDHive:          120,
}

// resolveScanTarget chooses auto aliases or a manual player name override.
func resolveScanTarget(identity PlayerIdentity, manualName string) ScanTarget {
	manualName = strings.TrimSpace(manualName)
	if manualName != "" {
		return ScanTarget{
			DisplayLabel: manualName,
			Names:        []string{manualName},
			ManualName:   manualName,
		}
	}

	display := identity.DisplayName
	if display == "" {
		display = "Current Player"
	}

	label := display
	if len(identity.Aliases) > 0 {
		label = fmt.Sprintf("Current Player (%s)", strings.Join(identity.Aliases, ", "))
	}

	return ScanTarget{
		DisplayLabel: label,
		Names:        append([]string(nil), identity.Aliases...),
	}
}

// scanMacroStats scans replays for the selected target and aggregates macro metrics.
func scanMacroStats(target ScanTarget, progressCallback func(float64)) (*MacroSummary, error) {
	repFiles, err := findReplayFiles(progressCallback)
	if err != nil {
		return nil, err
	}

	results := make([]ReplayMacroResult, 0, len(repFiles))
	skipped := 0

	for index, repFile := range repFiles {
		cfg := repparser.Config{Commands: true}
		rep, err := repparser.ParseFileConfig(repFile, cfg)
		if err != nil {
			skipped++
		} else {
			player := findMatchingPlayer(rep, target.Names)
			if player != nil {
				results = append(results, analyzeMatchedReplay(rep, player))
			}
		}

		if progressCallback != nil && len(repFiles) > 0 {
			progressCallback(float64(index+1) / float64(len(repFiles)))
		}
	}

	summary := aggregateMacroResults(target, results, skipped)
	summary.ScannedReplays = len(repFiles)
	return summary, nil
}

func findMatchingPlayer(rep *screp.Replay, names []string) *screp.Player {
	if rep == nil || rep.Header == nil {
		return nil
	}

	for _, player := range rep.Header.Players {
		for _, name := range names {
			if player.Name == name {
				return player
			}
		}
	}

	return nil
}

func analyzeMatchedReplay(rep *screp.Replay, player *screp.Player) ReplayMacroResult {
	result := ReplayMacroResult{
		Matched:     true,
		SupplyChart: make([]int, chartBucketCount),
		WorkerChart: make([]int, chartBucketCount),
	}
	if rep == nil || rep.Header == nil || rep.Commands == nil || player == nil || player.Race == nil {
		return result
	}

	config, ok := raceConfigs[player.Race.ID]
	if !ok {
		return result
	}

	events := groupCommandsBySecond(rep.Commands.Cmds, player.ID)
	duration := replayDurationSeconds(rep)
	if duration <= 0 {
		return result
	}

	state := replayState{
		config:                config,
		availableSupplyHalf:   config.startingSupplyHalf,
		usedSupplyHalf:        config.startingWorkers * 2,
		workerCount:           config.startingWorkers,
		workerProducerCount:   config.baseProducerCount,
		reachedWorkerCutoff:   false,
		activeWorkerTrainEnds: nil,
		pendingEvents:         map[int][]scheduledEvent{},
	}

	for second := 0; second <= duration; second++ {
		state.applyScheduledEvents(second)
		state.handleCommands(second, events[second])

		if state.isSupplyBlocked(second) {
			result.SupplyBlockedSeconds++
			addChartSecond(result.SupplyChart, second)
		}
		if state.isWorkerIdle() {
			result.WorkerIdleSeconds++
			addChartSecond(result.WorkerChart, second)
		}
	}

	return result
}

type replayState struct {
	config                raceConfig
	availableSupplyHalf   int
	usedSupplyHalf        int
	workerCount           int
	workerProducerCount   int
	reachedWorkerCutoff   bool
	supplyBlockedUntil    int
	activeWorkerTrainEnds []int
	pendingEvents         map[int][]scheduledEvent
}

func (s *replayState) handleCommands(second int, commands []commandEvent) {
	for _, event := range commands {
		switch cmd := event.command.(type) {
		case *repcmd.BuildCmd:
			s.handleBuildCommand(second, cmd.Unit.ID)
		case *repcmd.TrainCmd:
			s.handleTrainCommand(second, cmd.Unit.ID)
		case *repcmd.BuildingMorphCmd:
			s.handleBuildCompletion(second, cmd.Unit.ID)
		}
	}
}

func (s *replayState) handleBuildCommand(second int, unitID uint16) {
	if unitID == repcmd.UnitIDHatchery || unitID == repcmd.UnitIDLair || unitID == repcmd.UnitIDHive {
		// Drone morph frees its occupied supply when it starts a building.
		if s.config.workerUnitID == unitIDDrone && s.usedSupplyHalf >= 2 {
			s.usedSupplyHalf -= 2
			if s.workerCount > 0 {
				s.workerCount--
			}
		}
	}

	if buildSeconds, ok := unitBuildSeconds[unitID]; ok {
		s.schedule(second+buildSeconds, scheduledEvent{second: second + buildSeconds, kind: "supply", unitID: unitID})
		if unitID == repcmd.UnitIDCommandCenter || unitID == repcmd.UnitIDNexus || unitID == repcmd.UnitIDHatchery {
			s.schedule(second+buildSeconds, scheduledEvent{second: second + buildSeconds, kind: "producer", unitID: unitID})
		}
	}
}

func (s *replayState) handleTrainCommand(second int, unitID uint16) {
	supplyCost := unitSupplyCostHalf[unitID]
	if supplyCost > 0 && s.usedSupplyHalf >= s.availableSupplyHalf {
		s.supplyBlockedUntil = s.nextSupplyReliefSecond(second)
	}

	buildSeconds := unitBuildSeconds[unitID]
	if buildSeconds == 0 {
		buildSeconds = 30
	}

	if unitID == s.config.workerUnitID {
		if s.activeWorkerTrains() < s.workerProducerCount && s.usedSupplyHalf+supplyCost <= s.availableSupplyHalf {
			endSecond := second + buildSeconds
			s.activeWorkerTrainEnds = append(s.activeWorkerTrainEnds, endSecond)
			s.schedule(endSecond, scheduledEvent{second: endSecond, kind: "worker", unitID: unitID})
			s.schedule(endSecond, scheduledEvent{second: endSecond, kind: "supply-used", unitID: unitID})
		}
		return
	}

	if supplyCost > 0 {
		s.schedule(second+buildSeconds, scheduledEvent{second: second + buildSeconds, kind: "supply-used", unitID: unitID})
	}

	if unitID == unitIDOverlord {
		s.schedule(second+buildSeconds, scheduledEvent{second: second + buildSeconds, kind: "supply", unitID: unitID})
	}
}

func (s *replayState) handleBuildCompletion(second int, unitID uint16) {
	if unitID == repcmd.UnitIDLair || unitID == repcmd.UnitIDHive {
		s.schedule(second+unitBuildSeconds[unitID], scheduledEvent{second: second + unitBuildSeconds[unitID], kind: "supply", unitID: unitID})
	}
}

func (s *replayState) applyScheduledEvents(second int) {
	for _, event := range s.pendingEvents[second] {
		switch event.kind {
		case "supply":
			s.availableSupplyHalf += supplyProvidedHalf[event.unitID]
			if second >= s.supplyBlockedUntil {
				s.supplyBlockedUntil = 0
			}
		case "supply-used":
			s.usedSupplyHalf += unitSupplyCostHalf[event.unitID]
		case "worker":
			s.workerCount++
			s.activeWorkerTrainEnds = removeOneSecond(s.activeWorkerTrainEnds, second)
			if s.workerCount >= 60 {
				s.reachedWorkerCutoff = true
			}
		case "producer":
			s.workerProducerCount++
		}
	}
	delete(s.pendingEvents, second)
}

func (s *replayState) schedule(second int, event scheduledEvent) {
	s.pendingEvents[second] = append(s.pendingEvents[second], event)
}

func (s *replayState) activeWorkerTrains() int {
	return len(s.activeWorkerTrainEnds)
}

func (s *replayState) isSupplyBlocked(second int) bool {
	return s.supplyBlockedUntil > second
}

func (s *replayState) isWorkerIdle() bool {
	return !s.reachedWorkerCutoff && s.workerProducerCount > 0 && s.activeWorkerTrains() == 0
}

func (s *replayState) nextSupplyReliefSecond(current int) int {
	next := current + 1
	found := false
	for second, events := range s.pendingEvents {
		if second <= current {
			continue
		}
		for _, event := range events {
			if event.kind == "supply" && supplyProvidedHalf[event.unitID] > 0 {
				if !found || second < next {
					next = second
					found = true
				}
			}
		}
	}
	return next
}

func aggregateMacroResults(target ScanTarget, results []ReplayMacroResult, skippedReplays int) *MacroSummary {
	summary := &MacroSummary{
		TargetLabel:    target.DisplayLabel,
		SkippedReplays: skippedReplays,
		SupplyChart:    make([]int, chartBucketCount),
		WorkerChart:    make([]int, chartBucketCount),
	}

	for _, result := range results {
		if !result.Matched {
			continue
		}
		summary.MatchedReplays++
		summary.TotalSupplyBlockedSeconds += result.SupplyBlockedSeconds
		summary.TotalWorkerIdleSeconds += result.WorkerIdleSeconds
		addSeries(summary.SupplyChart, result.SupplyChart)
		addSeries(summary.WorkerChart, result.WorkerChart)
	}

	if summary.MatchedReplays > 0 {
		summary.AvgSupplyBlockedSeconds = float64(summary.TotalSupplyBlockedSeconds) / float64(summary.MatchedReplays)
		summary.AvgWorkerIdleSeconds = float64(summary.TotalWorkerIdleSeconds) / float64(summary.MatchedReplays)
		summary.SupplyRating = rateMetric(summary.AvgSupplyBlockedSeconds, supplyGreatThreshold, supplySolidThreshold)
		summary.WorkerRating = rateMetric(summary.AvgWorkerIdleSeconds, workerGreatThreshold, workerSolidThreshold)
	} else {
		summary.SupplyRating = "No Data"
		summary.WorkerRating = "No Data"
	}

	return summary
}

func rateMetric(avg, great, solid float64) string {
	switch {
	case avg <= great:
		return "Great"
	case avg <= solid:
		return "Solid"
	default:
		return "Needs Work"
	}
}

func groupCommandsBySecond(cmds []repcmd.Cmd, playerID byte) map[int][]commandEvent {
	result := make(map[int][]commandEvent)
	for _, cmd := range cmds {
		if cmd.BaseCmd().PlayerID != playerID {
			continue
		}
		second := int(cmd.BaseCmd().Frame.Duration().Seconds())
		result[second] = append(result[second], commandEvent{second: second, command: cmd})
	}
	return result
}

func replayDurationSeconds(rep *screp.Replay) int {
	if rep == nil || rep.Header == nil {
		return 0
	}
	return int(rep.Header.Frames.Duration().Seconds())
}

func addChartSecond(series []int, second int) {
	if second < 0 || second >= chartWindowSeconds {
		return
	}
	series[second/chartBucketSeconds]++
}

func addSeries(dst, src []int) {
	for i := range dst {
		if i < len(src) {
			dst[i] += src[i]
		}
	}
}

func removeOneSecond(values []int, target int) []int {
	for i, value := range values {
		if value == target {
			return append(values[:i], values[i+1:]...)
		}
	}
	return values
}

func seconds(n int) time.Duration {
	return time.Duration(n) * time.Second
}
