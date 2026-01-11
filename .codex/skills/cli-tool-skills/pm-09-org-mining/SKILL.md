---
name: pm-09-org-mining
description: Perform organisational mining and handover-of-work analysis with privacy-aware controls.
metadata:
  short-description: Organisational mining and handover analysis.
---

## Overview

This skill generates organisational mining artefacts such as handover-of-work networks.
It requires a resource column and must respect privacy settings.

## When to use this skill

Use after `pm-04-clean-filter` when organisational analysis is required and privacy constraints allow it.
Examples:
- “Generate handover-of-work metrics.”
- “Check resource handoffs with masking enabled.”

## Inputs required

- `output/stage_03_clean_filter/filtered_log.csv`
- Resource column mapping
- Privacy settings from the manifest
- Output directory

## Outputs produced

- `output/stage_08_org_mining/handover_of_work.csv`
- Optional network artefacts
- `output/notebooks/Rx.xx/08_org_mining.ipynb`
- `output/manifest.json` updated with stage status and hashes

## Workflow

1. Validate that a resource column is available or mapped.
2. Confirm privacy settings and masking rules.
3. Run organisational mining to generate handover artefacts.
4. Record any privacy constraints and outputs in the manifest.

## Decision checkpoints

Ask:
Choose whether to include resource-level analysis.

Complication:
Resource analysis can expose sensitive information and may violate privacy policies.

Options:
1) Include resource-level analysis with masking [preferred]
2) Include resource-level analysis without masking
3) Skip organisational mining

Impact:
- Option 1: protects privacy but may reduce interpretability.
- Option 2: highest fidelity with higher compliance risk.
- Option 3: avoids privacy risk but loses organisational insight.

## Commands

- `python .codex/skills/pm-09-org-mining/scripts/07_org_mining.py --use-filtered --output <dir>`

## Validations

- Confirm resource column exists and is not empty.
- Confirm handover artefacts exist and are hashed.
- Exit criteria: organisational artefacts exist or the stage is explicitly skipped.

## Issue handling

- On errors, check `references/index.json` for matching error signatures and follow the referenced runbook.
- If `references/index.json` or `references/README.md` is missing, create them using the pm-01-env issue pack format.
- When you resolve a new issue, add a runbook under `references/issue-fixes/` and register it in `references/index.json`.

## Failure modes and remediation

- Missing resource column: update mapping and re-run ingest.
- Privacy conflict: restrict analysis or skip this stage.

## Compatibility notes

- Can run after `pm-04-clean-filter` but must be recorded before reporting.
- Follows `pm-99` privacy and revisioning rules.

## Version history

- R1.00 Initial organisational mining skill.
