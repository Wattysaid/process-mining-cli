---
name: pm-02-ingest-profile
description: Ingest the event log, normalise schema, and generate an initial data profile with notebook and manifest updates.
metadata:
  short-description: Log ingestion, schema mapping, and profiling.
---

## Overview

This skill loads the event log, normalises it, and produces a profile to drive subsequent decisions.
It is the earliest data-dependent phase and must complete before any cleaning or mining choices.

## When to use this skill

Use after environment validation or when input mapping changes.
Examples:
- “Ingest this CSV and generate a profile.”
- “Re-map the case and activity columns and re-run ingest.”

## Inputs required

- Input file path and format
- CSV mappings for case, activity, timestamp, and optional resource
- Optional multi-source config for heterogeneous datasets
- Example config: `.codex/skills/cli-tool-skills/pm-02-ingest-profile/references/multi_source_config.example.json`
- Output directory
- Optional config file

## Outputs produced

- `output/stage_01_ingest_profile/normalised_log.csv`
- `output/stage_01_ingest_profile/ingest_profile.json`
- `output/stage_01_ingest_profile/sample_rows.csv`
- `output/notebooks/Rx.xx/01_ingest_profile.ipynb`
- `output/manifest.json` updated with stage status and hashes

## Workflow

1. Run the ingest script with mapping and format parameters.
2. Review the ingest profile for parsing failures and missingness.
3. Capture schema choices and masking settings in the manifest.
4. Generate the stage notebook and store hashes.

## Decision checkpoints

Ask:
Choose the schema mapping for case, activity, and timestamp.

Complication:
Incorrect mapping breaks event ordering and invalidates all downstream analysis.

Options:
1) Use the inferred mapping from the ingest profile [preferred]
2) Provide an explicit mapping override

Impact:
- Option 1: faster start but relies on inference accuracy.
- Option 2: more reliable but requires manual confirmation.

Ask:
Choose how to handle timestamp parsing issues detected during ingest.

Complication:
Parsing failures can remove events or skew throughput metrics.

Options:
1) Use the detected format and drop invalid rows [preferred]
2) Provide a custom parse format and re-run ingest
3) Pause for manual data repair

Impact:
- Option 1: fastest, but may lose data.
- Option 2: higher accuracy if format is known.
- Option 3: highest quality but delays the run.

Ask:
Choose whether to enable sensitive data masking.

Complication:
PII in the log affects privacy and may block organisational mining or sharing artefacts.

Options:
1) Enable masking using detected patterns [preferred]
2) Provide a custom pattern list
3) Disable masking for restricted internal use

Impact:
- Option 1: protects privacy with standard patterns.
- Option 2: tailored coverage for your data.
- Option 3: faster but higher compliance risk.

## Commands

- `python .codex/skills/cli-tool-skills/pm-02-ingest-profile/scripts/01_ingest.py --file <path> --format <csv|xes> --case <col> --activity <col> --timestamp <col> --output <dir>`
- `python .codex/skills/cli-tool-skills/pm-02-ingest-profile/scripts/01_ingest_multi.py --config .codex/skills/cli-tool-skills/pm-02-ingest-profile/references/multi_source_config.example.json --output <dir>`

## Validations

- Confirm `ingest_profile.json` includes parsed counts and inferred types.
- Confirm `normalised_log.csv` exists and row counts align with expectations.
- Exit criteria: normalised log and profile artefacts exist with a manifest update.

## Issue handling

- On errors, check `references/index.json` for matching error signatures and follow the referenced runbook.
- If `references/index.json` or `references/README.md` is missing, create them using the pm-01-env issue pack format.
- When you resolve a new issue, add a runbook under `references/issue-fixes/` and register it in `references/index.json`.

## Failure modes and remediation

- Schema mismatch: update mapping and re-run ingest.
- Parsing failures above threshold: provide explicit format or repair data.
- Missing file or format: validate path and format flag.
- Multi-source mismatch: align case correlation keys or switch to concat strategy.

## Compatibility notes

- Must run before `pm-03-data-quality`.
- Follows manifest and revisioning rules in `pm-99`.

## Version history

- R1.00 Initial ingest and profile skill.
