#!/usr/bin/env python3
"""Orchestrate the full process mining pipeline."""

import argparse
import logging
import os
import sys
import subprocess
from typing import Dict

import pandas as pd

COMMON_DIR = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "..", "..", "pm-99-utils-and-standards", "scripts")
)
if COMMON_DIR not in sys.path:
    sys.path.insert(0, COMMON_DIR)

from common import (
    ensure_output_dir,
    exit_with_error,
    load_config,
    merge_config,
    parse_list,
    require_file,
    save_json,
    setup_logging,
    write_manifest,
    ExitCodes,
)
from process_mining_steps import (
    apply_filters,
    clean_event_log,
    compute_arrival_metrics,
    compute_start_end,
    compute_statistics,
    compute_variant_stats,
    convert_dataframe_to_event_log,
    discover_models,
    evaluate_models,
    load_event_log,
    load_csv_dataframe,
    log_to_dataframe,
    organisational_analysis,
    performance_analysis,
    plot_activity_distributions,
    run_data_quality_checks,
    save_models,
)


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="End-to-end process mining CLI pipeline.")
    parser.add_argument("--file", required=True, help="Path to the event log file (CSV or XES).")
    parser.add_argument("--format", choices=["csv", "xes"], required=True, help="Input file format.")
    parser.add_argument("--case", default="case:concept:name", help="Case ID column name (CSV only).")
    parser.add_argument("--activity", default="concept:name", help="Activity column name (CSV only).")
    parser.add_argument("--timestamp", default="time:timestamp", help="Timestamp column name (CSV only).")
    parser.add_argument("--resource", help="Resource column name (CSV only).")
    parser.add_argument("--output", default="output", help="Directory to store analysis results.")
    parser.add_argument("--noise-threshold", type=float, default=0.0, help="Noise threshold for inductive miner.")
    parser.add_argument("--dependency-threshold", type=float, default=0.5, help="Dependency threshold for heuristic miner.")
    parser.add_argument("--frequency-threshold", type=float, default=0.0, help="Frequency threshold for heuristic miner.")
    parser.add_argument("--miner-selection", choices=["auto", "inductive", "heuristic", "both"], help="Miner selection strategy.")
    parser.add_argument("--variant-noise-threshold", type=float, help="Variant frequency threshold for auto selection.")
    parser.add_argument("--start-activities", help="Comma-separated start activities to retain.")
    parser.add_argument("--end-activities", help="Comma-separated end activities to retain.")
    parser.add_argument("--missing-value-threshold", type=float, help="Missing value threshold for dropping/imputation.")
    parser.add_argument("--timestamp-parse-threshold", type=float, help="Max allowed timestamp parse failure rate.")
    parser.add_argument("--duplicate-threshold", type=float, help="Warn on duplicate rate above threshold.")
    parser.add_argument("--timestamp-format", help="Explicit timestamp format string for parsing.")
    parser.add_argument("--timestamp-dayfirst", action="store_true", help="Parse timestamps with day-first format.")
    parser.add_argument("--timestamp-utc", action="store_true", help="Parse timestamps as UTC.")
    parser.add_argument("--timestamp-timezone", help="Timezone to localize/convert timestamps.")
    parser.add_argument("--dedupe-keys", help="Comma-separated columns to define duplicate events.")
    parser.add_argument("--order-violation-threshold", type=float, help="Case order violation threshold.")
    parser.add_argument("--fail-on-order-violations", action="store_true",
                        help="Fail if case order violation rate exceeds threshold.")
    parser.add_argument("--auto-filter-rare-activities", action="store_true", default=None, help="Filter low-frequency activities.")
    parser.add_argument("--min-activity-frequency", type=float, help="Min activity frequency for filtering.")
    parser.add_argument("--min-timestamp", help="Minimum timestamp (inclusive).")
    parser.add_argument("--max-timestamp", help="Maximum timestamp (inclusive).")
    parser.add_argument("--impute-missing-timestamps", action="store_true", default=None, help="Impute missing timestamps above threshold.")
    parser.add_argument("--auto-mask-sensitive", action="store_true", default=None, help="Mask detected sensitive columns.")
    parser.add_argument("--sensitive-column-patterns", help="Comma-separated patterns for sensitive columns.")
    parser.add_argument("--timestamp-impute-strategy", choices=["median", "mean"], help="Timestamp imputation strategy.")
    parser.add_argument("--mask-strategy", choices=["hash", "redact", "tokenize"], help="Masking strategy for sensitive columns.")
    parser.add_argument("--mask-salt", help="Optional salt for hash masking.")
    parser.add_argument("--lifecycle-column", default="lifecycle:transition", help="Lifecycle column name for summaries.")
    parser.add_argument("--config", help="Optional JSON/YAML config file.")
    parser.add_argument("-v", "--verbose", action="count", default=0, help="Increase logging verbosity.")
    return parser.parse_args()


