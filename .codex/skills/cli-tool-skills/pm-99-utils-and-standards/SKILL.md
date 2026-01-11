---
name: pm-99-utils-and-standards
description: Shared standards for decision checkpoints, manifest shape, revisioning, reproducibility, and privacy across the process mining skill suite.
metadata:
  short-description: Process mining standards and guardrails used by all suite skills.
---

## Overview

This skill centralises the rules that govern how the process mining suite behaves.
Use it to keep decision checkpoints, artefact handling, manifests, revisions, and notebook change detection consistent.
All phase skills and the orchestrator must reference these standards and avoid duplicating them.

## When to use this skill

Use when defining or updating any process mining skill workflow, decision checkpoint, or artefact handling rule.
Examples:
- “What is the required manifest structure for stage outputs?”
- “How do we handle notebook edits between stages?”
- “What is the decision checkpoint format?”

## Inputs required

- Existing output folder, if validating a run
- `manifest.json` when checking revisions and hashes
- Current stage artefacts and notebooks

## Outputs produced

- None directly
- Standards references live in:
  - `.codex/skills/cli-tool-skills/pm-99-utils-and-standards/references/interaction-patterns.md`
  - `.codex/skills/cli-tool-skills/pm-99-utils-and-standards/references/artefact-and-manifest-standard.md`
  - `.codex/skills/cli-tool-skills/pm-99-utils-and-standards/references/interaction-examples.md`
  - `.codex/skills/cli-tool-skills/pm-99-utils-and-standards/references/improved_prompt.md`
  - `.codex/skills/cli-tool-skills/pm-99-utils-and-standards/references/notebook-snippets/`
  - `.codex/skills/cli-tool-skills/pm-99-utils-and-standards/scripts/block_templates.py`

## Workflow

1. Read the relevant standards reference file for the request.
2. Apply the rules in the current skill or stage.
3. If a deviation is required, document it in the manifest revision reason.
4. Confirm that all dependent skills reference the same standards.

## Decision checkpoints

Use the decision format defined in `references/interaction-patterns.md`.
No additional checkpoints are defined here.

## Commands

None. This skill is documentation and governance only.

## Validations

- Confirm decision checkpoints follow Ask, Complication, Options, Impact.
- Confirm stage gating is respected and no future phase questions are asked.
- Confirm manifest fields and hashes match the standard.
- Exit criteria: the referenced skill or phase has been updated to comply with the standards.

## Issue handling

- On errors, check `references/index.json` for matching error signatures and follow the referenced runbook.
- If `references/index.json` or `references/README.md` is missing, create them using the pm-01-env issue pack format.
- When you resolve a new issue, add a runbook under `references/issue-fixes/` and register it in `references/index.json`.

## Failure modes and remediation

- Missing or inconsistent rules: update the relevant skill to reference the standards.
- Schema drift in `manifest.json`: align to the manifest standard and record a revision bump.

## Compatibility notes

- This skill is a dependency for all `pm-00` to `pm-10` skills.
- It must be referenced, not duplicated.

## Version history

- R1.00 Initial standards skill extracted from the monolithic process-mining-assistant.
