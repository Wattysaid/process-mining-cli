#!/usr/bin/env python3
"""Process mining pipeline steps."""

import json
import logging
import os
import hashlib
import zipfile
from typing import Dict, List, Optional, Tuple, Any

import numpy as np
import pandas as pd
import matplotlib.pyplot as plt

try:
    import pm4py
except ImportError:
    pm4py = None

if pm4py is not None:
    try:
        from pm4py.objects.conversion.log import converter as log_converter
        from pm4py.objects.log.util import dataframe_utils
        from pm4py.objects.log.importer.xes import importer as xes_importer
        from pm4py.algo.discovery.inductive import algorithm as inductive_miner
        from pm4py.algo.discovery.heuristics import algorithm as heuristic_miner
        from pm4py.statistics.start_activities.log import get as start_activities_get
        from pm4py.statistics.end_activities.log import get as end_activities_get
        from pm4py.statistics.variants.log import get as variants_get
        from pm4py.visualization.petri_net import visualizer as pn_vis
    except ImportError:
        log_converter = None
        dataframe_utils = None
        xes_importer = None
        inductive_miner = None
        heuristic_miner = None
        start_activities_get = None
        end_activities_get = None
        variants_get = None
        pn_vis = None

    try:
        from pm4py.evaluation.replay_fitness import evaluator as fitness_evaluator
        from pm4py.evaluation.precision import evaluator as precision_evaluator
        from pm4py.evaluation.generalization import evaluator as generalization_evaluator
        from pm4py.evaluation.simplicity import evaluator as simplicity_evaluator
        from pm4py.evaluation.soundness import algorithm as soundness_evaluator
    except ImportError:
        fitness_evaluator = None
        precision_evaluator = None
        generalization_evaluator = None
        simplicity_evaluator = None
        soundness_evaluator = None

    try:
        import pm4py.conformance as pm4py_conformance
    except ImportError:
        pm4py_conformance = None

    try:
        import pm4py.analysis as pm4py_analysis
    except ImportError:
        pm4py_analysis = None

    try:
        from pm4py.algo.conformance.tokenreplay import algorithm as token_replay
    except ImportError:
        token_replay = None
else:
    log_converter = None
    dataframe_utils = None
    xes_importer = None
    inductive_miner = None
    heuristic_miner = None
    start_activities_get = None
    end_activities_get = None
    variants_get = None
    pn_vis = None
    fitness_evaluator = None
    precision_evaluator = None
    generalization_evaluator = None
    simplicity_evaluator = None
    soundness_evaluator = None
    pm4py_conformance = None
    pm4py_analysis = None
    token_replay = None


def require_pm4py() -> None:
    if pm4py is None:
        raise RuntimeError("pm4py is required. Install it before running the pipeline.")


def normalize_timestamps(
    series: pd.Series,
    timestamp_format: Optional[str] = None,
    dayfirst: bool = False,
    utc: Optional[bool] = None,
    timezone: Optional[str] = None,
) -> pd.Series:
    parsed = pd.to_datetime(series, format=timestamp_format, errors="coerce", dayfirst=dayfirst, utc=utc)
    if timezone:
        try:
            if parsed.dt.tz is None:
                parsed = parsed.dt.tz_localize(timezone, nonexistent="shift_forward", ambiguous="NaT")
            else:
                parsed = parsed.dt.tz_convert(timezone)
        except Exception:
            return parsed
    return parsed


def load_csv_dataframe(
    file_path: str,
    case_col: str,
    activity_col: str,
    timestamp_col: str,
    resource_col: Optional[str] = None,
    timestamp_format: Optional[str] = None,
    timestamp_dayfirst: bool = False,
    timestamp_utc: Optional[bool] = None,
    timestamp_timezone: Optional[str] = None,
    delimiter: str = ",",
    encoding: Optional[str] = None,
) -> pd.DataFrame:
    df = pd.read_csv(file_path, sep=delimiter, encoding=encoding or "utf-8")
    rename_map = {
        case_col: "case:concept:name",
        activity_col: "concept:name",
        timestamp_col: "time:timestamp",
    }
    if resource_col:
        rename_map[resource_col] = "org:resource"
    df = df.rename(columns=rename_map)
    if "time:timestamp" in df.columns:
        df["time:timestamp"] = normalize_timestamps(
            df["time:timestamp"],
            timestamp_format=timestamp_format,
            dayfirst=timestamp_dayfirst,
            utc=timestamp_utc,
            timezone=timestamp_timezone,
        )
    return df


def load_excel_dataframe(
    file_path: str,
    case_col: str,
    activity_col: str,
    timestamp_col: str,
    resource_col: Optional[str] = None,
    timestamp_format: Optional[str] = None,
    timestamp_dayfirst: bool = False,
    timestamp_utc: Optional[bool] = None,
    timestamp_timezone: Optional[str] = None,
    sheet: Optional[str] = None,
) -> pd.DataFrame:
    df = pd.read_excel(file_path, sheet_name=sheet or 0)
    return normalize_dataframe_columns(
        df,
        case_col,
        activity_col,
        timestamp_col,
        resource_col=resource_col,
        timestamp_format=timestamp_format,
        timestamp_dayfirst=timestamp_dayfirst,
        timestamp_utc=timestamp_utc,
        timestamp_timezone=timestamp_timezone,
    )


