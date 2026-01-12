# Implementation Plan

**Document version:** R1.01 (2026-01-12)

## Phase 1: Core CLI and Orchestration
- Cobra command tree with global flags
- Startup splash screen and first-run bootstrap
- User and business profiles
- Project scaffolding templates
- Config load/save
- Bubble Tea prompt UI (lists, inputs, spinners, progress, file picker)
- Status dashboard and wizard entrypoint

Status: **Complete**

## Phase 2: Data Access and Ingest
- File connectors (CSV/Parquet)
- DB connector capture + read-only validation (Postgres/MySQL/MSSQL/Snowflake/BigQuery)
- Ingest pipeline wired to `cli-tool-skills`

Status: **Complete**

## Phase 3: Preparation and Mining
- Data quality + clean/filter flows
- EDA, discovery, conformance, performance
- Notebook append for each confirmed step

Status: **Complete**

## Phase 4: Reporting and QA
- Report generation
- QA review pack and backlog outputs

Status: **Complete**

## Phase 5: Enterprise Hardening
- Installer + self-update
- Packaging for IP protection
- Policy controls (disable LLM, connector restrictions)
- Auditable manifests and checksums

Status: **In progress** (Go 1.23 toolchain required; offline wheels/asset bundling pending)
