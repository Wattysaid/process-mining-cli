#!/usr/bin/env python3
"""Performance analysis for event logs."""

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

from common import ensure_notebook, ensure_output_dir, ensure_stage_dir, exit_with_error, infer_format_from_path, record_stage_failure, require_file, save_json, write_stage_manifest
from process_mining_steps import load_event_log, log_to_dataframe, performance_analysis


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Performance analysis for event logs.")
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
    parser.add_argument("--advanced", action="store_true", help="Generate advanced performance diagnostics.")
    parser.add_argument("--sla-hours", type=float, default=72.0, help="SLA threshold in hours.")
    parser.add_argument("--notebook-revision", default="R1.00", help="Notebook revision label.")
    return parser.parse_args()


def main() -> None:
    args = parse_arguments()
    ensure_output_dir(args.output)
    stage_dir = ensure_stage_dir(args.output, "stage_07_performance")
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
        perf_artifacts, perf_summary = performance_analysis(event_log, stage_dir)
        if perf_summary:
            summary_path = os.path.join(stage_dir, "performance_summary.json")
            save_json(perf_summary, summary_path)
        else:
            summary_path = None
        advanced_artifacts = {}
        if args.advanced:
            df = log_to_dataframe(event_log)
            df = df.sort_values(["case:concept:name", "time:timestamp"])
            df["time:timestamp"] = pd.to_datetime(df["time:timestamp"], errors="coerce")
            df["next_time"] = df.groupby("case:concept:name")["time:timestamp"].shift(-1)
            df["wait_hours"] = (df["next_time"] - df["time:timestamp"]).dt.total_seconds() / 3600.0
            wait_stats = df.groupby("concept:name")["wait_hours"].agg(
                mean="mean",
                median="median",
                p90=lambda x: x.quantile(0.9),
                count="count",
            ).sort_values("mean", ascending=False)
            wait_path = os.path.join(stage_dir, "activity_waiting_time_stats.csv")
            wait_stats.to_csv(wait_path)

            case_start = df.groupby("case:concept:name")["time:timestamp"].min()
            case_end = df.groupby("case:concept:name")["time:timestamp"].max()
            case_duration = (case_end - case_start).dt.total_seconds() / 3600.0
            duration_summary = {
                "mean_hours": float(case_duration.mean()),
                "median_hours": float(case_duration.median()),
                "p90_hours": float(case_duration.quantile(0.9)),
                "p95_hours": float(case_duration.quantile(0.95)),
                "sla_hours": float(args.sla_hours),
                "sla_breach_rate": float((case_duration > args.sla_hours).mean()),
            }
            duration_summary_path = os.path.join(stage_dir, "case_duration_summary.json")
            save_json(duration_summary, duration_summary_path)
            advanced_artifacts = {
                "activity_waiting_time_stats_csv": wait_path,
                "case_duration_summary_json": duration_summary_path,
            }
        notebook_path = ensure_notebook(
            args.output,
            args.notebook_revision,
            "07_performance.ipynb",
            "Performance Analysis",
            context_lines=[
                "",
                "Review case duration and sojourn time outputs.",
            ],
            code_lines=[
                "import pandas as pd",
                f"durations = pd.read_csv(r\"{stage_dir}/case_durations.csv\")",
                "durations.head()",
            ],
        )
        artifacts = {key: value for key, value in perf_artifacts.items()}
        if summary_path:
            artifacts["performance_summary_json"] = summary_path
        artifacts.update(advanced_artifacts)
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
                "Ensure the input log exists and matches the selected format.",
                "Prefer --use-filtered after running the clean/filter stage.",
                "Fix timestamp parsing issues or pass --timestamp-format.",
                "Re-run performance after corrections.",
            ],
        )
        exit_with_error(str(exc))


if __name__ == "__main__":
    main()
