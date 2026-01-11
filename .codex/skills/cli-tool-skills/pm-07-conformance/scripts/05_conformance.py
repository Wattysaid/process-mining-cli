#!/usr/bin/env python3
"""Conformance checking and model evaluation."""

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
from process_mining_steps import conformance_diagnostics, discover_models, evaluate_models, load_event_log, load_models


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Conformance checking for event logs.")
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
    parser.add_argument("--noise-threshold", type=float, default=0.0, help="Noise threshold for inductive miner.")
    parser.add_argument("--dependency-threshold", type=float, default=0.5, help="Dependency threshold for heuristic miner.")
    parser.add_argument("--frequency-threshold", type=float, default=0.0, help="Frequency threshold for heuristic miner.")
    parser.add_argument("--miner-selection", choices=["auto", "inductive", "heuristic", "both"], default="auto", help="Miner selection strategy.")
    parser.add_argument("--variant-noise-threshold", type=float, default=0.01, help="Variant frequency threshold for auto selection.")
    parser.add_argument("--conformance-method", choices=["alignments", "token"], default="alignments",
                        help="Conformance method: alignments or token replay.")
    parser.add_argument("--notebook-revision", default="R1.00", help="Notebook revision label.")
    return parser.parse_args()


def main() -> None:
    args = parse_arguments()
    ensure_output_dir(args.output)
    stage_dir = ensure_stage_dir(args.output, "stage_06_conformance")
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
        models_manifest = os.path.join(args.output, "stage_05_discover", "models_manifest.json")
        models = load_models(models_manifest) if os.path.isfile(models_manifest) else {}
        if not models:
            models = discover_models(
                event_log,
                stage_dir,
                args.noise_threshold,
                args.dependency_threshold,
                args.frequency_threshold,
                args.miner_selection,
                args.variant_noise_threshold,
            )
        evaluate_models(event_log, models, stage_dir)
        conformance_path = conformance_diagnostics(event_log, models, stage_dir, method=args.conformance_method)
        notebook_path = ensure_notebook(
            args.output,
            args.notebook_revision,
            "06_conformance.ipynb",
            "Conformance Checking",
            context_lines=[
                "",
                "This stage computes model fitness and precision metrics.",
            ],
            code_lines=[
                "import pandas as pd",
                f"metrics = pd.read_csv(r\"{stage_dir}/model_metrics.csv\")",
                "metrics",
            ],
        )
        artifacts = {
            "model_metrics_csv": f"{stage_dir}/model_metrics.csv",
            "inductive_petri_net_png": f"{stage_dir}/inductive_miner_petri_net.png",
            "heuristic_petri_net_png": f"{stage_dir}/heuristic_miner_petri_net.png",
        }
        if conformance_path:
            artifacts["conformance_metrics_csv"] = conformance_path
        per_case_path = os.path.join(stage_dir, "conformance_case_deviations.csv")
        if os.path.isfile(per_case_path):
            artifacts["conformance_case_deviations_csv"] = per_case_path
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
                "Ensure discovery has produced models (stage_05_discover/models_manifest.json).",
                "Prefer --use-filtered after running the clean/filter stage.",
                "If models cannot be loaded, re-run discovery then conformance.",
                "Re-run conformance after corrections.",
            ],
        )
        exit_with_error(str(exc))


if __name__ == "__main__":
    main()
