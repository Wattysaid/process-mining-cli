# Interaction Patterns

## Decision checkpoint format

All decision checkpoints must follow this structure and ask only about the current phase.

- Ask: short, direct question about the choice.
- Complication: why the decision affects correctness, privacy, interpretability, or downstream results.
- Options: 2 to 4 approaches. Mark the preferred option.
- Impact: what each choice changes in outputs, artefacts, and validation burden.

## Phase gating rules

- Ask questions only for the current phase.
- Do not ask about future phases in advance.
- After a choice is made, run the phase and produce artefacts before moving on.
- If the evidence does not justify a decision, collect more data first.

## Checkpoint template

Ask:
Choose how we should handle <decision> for this phase.

Complication:
The decision changes <quality, compliance, or analytical outcome> and will alter downstream metrics.

Options:
1) <Option A> [preferred]
2) <Option B>
3) <Option C>

Impact:
- Option A: <effect>
- Option B: <effect>
- Option C: <effect>

## Decision logging

- Record chosen options in `manifest.json` under `stages.<stage>.parameters`.
- If a decision changes outputs after a stage ran, bump the revision and re-run the stage.
