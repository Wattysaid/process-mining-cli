#!/usr/bin/env python3
"""Run data quality checks and output a cleaned CSV."""

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

from common import ensure_notebook, ensure_output_dir, ensure_stage_dir, exit_with_error, record_stage_failure, require_file, save_json, write_stage_manifest, ExitCodes
from process_mining_steps import load_csv_dataframe, run_data_quality_checks


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Data quality checks for CSV event logs.")
    parser.add_argument("--file", required=True, help="Path to the CSV event log.")
    parser.add_argument("--case", default="case:concept:name", help="Case ID column name.")
    parser.add_argument("--activity", default="concept:name", help="Activity column name.")
    parser.add_argument("--timestamp", default="time:timestamp", help="Timestamp column name.")
    parser.add_argument("--resource", help="Resource column name.")
    parser.add_argument("--output", default="output", help="Output directory.")
    parser.add_argument("--missing-value-threshold", type=float, default=0.05, help="Missing value threshold.")
    parser.add_argument("--timestamp-parse-threshold", type=float, default=0.02, help="Timestamp parse failure threshold.")
    parser.add_argument("--duplicate-threshold", type=float, default=0.02, help="Duplicate rate threshold.")
    parser.add_argument("--auto-filter-rare-activities", action="store_true", help="Filter low-frequency activities.")
    parser.add_argument("--min-activity-frequency", type=float, default=0.01, help="Min activity frequency for filtering.")
    parser.add_argument("--min-timestamp", help="Minimum timestamp (inclusive).")
    parser.add_argument("--max-timestamp", help="Maximum timestamp (inclusive).")
    parser.add_argument("--impute-missing-timestamps", action="store_true", help="Impute missing timestamps above threshold.")
    parser.add_argument("--timestamp-impute-strategy", choices=["median", "mean"], default="median", help="Timestamp imputation strategy.")
    parser.add_argument("--timestamp-format", help="Explicit timestamp format string for parsing.")
    parser.add_argument("--timestamp-dayfirst", action="store_true", help="Parse timestamps with day-first format.")
    parser.add_argument("--timestamp-utc", action="store_true", help="Parse timestamps as UTC.")
    parser.add_argument("--timestamp-timezone", help="Timezone to localize/convert timestamps.")
    parser.add_argument("--dedupe-keys", help="Comma-separated columns to define duplicate events.")
    parser.add_argument("--order-violation-threshold", type=float, default=0.02, help="Case order violation threshold.")
    parser.add_argument("--fail-on-order-violations", action="store_true",
                        help="Fail if case order violation rate exceeds threshold.")
    parser.add_argument("--auto-mask-sensitive", dest="auto_mask_sensitive", action="store_true",
                        help="Mask detected sensitive columns.")
    parser.add_argument("--no-auto-mask-sensitive", dest="auto_mask_sensitive", action="store_false",
                        help="Disable masking detected sensitive columns.")
    parser.set_defaults(auto_mask_sensitive=True)
    parser.add_argument("--sensitive-column-patterns", help="Comma-separated patterns for sensitive columns.")
    parser.add_argument("--mask-strategy", choices=["hash", "redact", "tokenize"], default="hash",
                        help="Masking strategy for sensitive columns.")
    parser.add_argument("--mask-salt", help="Optional salt for hash masking.")
    parser.add_argument("--lifecycle-column", default="lifecycle:transition",
                        help="Lifecycle column name for summary metrics.")
    parser.add_argument("--notebook-revision", default="R1.00", help="Notebook revision label.")
    return parser.parse_args()


def main() -> None:
    args = parse_arguments()
    ensure_output_dir(args.output)
    stage_dir = ensure_stage_dir(args.output, "stage_02_data_quality")
    try:
        require_file(args.file)
        df = load_csv_dataframe(
            args.file,
            args.case,
            args.activity,
            args.timestamp,
            resource_col=args.resource,
            timestamp_format=args.timestamp_format,
            timestamp_dayfirst=args.timestamp_dayfirst,
            timestamp_utc=args.timestamp_utc,
            timestamp_timezone=args.timestamp_timezone,
        )
        df, quality, recommendations = run_data_quality_checks(df, vars(args))
        cleaned_path = os.path.join(stage_dir, "cleaned_log.csv")
        df.to_csv(cleaned_path, index=False)
        quality_path = os.path.join(stage_dir, "data_quality.json")
        save_json(quality, quality_path)
        if recommendations:
            save_json(recommendations, os.path.join(stage_dir, "data_quality_recommendations.json"))
        notebook_path = ensure_notebook(
            args.output,
            args.notebook_revision,
            "02_data_quality.ipynb",
            "Data Quality Assessment",
            context_lines=[
                "",
                f"- Input: {args.file}",
                f"- Cleaned rows: {len(df)}",
            ],
            code_lines=[
                "import pandas as pd",
                f"df = pd.read_csv(r\"{cleaned_path}\")",
                "df.describe(include='all').head()",
            ],
        )
        artifacts = {
            "cleaned_log_csv": cleaned_path,
            "data_quality_json": quality_path,
        }
        if recommendations:
            artifacts["data_quality_recommendations_json"] = os.path.join(stage_dir, "data_quality_recommendations.json")
        write_stage_manifest(
            stage_dir,
            vars(args),
            artifacts,
            args.notebook_revision,
            notebook_path=notebook_path,
        )
    except ValueError as exc:
        message = str(exc)
        record_stage_failure(
            stage_dir,
            message,
            [
                "Check required columns and timestamp parsing settings.",
                "If parse failures exceed thresholds, fix upstream formats or pass --timestamp-format.",
                "Review case ordering and duplicate keys in data_quality_recommendations.json.",
                "Re-run data quality checks after corrections.",
            ],
        )
        if "Timestamp parse failure" in message:
            exit_with_error(message, ExitCodes.TIMESTAMP_ERROR)
        if "Missing required columns" in message:
            exit_with_error(message, ExitCodes.SCHEMA_ERROR)
        exit_with_error(message, ExitCodes.MISSING_VALUES_ERROR)
    except Exception as exc:
        record_stage_failure(
            stage_dir,
            str(exc),
            [
                "Check required columns and timestamp parsing settings.",
                "If parse failures exceed thresholds, fix upstream formats or pass --timestamp-format.",
                "Review case ordering and duplicate keys in data_quality_recommendations.json.",
                "Re-run data quality checks after corrections.",
            ],
        )
        exit_with_error(str(exc), ExitCodes.RUNTIME_ERROR)


if __name__ == "__main__":
    main()
