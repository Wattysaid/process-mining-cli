---
name: pm-07-conformance
description: Evaluate conformance of the event log to discovered models and generate deviation artefacts.
metadata:
  short-description: Conformance checking for process models.
---

## Overview

This skill measures fitness and precision using token-based replay or alignments.
It produces deviation summaries and conformance metrics for reporting.

## When to use this skill

Use after `pm-06-discovery` or when conformance method changes.
Examples:
- “Run token-based replay and generate conformance metrics.”
- “Use alignments for detailed deviations.”

## Inputs required

- `output/stage_05_discover/` model artefacts
- `output/stage_03_clean_filter/filtered_log.csv`
- Output directory

## Outputs produced

- `output/stage_06_conformance/conformance_metrics.csv`
- Deviation summaries and logs
- `output/notebooks/Rx.xx/06_conformance.ipynb`
- `output/manifest.json` updated with stage status and hashes

## Workflow

1. Run conformance checking with the chosen method.
2. Review fitness, precision, and deviation summaries.
3. Record method and parameters in the manifest.
4. Generate or update the notebook and store hashes.

## Decision checkpoints

Ask:
Choose the conformance checking method.

Complication:
Alignments provide more detail but are slower and more sensitive to noise.

Options:
1) Token-based replay [preferred]
2) Alignments

Impact:
- Option 1: faster and suitable for broad fitness checks.
- Option 2: deeper diagnostics with higher compute cost.

Ask:
Choose the deviation reporting detail level.

Complication:
The reporting level affects how much technical detail is generated and stored.

Options:
1) Executive summary level [preferred]
2) Technical detail with per-case deviations

Impact:
- Option 1: concise reporting, smaller artefacts.
- Option 2: deeper diagnostics, larger outputs.

## Commands

- `python .codex/skills/cli-tool-skills/pm-07-conformance/scripts/05_conformance.py --use-filtered --output <dir> --conformance-method <token|alignments>`

## Validations

- Confirm conformance artefacts exist and are hashed.
- Confirm method and reporting detail are logged in the manifest.
- Exit criteria: conformance metrics and deviations are available.

## Issue handling

- On errors, check `references/index.json` for matching error signatures and follow the referenced runbook.
- If `references/index.json` or `references/README.md` is missing, create them using the pm-01-env issue pack format.
- When you resolve a new issue, add a runbook under `references/issue-fixes/` and register it in `references/index.json`.

## Failure modes and remediation

- Missing model artefacts: re-run `pm-06-discovery`.
- Alignment failures: switch to token-based replay or adjust thresholds.

## Compatibility notes

- Feeds `pm-10-reporting` and may inform `pm-08-performance` narratives.
- Follows `pm-99` for revisioning and artefact hashing.

## Version history

- R1.00 Initial conformance skill.
