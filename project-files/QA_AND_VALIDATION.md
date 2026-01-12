# QA and Validation Pack

**Document version:** R1.01 (2026-01-12)

## 1. Goals
The QA pack ensures:
- data is fit for process mining
- assumptions are explicit
- results are interpretable
- outputs are reproducible

## 2. QA checks (MVP)
### Ingestion
- File encoding and delimiter sanity
- Required columns present
- Timestamp parse success rate
- Row count and uniqueness checks

### Event log readiness
- Missingness for required fields (case/activity/timestamp)
- Timestamp monotonicity within case
- Duplicate events (by case/activity/timestamp) and dedupe suggestions
- Case length distribution (min/median/p95)
- Activity cardinality and rare-activity warnings
- Time range and gaps

### Data science hygiene
- Outlier detection summaries (numeric columns)
- Type coercion counts
- Null handling report
- Feature extraction report (if enabled)

### Process mining sanity
- Model complexity indicators:
  - number of nodes/arcs
  - variants count
- Warning thresholds:
  - extremely high variant count
  - too sparse frequency thresholds
- Conformance compute warnings

## 3. Outputs
- `quality/qa_summary.md` (human-readable)
- `quality/qa_results.json` (machine-readable)
- `quality/issues_backlog.csv` (prioritised fix list)
- `run_manifest.json` includes QA step status and timestamps

## 4. Pass/fail rules
- Default is “warn not fail”, except for:
  - missing required fields
  - < configurable parse success for timestamps
  - empty dataset after required filtering

## 5. CLI behaviour
- In interactive mode:
  - show issues and ask if user wants to proceed
- In non-interactive mode:
  - fail on blocking rules only

## 6. Status
- QA pack is wired into `pm-assist review`.
- Summaries are appended to the notebook when the user confirms.
