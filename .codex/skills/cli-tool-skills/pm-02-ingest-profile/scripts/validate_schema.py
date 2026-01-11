#!/usr/bin/env python3
"""Validate schema for CSV event logs."""

import argparse
import os
import sys

import pandas as pd

COMMON_DIR = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "..", "..", "pm-99-utils-and-standards", "scripts")
)
if COMMON_DIR not in sys.path:
    sys.path.insert(0, COMMON_DIR)

from common import exit_with_error, require_file, validate_csv_columns


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Validate a CSV event log schema.")
    parser.add_argument("--file", required=True, help="Path to the CSV event log.")
    parser.add_argument("--case", default="case:concept:name", help="Case ID column name.")
    parser.add_argument("--activity", default="concept:name", help="Activity column name.")
    parser.add_argument("--timestamp", default="time:timestamp", help="Timestamp column name.")
    return parser.parse_args()


def main() -> None:
    args = parse_arguments()
    try:
        require_file(args.file)
        df = pd.read_csv(args.file)
        validate_csv_columns(df, (args.case, args.activity, args.timestamp))
    except Exception as exc:
        exit_with_error(str(exc))
    print("Schema validation passed.")


if __name__ == "__main__":
    main()
