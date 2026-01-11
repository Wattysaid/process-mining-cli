#!/usr/bin/env python3
"""Clean and filter an event log."""

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

from common import ensure_notebook, ensure_output_dir, ensure_stage_dir, exit_with_error, parse_list, record_stage_failure, require_file, save_json, write_stage_manifest
from process_mining_steps import (
    apply_filters,
    clean_event_log,
    compute_statistics,
    load_event_log,
    log_to_dataframe,
    require_pm4py,
)


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Clean and filter an event log.")
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
    parser.add_argument("--start-activities", help="Comma-separated start activities to retain.")
    parser.add_argument("--end-activities", help="Comma-separated end activities to retain.")
    parser.add_argument("--top-variants", type=int, help="Keep only the top N variants by frequency.")
    parser.add_argument("--auto-filter-rare-activities", action="store_true",
                        help="Filter activities below the minimum frequency threshold.")
    parser.add_argument("--min-activity-frequency", type=float, default=0.01,
                        help="Minimum activity frequency to retain when filtering.")
    parser.add_argument("--output", default="output", help="Output directory.")
    parser.add_argument("--notebook-revision", default="R1.00", help="Notebook revision label.")
    return parser.parse_args()


def main() -> None:
    args = parse_arguments()
    ensure_output_dir(args.output)
    stage_dir = ensure_stage_dir(args.output, "stage_03_clean_filter")
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
        before_stats = compute_statistics(event_log)
        event_log = clean_event_log(event_log)
        if args.auto_filter_rare_activities:
            df = log_to_dataframe(event_log)
            if "concept:name" in df.columns:
                freq = df["concept:name"].value_counts(normalize=True)
                keep = freq[freq >= args.min_activity_frequency].index.tolist()
                if keep:
                    require_pm4py()
                    import pm4py
                    event_log = pm4py.filter_event_attribute_values(
                        event_log, "concept:name", keep, level="event"
                    )
                    event_log = [trace for trace in event_log if len(trace) > 0]
        event_log = apply_filters(
            event_log,
            start_activities=parse_list(args.start_activities),
            end_activities=parse_list(args.end_activities),
        )
        if args.top_variants:
            require_pm4py()
            import pm4py
            event_log = pm4py.filter_variants_top_k(event_log, args.top_variants)
        after_stats = compute_statistics(event_log)
        df = log_to_dataframe(event_log)
        filtered_csv = os.path.join(stage_dir, "filtered_log.csv")
        df.to_csv(filtered_csv, index=False)
        require_pm4py()
        import pm4py
        filtered_xes = os.path.join(stage_dir, "filtered_log.xes")
        pm4py.write_xes(event_log, filtered_xes)
        summary = {
            "start_activities": parse_list(args.start_activities),
            "end_activities": parse_list(args.end_activities),
            "top_variants": args.top_variants,
            "auto_filter_rare_activities": args.auto_filter_rare_activities,
            "min_activity_frequency": args.min_activity_frequency,
            "before": before_stats,
            "after": after_stats,
        }
        summary_path = os.path.join(stage_dir, "filter_summary.json")
        save_json(summary, summary_path)
        notebook_path = ensure_notebook(
            args.output,
            args.notebook_revision,
            "03_clean_filter.ipynb",
            "Clean and Filter",
            context_lines=[
                "",
                f"- Start activities: {summary.get('start_activities')}",
                f"- End activities: {summary.get('end_activities')}",
            ],
            code_lines=[
                "import pandas as pd",
                f"df = pd.read_csv(r\"{filtered_csv}\")",
                "df.head()",
            ],
        )
        artifacts = {
            "filtered_log_csv": filtered_csv,
            "filtered_log_xes": filtered_xes,
            "filter_summary_json": summary_path,
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
                "Verify case/activity/timestamp mapping and parsing options.",
                "If filtering removes too much data, adjust start/end/top variants.",
                "Re-run clean/filter after correcting inputs.",
            ],
        )
        exit_with_error(str(exc))


if __name__ == "__main__":
    main()
