# Handover hotspots
import pandas as pd

handover = pd.read_csv(f"{OUTPUT_ROOT}/stage_08_org_mining/handover_of_work.csv")
handover.sort_values("count", ascending=False).head(10)