def load_json_dataframe(
    file_path: str,
    case_col: str,
    activity_col: str,
    timestamp_col: str,
    resource_col: Optional[str] = None,
    timestamp_format: Optional[str] = None,
    timestamp_dayfirst: bool = False,
    timestamp_utc: Optional[bool] = None,
    timestamp_timezone: Optional[str] = None,
    json_lines: bool = False,
) -> pd.DataFrame:
    try:
        df = pd.read_json(file_path, lines=json_lines)
    except ValueError:
        df = pd.read_json(file_path, lines=not json_lines)
    return normalize_dataframe_columns(
        df,
        case_col,
        activity_col,
        timestamp_col,
        resource_col=resource_col,
        timestamp_format=timestamp_format,
        timestamp_dayfirst=timestamp_dayfirst,
        timestamp_utc=timestamp_utc,
        timestamp_timezone=timestamp_timezone,
    )


def load_zip_csv_dataframe(
    file_path: str,
    case_col: str,
    activity_col: str,
    timestamp_col: str,
    resource_col: Optional[str] = None,
    timestamp_format: Optional[str] = None,
    timestamp_dayfirst: bool = False,
    timestamp_utc: Optional[bool] = None,
    timestamp_timezone: Optional[str] = None,
    delimiter: str = ",",
    encoding: Optional[str] = None,
    zip_member: Optional[str] = None,
) -> pd.DataFrame:
    if zip_member:
        path = f"zip://{file_path}::{zip_member}"
        df = pd.read_csv(path, sep=delimiter, encoding=encoding or "utf-8")
    else:
        df = pd.read_csv(file_path, sep=delimiter, encoding=encoding or "utf-8", compression="zip")
    return normalize_dataframe_columns(
        df,
        case_col,
        activity_col,
        timestamp_col,
        resource_col=resource_col,
        timestamp_format=timestamp_format,
        timestamp_dayfirst=timestamp_dayfirst,
        timestamp_utc=timestamp_utc,
        timestamp_timezone=timestamp_timezone,
    )


def normalize_dataframe_columns(
    df: pd.DataFrame,
    case_col: str,
    activity_col: str,
    timestamp_col: str,
    resource_col: Optional[str] = None,
    timestamp_format: Optional[str] = None,
    timestamp_dayfirst: bool = False,
    timestamp_utc: Optional[bool] = None,
    timestamp_timezone: Optional[str] = None,
) -> pd.DataFrame:
    rename_map = {
        case_col: "case:concept:name",
        activity_col: "concept:name",
        timestamp_col: "time:timestamp",
    }
    if resource_col:
        rename_map[resource_col] = "org:resource"
    df = df.rename(columns=rename_map)
    if "time:timestamp" in df.columns:
        df["time:timestamp"] = normalize_timestamps(
            df["time:timestamp"],
            timestamp_format=timestamp_format,
            dayfirst=timestamp_dayfirst,
            utc=timestamp_utc,
            timezone=timestamp_timezone,
        )
    return df


def convert_dataframe_to_event_log(df: pd.DataFrame) -> object:
    require_pm4py()
    if log_converter is None or dataframe_utils is None:
        raise RuntimeError("PM4Py conversion helpers are unavailable in this environment.")
    dataframe_utils.convert_timestamp_columns_in_df(df)
    return log_converter.apply(df)


def load_event_log(
    file_path: str,
    log_format: str,
    case_col: str,
    activity_col: str,
    timestamp_col: str,
    resource_col: Optional[str] = None,
    timestamp_format: Optional[str] = None,
    timestamp_dayfirst: bool = False,
    timestamp_utc: Optional[bool] = None,
    timestamp_timezone: Optional[str] = None,
) -> object:
    """Load an event log from XES or CSV."""
    require_pm4py()
    if log_format.lower() == "xes" and xes_importer is None:
        raise RuntimeError("PM4Py XES importer is unavailable in this environment.")
    if log_format.lower() == "xes":
        return xes_importer.apply(file_path)

    df = load_csv_dataframe(
        file_path,
        case_col,
        activity_col,
        timestamp_col,
        resource_col=resource_col,
        timestamp_format=timestamp_format,
        timestamp_dayfirst=timestamp_dayfirst,
        timestamp_utc=timestamp_utc,
        timestamp_timezone=timestamp_timezone,
    )
    df = df.dropna(subset=["case:concept:name", "concept:name", "time:timestamp"])
    df = df.drop_duplicates()
    df = sort_log_dataframe(df)
    return convert_dataframe_to_event_log(df)


def clean_event_log(event_log: object) -> object:
    """Placeholder for log cleaning; currently returns log unchanged."""
    return event_log


def check_case_order(df: pd.DataFrame) -> Tuple[float, int]:
    if df.empty:
        return 0.0, 0
    unsorted_cases = 0
    total_cases = 0
    for _, group in df.groupby("case:concept:name"):
        total_cases += 1
        if group["time:timestamp"].isna().any():
            continue
        if not group["time:timestamp"].is_monotonic_increasing:
            unsorted_cases += 1
    if total_cases == 0:
        return 0.0, 0
    return unsorted_cases / total_cases, unsorted_cases


