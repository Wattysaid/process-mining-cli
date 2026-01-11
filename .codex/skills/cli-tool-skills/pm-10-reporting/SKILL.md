---
name: pm-10-reporting
description: Generate final reports and packages from accumulated artefacts, with audience-specific options.
metadata:
  short-description: Reporting and packaging outputs.
---

## Overview

This skill compiles artefacts into reports and optional export bundles.
It must run after required analytical stages and uses the manifest to ensure consistency.

## When to use this skill

Use at the end of an engagement or when reports need regeneration after upstream changes.
Examples:
- “Generate the executive summary report.”
- “Export a zipped artefacts bundle.”

## Inputs required

- Manifest and stage artefacts from completed phases
- Output directory
- Reporting audience preferences

## Outputs produced

- `output/stage_09_report/process_mining_report.md`
- Optional packaged artefacts (zip)
- `output/notebooks/Rx.xx/09_report.ipynb`
- Finalised `output/manifest.json`
 - Includes data quality and conformance summaries when available

## Workflow

1. Validate that required upstream artefacts exist.
2. Select report format and audience.
3. Run report generation and optional export.
4. Update manifest with report artefacts and revision notes.

## Decision checkpoints

Ask:
Choose the reporting format and audience.

Complication:
Different audiences require different detail levels and artefact selections.

Options:
1) Executive summary report [preferred]
2) Technical report with full artefact references
3) Data quality and preparation log only

Impact:
- Option 1: concise outputs for leadership.
- Option 2: comprehensive documentation for analysts.
- Option 3: focused on data readiness and limitations.

Ask:
Choose whether to generate a zipped artefact bundle.

Complication:
Packaging changes how artefacts are distributed and may include sensitive outputs.

Options:
1) Generate a zipped artefact bundle [preferred]
2) Do not package artefacts

Impact:
- Option 1: easier sharing but higher privacy scrutiny.
- Option 2: keep artefacts local only.

## Commands

- `python .codex/skills/pm-10-reporting/scripts/08_report.py --output <dir>`
- `python .codex/skills/pm-10-reporting/scripts/export_artifacts.py --output <dir>`

## Validations

- Confirm report markdown exists and links to relevant artefacts.
- Confirm manifest is updated and hashes are current.
- Exit criteria: report generated and manifest finalised.

## Issue handling

- On errors, check `references/index.json` for matching error signatures and follow the referenced runbook.
- If `references/index.json` or `references/README.md` is missing, create them using the pm-01-env issue pack format.
- When you resolve a new issue, add a runbook under `references/issue-fixes/` and register it in `references/index.json`.

## Failure modes and remediation

- Missing upstream artefacts: re-run required stages and regenerate report.
- Privacy concerns: regenerate with masking or exclude sensitive sections.

## Compatibility notes

- Requires `pm-06`, `pm-07`, and `pm-08` outputs for full report coverage.
- Follows `pm-99` revisioning and privacy rules.

## Version history

- R1.00 Initial reporting skill.
