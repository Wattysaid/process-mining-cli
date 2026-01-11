# Implementation Status

**Document version:** R1.00 (2026-01-11)

## Completed
- CLI commands: init, connect, ingest, prepare, mine, report, review
- Startup splash + first-run bootstrap
- User profiles and business profiles
- File connectors + CSV preview
- DB connectors for Postgres/MySQL/MSSQL (read-only validation + schema/table listing)
- Notebook append per confirmed step
- QA pack outputs (md/json/csv)
- Installer script + self-update command
- Release workflow (builds tarballs + checksums)
- Skills path resolution with packaging fallback
- Policy controls (LLM enable/disable, offline-only, connector allow/deny)
- Structured logging + run manifest (config snapshot, inputs/outputs, hashes, step status)
- HTML report export (PDF export when pandoc is available)
- Config schema versioning + validation
- Non-interactive flags across core commands
- Column mapping (map command) with config persistence
- Report bundle output (`outputs/<run-id>/bundle/report_bundle_<run-id>.zip`)
- Config migration scaffolding (version 0 -> current)

## Partially Complete
- Packaging of Python assets: bundled skills + wheels; offline installs require wheel verification/testing.
- Policy enforcement: applied to connect + agent setup; offline-only enforced for python deps; warnings remain in ingest/prepare/mine/report.
- Snowflake/BigQuery connectors: config capture + explicit unsupported validation; no live connectivity checks yet.
- Automated tests: unit tests added for config/manifest/QA; smoke run wired via `scripts/smoke.sh`.
- Signed release verification hooks (cosign) added to installer/self-update; release workflow signs checksums when secrets are set.

## Not Started / Missing
- Snowflake/BigQuery live reachability checks

## Build Notes
- Go 1.22+ required
- Builds currently fetch modules from the network
- `.cache/` is intentionally excluded from git
