# Context Handoff

**Document version:** R1.00 (2026-01-11)

## Snapshot
- CLI commands implemented: `init`, `connect`, `ingest`, `prepare`, `mine`, `report`, `review`, `profile`, `business`, `agent setup`.
- Startup splash menu routes to `init`, `doctor`, `agent setup`, and profile management.
- Venv runner executes `cli-tool-skills` scripts only after explicit user confirmation.
- Notebook appends each confirmed step with Markdown + runnable command.
- Config is YAML (`pm-assist.yaml`) with connectors, profiles, business, and LLM provider.

## Key Paths
- CLI root: `internal/cli/root.go`
- Startup menu: `internal/cli/startup.go`
- Connect/Ingest/Prepare/Mine/Report/Review:
  - `internal/cli/commands/connect.go`
  - `internal/cli/commands/ingest.go`
  - `internal/cli/commands/prepare.go`
  - `internal/cli/commands/mine.go`
  - `internal/cli/commands/report.go`
  - `internal/cli/commands/review.go`
- Runner: `internal/runner/runner.go`
- QA pack: `internal/qa/qa.go`
- Notebook appender: `internal/notebook/notebook.go`
- Config model: `internal/config/config.go`

## Current Gaps
- DB connector validation for non-Postgres drivers
- Installer + self-update workflow
- Policy controls and license enforcement