def run_data_quality_checks(df: pd.DataFrame, config: Dict[str, Any]) -> Tuple[pd.DataFrame, Dict[str, Any], Dict[str, Any]]:
    required = ["case:concept:name", "concept:name", "time:timestamp"]
    missing_columns = [col for col in required if col not in df.columns]
    if missing_columns:
        raise ValueError(f"Missing required columns: {', '.join(missing_columns)}")
    missing = {col: float(df[col].isna().mean()) for col in required if col in df.columns}
    missing_threshold = float(config.get("missing_value_threshold", 0.05))
    timestamp_parse_threshold = float(config.get("timestamp_parse_threshold", 0.02))
    duplicate_threshold = float(config.get("duplicate_threshold", 0.02))
    impute_missing_timestamps = bool(config.get("impute_missing_timestamps", False))
    impute_strategy = config.get("timestamp_impute_strategy", "median")
    auto_mask_sensitive = bool(config.get("auto_mask_sensitive", True))
    sensitive_patterns = config.get(
        "sensitive_column_patterns",
        ["name", "email", "phone", "ssn", "address", "user", "customer", "patient", "employee", "resource"],
    )
    if isinstance(sensitive_patterns, str):
        sensitive_patterns = [item.strip() for item in sensitive_patterns.split(",") if item.strip()]

    df["case:concept:name"] = df["case:concept:name"].astype(str)
    df["concept:name"] = df["concept:name"].astype(str)

    parsed = normalize_timestamps(
        df["time:timestamp"],
        timestamp_format=config.get("timestamp_format"),
        dayfirst=bool(config.get("timestamp_dayfirst", False)),
        utc=config.get("timestamp_utc"),
        timezone=config.get("timestamp_timezone"),
    )
    parse_failure_rate = float(parsed.isna().mean())
    if parse_failure_rate > timestamp_parse_threshold:
        raise ValueError(f"Timestamp parse failure rate {parse_failure_rate:.2%} exceeds threshold {timestamp_parse_threshold:.2%}")
    df["time:timestamp"] = parsed

    recommendations = {}

    for col in ["case:concept:name", "concept:name"]:
        if missing.get(col, 0.0) > missing_threshold:
            raise ValueError(f"Missing values for {col} exceed threshold {missing_threshold:.2%}")
        if missing.get(col, 0.0) > 0:
            df = df.dropna(subset=[col])

    missing_timestamps = missing.get("time:timestamp", 0.0)
    if missing_timestamps > 0:
        if missing_timestamps <= missing_threshold:
            df = df.dropna(subset=["time:timestamp"])
        else:
            if not impute_missing_timestamps:
                raise ValueError("Missing timestamps exceed threshold; enable imputation or clean upstream data.")
            if impute_strategy == "median":
                medians = df.groupby("concept:name")["time:timestamp"].median()
                df["time:timestamp"] = df.apply(
                    lambda row: medians.get(row["concept:name"], pd.NaT) if pd.isna(row["time:timestamp"]) else row["time:timestamp"],
                    axis=1,
                )
                overall_median = df["time:timestamp"].median()
                df["time:timestamp"] = df["time:timestamp"].fillna(overall_median)
            elif impute_strategy == "mean":
                means = df.groupby("concept:name")["time:timestamp"].mean()
                df["time:timestamp"] = df.apply(
                    lambda row: means.get(row["concept:name"], pd.NaT) if pd.isna(row["time:timestamp"]) else row["time:timestamp"],
                    axis=1,
                )
                overall_mean = df["time:timestamp"].mean()
                df["time:timestamp"] = df["time:timestamp"].fillna(overall_mean)
            else:
                raise ValueError(f"Unsupported timestamp_impute_strategy: {impute_strategy}")
            recommendations["timestamp_imputation"] = impute_strategy

    dedupe_keys = config.get("dedupe_keys") or ["case:concept:name", "concept:name", "time:timestamp"]
    if isinstance(dedupe_keys, str):
        dedupe_keys = [item.strip() for item in dedupe_keys.split(",") if item.strip()]
    dedupe_keys = [key for key in dedupe_keys if key in df.columns]
    duplicate_rate = float(df.duplicated().mean())
    key_duplicate_rate = float(df.duplicated(subset=dedupe_keys).mean()) if dedupe_keys else duplicate_rate
    if duplicate_rate > 0:
        df = df.drop_duplicates()
    if key_duplicate_rate > duplicate_threshold:
        recommendations["high_duplicate_rate"] = key_duplicate_rate
        recommendations["duplicate_action"] = f"dropped_duplicates_on_{'+'.join(dedupe_keys) or 'row'}"

    if auto_mask_sensitive:
        mask_strategy = str(config.get("mask_strategy", "hash"))
        mask_salt = str(config.get("mask_salt", ""))
        lower_cols = {col: col.lower() for col in df.columns}
        sensitive_cols = [
            col for col, lower in lower_cols.items()
            if any(pattern in lower for pattern in sensitive_patterns)
        ]
        if sensitive_cols:
            for col in sensitive_cols:
                if mask_strategy == "redact":
                    df[col] = "***"
                elif mask_strategy == "tokenize":
                    tokens = {value: f"{col}_{idx}" for idx, value in enumerate(df[col].astype(str).unique())}
                    df[col] = df[col].astype(str).map(tokens)
                else:
                    df[col] = df[col].astype(str).apply(
                        lambda value: hashlib.sha256((mask_salt + value).encode("utf-8")).hexdigest()
                    )
            recommendations["masked_sensitive_columns"] = sensitive_cols
            recommendations["mask_strategy"] = mask_strategy

    if config.get("min_timestamp") or config.get("max_timestamp"):
        min_ts = pd.to_datetime(config.get("min_timestamp")) if config.get("min_timestamp") else None
        max_ts = pd.to_datetime(config.get("max_timestamp")) if config.get("max_timestamp") else None
        if min_ts is not None:
            df = df[df["time:timestamp"] >= min_ts]
        if max_ts is not None:
            df = df[df["time:timestamp"] <= max_ts]

    if config.get("auto_filter_rare_activities"):
        min_freq = float(config.get("min_activity_frequency", 0.01))
        freq = df["concept:name"].value_counts(normalize=True)
        keep = freq[freq >= min_freq].index
        df = df[df["concept:name"].isin(keep)]
        recommendations["activity_filter_min_freq"] = min_freq
    else:
        freq = df["concept:name"].value_counts(normalize=True)
        if not freq.empty:
            suggested = max(0.01, 1.0 / max(len(freq), 1))
            recommendations["suggested_min_activity_frequency"] = suggested

    order_violation_rate, unsorted_cases = check_case_order(df)
    order_violation_threshold = float(config.get("order_violation_threshold", 0.02))
    if order_violation_rate > order_violation_threshold:
        recommendations["case_order_violation_rate"] = order_violation_rate
        recommendations["case_order_action"] = "sort_by_case_timestamp"
        if config.get("fail_on_order_violations"):
            raise ValueError(
                f"Case order violation rate {order_violation_rate:.2%} exceeds threshold {order_violation_threshold:.2%}"
            )

    lifecycle_column = config.get("lifecycle_column", "lifecycle:transition")
    lifecycle_summary = {}
    if isinstance(lifecycle_column, str) and lifecycle_column in df.columns:
        lifecycle_counts = df[lifecycle_column].value_counts().head(10)
        lifecycle_summary = {str(k): int(v) for k, v in lifecycle_counts.items()}

    quality = {
        "missing_rates": missing,
        "timestamp_parse_failure_rate": parse_failure_rate,
        "duplicate_rate": duplicate_rate,
        "key_duplicate_rate": key_duplicate_rate,
        "case_order_violation_rate": order_violation_rate,
        "unsorted_cases": unsorted_cases,
        "lifecycle_summary": lifecycle_summary,
        "rows_after_cleaning": int(len(df)),
    }
    return df, quality, recommendations


