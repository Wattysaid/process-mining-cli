---
name: pm-05-eda
description: Run exploratory analysis on the filtered log to profile variants, distributions, and time patterns.
metadata:
  short-description: Exploratory data analysis for event logs.
---

## Overview

This skill generates descriptive statistics, variant analysis, and distribution charts.
It informs discovery and reporting decisions without changing the log itself.

## When to use this skill

Use after `pm-04-clean-filter` or when segmentation needs revisiting.
Examples:
- “Generate EDA charts for variants and throughput.”
- “Check whether we should segment by case type.”

## Inputs required

- `output/stage_03_clean_filter/filtered_log.csv`
- Output directory

## Outputs produced

- `output/stage_04_eda/summary_stats.json`
- `output/stage_04_eda/variant_counts.csv`
- `output/stage_04_eda/variant_pareto.png`
- Optional (with `--advanced`): `output/stage_04_eda/variant_coverage.csv`, `output/stage_04_eda/variant_entropy.json`, `output/stage_04_eda/case_length_distribution.csv`, `output/stage_04_eda/case_length_summary.json`
- `output/notebooks/Rx.xx/04_eda.ipynb`
- `output/manifest.json` updated with stage status and hashes

## Workflow

1. Run the EDA script to generate summary statistics and charts.
2. Review variant coverage and distribution patterns.
3. Record segmentation decisions in the manifest.
4. Generate or update the notebook and store hashes.

## Decision checkpoints

Ask:
Choose the segmentation approach for analysis.

Complication:
Segmentation changes comparability of metrics and can hide or reveal bottlenecks.

Options:
1) No segmentation, use full filtered log [preferred]
2) Segment by time window
3) Segment by case attribute or type

Impact:
- Option 1: broadest view with less granularity.
- Option 2: reveals seasonal patterns, increases complexity.
- Option 3: highlights differences across cohorts, requires clean attributes.

## Commands

- `python .codex/skills/cli-tool-skills/pm-05-eda/scripts/03_eda.py --use-filtered --output <dir>`
- `python .codex/skills/cli-tool-skills/pm-05-eda/scripts/03_eda.py --output <dir> --advanced` (adds variant coverage, entropy, and case length diagnostics)

## Validations

- Confirm EDA artefacts exist and charts are generated.
- Confirm any segmentation decisions are recorded.
- Exit criteria: summary and variant artefacts exist and are hashed.

## Issue handling

- On errors, check `references/index.json` for matching error signatures and follow the referenced runbook.
- If `references/index.json` or `references/README.md` is missing, create them using the pm-01-env issue pack format.
- When you resolve a new issue, add a runbook under `references/issue-fixes/` and register it in `references/index.json`.

## Failure modes and remediation

- Missing filtered log: re-run `pm-04-clean-filter`.
- Chart generation failures: verify plotting dependencies and re-run.

## Compatibility notes

- Precedes `pm-06-discovery` if EDA informs miner choices.
- Uses the `pm-99` standards for hashes and revisions.

## Version history

- R1.00 Initial EDA skill.
