#!/usr/bin/env python3
"""Ingest and normalize multiple event logs using a JSON config."""

import argparse
import os
import sys
from typing import Any, Dict, List

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

from common import ensure_notebook, ensure_output_dir, ensure_stage_dir, exit_with_error, load_json, record_stage_failure, save_json, write_stage_manifest
from process_mining_steps import load_csv_dataframe


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Multi-source ingest using a JSON config.")
    parser.add_argument("--config", required=True, help="Path to multi-source ingest config (JSON).")
    parser.add_argument("--output", default="output", help="Output directory.")
    parser.add_argument("--notebook-revision", default="R1.00", help="Notebook revision label.")
    return parser.parse_args()


def apply_column_map(df: pd.DataFrame, column_map: Dict[str, str]) -> pd.DataFrame:
    if column_map:
        df = df.rename(columns=column_map)
    return df


def add_prefix(df: pd.DataFrame, prefix: str, exclude: List[str]) -> pd.DataFrame:
    if not prefix:
        return df
    rename = {col: f"{prefix}{col}" for col in df.columns if col not in exclude}
    return df.rename(columns=rename)


def build_event_id(df: pd.DataFrame, columns: List[str], delimiter: str) -> pd.DataFrame:
    if not columns:
        return df
    existing = [col for col in columns if col in df.columns]
    if not existing:
        return df
    df["event_id"] = df[existing].astype(str).agg(delimiter.join, axis=1)
    return df


def ingest_source(source: Dict[str, Any]) -> pd.DataFrame:
    path = source["path"]
    fmt = source.get("format", "csv")
    if fmt != "csv":
        raise ValueError(f"Unsupported format for multi-source ingest: {fmt}")
    case_col = source.get("case", "case:concept:name")
    activity_col = source.get("activity", "concept:name")
    timestamp_col = source.get("timestamp", "time:timestamp")
    resource_col = source.get("resource")
    df = load_csv_dataframe(
        path,
        case_col,
        activity_col,
        timestamp_col,
        resource_col=resource_col,
        timestamp_format=source.get("timestamp_format"),
        timestamp_dayfirst=bool(source.get("timestamp_dayfirst", False)),
        timestamp_utc=source.get("timestamp_utc"),
        timestamp_timezone=source.get("timestamp_timezone"),
    )
    df = apply_column_map(df, source.get("column_map", {}))
    df = add_prefix(df, source.get("prefix", ""), ["case:concept:name", "concept:name", "time:timestamp", "org:resource"])
    df["source_system"] = source.get("name", os.path.basename(path))
    df = build_event_id(df, source.get("event_id_columns", []), source.get("event_id_delimiter", "::"))
    return df


def merge_sources(sources: List[pd.DataFrame], config: Dict[str, Any]) -> pd.DataFrame:
    if not sources:
        return pd.DataFrame()
    merge_cfg = config.get("merge", {"strategy": "concat"})
    strategy = merge_cfg.get("strategy", "concat")
    if strategy == "join":
        join_keys = merge_cfg.get("join_keys", ["case:concept:name"])
        how = merge_cfg.get("how", "outer")
        merged = sources[0]
        for df in sources[1:]:
            merged = merged.merge(df, on=join_keys, how=how, suffixes=("", "_dup"))
        return merged
    return pd.concat(sources, ignore_index=True)


def apply_case_strategy(df: pd.DataFrame, config: Dict[str, Any]) -> pd.DataFrame:
    strategy = config.get("case_id_strategy")
    if not strategy:
        return df
    if strategy.get("type") != "concat":
        return df
    columns = strategy.get("columns", [])
    delimiter = strategy.get("delimiter", "::")
    existing = [col for col in columns if col in df.columns]
    if not existing:
        raise ValueError("case_id_strategy columns not found in merged data.")
    df["case:concept:name"] = df[existing].astype(str).agg(delimiter.join, axis=1)
    return df


def main() -> None:
    args = parse_arguments()
    ensure_output_dir(args.output)
    stage_dir = ensure_stage_dir(args.output, "stage_01_ingest_profile")
    try:
        config = load_json(args.config)
        sources_cfg = config.get("sources", [])
        if not sources_cfg:
            raise ValueError("No sources defined in config.")
        sources = [ingest_source(source) for source in sources_cfg]
        combined = merge_sources(sources, config)
        combined = apply_case_strategy(combined, config)
        combined = combined.dropna(subset=["case:concept:name", "concept:name", "time:timestamp"])
        combined_path = os.path.join(stage_dir, "normalised_log.csv")
        combined.to_csv(combined_path, index=False)
        profile = {
            "row_count": int(len(combined)),
            "column_count": int(len(combined.columns)),
            "columns": list(combined.columns),
            "missing_rates": {col: float(combined[col].isna().mean()) for col in combined.columns},
            "source_systems": sorted(combined["source_system"].unique().tolist()) if "source_system" in combined.columns else [],
        }
        profile_path = os.path.join(stage_dir, "ingest_profile.json")
        save_json(profile, profile_path)
        notebook_path = ensure_notebook(
            args.output,
            args.notebook_revision,
            "01_ingest_profile.ipynb",
            "Ingest and Profile (Multi-source)",
            context_lines=[
                "",
                f"- Config: {args.config}",
                f"- Sources: {len(sources_cfg)}",
                f"- Rows: {profile['row_count']}",
            ],
            code_lines=[
                "import pandas as pd",
                f"df = pd.read_csv(r\"{combined_path}\")",
                "df.head()",
            ],
        )
        artifacts = {
            "normalised_log_csv": combined_path,
            "ingest_profile_json": profile_path,
            "ingest_config_json": args.config,
        }
        write_stage_manifest(
            stage_dir,
            {"output": args.output, "config": args.config, "notebook_revision": args.notebook_revision},
            artifacts,
            args.notebook_revision,
            notebook_path=notebook_path,
        )
    except Exception as exc:
        record_stage_failure(
            stage_dir,
            str(exc),
            [
                "Verify each source path and mapping fields in the config.",
                "Ensure case/activity/timestamp columns are mapped per source.",
                "Use merge.strategy=concat unless you have reliable join keys.",
                "Re-run multi-source ingest after corrections.",
            ],
        )
        exit_with_error(str(exc))


if __name__ == "__main__":
    main()
