---
name: pm-01-env
description: Validate environment readiness for the process mining pipeline and stop if dependencies are missing.
metadata:
  short-description: Environment validation for PM4Py pipeline.
---

## Overview

This skill verifies that Python and required libraries are available and compatible.
It is the first gate before any data ingestion or analysis.

## When to use this skill

Use at the start of any engagement or when a pipeline fails due to missing dependencies.
Examples:
- “Check whether the environment is ready.”
- “Validate PM4Py and pandas availability.”

## Inputs required

- Output directory path
- Virtual environment directory (default `.venv`)
- Requirements file path (default `.codex/skills/cli-tool-skills/pm-99-utils-and-standards/requirements.txt`)

## Outputs produced

- `output/stage_00_detect_env/detect_env.json`
- `output/stage_00_detect_env/detect_env.log`
- `output/stage_00_validate_env/validate_env.json`
- `output/stage_00_validate_env/validate_env.log`
- `output/manifest.json` updated with stage status

## Workflow

1. Run the environment detection script to capture OS, shell, and tooling.
2. Create or reuse the virtual environment and install requirements.
3. Run the environment validation script using the detected Python interpreter.
4. Review dependency checks and versions.
5. Update the manifest with the stage status and hashes.
6. If validation fails more than once, show manual `.venv` setup instructions for the detected OS/shell and re-run validation without `--setup-venv`.
7. When outputs are confirmed complete, remind the user to deactivate the virtual environment.

## Decision checkpoints

Ask:
Proceed after resolving any missing dependencies?

Complication:
Running later stages with missing or incompatible packages will corrupt outputs or fail mid-run.

Options:
1) Stop and remediate dependencies [preferred]
2) Continue in best-effort mode with warnings

Impact:
- Option 1: clean run with reproducible outputs.
- Option 2: risk of partial or inconsistent artefacts.

## Commands

- `python .codex/skills/cli-tool-skills/pm-01-env/scripts/00_detect_env.py --output <dir>`
- `python .codex/skills/cli-tool-skills/pm-01-env/scripts/00_validate_env.py --output <dir> --setup-venv --venv-dir .venv --requirements .codex/skills/cli-tool-skills/pm-99-utils-and-standards/requirements.txt --detect-env-json output/stage_00_detect_env/detect_env.json`

## Validations

- Confirm `validate_env.json` exists and reports required packages.
- Confirm manifest stage status is `success` before proceeding.
- Exit criteria: environment validation passes or remediation instructions are provided.

## Issue handling

- On errors, check `references/index.json` for matching error signatures and follow the referenced runbook.
- If `references/index.json` or `references/README.md` is missing, create them using the pm-01-env issue pack format.
- When you resolve a new issue, add a runbook under `references/issue-fixes/` and register it in `references/index.json`.

## Failure modes and remediation

- Missing packages: install from `requirements.txt` inside the venv and re-run validation.
- Version conflicts: record in manifest and resolve before proceeding.
- PEP 668 environment error on Ubuntu: follow `.codex/skills/cli-tool-skills/pm-01-env/references/issue-fixes/ISSUE-007-PEP-668-Ubuntu-Error.md`.
- Virtual environment creation failure: verify permissions and Python venv support, then re-run.
- Manual recovery: provide OS/shell-specific activation steps, reinstall requirements, re-run validation without `--setup-venv`, and deactivate with `deactivate` once finished.

## Compatibility notes

- Required before `pm-02-ingest-profile` and any downstream phase.
- Follows `pm-99` decision and manifest standards.

## Version history

- R1.00 Initial environment validation skill.