def apply_filters(event_log: object,
                  start_activities: Optional[List[str]] = None,
                  end_activities: Optional[List[str]] = None) -> object:
    """Filter the log by start/end activities using PM4Py helpers."""
    require_pm4py()
    if start_activities:
        event_log = pm4py.filter_start_activities(event_log, start_activities)
    if end_activities:
        event_log = pm4py.filter_end_activities(event_log, end_activities)
    return event_log


def compute_statistics(event_log: object) -> Dict[str, int]:
    """Compute basic log stats."""
    num_cases = len(event_log)
    num_events = sum(len(trace) for trace in event_log)
    variants = variants_get.get_variants(event_log)
    num_variants = len(variants)
    return {
        "num_events": num_events,
        "num_cases": num_cases,
        "num_variants": num_variants,
    }


def log_to_dataframe(event_log: object) -> pd.DataFrame:
    require_pm4py()
    return log_converter.apply(event_log, variant=log_converter.Variants.TO_DATA_FRAME)


def sort_log_dataframe(df: pd.DataFrame) -> pd.DataFrame:
    if "case:concept:name" in df.columns and "time:timestamp" in df.columns:
        return df.sort_values(["case:concept:name", "time:timestamp"])
    return df


def plot_activity_distributions(df: pd.DataFrame, output_dir: str) -> Dict[str, str]:
    """Plot activity distributions by hour, weekday, month, and throughput over time."""
    df = df.copy()
    df["time:timestamp"] = pd.to_datetime(df["time:timestamp"])
    df["hour"] = df["time:timestamp"].dt.hour
    df["weekday"] = df["time:timestamp"].dt.day_name()
    df["month_num"] = df["time:timestamp"].dt.month
    df["date"] = df["time:timestamp"].dt.date

    artifacts = {}
    activity_counts = df["concept:name"].value_counts()
    activity_counts.to_csv(os.path.join(output_dir, "activity_frequency.csv"), header=["count"])
    artifacts["activity_frequency"] = os.path.join(output_dir, "activity_frequency.csv")
    plt.figure(figsize=(10, 6))
    df.groupby("hour")["concept:name"].count().plot(kind="bar", color="skyblue")
    plt.xlabel("Hour of Day")
    plt.ylabel("Number of Events")
    plt.title("Activity Distribution by Hour")
    plt.tight_layout()
    path = os.path.join(output_dir, "activity_distribution_hour.png")
    plt.savefig(path)
    plt.close()
    artifacts["activity_distribution_hour"] = path

    plt.figure(figsize=(10, 6))
    df.groupby("weekday")["concept:name"].count().reindex(
        ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"]
    ).plot(kind="bar", color="teal")
    plt.xlabel("Day of Week")
    plt.ylabel("Number of Events")
    plt.title("Activity Distribution by Weekday")
    plt.tight_layout()
    path = os.path.join(output_dir, "activity_distribution_weekday.png")
    plt.savefig(path)
    plt.close()
    artifacts["activity_distribution_weekday"] = path

    plt.figure(figsize=(10, 6))
    month_counts = df.groupby("month_num")["concept:name"].count()
    month_counts.plot(kind="bar", color="slateblue")
    plt.xlabel("Month")
    plt.ylabel("Number of Events")
    plt.title("Activity Distribution by Month")
    plt.tight_layout()
    path = os.path.join(output_dir, "activity_distribution_month.png")
    plt.savefig(path)
    plt.close()
    artifacts["activity_distribution_month"] = path

    daily_counts = df.groupby("date")["concept:name"].count()
    plt.figure(figsize=(10, 5))
    daily_counts.plot(color="darkorange")
    plt.xlabel("Date")
    plt.ylabel("Events per Day")
    plt.title("Event Throughput Over Time")
    plt.tight_layout()
    path = os.path.join(output_dir, "event_throughput_timeseries.png")
    plt.savefig(path)
    plt.close()
    artifacts["event_throughput_timeseries"] = path

    case_starts = df.groupby("case:concept:name")["time:timestamp"].min().sort_values()
    case_start_counts = case_starts.dt.date.value_counts().sort_index()
    plt.figure(figsize=(10, 5))
    case_start_counts.plot(color="seagreen")
    plt.xlabel("Date")
    plt.ylabel("Case Arrivals")
    plt.title("Case Arrivals Over Time")
    plt.tight_layout()
    path = os.path.join(output_dir, "case_arrival_timeseries.png")
    plt.savefig(path)
    plt.close()
    artifacts["case_arrival_timeseries"] = path

    return artifacts


