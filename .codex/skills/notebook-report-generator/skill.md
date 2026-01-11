# notebook-report-generator

```yaml
{
  "skill": {
    "name": "notebook-report-generator",
    "version": "S1.00",
    "date": "2026-01-11",
    "purpose": "Generate reproducible notebooks and reports from templates and local artefacts.",
    "when_to_use": "Use when implementing notebook templates, report templates, or narrative generation.",
    "inputs": [
      "NOTEBOOK_AND_REPORTS.md",
      "OPENAI_INTEGRATION.md"
    ],
    "outputs": [
      "Notebook + report outputs"
    ],
    "checklist": [
      "Keep templates in versioned files (not inline strings)",
      "Reference saved figures and metrics from the run folder",
      "Support unexecuted notebooks by default",
      "Include placeholders for findings and recommendations",
      "If OpenAI narrative enabled, use summaries only and respect budgets"
    ],
    "references": [
      "Jupyter nbformat docs",
      "Pandoc/Markdown best practices"
    ]
  }
}
```

## Guidance
Generate reproducible notebooks and reports from templates and local artefacts.

## Checklist
- Keep templates in versioned files (not inline strings)
- Reference saved figures and metrics from the run folder
- Support unexecuted notebooks by default
- Include placeholders for findings and recommendations
- If OpenAI narrative enabled, use summaries only and respect budgets
