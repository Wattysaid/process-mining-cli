# Variant concentration check
import pandas as pd

df = pd.read_csv(f"{OUTPUT_ROOT}/stage_03_clean_filter/filtered_log.csv")
variant_counts = (
    df.groupby(["case:concept:name"])['concept:name']
    .apply(lambda x: ' -> '.join(x.astype(str)))
    .value_counts()
)
print("Top variants (sample):")
print(variant_counts.head(5))
