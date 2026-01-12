# Notebook and Report Generation

**Document version:** R1.01 (2026-01-12)

## 1. Output philosophy
PM Assist should deliver outputs that enterprises can:
- review
- audit
- share
- rerun

The CLI produces:
- a structured run folder
- a Jupyter notebook (template-based)
- a report (executive + technical appendices)

## 2. Notebook generation requirements
- Notebook must be generated from templates, not hand-edited.
- Notebook should include:
  - setup and environment checks
  - config snapshot
  - data validation results
  - EDA charts and tables
  - event log build steps
  - mining results (models and metrics)
  - findings placeholders (user edits)
- Support:
  - unexecuted notebooks (default)
  - executed notebooks (optional; can be expensive)
- Each confirmed CLI step appends to the notebook for reproducibility.

## 3. Report generation requirements
Report structure (MVP):
1. Executive summary
2. Scope and methodology
3. Data readiness and caveats
4. Process overview (high-level DFG)
5. Key findings:
   - throughput and bottlenecks
   - variant drivers
   - conformance issues
6. Recommendations:
   - quick wins
   - medium-term
   - governance and controls
7. Benefits estimate (simple, transparent assumptions)
8. Appendices:
   - metrics tables
   - parameter settings
   - QA results

Outputs:
- `report.md`
- `report.html`
- pdf export supported when pandoc is available
- `bundle/report_bundle_<run-id>.zip` (manifest, config snapshot, report assets)

## 4. LLM-assisted narrative (optional)
When enabled, LLMs (OpenAI/Anthropic/Gemini/Ollama) are used only for:
- drafting executive narrative from **local computed summaries**
- rephrasing and structuring recommendations
- creating stakeholder-ready wording

LLMs must not be used for:
- computing metrics
- deciding thresholds
- selecting algorithms without user approval

## 5. Visual assets
- Save figures to `outputs/<run-id>/figures/`
- Reference images in notebook and report
- Support a no-figures mode for restricted environments

## 6. Status
- Report bundle generation is implemented (`pm-assist report`).
- Notebook append is wired to user-confirmed steps.
