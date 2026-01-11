# Codex Agent Instructions: PM Assist CLI

**Document version:** R1.00 (2026-01-11)

## Operating mode
You are an AI coding agent implementing an enterprise-grade CLI and pipeline orchestrator for process mining.

### Primary objectives
- Implement the CLI and workflows as specified in:
  - `project-files/PRD.md`
  - `project-files/ARCHITECTURE.md`
  - `project-files/CLI_SPEC.md`
  - `project-files/WORKFLOWS.md`
- Integrate optional LLM capability (OpenAI/Anthropic/Gemini/Ollama) with strict cost controls and secure handling.

### Constraints
- The CLI must ask the user questions before:
  - dropping or imputing data
  - selecting algorithms
  - choosing thresholds
  - making external calls (LLM providers)
- Must be reproducible and deterministic given the same inputs.
- Must be safe by default (offline, no secrets in logs).

## Implementation plan (MECE)
1) Bootstrap repo structure and build system
2) Implement core CLI commands: version, doctor, init
3) Implement run management and config resolution
4) Implement python venv management and python runner
5) Implement connectors: file (CSV/Parquet)
6) Implement map + prepare pipeline
7) Implement mine pipeline
8) Implement notebook + report generation
9) Implement QA pack
10) Implement optional LLM narrative + agent command (OpenAI/Anthropic/Gemini/Ollama)
11) Package installer and release workflow
12) Add tests and smoke runs

## Quality gates
- `pm-assist doctor` passes on a clean machine with minimal prerequisites.
- `pm-assist init` creates a runnable scaffold.
- A synthetic dataset completes the full flow producing outputs and QA summary.
- No secrets are written to disk or logs.
- LLM calls are off unless user opts-in and configures a provider.

## Skill packs
Use the skill files under `.codex/skills/` to guide implementation decisions.
- When a task matches a skill “when_to_use”, read that skill first.
- Apply the “checklist” in each skill before finalising changes.
