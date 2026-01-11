---
name: pm-03-data-quality
description: Assess data quality for the normalised log and recommend remediation thresholds and privacy handling.
metadata:
  short-description: Data quality checks and recommendations.
---

## Overview

This skill evaluates data quality and produces recommendations for missing values, parse failures, duplicates, and privacy.
It must run after ingest and before cleaning and filtering.

## When to use this skill

Use after `pm-02-ingest-profile` or when data quality thresholds change.
Examples:
- “Assess data quality for the normalised log.”
- “Re-run quality checks after mapping changes.”

## Inputs required

- `output/stage_01_ingest_profile/normalised_log.csv`
- Output directory
- Optional config with thresholds

## Outputs produced

- `output/stage_02_data_quality/data_quality.json`
- `output/stage_02_data_quality/data_quality_recommendations.json`
- `output/notebooks/Rx.xx/02_data_quality.ipynb`
- `output/manifest.json` updated with stage status and hashes

## Workflow

1. Run the data quality script with thresholds.
2. Review missingness, parse failures, duplicates, and sensitive column detection.
3. Capture chosen remediation decisions in the manifest.
4. Generate or update the notebook and store hashes.

## Decision checkpoints

Ask:
Choose how to handle missing values and parse failures.

Complication:
Dropping or imputing events affects conformance and performance metrics.

Options:
1) Drop rows above thresholds and flag the rest [preferred]
2) Impute timestamps using a defined strategy
3) Pause for manual repair and re-ingest

Impact:
- Option 1: quick, but may reduce coverage.
- Option 2: preserves volume but risks imputation bias.
- Option 3: highest quality but delays the run.

Ask:
Choose how to handle duplicates.

Complication:
Duplicate events can inflate activity counts and distort model discovery.

Options:
1) Drop exact duplicates [preferred]
2) Deduplicate by case, activity, timestamp keys
3) Keep duplicates and document risk

Impact:
- Option 1: minimal distortion with low complexity.
- Option 2: stronger cleanup but may remove legitimate repeats.
- Option 3: preserves data but risks skewed results.

Ask:
Choose the privacy handling for sensitive columns.

Complication:
Resource and attribute values may contain PII and affect sharing or reporting.

Options:
1) Mask detected sensitive columns [preferred]
2) Mask with a custom pattern list
3) Disable masking for internal-only analysis

Impact:
- Option 1: balanced privacy and usability.
- Option 2: tailored protection.
- Option 3: highest analytical fidelity but higher compliance risk.

## Commands

- `python .codex/skills/cli-tool-skills/pm-03-data-quality/scripts/02_data_quality.py --output <dir> --missing-value-threshold <value> --timestamp-parse-threshold <value> --duplicate-threshold <value> --order-violation-threshold <value> --fail-on-order-violations`

## Validations

- Confirm `data_quality.json` and recommendations are present.
- Confirm manifest stage status is `success`.
- Exit criteria: data quality artefacts exist and decisions are recorded.

## Issue handling

- On errors, check `references/index.json` for matching error signatures and follow the referenced runbook.
- If `references/index.json` or `references/README.md` is missing, create them using the pm-01-env issue pack format.
- When you resolve a new issue, add a runbook under `references/issue-fixes/` and register it in `references/index.json`.

## Failure modes and remediation

- Missing normalised log: re-run ingest.
- Thresholds too strict: adjust config and re-run.
- Case order violations: sort by case/timestamp or enable `--fail-on-order-violations` for strict gating.

## Compatibility notes

- Must run before `pm-04-clean-filter`.
- Aligns with standards in `pm-99`.

## Version history

- R1.00 Initial data quality skill.
