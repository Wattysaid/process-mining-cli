# Coding Standards and Engineering Guardrails

**Document version:** R1.00 (2026-01-11)

## 1. General
- Prefer small, testable modules over monolith scripts.
- Every pipeline step must:
  - declare inputs/outputs
  - validate prerequisites
  - log parameters (excluding secrets)
  - produce a small “step summary” markdown

## 2. Go CLI
- Use Cobra for command structure.
- Use structured logging.
- No business logic in command handlers beyond validation and orchestration.

## 3. Python
- Use type hints and dataclasses for configs.
- Keep pm4py usage behind a thin service layer (for testability).
- Avoid notebook-only logic; notebooks should call library functions.

## 4. Testing
- Unit tests for:
  - config parsing and validation
  - pipeline step registry
  - “doctor” checks
- Smoke tests for end-to-end pipeline with a tiny synthetic dataset.

## 5. Documentation
- Every command has:
  - help text
  - examples
- Every pipeline step has:
  - a short README and a parameters table

