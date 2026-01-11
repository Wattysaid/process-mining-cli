# Bottleneck view: top sojourn times
import pandas as pd

sojourn = pd.read_csv(f"{OUTPUT_ROOT}/stage_07_performance/sojourn_times.csv")
sojourn.sort_values("avg_sojourn_hours", ascending=False).head(10)
