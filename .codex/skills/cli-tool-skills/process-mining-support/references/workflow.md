# Process Mining Support Workflow

Keep responses structured as:
1) Markdown cell (short narrative)
2) Code cell (runnable)
3) Decision prompt (what to confirm next)

## Step 0: Environment + Inputs

Markdown cell:
```
## Environment and Inputs
We will set up the environment and confirm the input log schema.
```

Code cell:
```
import pandas as pd

DATA_PATH = "data/your_log.csv"
df = pd.read_csv(DATA_PATH, nrows=5)
df.columns.tolist()
```

Decision prompt:
Confirm case/activity/timestamp/resource columns and output folder.

Notebook alignment check (run in terminal):
```
ls -t output/notebooks/R*.*/ | head -n 5
```
Open the latest notebook and confirm referenced output paths (e.g., output/stage_03_clean_filter).

## Step 1: Ingest + Schema Mapping

Markdown cell:
```
## Ingest and Normalize
We map raw columns to process mining schema and generate a normalized log.
```

Code cell (script-based):
```
!python .codex/skills/cli-tool-skills/pm-02-ingest-profile/scripts/01_ingest.py \
  --file data/your_log.csv --format csv \
  --case <case_col> --activity <activity_col> --timestamp <timestamp_col> \
  --resource <resource_col> --output output
```

Decision prompt:
Confirm parsing success and any timestamp format overrides.

## Step 2: Data Quality

Markdown cell:
```
## Data Quality
We assess missing values, duplicates, and ordering issues.
```

Code cell:
```
!python .codex/skills/cli-tool-skills/pm-03-data-quality/scripts/02_data_quality.py \
  --file output/stage_01_ingest_profile/normalised_log.csv \
  --case <case_col> --activity <activity_col> --timestamp <timestamp_col> \
  --resource <resource_col> --output output
```

Decision prompt:
Choose missing/duplicate handling and masking settings.

## Step 3: Clean + Filter

Markdown cell:
```
## Clean and Filter
We apply agreed filtering rules for rare activities and constraints.
```

Code cell:
```
!python .codex/skills/cli-tool-skills/pm-04-clean-filter/scripts/02_clean_filter.py \
  --file output/stage_01_ingest_profile/normalised_log.csv --format csv \
  --case <case_col> --activity <activity_col> --timestamp <timestamp_col> \
  --resource <resource_col> --output output \
  --auto-filter-rare-activities --min-activity-frequency <min_freq>
```

Decision prompt:
Confirm filtered volume and constraints.

## Step 4: EDA

Markdown cell:
```
## Exploratory Analysis
We profile variants, arrivals, and activity distributions.
```

Code cell:
```
!python .codex/skills/cli-tool-skills/pm-05-eda/scripts/03_eda.py \
  --use-filtered --output output --advanced
```

Decision prompt:
Confirm whether to segment by time or case attributes.

## Step 5: Discovery (Compare Models)

Markdown cell:
```
## Process Discovery
We discover process models and compare metrics.
```

Code cell:
```
!python .codex/skills/cli-tool-skills/pm-06-discovery/scripts/04_discover.py \
  --use-filtered --output output --miner-selection auto
```

Decision prompt:
Run inductive vs heuristic vs auto for comparison?

## Step 6: Conformance

Markdown cell:
```
## Conformance
We evaluate fitness and deviations using alignments.
```

Code cell:
```
!python .codex/skills/cli-tool-skills/pm-07-conformance/scripts/05_conformance.py \
  --use-filtered --output output --conformance-method alignments
```

Decision prompt:
Confirm whether per-case deviation output is needed.

## Step 7: Performance

Markdown cell:
```
## Performance Analysis
We analyze case durations and activity waiting times.
```

Code cell:
```
!python .codex/skills/cli-tool-skills/pm-08-performance/scripts/06_performance.py \
  --use-filtered --output output --advanced --sla-hours 72
```

Decision prompt:
Confirm SLA threshold and bottleneck focus.

## Step 8: Org Mining

Markdown cell:
```
## Organizational Mining
We analyze handovers between resources/agents.
```

Code cell:
```
!python .codex/skills/cli-tool-skills/pm-09-org-mining/scripts/07_org_mining.py \
  --use-filtered --output output
```

Decision prompt:
Confirm privacy constraints for resource names.

## Step 9: Reporting

Markdown cell:
```
## Reporting
We compile narrative insights and link to artifacts.
```

Code cell:
```
!python .codex/skills/cli-tool-skills/pm-10-reporting/scripts/08_report.py --output output
```

Decision prompt:
Confirm final report scope and packaging.