def compute_variant_stats(event_log: object, output_dir: str, top_n: int = 10) -> Dict[str, str]:
    variants = variants_get.get_variants(event_log)
    counts = {variant: len(traces) for variant, traces in variants.items()}
    df = pd.DataFrame(list(counts.items()), columns=["variant", "count"]).sort_values(
        "count", ascending=False
    )
    df["percent"] = df["count"] / df["count"].sum() * 100
    df["cum_percent"] = df["percent"].cumsum()
    df.to_csv(os.path.join(output_dir, "variant_counts.csv"), index=False)

    # Pareto chart
    plt.figure(figsize=(10, 6))
    top_df = df.head(top_n)
    ax = top_df.plot(kind="bar", x="variant", y="count", legend=False, color="coral")
    ax2 = ax.twinx()
    ax2.plot(top_df["cum_percent"].values, color="black", marker="o")
    ax2.set_ylabel("Cumulative %")
    ax.set_xlabel("Variant")
    ax.set_ylabel("Case Count")
    plt.title("Top Variants Pareto")
    plt.tight_layout()
    pareto_path = os.path.join(output_dir, "variant_pareto.png")
    plt.savefig(pareto_path)
    plt.close()

    return {
        "variant_counts": os.path.join(output_dir, "variant_counts.csv"),
        "variant_pareto": pareto_path,
    }


def compute_arrival_metrics(event_log: object) -> Dict[str, float]:
    start_times = []
    for trace in event_log:
        if trace:
            start_times.append(trace[0]["time:timestamp"])
    if len(start_times) < 2:
        return {"mean_interarrival_hours": float("nan")}
    start_times = sorted(start_times)
    inter_arrivals = [
        (start_times[i] - start_times[i - 1]).total_seconds() / 3600.0
        for i in range(1, len(start_times))
    ]
    return {
        "mean_interarrival_hours": float(np.mean(inter_arrivals)),
        "median_interarrival_hours": float(np.median(inter_arrivals)),
    }


def compute_case_duration_stats(event_log: object) -> Dict[str, float]:
    durations = []
    for trace in event_log:
        if not trace:
            continue
        duration = (trace[-1]["time:timestamp"] - trace[0]["time:timestamp"]).total_seconds() / 3600.0
        durations.append(duration)
    if not durations:
        return {}
    series = pd.Series(durations)
    return {
        "mean_hours": float(series.mean()),
        "median_hours": float(series.median()),
        "p95_hours": float(series.quantile(0.95)),
        "max_hours": float(series.max()),
    }