def run_env_detection(output_dir: str) -> None:
    detect_script = os.path.join(
        os.path.dirname(__file__),
        "..",
        "..",
        "pm-01-env",
        "scripts",
        "00_detect_env.py",
    )
    detect_script = os.path.abspath(detect_script)
    subprocess.check_call([sys.executable, detect_script, "--output", output_dir])


def generate_report(stats: Dict[str, int],
                    model_metrics: pd.DataFrame,
                    arrival_metrics: Dict[str, float],
                    start_end: Dict[str, Dict[str, int]],
                    data_quality: Dict[str, object],
                    performance_summary: Dict[str, object],
                    output_dir: str,
                    report_path: str) -> None:
    with open(report_path, "w", encoding="utf-8") as handle:
        handle.write("# Process Mining CLI Report\n\n")
        handle.write("## Summary Statistics\n")
        handle.write(f"- Number of events: {stats['num_events']}\n")
        handle.write(f"- Number of cases: {stats['num_cases']}\n")
        handle.write(f"- Number of variants: {stats['num_variants']}\n\n")
        handle.write("## Arrival Metrics\n")
        handle.write(f"- Mean inter-arrival (hours): {arrival_metrics.get('mean_interarrival_hours')}\n")
        handle.write(f"- Median inter-arrival (hours): {arrival_metrics.get('median_interarrival_hours')}\n\n")
        if data_quality:
            handle.write("## Data Quality\n")
            missing = data_quality.get("missing_rates", {})
            handle.write(f"- Missing case IDs: {missing.get('case:concept:name', 0):.2%}\n")
            handle.write(f"- Missing activities: {missing.get('concept:name', 0):.2%}\n")
            handle.write(f"- Missing timestamps: {missing.get('time:timestamp', 0):.2%}\n")
            handle.write(f"- Timestamp parse failure rate: {data_quality.get('timestamp_parse_failure_rate', 0):.2%}\n")
            handle.write(f"- Duplicate rate: {data_quality.get('duplicate_rate', 0):.2%}\n\n")
        if performance_summary:
            handle.write("## Performance Summary\n")
            duration_stats = performance_summary.get("duration_stats", {})
            if duration_stats:
                handle.write(f"- Mean case duration (hours): {duration_stats.get('mean_hours')}\n")
                handle.write(f"- Median case duration (hours): {duration_stats.get('median_hours')}\n")
                handle.write(f"- P95 case duration (hours): {duration_stats.get('p95_hours')}\n")
                handle.write(f"- Max case duration (hours): {duration_stats.get('max_hours')}\n")
            skew_ratio = performance_summary.get("p95_to_median_ratio")
            if skew_ratio:
                handle.write(f"- Heavy tail ratio (P95/Median): {skew_ratio:.2f}\n")
            handle.write("\n")
        handle.write("## Start Activities\n")
        handle.write(pd.DataFrame(list(start_end["start_activities"].items()), columns=["activity", "count"]).to_markdown(index=False))
        handle.write("\n\n")
        handle.write("## End Activities\n")
        handle.write(pd.DataFrame(list(start_end["end_activities"].items()), columns=["activity", "count"]).to_markdown(index=False))
        handle.write("\n\n")
        handle.write("## Model Evaluation\n\n")
        handle.write(model_metrics.to_markdown(index=False))
        handle.write("\n\n")
        handle.write("The output directory includes plots and CSVs for variants, activity distributions, case durations, sojourn times, and organisational handovers. Use these artifacts to identify bottlenecks, deviations, and improvement opportunities.\n")


