import pandas as pd

if RUN_ADVANCED_DIAGNOSTICS and RUN_SEGMENT_DIAGNOSTICS:
    df = pd.read_csv(f"{OUTPUT_ROOT}/stage_03_clean_filter/filtered_log.csv")
    df["time:timestamp"] = pd.to_datetime(df["time:timestamp"], errors="coerce")
    if SEGMENT_COL in df.columns:
        case_start = df.groupby(["case:concept:name", SEGMENT_COL])["time:timestamp"].min().reset_index()
        case_end = df.groupby(["case:concept:name", SEGMENT_COL])["time:timestamp"].max().reset_index()
        durations = case_start.merge(case_end, on=["case:concept:name", SEGMENT_COL], suffixes=("_start", "_end"))
        durations["duration_hours"] = (durations["time:timestamp_end"] - durations["time:timestamp_start"]).dt.total_seconds() / 3600.0
        segment_summary = durations.groupby(SEGMENT_COL)["duration_hours"].describe()
        display(segment_summary)
        sla_rate = durations.groupby(SEGMENT_COL)["duration_hours"].apply(lambda x: (x > SLA_HOURS).mean())
        print("SLA breach rate by segment:")
        display(sla_rate)
    else:
        print(f"SEGMENT_COL '{SEGMENT_COL}' not found in data.")
else:
    print("Segment diagnostics disabled.")
