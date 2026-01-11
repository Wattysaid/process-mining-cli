# Roadmap

**Document version:** R1.00 (2026-01-11)

## MVP (v0.1)
- Go CLI skeleton + installer
- File connectors: CSV/Parquet
- Project scaffold and run folder structure
- Data prep pipeline (interactive)
- Core mining: DFG, Inductive, Heuristic
- Conformance: optional alignments (guarded by compute warning)
- Notebook generation (template) and report generation (md/html)
- QA pack
- Optional OpenAI narrative with cost caps

## v0.2
- DB connectors (read-only)
- Drift analysis (time windows)
- Better packaging: offline wheels support
- Self-update
- Policy controls (disable OpenAI, restrict samples)

## v0.3
- Predictive monitoring (remaining time, SLA risk) with model cards
- Simulation/what-if (bounded)
- Export packs for ARIS/Signavio style consumption (where feasible)

## v1.0
- Enterprise installer variants
- Signed releases, audit-grade manifests
- Pluggable connectors and step marketplace

