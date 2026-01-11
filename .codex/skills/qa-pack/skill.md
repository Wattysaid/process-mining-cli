# qa-pack

```yaml
{
  "skill": {
    "name": "qa-pack",
    "version": "S1.00",
    "date": "2026-01-11",
    "purpose": "Build a QA pack that validates data and analysis assumptions and produces an actionable backlog.",
    "when_to_use": "Use when implementing ingest checks, event log readiness scoring, and the review command.",
    "inputs": [
      "QA_AND_VALIDATION.md"
    ],
    "outputs": [
      "QA outputs + review command"
    ],
    "checklist": [
      "Implement blocking vs warning checks",
      "Generate both markdown and JSON outputs",
      "Produce an issues backlog (CSV) with severity and suggested fix",
      "In interactive mode, ask if user wants to proceed",
      "In non-interactive mode, fail only on blocking rules"
    ],
    "references": [
      "Great Expectations (conceptually)",
      "Data quality frameworks"
    ]
  }
}
```

## Guidance
Build a QA pack that validates data and analysis assumptions and produces an actionable backlog.

## Checklist
- Implement blocking vs warning checks
- Generate both markdown and JSON outputs
- Produce an issues backlog (CSV) with severity and suggested fix
- In interactive mode, ask if user wants to proceed
- In non-interactive mode, fail only on blocking rules
