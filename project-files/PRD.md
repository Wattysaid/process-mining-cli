# PRD: PM Assist CLI

**Document version:** R1.00 (2026-01-11)  
**Owner:** Product / Engineering  
**Status:** Draft  
**Primary KPI:** Time-to-first-insight for a new dataset (target: < 60 minutes for a competent analyst)

## 1. Problem statement
Enterprise process mining programmes often require:
- multiple specialist roles (data engineering, data science, process mining, reporting, stakeholder comms)
- expensive platforms and long implementation cycles
- fragmented tooling and non-reproducible, analyst-specific notebooks

PM Assist reduces time and dependency on large teams by providing a guided, reproducible CLI workflow that produces standard artefacts (event logs, models, notebooks, reports) and uses optional LLM support for orchestration and narrative work (OpenAI, Anthropic, Gemini, or local Ollama). The tool never makes analytical decisions on behalf of the user; it prompts for each decision and adapts guidance to user aptitude.

## 2. Target users
Primary:
- Process mining analysts, process architects, business analysts, transformation consultants
Secondary:
- Data engineers supporting read-only extraction
- Enterprise architects and governance teams validating outputs

## 3. Use cases (MVP)
- Import event data from CSV/Parquet and validate readiness
- Clean and standardise into a canonical event log schema
- Perform discovery (DFG, Inductive, Heuristic), conformance (alignments), performance analysis, variants
- Generate:
  - a notebook (template-based, with placeholders for findings)
  - an executive report (templated narrative + charts)
- Run a QA pack (schema checks, duplicates, sorting, timestamp sanity, basic statistical checks)

## 4. Use cases (post-MVP)
- Read-only connectors: Postgres/SQL Server/Snowflake/BigQuery, S3/ADLS, SharePoint
- Drift detection across windows
- Predictive monitoring (remaining time, SLA breach risk)
- Simulation / what-if (bounded scope, explicit assumptions)
- Export to BPMN / collaboration artefacts for process teams

## 5. Non-functional requirements
- Installability: curl installer, minimal dependencies, deterministic versions; first-run creates a project-local `.venv`
- Security: secrets never logged, least privilege, auditable runs
- Performance: handle large datasets via chunking and columnar storage
- Portability: Linux/macOS/Windows (native where feasible; WSL2 acceptable for MVP)
- Extensibility: pipeline steps are plug-in style and backed by the `cli-tool-skills` Python library; new steps can be added without rewriting CLI
- IP protection: distribute as a standalone app with bundled Python assets; do not expose source or allow user edits to the tool code

## 6. Competitive differentiation (vs Celonis/Signavio/ARIS)
- Faster start, lower overhead, offline-capable, local-first
- Artefact-centric delivery: notebooks and reports are standard outputs
- Better for consulting engagements and internal analytics teams that want control
- Integrates with existing data science stack and enterprise repositories

## 7. Pricing and packaging assumptions (for engineering decisions)
- Enterprise licence: per seat and/or per environment
- Deployment modes: local, on-prem, VPC, optionally managed SaaS later
- Feature gating: connectors, predictive modules, and agent features can be licensed separately

## 8. MVP acceptance criteria
- A new user can run: init -> ingest -> prepare -> mine -> report
- Outputs are produced in a run folder with a stable structure
- LLM integration is optional and requires explicit credentials + opt-in per run (or local Ollama)
- QA report clearly flags data issues and model caveats
- User profile is captured and used to adjust prompt depth without changing decision rules
- Business profile is captured for repeatable setup across engagements
- Notebook is appended with each user-confirmed step for transparent reproducibility