def discover_models(event_log: object, output_dir: str, noise_threshold: float,
                    dependency_threshold: float, frequency_threshold: float,
                    miner_selection: str = "auto", variant_noise_threshold: float = 0.01) -> Dict[str, object]:
    require_pm4py()
    if variants_get is None or pn_vis is None:
        raise RuntimeError("PM4Py discovery helpers are unavailable in this environment.")
    models = {}
    selection = miner_selection.lower()
    if selection == "auto":
        variants = variants_get.get_variants(event_log)
        num_cases = max(len(event_log), 1)
        variant_count = len(variants)
        low_freq_variants = sum(
            1 for traces in variants.values()
            if len(traces) / num_cases < variant_noise_threshold
        )
        noisy = variant_count / num_cases > 0.5 or (variant_count > 0 and low_freq_variants / variant_count > 0.5)
        selection = "heuristic" if noisy else "inductive"
        if selection == "heuristic" and frequency_threshold < 0.02:
            frequency_threshold = max(frequency_threshold, 0.02)

    if selection in ("inductive", "both"):
        try:
            if hasattr(pm4py, "discover_petri_net_inductive"):
                net_ind, im_ind, fm_ind = pm4py.discover_petri_net_inductive(
                    event_log,
                    noise_threshold=noise_threshold,
                )
            else:
                net_ind, im_ind, fm_ind = inductive_miner.apply(
                    event_log,
                    parameters={"noise_threshold": noise_threshold},
                )
            models["inductive"] = (net_ind, im_ind, fm_ind)
            try:
                gviz_ind = pn_vis.apply(net_ind, im_ind, fm_ind)
                pn_vis.save(gviz_ind, os.path.join(output_dir, "inductive_miner_petri_net.png"))
            except Exception as exc:
                logging.warning("Inductive miner visualization failed: %s", exc)
        except Exception as exc:
            logging.warning("Inductive miner failed: %s", exc)

    if selection in ("heuristic", "both"):
        try:
            if hasattr(pm4py, "discover_petri_net_heuristics"):
                net_heu, im_heu, fm_heu = pm4py.discover_petri_net_heuristics(
                    event_log,
                    dependency_threshold=dependency_threshold,
                )
            else:
                net_heu, im_heu, fm_heu = heuristic_miner.apply_heu(
                    event_log,
                    dependency_threshold=dependency_threshold,
                    frequency_threshold=frequency_threshold,
                )
            models["heuristic"] = (net_heu, im_heu, fm_heu)
            try:
                gviz_heu = pn_vis.apply(net_heu, im_heu, fm_heu)
                pn_vis.save(gviz_heu, os.path.join(output_dir, "heuristic_miner_petri_net.png"))
            except Exception as exc:
                logging.warning("Heuristic miner visualization failed: %s", exc)
        except Exception as exc:
            logging.warning("Heuristic miner failed: %s", exc)
    return models


def save_models(models: Dict[str, Tuple], output_dir: str) -> Dict[str, str]:
    require_pm4py()
    saved = {}
    for name, (net, im, fm) in models.items():
        pnml_path = os.path.join(output_dir, f"{name}_petri_net.pnml")
        try:
            if hasattr(pm4py, "write_pnml"):
                pm4py.write_pnml(net, im, fm, pnml_path)
            else:
                from pm4py.objects.petri_net.exporter import exporter as pnml_exporter
                pnml_exporter.apply(net, im, pnml_path)
            saved[name] = pnml_path
        except Exception:
            continue
    if saved:
        with open(os.path.join(output_dir, "models_manifest.json"), "w", encoding="utf-8") as handle:
            json.dump(saved, handle, indent=2)
    return saved


def load_models(models_manifest: str) -> Dict[str, Tuple]:
    require_pm4py()
    if not os.path.isfile(models_manifest):
        return {}
    with open(models_manifest, "r", encoding="utf-8") as handle:
        saved = json.load(handle)
    models = {}
    for name, pnml_path in saved.items():
        if not os.path.isfile(pnml_path):
            continue
        try:
            if hasattr(pm4py, "read_pnml"):
                net, im, fm = pm4py.read_pnml(pnml_path)
            else:
                from pm4py.objects.petri_net.importer import importer as pnml_importer
                net, im, fm = pnml_importer.apply(pnml_path)
            models[name] = (net, im, fm)
        except Exception:
            continue
    return models


