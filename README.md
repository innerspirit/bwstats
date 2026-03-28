# bwstats
BWStats is a small Windows desktop tool for analyzing StarCraft: Brood War / Remastered replay files.

## What it does
- auto-detects the Windows replay autosave folder at `~/Documents/StarCraft/Maps/Replays/AutoSave`
- reads `CSettings.json` from `~/Documents/StarCraft` and uses all `Gateway History` accounts as the current player's aliases
- lets you override that with a manual player name input
- scans matching replays and estimates two macro metrics:
  - supply-block time
  - worker-production idle time until the replay first reaches 60 workers
- shows a compact summary with ratings plus two small charts for the first 15 minutes:
  - supply-block seconds per 30-second bucket
  - worker-idle seconds per 30-second bucket
- shows progress and sends a desktop notification when the scan completes

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
- metrics are command-based estimates, not exact reconstructed game state.
- supply uses Brood War rules, not StarCraft II rules.
- worker idle stops being counted after a replay first reaches 60 workers, even if worker count later drops.
