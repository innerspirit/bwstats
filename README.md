# bwstats
BWStats is a small desktop tool for analyzing StarCraft: Remastered and Brood War replay files.

## What it does
- auto-detects the Windows replay autosave folder at `~/Documents/StarCraft/Maps/Replays/AutoSave`
- reads `CSettings.json` from `~/Documents/StarCraft` to infer the current player nickname
- summarizes the player's race distribution: Terran, Zerg, Protoss
- collects per-game timeline counts for player building commands:
  - Supply Depots (Terran)
  - Overlords (Zerg)
  - Pylons (Protoss)
- shows progress and results in a simple GUI, and triggers desktop notifications when scan completes

## Running the app
1. Install Go 1.19+.
2. Run in project folder:
   - `go run .`

## Compiling and running tests
1. In project folder, run:
   - `go test ./...`
2. To build the executable:
   - `go build -o bwstats .`
3. Start the executable (Windows path assumptions still apply for replay folder):
   - `./bwstats`

## Notes
- replay folder path is currently Windows-centric (`USERPROFILE` based);
  on non-Windows systems, set `USERPROFILE` or adjust code for cross-platform paths.
- unit tests include data-path checks for replay file discovery and settings load,
  plus UI label update logic for race/building summaries.

