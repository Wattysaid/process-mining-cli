import pandas as pd

if RUN_ADVANCED_DIAGNOSTICS and RUN_REWORK_DIAGNOSTICS:
    df = pd.read_csv(f"{OUTPUT_ROOT}/stage_03_clean_filter/filtered_log.csv")
    rework = (
        df.groupby(["case:concept:name", "concept:name"]).size().reset_index(name="count")
    )
    rework_rate = (rework["count"] > 1).mean()
    print(f"Rework rate (activity repeats): {rework_rate:.2%}")
    top_rework = rework.sort_values("count", ascending=False).head(15)
    display(top_rework)
else:
    print("Rework diagnostics disabled.")
