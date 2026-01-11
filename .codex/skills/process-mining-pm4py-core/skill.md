# process-mining-pm4py-core

```yaml
{
  "skill": {
    "name": "process-mining-pm4py-core",
    "version": "S1.00",
    "date": "2026-01-11",
    "purpose": "Implement process mining analysis steps using pm4py in a modular, testable way.",
    "when_to_use": "Use when implementing event log creation, discovery algorithms, conformance checking, performance analysis, and variant analysis.",
    "inputs": [
      "WORKFLOWS.md",
      "QA_AND_VALIDATION.md"
    ],
    "outputs": [
      "Python pipelines for mining + exported artefacts"
    ],
    "checklist": [
      "Validate event log readiness before mining",
      "Support DFG, Inductive Miner, Heuristic Miner as MVP",
      "Gate compute-heavy steps behind explicit prompts (alignments)",
      "Export models and metrics to the run folder",
      "Document algorithm choices and parameters in the report appendix"
    ],
    "references": [
      "pm4py documentation",
      "Process Mining Handbook (van der Aalst)"
    ]
  }
}
```

## Guidance
Implement process mining analysis steps using pm4py in a modular, testable way.

## Checklist
- Validate event log readiness before mining
- Support DFG, Inductive Miner, Heuristic Miner as MVP
- Gate compute-heavy steps behind explicit prompts (alignments)
- Export models and metrics to the run folder
- Document algorithm choices and parameters in the report appendix
