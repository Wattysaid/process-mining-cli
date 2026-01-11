# cli-go-cobra

```yaml
{
  "skill": {
    "name": "cli-go-cobra",
    "version": "S1.00",
    "date": "2026-01-11",
    "purpose": "Implement a production-grade Go CLI with Cobra, consistent logging, prompts, and exit codes.",
    "when_to_use": "Use when creating or modifying CLI commands, flags, prompts, error handling, or help text.",
    "inputs": [
      "CLI_SPEC.md",
      "ARCHITECTURE.md"
    ],
    "outputs": [
      "Go CLI skeleton with commands and tests"
    ],
    "checklist": [
      "Use Cobra command tree aligned to CLI_SPEC.md",
      "Implement global flags and consistent exit codes",
      "Avoid business logic in handlers; delegate to internal packages",
      "Use structured logging and do not log secrets",
      "Provide examples in help text for each command"
    ],
    "references": [
      "Cobra documentation",
      "Command Line Interface Guidelines"
    ]
  }
}
```

## Guidance
Implement a production-grade Go CLI with Cobra, consistent logging, prompts, and exit codes.

## Checklist
- Use Cobra command tree aligned to CLI_SPEC.md
- Implement global flags and consistent exit codes
- Avoid business logic in handlers; delegate to internal packages
- Use structured logging and do not log secrets
- Provide examples in help text for each command
