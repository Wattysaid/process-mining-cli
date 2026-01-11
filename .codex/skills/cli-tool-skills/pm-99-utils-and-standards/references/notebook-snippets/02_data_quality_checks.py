# Duplicate rate check
import pandas as pd

df = pd.read_csv(f"{OUTPUT_ROOT}/stage_02_data_quality/cleaned_log.csv")
print("Duplicate rate:", df.duplicated().mean())
