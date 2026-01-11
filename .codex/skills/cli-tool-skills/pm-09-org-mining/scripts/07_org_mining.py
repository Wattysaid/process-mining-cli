#!/usr/bin/env python3
"""Organisational mining for event logs."""

import argparse
import os
import sys

COMMON_DIR = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "..", "..", "pm-99-utils-and-standards", "scripts")
)
ORCHESTRATOR_DIR = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "..", "..", "pm-00-orchestrator", "scripts")
)
for path in (COMMON_DIR, ORCHESTRATOR_DIR):
    if path not in sys.path:
        sys.path.insert(0, path)

from common import ensure_notebook, ensure_output_dir, ensure_stage_dir, exit_with_error, infer_format_from_path, record_stage_failure, require_file, write_stage_manifest
from process_mining_steps import load_event_log, organisational_analysis


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Organisational analysis for event logs.")
    parser.add_argument("--file", help="Path to the event log file (CSV or XES).")
    parser.add_argument("--format", choices=["csv", "xes"], help="Input file format.")
    parser.add_argument("--input-log", help="Override input log path (defaults to --file).")
    parser.add_argument("--input-format", choices=["csv", "xes"], help="Override input format (defaults to --format).")
    parser.add_argument("--use-filtered", action="store_true", help="Use output/stage_03_clean_filter/filtered_log.csv when available.")
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
    stage_dir = ensure_stage_dir(args.output, "stage_08_org_mining")
    try:
        if args.use_filtered or (not args.input_log and not args.file):
            candidate = os.path.join(args.output, "stage_03_clean_filter", "filtered_log.csv")
            if os.path.isfile(candidate):
                args.input_log = candidate
        input_log = args.input_log or args.file
        input_format = args.input_format or args.format
        if not input_log:
            raise FileNotFoundError("Input log not provided. Use --file or --input-log.")
        if not input_format:
            input_format = infer_format_from_path(input_log)
        if not input_format:
            raise ValueError("Unable to infer input format; set --format or --input-format.")
        require_file(input_log)
        event_log = load_event_log(
            input_log,
            input_format,
            args.case,
            args.activity,
            args.timestamp,
            resource_col=args.resource,
            timestamp_format=args.timestamp_format,
            timestamp_dayfirst=args.timestamp_dayfirst,
            timestamp_utc=args.timestamp_utc,
            timestamp_timezone=args.timestamp_timezone,
        )
        handover_path = organisational_analysis(event_log, stage_dir)
        notebook_path = ensure_notebook(
            args.output,
            args.notebook_revision,
            "08_org_mining.ipynb",
            "Organisational Mining",
            context_lines=[
                "",
                "Review handover-of-work results.",
            ],
            code_lines=[
                "import pandas as pd",
                f"handover = pd.read_csv(r\"{handover_path}\")",
                "handover.head()",
            ],
        )
        artifacts = {"handover_of_work_csv": handover_path}
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
                "Ensure resource identifiers are present or map them with --resource.",
                "Prefer --use-filtered after running the clean/filter stage.",
                "Fix timestamp parsing issues or pass --timestamp-format.",
                "Re-run organisational mining after corrections.",
            ],
        )
        exit_with_error(str(exc))


if __name__ == "__main__":
    main()
