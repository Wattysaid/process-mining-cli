# Plot throughput over time
import pandas as pd
import matplotlib.pyplot as plt

df = pd.read_csv(f"{OUTPUT_ROOT}/stage_03_clean_filter/filtered_log.csv")
df["time:timestamp"] = pd.to_datetime(df["time:timestamp"], errors="coerce")
throughput = df.groupby(df["time:timestamp"].dt.date)["concept:name"].count()
throughput.plot(figsize=(10, 4), title="Events Per Day")
plt.tight_layout()
plt.show()
