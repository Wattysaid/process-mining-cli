#!/usr/bin/env python3
"""Exploratory data analysis for event logs."""

import argparse
import os
import sys

import numpy as np

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
from process_mining_steps import (
    compute_arrival_metrics,
    compute_start_end,
    compute_statistics,
    compute_variant_stats,
    load_event_log,
    log_to_dataframe,
    plot_activity_distributions,
)


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="EDA for an event log.")
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
    parser.add_argument("--advanced", action="store_true", help="Generate advanced diagnostics artifacts.")
    parser.add_argument("--notebook-revision", default="R1.00", help="Notebook revision label.")
    return parser.parse_args()


def main() -> None:
    args = parse_arguments()
    ensure_output_dir(args.output)
    stage_dir = ensure_stage_dir(args.output, "stage_04_eda")
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
        stats = compute_statistics(event_log)
        start_end = compute_start_end(event_log)
        arrival_metrics = compute_arrival_metrics(event_log)
        summary_path = os.path.join(stage_dir, "summary_stats.json")
        save_json({"stats": stats, "arrival_metrics": arrival_metrics, "start_end": start_end},
                  summary_path)
        df = log_to_dataframe(event_log)
        plot_activity_distributions(df, stage_dir)
        compute_variant_stats(event_log, stage_dir, top_n=10)
        advanced_artifacts = {}
        if args.advanced:
            case_lengths = df.groupby("case:concept:name")["concept:name"].size()
            case_length_path = os.path.join(stage_dir, "case_length_distribution.csv")
            case_lengths.to_csv(case_length_path, header=["event_count"])
            case_length_summary = {
                "mean": float(case_lengths.mean()),
                "median": float(case_lengths.median()),
                "p95": float(case_lengths.quantile(0.95)),
                "max": int(case_lengths.max()),
            }
            case_length_summary_path = os.path.join(stage_dir, "case_length_summary.json")
            save_json(case_length_summary, case_length_summary_path)

            variants = (
                df.groupby("case:concept:name")["concept:name"]
                .apply(lambda x: " -> ".join(x.astype(str)))
                .value_counts()
                .reset_index()
            )
            variants.columns = ["variant", "count"]
            variants["percent"] = variants["count"] / variants["count"].sum() * 100
            variants["cum_percent"] = variants["percent"].cumsum()
            variant_coverage_path = os.path.join(stage_dir, "variant_coverage.csv")
            variants.to_csv(variant_coverage_path, index=False)
            probs = variants["count"] / variants["count"].sum()
            entropy = float(-(probs * np.log2(probs)).sum())
            variant_entropy_path = os.path.join(stage_dir, "variant_entropy.json")
            save_json({"variant_entropy": entropy}, variant_entropy_path)
            advanced_artifacts = {
                "case_length_distribution_csv": case_length_path,
                "case_length_summary_json": case_length_summary_path,
                "variant_coverage_csv": variant_coverage_path,
                "variant_entropy_json": variant_entropy_path,
            }
        notebook_path = ensure_notebook(
            args.output,
            args.notebook_revision,
            "04_eda.ipynb",
            "Exploratory Data Analysis",
            context_lines=[
                "",
                f"- Events: {stats.get('num_events')}",
                f"- Cases: {stats.get('num_cases')}",
                f"- Variants: {stats.get('num_variants')}",
            ],
            code_lines=[
                "import json",
                f"with open(r\"{summary_path}\", \"r\", encoding=\"utf-8\") as handle:",
                "    summary = json.load(handle)",
                "summary",
            ],
        )
        artifacts = {
            "summary_stats_json": summary_path,
            "variant_counts_csv": os.path.join(stage_dir, "variant_counts.csv"),
            "variant_pareto_png": os.path.join(stage_dir, "variant_pareto.png"),
        }
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
                "Re-run EDA after corrections.",
            ],
        )
        exit_with_error(str(exc))


if __name__ == "__main__":
    main()
