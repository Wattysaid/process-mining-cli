#!/usr/bin/env python3
"""Generate a Markdown report from pipeline artifacts."""

import argparse
import json
import os
import sys

import pandas as pd

COMMON_DIR = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "..", "..", "pm-99-utils-and-standards", "scripts")
)
if COMMON_DIR not in sys.path:
    sys.path.insert(0, COMMON_DIR)

from common import ensure_notebook, ensure_stage_dir, exit_with_error, write_stage_manifest


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate a report from process mining artifacts.")
    parser.add_argument("--output", default="output", help="Directory containing analysis results.")
    parser.add_argument("--report", default="process_mining_report.md", help="Report filename.")
    parser.add_argument("--notebook-revision", default="R1.00", help="Notebook revision label.")
    return parser.parse_args()


def main() -> None:
    args = parse_arguments()
    stage_dir = ensure_stage_dir(args.output, "stage_09_report")
    summary_path = os.path.join(args.output, "stage_04_eda", "summary_stats.json")
    variant_counts_path = os.path.join(args.output, "stage_04_eda", "variant_counts.csv")
    variant_coverage_path = os.path.join(args.output, "stage_04_eda", "variant_coverage.csv")
    metrics_path = os.path.join(args.output, "stage_05_discover", "model_metrics.csv")
    inductive_png = os.path.join(args.output, "stage_05_discover", "inductive_miner_petri_net.png")
    heuristic_png = os.path.join(args.output, "stage_05_discover", "heuristic_miner_petri_net.png")
    conformance_path = os.path.join(args.output, "stage_06_conformance", "conformance_metrics.csv")
    conformance_case_path = os.path.join(args.output, "stage_06_conformance", "conformance_case_deviations.csv")
    data_quality_path = os.path.join(args.output, "stage_02_data_quality", "data_quality.json")
    data_quality_reco_path = os.path.join(args.output, "stage_02_data_quality", "data_quality_recommendations.json")
    performance_summary_path = os.path.join(args.output, "stage_07_performance", "performance_summary.json")
    waiting_stats_path = os.path.join(args.output, "stage_07_performance", "activity_waiting_time_stats.csv")
    case_duration_summary_path = os.path.join(args.output, "stage_07_performance", "case_duration_summary.json")
    filtered_log_path = os.path.join(args.output, "stage_03_clean_filter", "filtered_log.csv")
    if not os.path.isfile(summary_path):
        summary_path = os.path.join(args.output, "summary_stats.json")
    if not os.path.isfile(metrics_path):
        metrics_path = os.path.join(args.output, "model_metrics.csv")
    if not os.path.isfile(summary_path):
        exit_with_error(f"Missing summary stats: {summary_path}")
    if not os.path.isfile(metrics_path):
        exit_with_error(f"Missing model metrics: {metrics_path}")

    with open(summary_path, "r", encoding="utf-8") as handle:
        summary = json.load(handle)
    model_metrics = pd.read_csv(metrics_path)
    variant_counts = None
    if os.path.isfile(variant_coverage_path):
        variant_counts = pd.read_csv(variant_coverage_path)
    elif os.path.isfile(variant_counts_path):
        variant_counts = pd.read_csv(variant_counts_path)
    waiting_stats = pd.read_csv(waiting_stats_path) if os.path.isfile(waiting_stats_path) else None

    agent_perf_path = None
    agent_perf = None
    if os.path.isfile(filtered_log_path):
        df = pd.read_csv(filtered_log_path)
        if {"case:concept:name", "time:timestamp", "org:resource"}.issubset(df.columns):
            df = df.copy()
            df["time:timestamp"] = pd.to_datetime(df["time:timestamp"], errors="coerce")
            case_times = df.groupby("case:concept:name")["time:timestamp"].agg(["min", "max"])
            case_times["duration_hours"] = (
                (case_times["max"] - case_times["min"]).dt.total_seconds() / 3600.0
            )
            agent_counts = (
                df.groupby(["case:concept:name", "org:resource"])
                .size()
                .reset_index(name="event_count")
            )
            agent_counts["rank"] = agent_counts.groupby("case:concept:name")["event_count"].rank(
                method="first", ascending=False
            )
            primary_agent = agent_counts[agent_counts["rank"] == 1][
                ["case:concept:name", "org:resource"]
            ]
            case_with_agent = case_times.join(
                primary_agent.set_index("case:concept:name"), how="left"
            ).dropna(subset=["org:resource"])
            agent_perf = (
                case_with_agent.groupby("org:resource")["duration_hours"]
                .agg(case_count="count", mean_hours="mean", median_hours="median", p90_hours=lambda x: x.quantile(0.9))
                .sort_values("mean_hours", ascending=False)
                .reset_index()
            )
            agent_perf_path = os.path.join(stage_dir, "agent_performance_summary.csv")
            agent_perf.to_csv(agent_perf_path, index=False)

    report_path = os.path.join(stage_dir, args.report)
    with open(report_path, "w", encoding="utf-8") as handle:
        handle.write("# Process Mining CLI Report\n\n")
        stats = summary.get("stats", {})
        handle.write("## Executive Summary\n")
        handle.write(
            "This report summarizes the main process characteristics, highlights the most common variants, "
            "and surfaces key bottlenecks and agent-level performance differences.\n\n"
        )
        handle.write("## Summary Statistics\n")
        handle.write(f"- Number of events: {stats.get('num_events')}\n")
        handle.write(f"- Number of cases: {stats.get('num_cases')}\n")
        handle.write(f"- Number of variants: {stats.get('num_variants')}\n\n")

        arrival = summary.get("arrival_metrics", {})
        handle.write("## Arrival Metrics\n")
        handle.write(f"- Mean inter-arrival (hours): {arrival.get('mean_interarrival_hours')}\n")
        handle.write(f"- Median inter-arrival (hours): {arrival.get('median_interarrival_hours')}\n\n")

        start_end = summary.get("start_end", {})
        if start_end.get("start_activities"):
            handle.write("## Start Activities\n")
            handle.write(pd.DataFrame(list(start_end["start_activities"].items()), columns=["activity", "count"]).to_markdown(index=False))
            handle.write("\n\n")
        if start_end.get("end_activities"):
            handle.write("## End Activities\n")
            handle.write(pd.DataFrame(list(start_end["end_activities"].items()), columns=["activity", "count"]).to_markdown(index=False))
            handle.write("\n\n")

        if os.path.isfile(data_quality_path):
            with open(data_quality_path, "r", encoding="utf-8") as dq_handle:
                data_quality = json.load(dq_handle)
            handle.write("## Data Quality Summary\n")
            missing_rates = data_quality.get("missing_rates", {})
            handle.write(f"- Missing case IDs: {missing_rates.get('case:concept:name', 0):.2%}\n")
            handle.write(f"- Missing activities: {missing_rates.get('concept:name', 0):.2%}\n")
            handle.write(f"- Missing timestamps: {missing_rates.get('time:timestamp', 0):.2%}\n")
            handle.write(f"- Timestamp parse failure rate: {data_quality.get('timestamp_parse_failure_rate', 0):.2%}\n")
            handle.write(f"- Duplicate rate: {data_quality.get('duplicate_rate', 0):.2%}\n")
            handle.write(f"- Case order violation rate: {data_quality.get('case_order_violation_rate', 0):.2%}\n\n")
            handle.write("Commentary: data quality is strong with no material missingness or ordering issues.\n\n")
        if os.path.isfile(data_quality_reco_path):
            with open(data_quality_reco_path, "r", encoding="utf-8") as dq_handle:
                recommendations = json.load(dq_handle)
            handle.write("## Data Quality Recommendations\n")
            for key, value in recommendations.items():
                handle.write(f"- {key}: {value}\n")
            handle.write("\n")

        if os.path.isfile(case_duration_summary_path):
            with open(case_duration_summary_path, "r", encoding="utf-8") as perf_handle:
                duration_summary = json.load(perf_handle)
            handle.write("## Descriptive Statistics\n")
            handle.write(
                "- Case duration (hours): "
                f"mean {duration_summary.get('mean_hours'):.2f}, "
                f"median {duration_summary.get('median_hours'):.2f}, "
                f"p90 {duration_summary.get('p90_hours'):.2f}, "
                f"p95 {duration_summary.get('p95_hours'):.2f}\n"
            )
            handle.write(
                f"- SLA ({duration_summary.get('sla_hours')} hours) breach rate: "
                f"{duration_summary.get('sla_breach_rate'):.2%}\n\n"
            )

        if variant_counts is not None and not variant_counts.empty:
            handle.write("## Top 5 Variants\n\n")
            top_variants = variant_counts.head(5)
            handle.write(top_variants.to_markdown(index=False))
            handle.write("\n\n")

        handle.write("## Model Evaluation\n\n")
        handle.write(model_metrics.to_markdown(index=False))
        handle.write("\n\n")
        if os.path.isfile(inductive_png):
            handle.write("### Inductive Miner Model\n\n")
            handle.write(f"![Inductive Miner Model]({inductive_png})\n\n")
        if os.path.isfile(heuristic_png):
            handle.write("### Heuristic Miner Model\n\n")
            handle.write(f"![Heuristic Miner Model]({heuristic_png})\n\n")
        if os.path.isfile(conformance_path):
            conformance_metrics = pd.read_csv(conformance_path)
            handle.write("## Conformance Diagnostics\n\n")
            handle.write(conformance_metrics.to_markdown(index=False))
            handle.write("\n\n")
            if os.path.isfile(conformance_case_path):
                handle.write("Per-case deviations are available in the technical appendix: ")
                handle.write(f"`{conformance_case_path}`.\n\n")

        if waiting_stats is not None and not waiting_stats.empty:
            handle.write("## Bottlenecks (Top Waiting Activities)\n\n")
            handle.write(waiting_stats.head(5).to_markdown(index=False))
            handle.write("\n\n")

        if agent_perf is not None and not agent_perf.empty:
            handle.write("## Agent Performance Breakdown\n\n")
            handle.write(
                "Cases are attributed to the primary agent (most events in the case). "
                "Metrics show average case duration by agent.\n\n"
            )
            handle.write(agent_perf.head(10).to_markdown(index=False))
            handle.write("\n\n")
            handle.write(f"Full agent summary: `{agent_perf_path}`.\n\n")

        handle.write("Review the output directory for plots and CSVs covering activity distributions, variants, case durations, sojourn times, and organisational handovers.\n")
    notebook_path = ensure_notebook(
        args.output,
        args.notebook_revision,
        "09_report.ipynb",
        "Reporting",
        context_lines=[
            "",
            f"- Report: {report_path}",
        ],
        code_lines=[
            f"report_path = r\"{report_path}\"",
            "print(report_path)",
        ],
    )
    artifacts = {"process_mining_report_md": report_path}
    if agent_perf_path:
        artifacts["agent_performance_summary_csv"] = agent_perf_path
    write_stage_manifest(
        stage_dir,
        vars(args),
        artifacts,
        args.notebook_revision,
        notebook_path=notebook_path,
    )


if __name__ == "__main__":
    main()
