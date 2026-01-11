---
name: pm-04-clean-filter
description: Apply cleaning and filtering actions based on data quality decisions and generate filtered log artefacts.
metadata:
  short-description: Cleaning and filtering of the event log.
---

## Overview

This skill applies cleaning and filtering actions, producing a filtered log for downstream analysis.
It must run after data quality decisions are confirmed.

## When to use this skill

Use after `pm-03-data-quality` or when filtering rules change.
Examples:
- “Filter rare activities and apply time window constraints.”
- “Re-run cleaning with new thresholds.”

## Inputs required

- `output/stage_01_ingest_profile/normalised_log.csv`
- Data quality decisions from the manifest
- Output directory

## Outputs produced

- `output/stage_03_clean_filter/filtered_log.csv`
- `output/stage_03_clean_filter/filter_summary.json`
- `output/notebooks/Rx.xx/03_clean_filter.ipynb`
- `output/manifest.json` updated with stage status and hashes

## Workflow

1. Apply cleaning actions from data quality decisions.
2. Apply filters for rare activities, start and end activities, and time windows.
3. Record filtering parameters in the manifest.
4. Generate or update the stage notebook and store hashes.

## Decision checkpoints

Ask:
Choose the rare activity filtering strategy.

Complication:
Filtering changes variant frequency and may remove valid but infrequent paths.

Options:
1) No rare activity filtering [preferred]
2) Filter by minimum activity frequency threshold
3) Filter by top variants only

Impact:
- Option 1: preserves full behaviour but can be noisy.
- Option 2: reduces noise but risks losing minority cases.
- Option 3: focuses on dominant paths, least comprehensive.

Ask:
Choose start and end activity constraints.

Complication:
Constraining start or end activities can improve comparability but may exclude valid cases.

Options:
1) No constraints [preferred]
2) Constrain to specified start and end activities

Impact:
- Option 1: highest coverage.
- Option 2: cleaner comparisons, but narrower scope.

Ask:
Choose time window filtering.

Complication:
Time windows affect seasonality and performance baselines.

Options:
1) No time window filtering [preferred]
2) Apply a fixed time range

Impact:
- Option 1: full period coverage.
- Option 2: focused analysis for a specific period.

## Commands

- `python .codex/skills/pm-04-clean-filter/scripts/02_clean_filter.py --output <dir> --auto-filter-rare-activities <true|false> --min-activity-frequency <value>`

## Validations

- Confirm `filtered_log.csv` exists and record counts align with expectations.
- Confirm filter parameters are recorded in the manifest.
- Exit criteria: filtered log and filter summary artefacts are present.

## Issue handling

- On errors, check `references/index.json` for matching error signatures and follow the referenced runbook.
- If `references/index.json` or `references/README.md` is missing, create them using the pm-01-env issue pack format.
- When you resolve a new issue, add a runbook under `references/issue-fixes/` and register it in `references/index.json`.

## Failure modes and remediation

- Empty filtered log: relax filters and re-run.
- Missing input log: re-run ingest and data quality first.

## Compatibility notes

- Must run before `pm-05-eda`.
- Follows `pm-99` artefact and revisioning rules.

## Version history

- R1.00 Initial clean and filter skill.
