---
name: pm-08-performance
description: Analyse throughput, case duration, and activity sojourn times from the filtered log.
metadata:
  short-description: Performance and time analysis.
---

## Overview

This skill computes performance metrics such as case durations, sojourn times, and throughput trends.
It supports bottleneck analysis and reporting.

## When to use this skill

Use after `pm-06-discovery` or when performance lens decisions change.
Examples:
- “Generate case duration distributions.”
- “Focus on sojourn times by activity.”

## Inputs required

- `output/stage_03_clean_filter/filtered_log.csv`
- Output directory

## Outputs produced

- `output/stage_07_performance/case_durations.csv`
- `output/stage_07_performance/sojourn_times.csv`
- `output/stage_07_performance/performance_summary.json`
- Optional (with `--advanced`): `output/stage_07_performance/activity_waiting_time_stats.csv`, `output/stage_07_performance/case_duration_summary.json`
- `output/notebooks/Rx.xx/07_performance.ipynb`
- `output/manifest.json` updated with stage status and hashes

## Workflow

1. Run the performance script on the filtered log.
2. Review duration distributions and trend charts.
3. Record the chosen performance lens in the manifest.
4. Generate or update the notebook and store hashes.

## Decision checkpoints

Ask:
Choose the performance lens to prioritise.

Complication:
Different lenses change which bottlenecks are highlighted and which charts are generated.

Options:
1) Case duration focus [preferred]
2) Activity sojourn focus
3) Both case and activity lenses

Impact:
- Option 1: emphasises end-to-end cycle time.
- Option 2: emphasises waiting time and queueing.
- Option 3: more complete but heavier artefact set.

## Commands

- `python .codex/skills/pm-08-performance/scripts/06_performance.py --use-filtered --output <dir>`
- `python .codex/skills/pm-08-performance/scripts/06_performance.py --output <dir> --advanced --sla-hours <hours>` (adds waiting time stats and SLA breach summary)

## Validations

- Confirm performance artefacts exist and are hashed.
- Confirm lens choice is recorded in the manifest.
- Exit criteria: performance summary and charts are present.

## Issue handling

- On errors, check `references/index.json` for matching error signatures and follow the referenced runbook.
- If `references/index.json` or `references/README.md` is missing, create them using the pm-01-env issue pack format.
- When you resolve a new issue, add a runbook under `references/issue-fixes/` and register it in `references/index.json`.

## Failure modes and remediation

- Missing filtered log: re-run `pm-04-clean-filter`.
- Missing timestamps: re-run `pm-02-ingest-profile` with correct mapping.

## Compatibility notes

- Feeds `pm-10-reporting` and may be run in parallel with `pm-09-org-mining` if privacy allows.
- Follows `pm-99` standards.

## Version history

- R1.00 Initial performance analysis skill.
