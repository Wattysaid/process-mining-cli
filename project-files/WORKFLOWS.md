# Workflows: End-to-end process mining delivery

**Document version:** R1.00 (2026-01-11)

## 1. Engagement workflow (consulting-style)
PM Assist should guide the user through the same phases a mature process mining team would follow.

### Phase A: Define
- Confirm business process scope and boundaries
- Define key questions:
  - where are the delays?
  - what drives rework?
  - what variants cause SLA breaches?
- Identify required attributes:
  - customer segment, region, product, channel, priority, cost rates

CLI support:
- `pm-assist init` prompts for process name, objectives, stakeholders, and KPIs
- Generates a scope file (YAML) used in reports

### Phase B: Extract (read-only)
- Identify systems of record
- Extract event data and reference data (read-only)
- Capture lineage and query metadata

CLI support:
- `pm-assist connect` registers connectors
- `pm-assist ingest` snapshots a reproducible staging dataset
- Always writes a lineage summary (what, when, from where)

### Phase C: Prepare
- Data quality checks and repairs
- Canonical event log construction
- Readiness score and caveats

CLI support:
- `pm-assist map` for mapping
- `pm-assist prepare` for cleaning pipeline (step-by-step opt-in)
- Produces:
  - readiness scorecard
  - key assumptions list
  - issues backlog (what to fix upstream)

### Phase D: Mine and analyse
- Discovery: DFG + petri/BPMN via selected algorithms
- Conformance: deviations, non-compliant traces, fit/precision metrics
- Performance: throughput, waiting time, bottlenecks, resource
- Variants: top variants, long-tail, segmentation
- Drift: compare windows (post-MVP)

CLI support:
- `pm-assist mine` prompts which analyses to run
- Stores intermediate outputs and metrics for reproducibility

### Phase E: Recommend and forecast
- Translate analytics into:
  - intervention points
  - expected benefits
  - implementation plan
- Validate with stakeholders
- Create a business case and ROI

CLI support:
- `pm-assist report` generates a recommended-actions section
- Optional: simulation / what-if module (explicit assumptions)
- Optional: predictive monitoring (model cards and validation)

### Phase F: Package and handover
- Notebook for analysts
- Executive report for sponsors
- Artefacts for audit
- A “next run” recipe

CLI support:
- `pm-assist report` and `pm-assist review`
- Outputs:
  - run bundle
  - run manifest
  - reproducibility checklist

## 2. Pipeline step catalogue (MVP)
Each step must be:
- idempotent (safe to rerun)
- deterministic given the same inputs
- documented with assumptions and parameters

### Data preparation steps (must align with the user’s preferred order)
1. Handling missing values (imputation or removal)
2. Converting data types to appropriate formats
3. Removing duplicate records
4. Detecting and handling outliers
5. Standardising and normalising data
6. Encoding categorical variables
7. Cleaning and preprocessing string data
8. Extracting features from date columns

### Process mining steps
- Event log creation and validation
- DFG discovery
- Inductive Miner discovery
- Heuristic Miner discovery
- BPMN conversion (where feasible)
- Conformance checking (alignments optional due to compute cost)
- Performance annotation
- Variant analysis
- Segmentation filters

## 3. “Question-first” prompting rules
The CLI must ask before:
- dropping rows or columns
- imputing or normalising values
- selecting a mining algorithm
- choosing thresholds (noise, frequency cut-offs)
- running compute-expensive steps (alignments, simulations, profiling)

Prompt design:
- show options
- show defaults
- provide “learn more” text
- allow the user to override via flags for automation

