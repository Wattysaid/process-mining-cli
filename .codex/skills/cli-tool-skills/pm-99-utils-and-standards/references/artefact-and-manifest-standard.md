# Artefact and Manifest Standard

## Artefact folder conventions

Output root:
`output/`

Required layout:
```
output/
  manifest.json
  run_log.txt
  stage_00_validate_env/
  stage_01_ingest_profile/
  stage_02_data_quality/
  stage_03_clean_filter/
  stage_04_eda/
  stage_05_discover/
  stage_06_conformance/
  stage_07_performance/
  stage_08_org_mining/
  stage_09_report/
  notebooks/
    R1.00/
      01_ingest_profile.ipynb
      02_data_quality.ipynb
      03_clean_filter.ipynb
      04_eda.ipynb
      05_discover.ipynb
      06_conformance.ipynb
      07_performance.ipynb
      08_org_mining.ipynb
      09_report.ipynb
```

Stage folders must contain:
- Stage artefacts and logs
- A stage summary JSON where applicable

## Manifest requirements

Minimum keys in `manifest.json`:
- `run_id`
- `created_at`
- `input` (file, format, mapping, time window)
- `config` (resolved configuration used)
- `revisions` (array of revision objects)
- `stages` (per-stage status, parameters, artefacts, notebook path, hashes)
- `hashes` (hash per artefact and notebook)
- `privacy` (masking settings and patterns used)

### Informal JSON schema sketch

```
{
  "run_id": "string",
  "created_at": "ISO-8601 timestamp",
  "input": {
    "file": "path",
    "format": "csv|xes",
    "mapping": {
      "case": "string",
      "activity": "string",
      "timestamp": "string",
      "resource": "string|optional"
    },
    "time_window": "string|optional"
  },
  "config": { "resolved": "object" },
  "revisions": [
    {
      "revision": "R1.00",
      "timestamp": "ISO-8601 timestamp",
      "reason": "string",
      "changed_files": ["path"]
    }
  ],
  "stages": {
    "stage_01_ingest_profile": {
      "status": "success|failed|skipped",
      "parameters": { "object" },
      "artefacts": ["path"],
      "notebook": "path",
      "hashes": { "path": "sha256" }
    }
  },
  "hashes": {
    "path": "sha256"
  },
  "privacy": {
    "masking_enabled": true,
    "patterns": ["string"],
    "columns_masked": ["string"]
  }
}
```

## Revisioning rules

- Use R1.00, R1.01, R1.02 and so on.
- Bump the revision when:
  - a notebook is edited,
  - a stage is re-run with different parameters,
  - config or input mappings change,
  - an upstream artefact changes and invalidates downstream outputs.
- Record revision rationale in `revisions` with a short reason and changed file list.

## Notebook change detection

- Hash notebooks and key artefacts at the end of each stage.
- Store hashes in `manifest.json` under `stages.<stage>.hashes` and `hashes`.
- On stage transition, compare current hashes to the manifest.
- If a mismatch is found:
  - bump the revision,
  - re-run only impacted stages,
  - invalidate downstream artefacts that depend on the changed output.

### Dependency rules

- If `stage_01_ingest_profile` changes, stages 02 to 09 must be re-run.
- If `stage_02_data_quality` changes, stages 03 to 09 must be re-run.
- If `stage_03_clean_filter` changes, stages 04 to 09 must be re-run.
- If `stage_04_eda` changes, stages 05 to 09 must be re-run when EDA drives miner choices.
- If `stage_05_discover` changes, stages 06 to 09 must be re-run.
- If `stage_06_conformance` changes, stages 07 to 09 must be re-run only if reporting depends on conformance outputs.
- If `stage_07_performance` changes, stages 08 to 09 must be re-run only if reporting depends on performance outputs.
- If `stage_08_org_mining` changes, stage 09 must be re-run.

## Strict reproducibility mode

- When enabled, the run must halt if any artefact or notebook hash does not match the manifest.
- The assistant must request re-run or explicit override.
- No silent regeneration or continuation is permitted.

## Privacy guidance

- Default to masking when PII-like patterns are detected.
- Record masking settings and columns in the manifest.
- If privacy constraints conflict with org mining, ask for explicit permission before proceeding.
