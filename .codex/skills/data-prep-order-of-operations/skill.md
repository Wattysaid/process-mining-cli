# data-prep-order-of-operations

```yaml
{
  "skill": {
    "name": "data-prep-order-of-operations",
    "version": "S1.00",
    "date": "2026-01-11",
    "purpose": "Implement data preparation in the required order with user-controlled decisions.",
    "when_to_use": "Use when building the prepare pipeline and any automated data cleaning routines.",
    "inputs": [
      "WORKFLOWS.md"
    ],
    "outputs": [
      "Data prep pipeline modules + summaries"
    ],
    "checklist": [
      "Follow the ordered steps: missingness, types, duplicates, outliers, standardise/normalise, encode categoricals, clean strings, date features",
      "Prompt user for each step with recommended defaults",
      "Record all decisions in the config snapshot and step summary",
      "Avoid silent row drops; always show impact counts",
      "Produce before/after quality metrics"
    ],
    "references": [
      "CRISP-DM",
      "KDD",
      "Data quality dimensions"
    ]
  }
}
```

## Guidance
Implement data preparation in the required order with user-controlled decisions.

## Checklist
- Follow the ordered steps: missingness, types, duplicates, outliers, standardise/normalise, encode categoricals, clean strings, date features
- Prompt user for each step with recommended defaults
- Record all decisions in the config snapshot and step summary
- Avoid silent row drops; always show impact counts
- Produce before/after quality metrics
