# Inspect discovered model metrics
import pandas as pd

metrics = pd.read_csv(f"{OUTPUT_ROOT}/stage_05_discover/model_metrics.csv")
metrics
