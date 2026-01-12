# Context Handoff

**Document version:** R1.01 (2026-01-12)

## Snapshot
- CLI commands implemented: `init`, `connect`, `ingest`, `map`, `prepare`, `mine`, `report`, `review`, `status`, `start`, `connectors list`, `doctor`, `self-update`, `profile`, `business`, `agent setup`.
- Splash screen shown on startup and after successful commands; command frames show purpose, step, outputs, and next step.
- Bubble Tea UI for prompts (arrow-key lists, text input, textarea, table, file picker, spinner, progress) with a theme file.
- Venv runner executes `cli-tool-skills` scripts only after explicit user confirmation.
- Notebook appends each confirmed step with Markdown + runnable command.
- Config is YAML (`pm-assist.yaml`) with connectors, profiles, business, policy, and LLM provider.
- Run manifests and structured logs are written under `outputs/<run-id>/`.
- Column mapping (`pm-assist map`) stores mappings in config.
- DB connectors include Postgres/MySQL/MSSQL/Snowflake/BigQuery with read-only validation and schema/table listing.

## Key Paths
- CLI root: `internal/cli/root.go`
- Startup menu: `internal/cli/startup.go`
- UI theme + widgets: `internal/ui/theme.go`, `internal/ui/tui.go`, `internal/ui/splash.go`, `internal/ui/frame.go`
- Connect/Ingest/Prepare/Mine/Report/Review:
  - `internal/cli/commands/connect.go`
  - `internal/cli/commands/ingest.go`
  - `internal/cli/commands/map.go`
  - `internal/cli/commands/prepare.go`
  - `internal/cli/commands/mine.go`
  - `internal/cli/commands/report.go`
  - `internal/cli/commands/review.go`
- Status + wizard:
  - `internal/cli/commands/status.go`
  - `internal/cli/commands/start.go`
- Connector list:
  - `internal/cli/commands/connectors.go`
- Runner: `internal/runner/runner.go`
- QA pack: `internal/qa/qa.go`
- Notebook appender: `internal/notebook/notebook.go`
- Config model: `internal/config/config.go`
- Snowflake/BigQuery: `internal/db/snowflake.go`, `internal/db/bigquery.go`

## Tracking
- Task backlog: `project-files/tasks/`
- Review findings: `Review/` (use `Review/INDEX.md` as entry point)

## Current Gaps
- Windows venv path handling (currently Linux/macOS `.venv/bin/*` only).
- Non-interactive safety: disable Bubble Tea prompts when stdin/stdout are not TTYs.
- Profile/business `set/show` should resolve sanitized filenames.
- Directory picker should support directory selection (not file-only).
- Run-id reuse behavior when not explicitly provided.
