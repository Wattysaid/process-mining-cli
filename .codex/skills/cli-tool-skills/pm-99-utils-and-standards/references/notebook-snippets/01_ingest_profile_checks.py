# Quick schema validation
import pandas as pd
import numpy as np

df = pd.read_csv(f"{OUTPUT_ROOT}/stage_01_ingest_profile/normalised_log.csv")
print("Columns:", df.columns.tolist())
print("Missing %:")
print(df.isna().mean().sort_values(ascending=False).head(10))