def conformance_diagnostics(event_log: object, models: Dict[str, Tuple], output_dir: str,
                            method: str = "alignments") -> Optional[str]:
    if not models:
        return None
    case_ids = []
    try:
        for idx, trace in enumerate(event_log):
            case_id = None
            if hasattr(trace, "attributes"):
                case_id = trace.attributes.get("concept:name") or trace.attributes.get("case:concept:name")
            case_ids.append(str(case_id) if case_id is not None else str(idx))
    except Exception:
        case_ids = [str(idx) for idx, _ in enumerate(event_log)]
    rows = []
    per_case_rows = []
    for name, (net, im, fm) in models.items():
        try:
            method_key = method.lower()
            if method_key == "token" and token_replay is not None:
                replay = token_replay.apply(event_log, net, im, fm)
                fitness = [item.get("fitness", 0) for item in replay if isinstance(item, dict)]
                row = {
                    "model": name,
                    "method": "token",
                    "cases": len(replay),
                    "avg_fitness": float(np.mean(fitness)) if fitness else 0.0,
                    "min_fitness": float(np.min(fitness)) if fitness else 0.0,
                }
                rows.append(row)
            else:
                if pm4py_conformance is None:
                    continue
                alignments = pm4py_conformance.conformance_diagnostics_alignments(event_log, net, im, fm)
                costs = [item.get("cost", 0) for item in alignments if isinstance(item, dict)]
                row = {
                    "model": name,
                    "method": "alignments",
                    "cases": len(alignments),
                    "avg_cost": float(np.mean(costs)) if costs else 0.0,
                    "max_cost": float(np.max(costs)) if costs else 0.0,
                }
                rows.append(row)
                for idx, item in enumerate(alignments):
                    if not isinstance(item, dict):
                        continue
                    alignment = item.get("alignment") or []
                    log_moves = 0
                    model_moves = 0
                    for move in alignment:
                        if not isinstance(move, (list, tuple)) or len(move) < 2:
                            continue
                        log_move, model_move = move[0], move[1]
                        if log_move == ">>" and model_move != ">>":
                            model_moves += 1
                        elif model_move == ">>" and log_move != ">>":
                            log_moves += 1
                    per_case_rows.append({
                        "model": name,
                        "case_id": case_ids[idx] if idx < len(case_ids) else str(idx),
                        "alignment_cost": item.get("cost"),
                        "fitness": item.get("fitness"),
                        "deviation_count": log_moves + model_moves,
                        "log_move_count": log_moves,
                        "model_move_count": model_moves,
                    })
        except Exception:
            continue
    if not rows:
        return None
    df = pd.DataFrame(rows)
    path = os.path.join(output_dir, "conformance_metrics.csv")
    df.to_csv(path, index=False)
    if per_case_rows:
        per_case_df = pd.DataFrame(per_case_rows)
        per_case_path = os.path.join(output_dir, "conformance_case_deviations.csv")
        per_case_df.to_csv(per_case_path, index=False)
    return path


def evaluate_models(event_log: object, models: Dict[str, Tuple], output_dir: str) -> pd.DataFrame:
    rows = []
    for name, (net, im, fm) in models.items():
        try:
            if fitness_evaluator is not None:
                fitness = fitness_evaluator.apply(event_log, net, im, fm)
            elif pm4py_conformance is not None:
                fitness = pm4py_conformance.fitness_alignments(event_log, net, im, fm)
            else:
                fitness = {"averageFitness": np.nan}
        except Exception:
            fitness = {"averageFitness": np.nan}
        try:
            if precision_evaluator is not None:
                precision = precision_evaluator.apply(event_log, net, im, fm)
            elif pm4py_conformance is not None:
                precision = pm4py_conformance.precision_alignments(event_log, net, im, fm)
            else:
                precision = np.nan
        except Exception:
            precision = np.nan
        try:
            if generalization_evaluator is not None:
                generalisation = generalization_evaluator.apply(event_log, net, im, fm)
            elif pm4py_conformance is not None:
                generalisation = pm4py_conformance.generalization_tbr(event_log, net, im, fm)
            else:
                generalisation = np.nan
        except Exception:
            generalisation = np.nan
        try:
            if simplicity_evaluator is not None:
                simplicity = simplicity_evaluator.apply(net)
            elif pm4py_analysis is not None:
                simplicity = pm4py_analysis.simplicity_petri_net(net, im, fm)
            else:
                simplicity = np.nan
        except Exception:
            simplicity = np.nan
        try:
            if soundness_evaluator is not None:
                soundness = soundness_evaluator.apply(net)
            elif pm4py_analysis is not None:
                soundness = pm4py_analysis.check_soundness(net, im, fm)[0]
            else:
                soundness = np.nan
        except Exception:
            soundness = np.nan
        rows.append({
            "model": name,
            "fitness": (
                fitness.get("averageFitness")
                if isinstance(fitness, dict) and "averageFitness" in fitness
                else fitness.get("log_fitness")
                if isinstance(fitness, dict) and "log_fitness" in fitness
                else fitness.get("fitness")
                if isinstance(fitness, dict) and "fitness" in fitness
                else fitness
            ),
            "precision": precision,
            "generalisation": generalisation,
            "simplicity": simplicity,
            "soundness": soundness,
        })
    df = pd.DataFrame(rows)
    df.to_csv(os.path.join(output_dir, "model_metrics.csv"), index=False)
    return df


