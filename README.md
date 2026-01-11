# PM Assist CLI

PM Assist CLI is an interactive, enterprise-ready process mining assistant that guides analysts through end-to-end engagements from data access to discovery, conformance, performance analysis, and reporting. It is a local CLI tool with optional AI assistance (OpenAI, Anthropic, Gemini, or local Ollama) that never makes decisions for the user and never executes destructive actions without confirmation.

The tool is designed to be a trusted, auditable companion: it asks the right questions, proposes next steps, generates reproducible code and notebooks, and keeps an immutable trail of what was run.

## Highlights
- **User-led decisions**: prompts for approval at each step; no silent actions.
- **Profiles & business context**: stores user and business profiles in `.profiles/` and `.business/`.
- **Optional AI**: use AI for summaries, interpretation, and task translation; keep data processing deterministic.
- **Cross-platform**: Linux, macOS, Windows (WSL2). One-line installer.
- **Enterprise-safe**: read-only connectors, secrets via env vars, audit-friendly outputs.

## Quick Start

### Install (Linux/macOS/WSL2)
```bash
curl -fsSL https://YOUR_RELEASES_HOST/install.sh | sh
```

### Verify
```bash
pm-assist doctor
pm-assist version
```

### Create a new project
```bash
pm-assist init
```

### Run a guided workflow
```bash
pm-assist connect
pm-assist ingest
pm-assist prepare
pm-assist mine
pm-assist report
```

## How It Works

PM Assist is a local CLI that orchestrates Python-based skills. It creates a project scaffold, registers data sources, and builds a notebook as you confirm each step. The CLI uses AI only where it adds value (summaries, interpretation, and intent translation), and relies on deterministic Python scripts for data processing.

Core flow:
1. **Initialize**: create project layout and profiles.
2. **Connect**: register read-only data sources.
3. **Ingest**: load and validate data.
4. **Prepare**: clean and transform event logs.
5. **Mine**: run discovery, conformance, performance, and variant analysis.
6. **Report**: generate narrative insights and artifacts.
7. **Review**: automated QA checks and audit summary.

## Example Scenarios

### Scenario 1: Procurement bottlenecks
You are asked to identify bottlenecks in the procurement process.
1. `pm-assist init` creates a project and asks about your role, experience level, and target business.
2. `pm-assist connect` asks how to access procurement data (ERP, CSV exports, database).
3. `pm-assist ingest` validates tables and schema.
4. `pm-assist prepare` cleans timestamps, filters invalid rows, and normalizes activity names.
5. `pm-assist mine` highlights the longest waiting steps and most frequent rework loops.
6. `pm-assist report` creates a summary report and appends all code to the notebook.

### Scenario 2: Duplicate events in a dataset
You want to check a data extract for duplicates.
1. `pm-assist connect` registers the CSV.
2. `pm-assist ingest` previews rows and verifies schema.
3. `pm-assist prepare` scans duplicates and proposes a fix.
4. You approve, and the notebook is updated with the exact code used.

### Scenario 3: Executive-ready report
You need a short executive summary for leadership.
1. Run your mining steps.
2. `pm-assist report` generates a concise Markdown report.
3. If AI is enabled, it drafts a narrative aligned to your industry.

## Configuration and Profiles
- **User profiles**: stored in `.profiles/` as YAML.
- **Business profiles**: stored in `.business/` as YAML.
- **Project config**: `pm-assist.yaml` in the project root.

The CLI uses these profiles to tailor prompts, defaults, and recommendations.

## Security & Privacy
- Credentials are never stored in plaintext; use environment variables.
- Connectors are read-only by default.
- AI use is optional and policy-controlled.
- All outputs are local, deterministic, and audit-friendly.

## Repository Layout
- `cmd/` and `internal/`: Go CLI source code.
- `python/`: Python workflows and scripts.
- `.codex/skills/cli-tool-skills/`: minimum skills library.
- `project-files/`: product specs and architecture docs.
- `scripts/`: installer and packaging scripts.

## Development
```bash
go build ./cmd/pm-assist
./pm-assist version
```

## License
Proprietary. Do not distribute without permission.
