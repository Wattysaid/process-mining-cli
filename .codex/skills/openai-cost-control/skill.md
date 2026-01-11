# openai-cost-control

```yaml
{
  "skill": {
    "name": "openai-cost-control",
    "version": "S1.00",
    "date": "2026-01-11",
    "purpose": "Implement OpenAI integration with strict token budgets, redaction, and explicit user consent.",
    "when_to_use": "Use when implementing pm-assist agent or any narrative/report drafting features that call OpenAI.",
    "inputs": [
      "OPENAI_INTEGRATION.md",
      "PRIVACY.md",
      "SECURITY_MODEL.md"
    ],
    "outputs": [
      "LLM client abstraction + budgets + redaction"
    ],
    "checklist": [
      "OpenAI is disabled by default and requires explicit opt-in",
      "Require OPENAI_API_KEY and never store it",
      "Implement per-run budgets and enforce hard stops",
      "Minimise data sent: prefer schema and aggregates",
      "Write a per-run ledger of call counts and token estimates"
    ],
    "references": [
      "OpenAI API docs",
      "OWASP LLM guidance"
    ]
  }
}
```

## Guidance
Implement OpenAI integration with strict token budgets, redaction, and explicit user consent.

## Checklist
- OpenAI is disabled by default and requires explicit opt-in
- Require OPENAI_API_KEY and never store it
- Implement per-run budgets and enforce hard stops
- Minimise data sent: prefer schema and aggregates
- Write a per-run ledger of call counts and token estimates
