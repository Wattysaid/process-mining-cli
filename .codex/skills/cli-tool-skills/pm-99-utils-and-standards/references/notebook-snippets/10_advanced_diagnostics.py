import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import networkx as nx

if not RUN_ADVANCED_DIAGNOSTICS:
    print("Advanced diagnostics are disabled. Set RUN_ADVANCED_DIAGNOSTICS = True to enable.")
else:
    df = pd.read_csv(f"{OUTPUT_ROOT}/stage_03_clean_filter/filtered_log.csv")
    df["time:timestamp"] = pd.to_datetime(df["time:timestamp"], errors="coerce")

    # Variant coverage and entropy
    variants = (
        df.groupby(["case:concept:name"])['concept:name']
        .apply(lambda x: ' -> '.join(x.astype(str)))
        .value_counts()
        .reset_index()
        .rename(columns={"concept:name": "count", "index": "variant"})
    )
    variants["percent"] = variants["count"] / variants["count"].sum() * 100
    variants["cum_percent"] = variants["percent"].cumsum()
    entropy = -(variants["count"] / variants["count"].sum()).apply(lambda p: p * np.log2(p)).sum()
    print("Variant entropy:", entropy)
    display(variants.head(10))

    # Case length distribution
    case_lengths = df.groupby("case:concept:name")["concept:name"].size()
    print("Case length summary:")
    print(case_lengths.describe())
    case_lengths.hist(bins=30)
    plt.title("Case Length Distribution")
    plt.xlabel("Events per Case")
    plt.ylabel("Count")
    plt.tight_layout()
    plt.show()

    # Waiting time per activity
    df_sorted = df.sort_values(["case:concept:name", "time:timestamp"])
    df_sorted["next_time"] = df_sorted.groupby("case:concept:name")["time:timestamp"].shift(-1)
    df_sorted["wait_hours"] = (df_sorted["next_time"] - df_sorted["time:timestamp"]).dt.total_seconds() / 3600.0
    wait_stats = df_sorted.groupby("concept:name")["wait_hours"].agg(["mean", "median", "count"]).sort_values("mean", ascending=False)
    display(wait_stats.head(10))

    # SLA breach analysis
    case_start = df_sorted.groupby("case:concept:name")["time:timestamp"].min()
    case_end = df_sorted.groupby("case:concept:name")["time:timestamp"].max()
    case_duration = (case_end - case_start).dt.total_seconds() / 3600.0
    breaches = (case_duration > SLA_HOURS).mean()
    print(f"SLA breach rate (>{SLA_HOURS}h): {breaches:.2%}")

    # Segmentation analysis
    if SEGMENT_COL in df.columns:
        segment_case_duration = (
            df_sorted.groupby(["case:concept:name", SEGMENT_COL])["time:timestamp"]
            .agg(["min", "max"])
            .reset_index()
        )
        segment_case_duration["duration_hours"] = (
            segment_case_duration["max"] - segment_case_duration["min"]
        ).dt.total_seconds() / 3600.0
        segment_summary = segment_case_duration.groupby(SEGMENT_COL)["duration_hours"].describe()
        display(segment_summary)
    else:
        print(f"SEGMENT_COL '{SEGMENT_COL}' not found in data.")

    # Handover network
    if RESOURCE_COL in df.columns:
        df_sorted["prev_resource"] = df_sorted.groupby("case:concept:name")[RESOURCE_COL].shift(1)
        handovers = df_sorted.dropna(subset=["prev_resource", RESOURCE_COL])
        edges = handovers[handovers["prev_resource"] != handovers[RESOURCE_COL]]
        edge_counts = edges.groupby(["prev_resource", RESOURCE_COL]).size().reset_index(name="count")
        G = nx.from_pandas_edgelist(edge_counts, "prev_resource", RESOURCE_COL, ["count"], create_using=nx.DiGraph)
        centrality = nx.betweenness_centrality(G)
        top_nodes = sorted(centrality.items(), key=lambda x: x[1], reverse=True)[:10]
        print("Top handover brokers:")
        print(top_nodes)
    else:
        print(f"RESOURCE_COL '{RESOURCE_COL}' not found in data.")