def performance_analysis(event_log: object, output_dir: str) -> Tuple[Dict[str, str], Dict[str, Any]]:
    case_durations = []
    case_duration_by_start = []
    sojourn_times: Dict[str, List[float]] = {}
    for trace in event_log:
        if not trace:
            continue
        start_time = trace[0]["time:timestamp"]
        end_time = trace[-1]["time:timestamp"]
        duration = (end_time - start_time).total_seconds() / 3600.0
        case_durations.append(duration)
        case_duration_by_start.append((start_time, duration))
        for idx, event in enumerate(trace):
            act = event["concept:name"]
            if idx < len(trace) - 1:
                next_time = trace[idx + 1]["time:timestamp"]
                sojourn = (next_time - event["time:timestamp"]).total_seconds() / 3600.0
                sojourn_times.setdefault(act, []).append(sojourn)

    duration_df = pd.DataFrame({"case_duration_hours": case_durations})
    duration_df.to_csv(
        os.path.join(output_dir, "case_durations.csv"), index=False
    )

    plt.figure(figsize=(8, 5))
    plt.hist(case_durations, bins=30, color="salmon", edgecolor="black")
    plt.title("Distribution of Case Durations")
    plt.xlabel("Duration (hours)")
    plt.ylabel("Number of Cases")
    plt.tight_layout()
    duration_chart = os.path.join(output_dir, "case_duration_distribution.png")
    plt.savefig(duration_chart)
    plt.close()

    plt.figure(figsize=(6, 6))
    plt.boxplot(case_durations, vert=True, patch_artist=True)
    plt.title("Case Duration Boxplot")
    plt.ylabel("Duration (hours)")
    plt.tight_layout()
    boxplot_path = os.path.join(output_dir, "case_duration_boxplot.png")
    plt.savefig(boxplot_path)
    plt.close()

    if case_durations:
        sorted_durations = [item[1] for item in sorted(case_duration_by_start, key=lambda x: x[0])]
        mean = np.mean(sorted_durations)
        std = np.std(sorted_durations)
        ucl = mean + 3 * std
        lcl = max(0.0, mean - 3 * std)
        plt.figure(figsize=(10, 5))
        plt.plot(sorted_durations, marker="o", linestyle="-", color="steelblue")
        plt.axhline(mean, color="green", linestyle="--", label="Mean")
        plt.axhline(ucl, color="red", linestyle="--", label="UCL")
        plt.axhline(lcl, color="red", linestyle="--", label="LCL")
        plt.title("Case Duration SPC Chart")
        plt.xlabel("Case Index")
        plt.ylabel("Duration (hours)")
        plt.legend()
        plt.tight_layout()
        spc_path = os.path.join(output_dir, "case_duration_spc.png")
        plt.savefig(spc_path)
        plt.close()
    else:
        spc_path = os.path.join(output_dir, "case_duration_spc.png")

    avg_sojourn = {act: float(np.mean(times)) for act, times in sojourn_times.items()}
    df_sojourn = pd.DataFrame(list(avg_sojourn.items()), columns=["activity", "avg_sojourn_hours"])
    df_sojourn.to_csv(os.path.join(output_dir, "sojourn_times.csv"), index=False)

    plt.figure(figsize=(10, 6))
    df_sojourn.sort_values("avg_sojourn_hours", ascending=False).plot.bar(
        x="activity", y="avg_sojourn_hours", color="olive", legend=False
    )
    plt.title("Average Sojourn Time per Activity")
    plt.xlabel("Activity")
    plt.ylabel("Sojourn Time (hours)")
    plt.xticks(rotation=45, ha="right")
    plt.tight_layout()
    sojourn_chart = os.path.join(output_dir, "sojourn_time_chart.png")
    plt.savefig(sojourn_chart)
    plt.close()

    metrics = compute_case_duration_stats(event_log)
    skew_flag = None
    if metrics.get("p95_hours") and metrics.get("median_hours"):
        ratio = metrics["p95_hours"] / max(metrics["median_hours"], 0.0001)
        if ratio > 5:
            skew_flag = ratio
    recommendations = []
    if skew_flag:
        recommendations.append("Heavy tail detected in case durations; consider trimming, winsorizing, or sampling.")
    summary = {
        "duration_stats": metrics,
        "p95_to_median_ratio": skew_flag,
        "recommendations": recommendations,
    }
    return {
        "case_duration_distribution": duration_chart,
        "case_duration_boxplot": boxplot_path,
        "case_duration_spc": spc_path,
        "sojourn_time_chart": sojourn_chart,
    }, summary


def organisational_analysis(event_log: object, output_dir: str) -> str:
    resource_key = None
    candidate_keys = ["org:resource", "agent_name", "adjuster_name", "user", "user_type", "resource"]
    for trace in event_log:
        for event in trace:
            for key in candidate_keys:
                if key in event and event.get(key):
                    resource_key = key
                    break
            if resource_key:
                break
        if resource_key:
            break

    handover_counts = {}
    for trace in event_log:
        prev_res = None
        for event in trace:
            res = event.get(resource_key) if resource_key else None
            if prev_res and res and res != prev_res:
                key = (prev_res, res)
                handover_counts[key] = handover_counts.get(key, 0) + 1
            prev_res = res
    df_handovers = pd.DataFrame([
        {"from": k[0], "to": k[1], "count": v}
        for k, v in handover_counts.items()
    ])
    output_path = os.path.join(output_dir, "handover_of_work.csv")
    df_handovers.to_csv(output_path, index=False)
    return output_path


def compute_start_end(event_log: object) -> Dict[str, Dict[str, int]]:
    if start_activities_get is None or end_activities_get is None:
        raise RuntimeError("PM4Py start/end activity helpers are unavailable in this environment.")
    start_fn = (
        start_activities_get.get_start_activities
        if hasattr(start_activities_get, "get_start_activities")
        else start_activities_get
    )
    end_fn = (
        end_activities_get.get_end_activities
        if hasattr(end_activities_get, "get_end_activities")
        else end_activities_get
    )
    return {
        "start_activities": start_fn(event_log),
        "end_activities": end_fn(event_log),
    }
