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

## Partially Complete
- Packaging of Python assets: scripts exist, but integration into release workflow needs verification and tests.
- Policy enforcement: applied to connect + agent setup; warnings in ingest/prepare/mine/report; no hard enforcement yet.
- Doctor command: basic env checks only, lacks connector reachability checks.

## Not Started / Missing
- Structured logging + secrets redaction
- Run manifests with hashes, config snapshot, step status
- Non-interactive flags for all command prompts
- Snowflake/BigQuery connectors (capture + validation stubs)
- Report HTML/PDF output
- Config validation and schema versioning
- Automated tests + smoke run
- Signed releases + verification

## Build Notes
- Go 1.22+ required
- Builds currently fetch modules from the network
- `.cache/` is intentionally excluded from git
