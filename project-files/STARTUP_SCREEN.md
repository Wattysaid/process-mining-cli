# Startup Screen and First-Run Experience

**Document version:** R1.00 (2026-01-11)

This file defines the terminal starting screen, first-run experience, and menu-driven entry flow for the PM Assist CLI.  
It is intentionally **stdout-based (not full-screen TUI)** to remain compatible with SSH, CI, and log capture.

## 1. When the startup screen is shown

Show the startup screen when:
- The user runs `pm-assist` with no subcommand, AND
- `--non-interactive` is NOT set

Skip the startup screen when:
- Any subcommand is provided (e.g. `pm-assist init`, `pm-assist mine`)
- `--non-interactive` is set

## 2. Visual layout (ASCII banner + status + menu)

### Banner

```
            ____  __  __        ___              __
  (\_/)    |  _ \|  \/  |      / _ \ ___ ___ ___/ _\___
  ( •_•)   | |_) | |\/| |_____/ /_)/ __/ __/ _ \ \ / __|
   />[_]   |  __/| |  | |_____/ ___/ (_| (_|  __/\ \__ \
           |_|   |_|  |_|      \/    \___\___\___\__/___/

                PM Assist · Enterprise Process Mining CLI
                -----------------------------------------
```

### Status line

```
Version: 0.1.0 | Python: ready | LLM: not configured | Graphviz: ready
```

Status logic:
- **Python**
  - `ready` -> venv exists and core imports succeed
  - `missing` -> python not found
  - `deps missing` -> pm4py or required libs missing

- **LLM**
  - `not configured` -> no provider or key present
  - `configured` -> provider + key detected (or local Ollama reachable)
  - `disabled by policy` -> config forbids external calls

- **Graphviz**
  - `ready` or `missing` (warn if missing because model export may fail)

### Environment check message

If all critical checks pass:
```
[SUCCESS] Environment check passed.
```

If degraded:
```
[WARN] Some features are unavailable. Run `pm-assist doctor` for details.
```

### Main menu

```
What would you like to do?

  1) Start a new process mining project
  2) Continue an existing project
  3) Run environment diagnostics (doctor)
  4) Configure LLM integration
  5) Exit

Select an option (1–5):
```

## 3. First-time user flow

Trigger when:
- No global config exists AND
- No project detected in current directory

Prompt:

```
It looks like this is your first time using PM Assist.

Before we start, a few quick questions will help set things up.
This takes about 2 minutes.

Continue? (Y/n):
```

Questions:
- Name and role
- Aptitude level (beginner/intermediate/expert)
- Default workspace directory
- Default output formats (notebook, html report, both)
- Whether LLM features should be enabled by default

Creates:
- `~/.config/pm-assist/config.yaml`
- `.profiles/<name>.yaml` in the project directory
- Example project scaffold (optional, user-approved)

## 4. Existing project detection

If `pm-assist.yaml` found in current directory:

```
Active project detected:
  Name: Order-to-Cash Analysis
  Path: /mnt/d/projects/o2c-q4

Last run:
  2026-01-10 18:42 | Completed with warnings

What would you like to do?

  1) Continue pipeline (next recommended step: prepare)
  2) Re-run previous step
  3) Start a new run
  4) View previous results
  5) Back to main menu
```

The CLI should determine the **next recommended step** from the run manifest.

## 5. LLM configuration screen

Menu option:

```
LLM Integration Setup
---------------------

PM Assist can use LLMs (OpenAI, Anthropic, Gemini, or local Ollama) for:
  - guided decision support
  - drafting executive narrative
  - summarising findings

It will NEVER:
  - upload full datasets by default
  - run analysis on your behalf
  - make decisions without your approval

To enable, choose a provider and set its API key (or use local Ollama).

Choose setup method:

  1) Set environment variable now
  2) Add to local .env file (not recommended for enterprise)
  3) Use local Ollama (no key)
  4) Skip for now

Select option:
```

If user chooses option 1:
- prompt for hidden input
- export for current session only
- print instructions for permanent setup

## 6. Interaction rules

- Never block automation:
  - if `--non-interactive`, do not show menus
- Always provide:
  - escape path to exit
  - CLI equivalents of menu actions

Menu options must map to commands:
- New project -> `pm-assist init`
- Continue project -> guided pipeline command
- Doctor -> `pm-assist doctor`
- Configure LLM -> `pm-assist agent setup` (or similar)

## 7. Implementation guidance (Go)

Recommended libraries:
- Cobra (commands)
- fatih/color or charmbracelet/lipgloss (styling)
- AlecAivazis/survey (menus)

Implementation notes:
- Banner printed before Cobra command routing when no subcommand is provided
- Menu routing handled before executing default command
- All menu actions call the same internal handlers as CLI commands

## 8. Enterprise considerations

- Startup screen must be disableable via policy/config
- Must not leak environment details beyond readiness states
- Must be safe over SSH and in restricted terminals
