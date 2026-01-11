---
name: pm-00-orchestrator
description: Orchestrate an end-to-end process mining engagement by chaining phase skills with decision checkpoints and reproducibility safeguards.
metadata:
  short-description: End-to-end orchestrator for the process mining skill suite.
---

## Overview

This skill coordinates the full process mining workflow from environment validation to reporting.
It enforces decision checkpoints, stage gating, and reproducibility as defined in `pm-99-utils-and-standards`.
Use it when the user wants an end-to-end run or a guided resume from a specific stage.

## When to use this skill

Use when the user requests an end-to-end process mining run or wants help sequencing phases.
Examples:
- “Run the full pipeline for this log.”
- “Resume from performance analysis.”

## Inputs required

- Input log file path and format
- Optional config file (YAML or JSON)
- Output directory
- Stage resume point if resuming

## Outputs produced

- Stage artefacts in `output/stage_*/`
- Versioned notebooks in `output/notebooks/Rx.xx/`
- Updated `output/manifest.json`
- Final report artefacts in `output/stage_09_report/`

## Workflow

1. Confirm objective, scope, and success criteria.
2. Run environment detection and validation via `pm-01-env`.
3. Run phases sequentially by delegating to `pm-02` through `pm-10`.
4. Enforce decision checkpoints and apply phase gating.
5. On each transition, check notebook and artefact hashes and update the manifest.
6. If strict reproducibility is enabled, stop on any mismatch and request action.

## Decision checkpoints

Ask:
Choose the run mode for this engagement.

Complication:
The run mode affects reproducibility and how we handle changes between stages.

Options:
1) Strict reproducibility with hash enforcement [preferred]
2) Standard reproducibility with warnings
3) Ad-hoc exploratory run with no hash enforcement

Impact:
- Option 1: stop on any mismatch, enforce revision bumps and re-runs.
- Option 2: allow continuation with logged warnings and revision bumps.
- Option 3: fastest iteration but lowest auditability.

Ask:
Choose the execution approach for the pipeline.

Complication:
Using stage-by-stage execution gives tighter control, while full pipeline runs are faster but less flexible.

Options:
1) Stage-by-stage with checkpoints [preferred]
2) Full pipeline run via `run_pipeline.py`

Impact:
- Option 1: explicit decisions between stages, higher governance.
- Option 2: faster execution, fewer intervention points.

## Commands

- End-to-end pipeline:
  - `python .codex/skills/cli-tool-skills/pm-00-orchestrator/scripts/run_pipeline.py --file <path> --format <csv|xes> --output <dir>`
- Resume from stage (manual sequence):
  - `python .codex/skills/cli-tool-skills/pm-01-env/scripts/00_detect_env.py --output <dir>`
  - `python .codex/skills/cli-tool-skills/pm-01-env/scripts/00_validate_env.py --output <dir> --setup-venv --venv-dir .venv --requirements .codex/skills/cli-tool-skills/pm-99-utils-and-standards/requirements.txt`
  - Follow phase skills in order with their command patterns.

## Validations

- Confirm `manifest.json` exists or will be created in the output root.
- Confirm stage outputs match required artefact folders.
- Exit criteria: all requested phases complete with artefacts and an updated manifest.

## Issue handling

- On errors, check `references/index.json` for matching error signatures and follow the referenced runbook.
- If `references/index.json` or `references/README.md` is missing, create them using the pm-01-env issue pack format.
- When you resolve a new issue, add a runbook under `references/issue-fixes/` and register it in `references/index.json`.

## Failure modes and remediation

- Missing dependencies: stop and route to `pm-01-env` remediation.
- Hash mismatch in strict mode: stop, bump revision, re-run impacted stages.
- Missing stage artefacts: re-run the missing stage via its skill.

## Compatibility notes

- Delegates to `pm-01` through `pm-10` and follows `pm-99` standards.
- Compatible with `process_mining_cli.py` wrapper if present.

## Version history

- R1.00 Initial orchestrator split from the monolithic process-mining-assistant.
