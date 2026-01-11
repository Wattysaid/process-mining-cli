# Architecture: PM Assist CLI

**Document version:** R1.00 (2026-01-11)

## 1. Architecture overview
PM Assist is a two-layer application:

1) **CLI Orchestrator (Go recommended)**
- Handles commands, prompting, validation, configuration, logging, and run management
- Manages a dedicated Python virtual environment for the project
- Calls into Python modules for heavy lifting
- Provides a stable interface for enterprise deployment

2) **Python Pipeline Library**
- Implements data prep, event log construction, pm4py analysis, notebook/report generation
- Designed as composable pipeline steps
- Supports a "snippet registry" of reusable, tested components

## 2. Why a Go CLI + Python pipelines
- Users get a single “tool” with a reliable UX and predictable installation.
- Python remains the best ecosystem fit for pm4py and data science.
- Separation allows secure secret handling and clean enterprise packaging.

## 3. Repository layout (target)
```text
pm-assist/
  cmd/pm-assist/                 # Go entrypoint
  internal/
    cli/                         # command handlers, prompts
    config/                      # config model + merge/validate
    runner/                      # python env + module execution
    telemetry/                   # optional metrics, local only by default
  python/
    pm_assist/                   # python package (pip-installable)
      pipelines/
      snippets/
      report/
      notebook/
      connectors/
      qa/
  scripts/
    install.sh
  .github/workflows/
    release.yml
  project-files/                 # instructions for Codex
  .codex/                        # agent skill packs for Codex
```

## 4. Runtime behaviour
- The CLI creates a project workspace and outputs folder.
- The CLI resolves configuration:
  - defaults -> project config -> command flags -> prompt answers
- The CLI creates (or reuses) a Python venv:
  - stored under `~/.local/share/pm-assist/venv` or within the project (configurable)
- The CLI runs a Python module entrypoint, passing:
  - run id, paths, config, and redacted secrets via environment variables
- Python writes outputs; CLI surfaces progress and summarises results.

## 5. Configuration
- Human-readable YAML (`pm-assist.yaml`)
- Must support:
  - columns mapping (case_id, activity, timestamp, resource)
  - timestamp parsing rules and timezones
  - connector definitions (read-only)
  - pipeline step selection and parameters
  - OpenAI settings (enabled flag, model, budget caps)
  - output formats (notebook/report)

## 6. Logging and audit
- CLI logs: structured, levels, no secrets
- Run artefacts: `outputs/<run-id>/run.log` plus `config.yaml` snapshot
- QA summary: `outputs/<run-id>/quality/qa_summary.md`

## 7. Extensibility model
- Pipeline steps described in a registry:
  - name, description, inputs, outputs, parameters, compatibility constraints
- CLI reads registry to:
  - prompt user which steps to run
  - validate prerequisites
  - execute steps in order
- New steps can be added without changing core CLI logic.

## 8. Security boundaries
- Secrets only read from:
  - environment variables
  - OS keychain integration (post-MVP)
  - project `.env` file (discouraged for enterprise; allowed for dev)
- Secrets never written to disk, never logged
- OpenAI calls can be disabled globally and per project

