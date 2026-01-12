# Roadmap

**Document version:** R1.01 (2026-01-12)

## Current build (v0.1.x in repo)
- Bubble Tea prompt UI (lists, inputs, spinners, progress, file picker)
- Wizard flow (`pm-assist start`) + status dashboard (`pm-assist status`)
- Connector list command (`pm-assist connectors list`)
- DB connectors with read-only validation: Postgres/MySQL/MSSQL/Snowflake/BigQuery
- Command framing + splash screens before/after commands

## MVP (v0.1)
- Go CLI skeleton + installer
- File connectors: CSV/Parquet
- Project scaffold and run folder structure
- Data prep pipeline (interactive)
- Core mining: DFG, Inductive, Heuristic
- Conformance: optional alignments (guarded by compute warning)
- Notebook generation (template) and report generation (md/html)
- QA pack
- Optional LLM narrative (OpenAI/Anthropic/Gemini/Ollama) with cost caps

## v0.2
- DB connectors (read-only)
- Drift analysis (time windows)
- Better packaging: offline wheels support
- Self-update
- Policy controls (disable LLM, restrict samples)

## v0.2.1 (hardening)
- Windows native venv handling
- Non-interactive TUI fallback (TTY detection)
- Sanitized profile/business lookups in `set/show`

## v0.3
- Predictive monitoring (remaining time, SLA risk) with model cards
- Simulation/what-if (bounded)
- Export packs for ARIS/Signavio style consumption (where feasible)

## v1.0
- Enterprise installer variants
- Signed releases, audit-grade manifests
- Pluggable connectors and step marketplace
