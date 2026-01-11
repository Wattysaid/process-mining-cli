# CLI Specification: Commands, prompts, and outputs

**Document version:** R1.00 (2026-01-11)

## 1. CLI name and philosophy
Tool name: `pm-assist`  
Design goals:
- “Guided but not paternalistic”: always ask before making a choice that affects results.
- “Explain like a consultant”: show trade-offs and caveats, not just outputs.
- “Stable automation”: every command can be run non-interactively with flags for CI/CD.

## 2. Startup screen (no subcommand)
When the user runs `pm-assist` with no subcommand, show the startup screen and menu described in `project-files/STARTUP_SCREEN.md` (unless `--non-interactive` is set).

## 2. Global flags
- `--config <path>` path to config file (default: `./pm-assist.yaml` if present)
- `--project <path>` project root (default: current directory)
- `--run-id <id>` reuse a run folder
- `--non-interactive` fail if required inputs are missing
- `--log-level debug|info|warn|error`
- `--json` machine-readable output (for automation)
- `--yes` assume “yes” for safe prompts (never for destructive prompts)
- `--llm-provider <openai|anthropic|gemini|ollama|none>` override configured LLM provider for this run
- `--profile <name>` use a specific user profile from `.profiles/`

## 3. Command tree (MVP)
### `pm-assist version`
- Prints version, build metadata, python runtime status.

### `pm-assist doctor`
- Checks environment readiness:
  - Python, venv, required OS deps (graphviz), disk space
  - connector reachability (if configured)

### `pm-assist init`
Interactive:
- Project name
- Default folders
- Output formats desired
- LLM provider preference (optional, can be skipped)
- User profile setup (name, role, aptitude level, preferred depth)
Creates:
- `pm-assist.yaml` template
- `.gitignore` suggestions
- `outputs/` folder
- `data/` folder placeholder
- `docs/` folder for reports
- `.venv/` (project-local Python environment, unless disabled)
- `.profiles/<name>.yaml` (user profile)

### `pm-assist connect`
MVP connectors:
- `file` (CSV/Parquet)
Prompts:
- Path(s)
- Delimiter, encoding
- Row count estimation and sampling approach
Outputs:
- `connectors.yaml` or embedded config in `pm-assist.yaml`

### `pm-assist ingest`
- Loads data (full or sampled) into a canonical staging dataset
Prompts:
- Choose input connector
- Select dataset/table/file
- Sampling options (rows, time window)
Outputs:
- `outputs/<run-id>/staging/` (parquet)
- `outputs/<run-id>/quality/ingest_checks.md`

### `pm-assist map`
- Column mapping and schema validation
Prompts:
- Choose columns for case_id, activity, timestamp, resource (optional)
- Timestamp format and timezone handling
Outputs:
- saved mapping in config snapshot
- column profiling summary

### `pm-assist prepare`
- Data preparation pipeline
Prompts (each step is opt-in with defaults):
- Missing values: drop vs impute (and which strategy)
- Types: coercions and invalid value handling
- Duplicates: definition and dedupe rules
- Outliers: detection method and action
- Standardise/normalise numeric fields
- Encode categorical variables (if needed for predictive steps)
- Clean string columns
- Date feature extraction
Outputs:
- `outputs/<run-id>/event_log/event_log.parquet`
- `outputs/<run-id>/quality/data_prep_summary.md`

### `pm-assist mine`
Prompts:
- Choose discovery algorithms to run:
  - DFG, Inductive Miner, Heuristic Miner (Alpha optional)
- Choose conformance:
  - none, token replay, alignments
- Choose performance analysis:
  - throughput, bottlenecks, resource workload
- Variant analysis depth
Outputs:
- models and plots in `outputs/<run-id>/models/` and `outputs/<run-id>/figures/`
- `outputs/<run-id>/analysis/metrics.json`

### `pm-assist report`
Prompts:
- Notebook: create, execute, or create-only
- Report: Markdown/HTML, executive vs technical depth
- Include LLM narrative generation? (explicit opt-in; provider set in config)
Outputs:
- `outputs/<run-id>/analysis_notebook.ipynb`
- `outputs/<run-id>/report.md` and `report.html` (optional pdf post-MVP)

### `pm-assist review`
- Runs QA suite and produces a single summary:
  - dataset quality
  - modelling caveats
  - assumptions list
  - reproducibility checklist
Outputs:
- `outputs/<run-id>/quality/qa_summary.md`

### `pm-assist agent setup`
- Guides the user through LLM provider configuration
Prompts:
- Provider: OpenAI, Anthropic, Gemini, Ollama, or none
- Model selection (provider-specific defaults)
- Token budget and cost caps
- Offline mode preference
Outputs:
- Updates `pm-assist.yaml` (provider + model, never store API keys)

### `pm-assist profile`
- Create or update user profiles stored in `.profiles/`
Subcommands:
- `pm-assist profile init` (interactive setup)
- `pm-assist profile set --name <name>` (activate profile)
- `pm-assist profile show --name <name>`
Profile fields:
- name, role, aptitude (beginner|intermediate|expert), preferences (prompt depth, defaults)

## 4. `pm-assist agent` (optional, LLM-enabled)
Behaviour:
- Provides guided decision support and report narrative drafting.
 - Asks clarifying questions to gauge user aptitude and goals.
Constraints:
- Must respect token budgets and a per-run cost cap.
- Must avoid sending raw sensitive data by default (use summaries and aggregates).
- Must support “offline mode” with no external LLM calls.
- Must always ask for analytical decisions; never auto-select algorithms, thresholds, or destructive actions.

## 5. Exit codes (standardise)
- 0: success
- 1: unknown error
- 2: invalid arguments / config
- 3: environment missing dependency
- 4: connector error (auth/connection)
- 5: data validation failed (blocking)
- 6: pipeline step failed
- 7: LLM error (only if LLM-enabled steps executed)
