#!/usr/bin/env python3
"""Load and normalize an event log."""

import argparse
import os
import sys

import pandas as pd

COMMON_DIR = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "..", "..", "pm-99-utils-and-standards", "scripts")
)
ORCHESTRATOR_DIR = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "..", "..", "pm-00-orchestrator", "scripts")
)
for path in (COMMON_DIR, ORCHESTRATOR_DIR):
    if path not in sys.path:
        sys.path.insert(0, path)

from common import ensure_notebook, ensure_output_dir, ensure_stage_dir, exit_with_error, record_stage_failure, require_file, save_json, write_stage_manifest
from process_mining_steps import load_event_log, log_to_dataframe, require_pm4py


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Ingest and normalize an event log.")
    parser.add_argument("--file", required=True, help="Path to the event log file (CSV or XES).")
    parser.add_argument("--format", choices=["csv", "xes"], required=True, help="Input file format.")
    parser.add_argument("--case", default="case:concept:name", help="Case ID column name (CSV only).")
    parser.add_argument("--activity", default="concept:name", help="Activity column name (CSV only).")
    parser.add_argument("--timestamp", default="time:timestamp", help="Timestamp column name (CSV only).")
    parser.add_argument("--resource", help="Resource column name (CSV only).")
    parser.add_argument("--timestamp-format", help="Explicit timestamp format string for parsing.")
    parser.add_argument("--timestamp-dayfirst", action="store_true", help="Parse timestamps with day-first format.")
    parser.add_argument("--timestamp-utc", action="store_true", help="Parse timestamps as UTC.")
    parser.add_argument("--timestamp-timezone", help="Timezone to localize/convert timestamps.")
    parser.add_argument("--output", default="output", help="Output directory.")
    parser.add_argument("--notebook-revision", default="R1.00", help="Notebook revision label.")
    return parser.parse_args()


def main() -> None:
    args = parse_arguments()
    ensure_output_dir(args.output)
    stage_dir = ensure_stage_dir(args.output, "stage_01_ingest_profile")
    try:
        require_file(args.file)
        event_log = load_event_log(
            args.file,
            args.format,
            args.case,
            args.activity,
            args.timestamp,
            resource_col=args.resource,
            timestamp_format=args.timestamp_format,
            timestamp_dayfirst=args.timestamp_dayfirst,
            timestamp_utc=args.timestamp_utc,
            timestamp_timezone=args.timestamp_timezone,
        )
        df = log_to_dataframe(event_log)
        df.to_csv(os.path.join(stage_dir, "normalised_log.csv"), index=False)
        df.head(50).to_csv(os.path.join(stage_dir, "sample_rows.csv"), index=False)
        ingest_profile = {
            "row_count": int(len(df)),
            "column_count": int(len(df.columns)),
            "columns": list(df.columns),
            "dtypes": {col: str(dtype) for col, dtype in df.dtypes.items()},
            "missing_rates": {col: float(df[col].isna().mean()) for col in df.columns},
            "duplicate_rate": float(df.duplicated().mean()),
        }
        if "time:timestamp" in df.columns:
            parsed = pd.to_datetime(df["time:timestamp"], errors="coerce")
            ingest_profile["timestamp_parse_failure_rate"] = float(parsed.isna().mean())
        profile_path = os.path.join(stage_dir, "ingest_profile.json")
        save_json(ingest_profile, profile_path)
        require_pm4py()
        import pm4py
        normalised_xes = os.path.join(stage_dir, "normalised_log.xes")
        pm4py.write_xes(event_log, normalised_xes)
        notebook_path = ensure_notebook(
            args.output,
            args.notebook_revision,
            "01_ingest_profile.ipynb",
            "Ingest and Profile",
            context_lines=[
                "",
                f"- Input: {args.file}",
                f"- Format: {args.format}",
                f"- Rows: {ingest_profile['row_count']}",
                f"- Columns: {ingest_profile['column_count']}",
            ],
            code_lines=[
                "import pandas as pd",
                f"df = pd.read_csv(r\"{os.path.join(stage_dir, 'normalised_log.csv')}\")",
                "df.head()",
            ],
        )
        artifacts = {
            "normalised_log_csv": os.path.join(stage_dir, "normalised_log.csv"),
            "normalised_log_xes": normalised_xes,
            "sample_rows_csv": os.path.join(stage_dir, "sample_rows.csv"),
            "ingest_profile_json": profile_path,
        }
        write_stage_manifest(
            stage_dir,
            vars(args),
            artifacts,
            args.notebook_revision,
            notebook_path=notebook_path,
        )
    except Exception as exc:
        record_stage_failure(
            stage_dir,
            str(exc),
            [
                "Confirm the input file path and format are correct.",
                "Verify case/activity/timestamp column mappings and timestamp formats.",
                "If timestamps are inconsistent, set --timestamp-format or fix upstream data.",
                "Re-run this stage after correcting the schema or file.",
            ],
        )
        exit_with_error(str(exc))


if __name__ == "__main__":
    main()