def main() -> None:
    args = parse_arguments()
    setup_logging(args.verbose)
    run_env_detection(args.output)
    try:
        require_file(args.file)
        config = load_config(args.config)
        params = merge_config(args, config)
        ensure_output_dir(params["output"])
    except Exception as exc:
        exit_with_error(str(exc))

    data_quality = {}
    quality_recommendations = {}
    data_quality_path = None
    try:
        if params["format"] == "csv":
            df = load_csv_dataframe(
                params["file"],
                params["case"],
                params["activity"],
                params["timestamp"],
                resource_col=params.get("resource"),
                timestamp_format=params.get("timestamp_format"),
                timestamp_dayfirst=bool(params.get("timestamp_dayfirst", False)),
                timestamp_utc=params.get("timestamp_utc"),
                timestamp_timezone=params.get("timestamp_timezone"),
            )
            df, data_quality, quality_recommendations = run_data_quality_checks(df, params)
            data_quality_path = os.path.join(params["output"], "data_quality.json")
            save_json(data_quality, data_quality_path)
            if quality_recommendations:
                save_json(quality_recommendations, os.path.join(params["output"], "data_quality_recommendations.json"))
            event_log = convert_dataframe_to_event_log(df)
        else:
            event_log = load_event_log(
                params["file"],
                params["format"],
                params["case"],
                params["activity"],
                params["timestamp"],
                resource_col=params.get("resource"),
                timestamp_format=params.get("timestamp_format"),
                timestamp_dayfirst=bool(params.get("timestamp_dayfirst", False)),
                timestamp_utc=params.get("timestamp_utc"),
                timestamp_timezone=params.get("timestamp_timezone"),
            )
        event_log = clean_event_log(event_log)
        event_log = apply_filters(
            event_log,
            start_activities=parse_list(params.get("start_activities")),
            end_activities=parse_list(params.get("end_activities")),
        )
    except ValueError as exc:
        message = str(exc)
        if "Timestamp parse failure" in message:
            exit_with_error(f"Validation error: {exc}", ExitCodes.TIMESTAMP_ERROR)
        if "Missing required columns" in message:
            exit_with_error(f"Validation error: {exc}", ExitCodes.SCHEMA_ERROR)
        exit_with_error(f"Validation error: {exc}", ExitCodes.MISSING_VALUES_ERROR)
    except Exception as exc:
        exit_with_error(f"Failed to load or filter event log: {exc}", ExitCodes.RUNTIME_ERROR)

    stats = compute_statistics(event_log)
    start_end = compute_start_end(event_log)
    arrival_metrics = compute_arrival_metrics(event_log)
    save_json({"stats": stats, "arrival_metrics": arrival_metrics, "start_end": start_end},
              os.path.join(params["output"], "summary_stats.json"))

    df = log_to_dataframe(event_log)
    dist_artifacts = plot_activity_distributions(df, params["output"])
    variant_artifacts = compute_variant_stats(event_log, params["output"], top_n=10)

    models = discover_models(
        event_log,
        params["output"],
        params["noise_threshold"],
        params["dependency_threshold"],
        params["frequency_threshold"],
        params.get("miner_selection", "auto"),
        float(params.get("variant_noise_threshold", 0.01)),
    )
    saved_models = save_models(models, params["output"])
    model_metrics = evaluate_models(event_log, models, params["output"])
    perf_artifacts, perf_summary = performance_analysis(event_log, params["output"])
    if perf_summary.get("recommendations"):
        save_json(perf_summary, os.path.join(params["output"], "performance_summary.json"))
    org_artifact = organisational_analysis(event_log, params["output"])

    report_path = os.path.join(params["output"], "process_mining_report.md")
    generate_report(stats, model_metrics, arrival_metrics, start_end, data_quality, perf_summary, params["output"], report_path)

    artifacts = {
        **dist_artifacts,
        **variant_artifacts,
        **perf_artifacts,
        "org_handover": org_artifact,
        "model_metrics": os.path.join(params["output"], "model_metrics.csv"),
        "report": report_path,
    }
    if saved_models:
        artifacts["models_manifest"] = os.path.join(params["output"], "models_manifest.json")
    if data_quality_path:
        artifacts["data_quality"] = data_quality_path
    if quality_recommendations:
        artifacts["data_quality_recommendations"] = os.path.join(params["output"], "data_quality_recommendations.json")
    if perf_summary.get("recommendations"):
        artifacts["performance_summary"] = os.path.join(params["output"], "performance_summary.json")
    write_manifest(params["output"], params, artifacts)

    logging.info("Process mining analysis complete. Results saved in %s", params["output"])


if __name__ == "__main__":
    main()
