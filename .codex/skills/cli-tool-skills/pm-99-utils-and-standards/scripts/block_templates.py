#!/usr/bin/env python3
"""Reusable block-style templates for process mining runs."""

MULTI_SOURCE_CONFIG_BLOCK = """{
  "sources": [
    {
      "name": "source_a",
      "path": "path/to/source_a.csv",
      "case": "CaseId",
      "activity": "Activity",
      "timestamp": "Timestamp",
      "resource": "Owner",
      "timestamp_format": "%Y-%m-%d %H:%M:%S"
    },
    {
      "name": "source_b",
      "path": "path/to/source_b.csv",
      "case": "TicketId",
      "activity": "Stage",
      "timestamp": "CreatedAt",
      "resource": "UserId",
      "timestamp_timezone": "UTC"
    }
  ],
  "merge": {"strategy": "concat"}
}"""

RUN_BLOCKS = [
    "python .codex/skills/cli-tool-skills/pm-01-env/scripts/00_detect_env.py --output output",
    "python .codex/skills/cli-tool-skills/pm-01-env/scripts/00_validate_env.py --output output --setup-venv --venv-dir .venv "
    "--requirements .codex/skills/cli-tool-skills/pm-99-utils-and-standards/requirements.txt",
    "python .codex/skills/cli-tool-skills/pm-02-ingest-profile/scripts/01_ingest_multi.py --config path/to/config.json --output output",
    "python .codex/skills/cli-tool-skills/pm-03-data-quality/scripts/02_data_quality.py --file output/stage_01_ingest_profile/normalised_log.csv --output output",
    "python .codex/skills/cli-tool-skills/pm-04-clean-filter/scripts/02_clean_filter.py --file output/stage_02_data_quality/cleaned_log.csv --format csv --output output",
    "python .codex/skills/cli-tool-skills/pm-05-eda/scripts/03_eda.py --use-filtered --output output --format csv --file output/stage_03_clean_filter/filtered_log.csv",
    "python .codex/skills/cli-tool-skills/pm-06-discovery/scripts/04_discover.py --use-filtered --output output --format csv --file output/stage_03_clean_filter/filtered_log.csv",
    "python .codex/skills/cli-tool-skills/pm-07-conformance/scripts/05_conformance.py --use-filtered --output output --format csv --file output/stage_03_clean_filter/filtered_log.csv",
    "python .codex/skills/cli-tool-skills/pm-08-performance/scripts/06_performance.py --use-filtered --output output --format csv --file output/stage_03_clean_filter/filtered_log.csv",
    "python .codex/skills/cli-tool-skills/pm-09-org-mining/scripts/07_org_mining.py --use-filtered --output output --format csv --file output/stage_03_clean_filter/filtered_log.csv",
    "python .codex/skills/cli-tool-skills/pm-10-reporting/scripts/08_report.py --output output"
]
