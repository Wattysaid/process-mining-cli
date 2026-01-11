# Architecture: PM Assist CLI

**Document version:** R1.00 (2026-01-11)

## 1. Architecture overview
PM Assist is a two-layer application:

1) **CLI Orchestrator (Go recommended)**
- Handles commands, prompting, validation, configuration, logging, and run management
- Manages a dedicated Python virtual environment for the project
- Calls into Python modules for heavy lifting
- Provides a stable interface for enterprise deployment
- Maintains user profiles in `.profiles/` and adapts prompts to user aptitude

2) **Python Pipeline Library**
- Implements data prep, event log construction, pm4py analysis, notebook/report generation
- Designed as composable pipeline steps
- Supports a "skill registry" of reusable, tested components (sourced from `cli-tool-skills`)

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
  .codex/                        # agent skill packs for Codex + cli-tool-skills library
  .profiles/                     # user profiles (YAML, user-editable)
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
- Assistant-driven edits are limited to `.profiles/` and output artefacts; the tool code and bundled Python assets are not modified at runtime.

## 5. Configuration
- Human-readable YAML (`pm-assist.yaml`)
- Must support:
  - columns mapping (case_id, activity, timestamp, resource)
  - timestamp parsing rules and timezones
  - connector definitions (read-only)
  - pipeline step selection and parameters
  - LLM settings (provider, enabled flag, model, budget caps, offline policy)
  - output formats (notebook/report)
  - profile preferences (prompt level, defaults, UI hints)

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
- The initial registry should map to the bundled `cli-tool-skills` Python scripts to guarantee minimum capability coverage.
- The Python assets should be bundled with the CLI and executed as packaged modules to avoid exposing editable source code.

## 8. Security boundaries
- Secrets only read from:
  - environment variables
  - OS keychain integration (post-MVP)
  - project `.env` file (discouraged for enterprise; allowed for dev)
- Secrets never written to disk, never logged
- LLM calls can be disabled globally and per project


## 9. UX entrypoint
- When invoked without a subcommand, the CLI displays a stdout-based startup screen and menu (see `project-files/STARTUP_SCREEN.md`).
- Menu actions must route to the same internal handlers as the equivalent commands to avoid logic drift.
