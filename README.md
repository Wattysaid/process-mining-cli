# PM Assist CLI (Process Mining Assistant)

**Document version:** R1.00 (2026-01-11)  
**Audience:** Engineering team (Codex), product, security, delivery  
**Goal:** Build an enterprise-grade CLI that enables a single practitioner to run an end-to-end process mining engagement (data engineering through mining through reporting), with guided workflows and optional OpenAI-powered assistance that is token-cost controlled.

## What this repo provides (instructions only)
This repository folder contains **MD instruction files** that an AI coding agent can use to implement the application. You (the human) will add your existing Python scripts and skill library separately; the agent should decide what to keep, refactor, or replace.

## Product positioning
PM Assist is a CLI-based process mining copilot that competes with Celonis, Signavio and ARIS on *time-to-insight* for single analysts and small teams, while remaining deployable inside regulated enterprise environments (on-prem, VPC, private networks).

## Key principles
- **User-led decisions**: the CLI must ask questions and present options; it must not make unilateral analytical decisions.
- **Reproducibility**: each run produces a deterministic artefact bundle (config, logs, notebooks, reports, model outputs).
- **Enterprise hygiene**: secure secret handling, audit logs, data minimisation, read-only connectors by default.
- **Cost control**: OpenAI usage is optional, measurable, capped, and used mainly for orchestration and narrative generation, not heavy computation.
- **Composable pipelines**: pipelines are built from reusable Python snippets and modules (data science + process mining).

## High-level workflow (happy path)
1. `pm-assist init` creates a project scaffold.
2. `pm-assist connect` registers read-only data sources (CSV/Parquet first; DB and S3/SharePoint later).
3. `pm-assist ingest` imports data and validates schema.
4. `pm-assist prepare` runs data prep (missingness, type fixes, duplicates, outliers, normalisation, encoding, text cleanup, date features).
5. `pm-assist mine` performs discovery, conformance, performance, variants, drift, and (optional) predictive monitoring.
6. `pm-assist report` generates:
   - a Jupyter notebook (executed or unexecuted, configurable)
   - an executive report (Markdown/HTML/PDF roadmap)
   - an artefact bundle for auditability
7. `pm-assist review` runs automated checks and produces a QA summary.
8. Optional: `pm-assist agent` provides guided Q&A and narrative drafting using OpenAI.

## Project outputs (per run)
- `./outputs/<run-id>/`
  - `config.yaml` (resolved configuration snapshot)
  - `run.log` (structured log)
  - `data_profile/` (optional profiling reports)
  - `event_log/` (canonical event log in parquet)
  - `models/` (DFG, Petri, BPMN exports)
  - `analysis_notebook.ipynb`
  - `report.md` + `report.html` (+ pdf if enabled)
  - `quality/` (validation and test results)

## Success criteria
- Install via a single command (curl installer) on Linux/macOS; Windows via WSL2 is supported.
- CLI is intuitive, guides the user, and fails safely with actionable messages.
- Pipelines can run fully offline except for optional OpenAI calls.
- The implementation supports enterprise governance: secrets, audit logs, deterministic outputs.

## Required deliverables for MVP
- CLI binary (recommended: Go) that orchestrates Python pipelines in a managed venv
- Installer script (`curl | sh`)
- OpenAI integration (optional, gated by explicit opt-in)
- Template-based notebook and report generation
- Data validation + QA checks
- Extensible skill library (`.codex/skills/…`) used by Codex while coding, and optionally by the product for runtime guidance

## Where to start
Read these files in order:
1. `project-files/PRD.md`
2. `project-files/ARCHITECTURE.md`
3. `project-files/CLI_SPEC.md`
4. `project-files/INSTALLATION_AND_RELEASE.md`
5. `project-files/OPENAI_INTEGRATION.md`
6. `project-files/WORKFLOWS.md`
7. `.codex/agent.md` and `.codex/skills/*/skill.md`

---

## Notes on your existing assets
You already have an end-to-end pm4py notebook and guidance content. The agent should extract reusable code blocks and convert them into modular pipeline steps (see `project-files/WORKFLOWS.md` and `project-files/NOTEBOOK_AND_REPORTS.md`). 

