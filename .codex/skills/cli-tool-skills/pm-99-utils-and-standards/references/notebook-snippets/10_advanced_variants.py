import pandas as pd
import numpy as np

if RUN_ADVANCED_DIAGNOSTICS and RUN_VARIANT_DIAGNOSTICS:
    df = pd.read_csv(f"{OUTPUT_ROOT}/stage_03_clean_filter/filtered_log.csv")
    variants = (
        df.groupby(["case:concept:name"])['concept:name']
        .apply(lambda x: ' -> '.join(x.astype(str)))
        .value_counts()
        .reset_index()
    )
    variants.columns = ["variant", "count"]
    variants["percent"] = variants["count"] / variants["count"].sum() * 100
    variants["cum_percent"] = variants["percent"].cumsum()
    probs = variants["count"] / variants["count"].sum()
    entropy = float(-(probs * np.log2(probs)).sum())
    print("Variant entropy:", entropy)
    display(variants.head(15))
else:
    print("Variant diagnostics disabled.")
