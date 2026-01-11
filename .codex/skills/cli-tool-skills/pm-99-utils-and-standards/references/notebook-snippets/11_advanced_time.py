import pandas as pd
import matplotlib.pyplot as plt

if RUN_ADVANCED_DIAGNOSTICS and RUN_TIME_DIAGNOSTICS:
    df = pd.read_csv(f"{OUTPUT_ROOT}/stage_03_clean_filter/filtered_log.csv")
    df["time:timestamp"] = pd.to_datetime(df["time:timestamp"], errors="coerce")

    # Case length distribution
    case_lengths = df.groupby("case:concept:name")["concept:name"].size()
    print(case_lengths.describe())
    case_lengths.hist(bins=30)
    plt.title("Case Length Distribution")
    plt.xlabel("Events per Case")
    plt.ylabel("Count")
    plt.tight_layout()
    plt.show()

    # Weekday and hour patterns
    df["weekday"] = df["time:timestamp"].dt.day_name()
    df["hour"] = df["time:timestamp"].dt.hour
    weekday_counts = df["weekday"].value_counts()
    hour_counts = df["hour"].value_counts().sort_index()
    weekday_counts.plot(kind="bar", title="Events by Weekday")
    plt.tight_layout()
    plt.show()
    hour_counts.plot(kind="bar", title="Events by Hour")
    plt.tight_layout()
    plt.show()
else:
    print("Time diagnostics disabled.")
