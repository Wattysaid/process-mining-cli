# Process Mining Assistant: Interaction Examples

This reference provides realistic, phase-by-phase example situations and question patterns.
Use it to guide a conversational, data-dependent workflow without front-loading all decisions.

## Guiding Principles

- Work one phase at a time; do not ask future-phase questions early.
- Within a phase, it is acceptable to ask multiple, tightly-related questions.
- Make questions verbose and decision-supportive: state context, why it matters, and what happens next.
- Ask only questions that are necessary based on the data seen so far.

## Phase 0: Environment and Input Discovery

### Situation A: No Python or missing dependencies
Context: The user wants an analysis, but Python or PM4Py is missing.

Ask (env setup decision):
"Choose how we should handle the runtime environment.
Complication: without a working Python environment, the pipeline cannot run and later steps will fail.
Options (preferred marked):
1. Create a project-local virtual environment and install dependencies (preferred).
2. Use the system Python and install dependencies globally.
3. Skip installs and I will only outline the steps.
Impact: option 1 is reproducible and isolated; option 2 is faster but affects the system; option 3 yields guidance only."

Follow-up (only if user chooses option 1):
"To avoid repeated install issues, do you want me to verify that pip is available and run a clean install from `requirements.txt` before any analysis?"

### Situation B: Input format is unclear
Context: The user says "use the data folder" but no format specified.

Ask (input format decision):
"Choose how we should identify the input file and format.
Complication: CSV and XES require different loading steps and column mapping.
Options (preferred marked):
1. I will list files in the data folder and propose the best candidate (preferred).
2. You will specify the exact file name and format.
Impact: option 1 is faster but assumes the best match; option 2 removes ambiguity."

## Phase 1: Schema and Resource Mapping

### Situation C: CSV with likely columns
Context: A CSV exists with columns resembling case/activity/timestamp.

Ask (column mapping decisions in one phase):
"Choose how we should map CSV columns to the event log schema.
Complication: incorrect mapping invalidates all downstream metrics.
Options (preferred marked):
1. Use `case_id` -> case, `activity_name` -> activity, `timestamp` -> timestamp (preferred).
2. Preview 20 rows to confirm or override the mapping.
Impact: option 1 is faster; option 2 reduces risk of mis-mapping."

If user picks option 2, then ask (still in schema phase):
"Choose the resource dimension for 'who is doing it.'
Complication: different resource columns change handover-of-work and bottleneck attribution.
Options (preferred marked):
1. Use `agent_name` as the resource (preferred).
2. Use `adjuster_name` as the resource.
3. Use `user_type` for role-based segmentation.
Impact: this determines organizational mining and role-based performance."

### Situation D: XES input
Context: XES logs have canonical mappings, but resource attributes may vary.

Ask:
"Choose how we should handle resource attributes in the XES log.
Complication: organizational mining needs a clear resource key.
Options (preferred marked):
1. Use `org:resource` if present (preferred).
2. If absent, skip org mining for now and revisit after discovery.
Impact: option 1 enables handover and workload metrics; option 2 avoids incorrect attribution."

## Phase 2: Data Quality and Privacy

### Situation E: Missing timestamps or parse failures detected
Context: Timestamp parse failures exceed thresholds.

Ask (data quality decisions):
"Choose how we should handle timestamp issues.
Complication: timestamp errors distort throughput and bottleneck metrics.
Options (preferred marked):
1. Drop events with invalid timestamps and continue (preferred).
2. Impute timestamps using median by activity.
3. Stop and request upstream data cleanup.
Impact: option 1 may reduce volume; option 2 preserves volume but adds assumptions; option 3 delays analysis."

### Situation F: Sensitive data detected
Context: Columns look like PII.

Ask:
"Choose how we should handle sensitive columns.
Complication: outputs may expose personal data in reports and charts.
Options (preferred marked):
1. Mask sensitive columns by hashing (preferred).
2. Drop sensitive columns entirely.
3. Keep all columns unmodified.
Impact: option 1 preserves structure; option 2 reduces risk but loses detail; option 3 carries privacy risk."

## Phase 3: Filtering and Variant Noise

### Situation G: Many rare activities/variants
Context: Variant count is high; tail variants dominate.

Ask (filtering decisions):
"Choose how we should handle rare activities and variants.
Complication: heavy tails can produce noisy or unreadable models.
Options (preferred marked):
1. Filter activities <1% frequency and variants <1% frequency (preferred).
2. Keep all activities and variants.
3. Use custom thresholds you specify.
Impact: option 1 yields clearer models; option 2 preserves completeness; option 3 customizes sensitivity."

### Situation H: Known start/end constraints
Context: User knows valid start/end activities.

Ask:
"Choose whether to enforce start/end activity filters.
Complication: filters can improve model clarity but may exclude valid behavior.
Options (preferred marked):
1. Apply start/end filters you provide (preferred if known).
2. Do not filter and infer start/end from data.
Impact: option 1 improves interpretability; option 2 preserves all behavior."

## Phase 4: Mining and Conformance

### Situation I: Uncertain miner choice
Context: User wants a model but is unsure which miner to use.

Ask (miner selection decisions):
"Choose the discovery approach.
Complication: miner choice changes model quality and interpretability.
Options (preferred marked):
1. Auto-select miner based on noise/variant metrics (preferred).
2. Inductive miner only (sound model, may overfit).
3. Heuristic miner only (noise-robust, may be unsound).
Impact: this determines the discovered process model and conformance results."

### Situation J: Conformance depth preference
Context: User wants deviations, but time is limited.

Ask:
"Choose the conformance method depth.
Complication: alignment-based conformance is more accurate but slower.
Options (preferred marked):
1. Token-based replay only (preferred for speed).
2. Alignments only (preferred for accuracy).
3. Both token-based and alignments.
Impact: affects runtime and precision of deviation analysis."

## Phase 5: Reporting and Audience

### Situation K: Unsure about audience
Context: The user wants results but hasn't chosen the audience.

Ask (reporting decisions):
"Choose the reporting format and audience.
Complication: report depth and visuals depend on the audience.
Options (preferred marked):
1. Technical report + artifacts (preferred).
2. Executive summary + key charts.
3. Both technical and executive.
Impact: controls report scope and detail level."

### Situation L: Need a decision log
Context: The user wants traceability for decisions.

Ask:
"Do you want a decision log captured in the report outputs?
Complication: decision logs improve auditability but add overhead.
Options (preferred marked):
1. Yes, include a decision log table (preferred).
2. No, keep reports concise.
Impact: determines whether the report records choices and rationale."

## Common Pitfalls to Avoid

- Do not ask schema, privacy, and mining questions in the same turn.
- Do not request all option numbers in a single reply.
- Do not proceed to mining until data quality and filtering choices are resolved.
- Do not mask or drop columns without an explicit user choice.

## Between Phases: Notebook Review Gate

### Situation M: User may have edited the stage notebook
Context: The previous stage produced a notebook and the user could have edited it.

Ask:
"Before we proceed, I checked the stage notebook and it looks like it was updated after the last run.
Complication: if we skip these edits, the next stage may not reflect your changes.
Options (preferred marked):
1. I will re-load the notebook changes and continue (preferred).
2. I will re-run the previous stage using the updated notebook as the source of truth.
Impact: option 1 is faster; option 2 ensures outputs are regenerated from your edits."
