# python-venv-runner

```yaml
{
  "skill": {
    "name": "python-venv-runner",
    "version": "S1.00",
    "date": "2026-01-11",
    "purpose": "Manage a dedicated Python venv and run pipeline entrypoints reliably from the CLI.",
    "when_to_use": "Use when implementing the Go runner that provisions a venv, installs deps, and executes python modules.",
    "inputs": [
      "ARCHITECTURE.md"
    ],
    "outputs": [
      "Runner module in Go + python package entrypoints"
    ],
    "checklist": [
      "Create venv in a predictable location (XDG preferred)",
      "Pin dependencies and record versions in outputs",
      "Pass secrets via environment variables only",
      "Capture stdout/stderr to run logs",
      "Provide clear error messages for missing python/graphviz/pm4py"
    ],
    "references": [
      "PEP 668 guidance",
      "Python venv documentation"
    ]
  }
}
```

## Guidance
Manage a dedicated Python venv and run pipeline entrypoints reliably from the CLI.

## Checklist
- Create venv in a predictable location (XDG preferred)
- Pin dependencies and record versions in outputs
- Pass secrets via environment variables only
- Capture stdout/stderr to run logs
- Provide clear error messages for missing python/graphviz/pm4py
