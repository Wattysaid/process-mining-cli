import pandas as pd

if RUN_ADVANCED_DIAGNOSTICS and RUN_RESOURCE_DIAGNOSTICS:
    df = pd.read_csv(f"{OUTPUT_ROOT}/stage_03_clean_filter/filtered_log.csv")
    if RESOURCE_COL in df.columns:
        events_per_resource = df[RESOURCE_COL].value_counts().head(20)
        print("Events per resource (top 20):")
        display(events_per_resource)
        cases_per_resource = df.groupby(RESOURCE_COL)["case:concept:name"].nunique().sort_values(ascending=False).head(20)
        print("Cases per resource (top 20):")
        display(cases_per_resource)
    else:
        print(f"RESOURCE_COL '{RESOURCE_COL}' not found in data.")
else:
    print("Resource diagnostics disabled.")
