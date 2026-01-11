---
name: pm-06-discovery
description: Discover process models using selected miner strategies and generate model artefacts.
metadata:
  short-description: Process discovery using PM4Py miners.
---

## Overview

This skill discovers process models from the filtered log and records miner parameters.
It produces visual models and metrics for downstream conformance and reporting.

## When to use this skill

Use after `pm-05-eda` or when miner strategy changes.
Examples:
- “Run inductive miner for discovery.”
- “Compare inductive and heuristic miners.”

## Inputs required

- `output/stage_03_clean_filter/filtered_log.csv`
- Optional segmentation decision from EDA
- Output directory

## Outputs produced

- `output/stage_05_discover/model_metrics.csv`
- `output/stage_05_discover/models_manifest.json` (if models saved)
- Model artefacts and visualisations (DFG or Petri net files)
- `output/notebooks/Rx.xx/05_discover.ipynb`
- `output/manifest.json` updated with stage status and hashes

## Workflow

1. Run discovery with the chosen miner strategy and thresholds.
2. Review model metrics and visualisations.
3. Record miner parameters in the manifest.
4. Generate or update the notebook and store hashes.

## Decision checkpoints

Ask:
Choose the miner strategy for discovery.

Complication:
Different miners balance noise handling and model complexity, which affects conformance outcomes.

Options:
1) Auto selection with defaults [preferred]
2) Inductive miner with specified noise threshold
3) Heuristic miner with dependency and frequency thresholds

Impact:
- Option 1: balanced, low effort selection.
- Option 2: robust to noise, may simplify behaviour.
- Option 3: richer detail but more sensitive to noise.

## Commands

- `python .codex/skills/pm-06-discovery/scripts/04_discover.py --use-filtered --output <dir> --miner-selection <auto|inductive|heuristic>`

## Validations

- Confirm model artefacts and `model_metrics.csv` exist.
- Confirm miner parameters are logged in the manifest.
- Exit criteria: discovery artefacts exist and are hashed.

## Issue handling

- On errors, check `references/index.json` for matching error signatures and follow the referenced runbook.
- If `references/index.json` or `references/README.md` is missing, create them using the pm-01-env issue pack format.
- When you resolve a new issue, add a runbook under `references/issue-fixes/` and register it in `references/index.json`.

## Failure modes and remediation

- Missing filtered log: re-run `pm-04-clean-filter`.
- Model generation errors: adjust thresholds and re-run.

## Compatibility notes

- Must run before `pm-07-conformance`.
- Follows `pm-99` for revisioning and artefact hashing.

## Version history

- R1.00 Initial process discovery skill.
