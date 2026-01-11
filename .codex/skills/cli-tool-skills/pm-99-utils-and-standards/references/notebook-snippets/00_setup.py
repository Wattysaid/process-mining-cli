import importlib
import os
import sys
import subprocess

os.environ.setdefault("MPLCONFIGDIR", os.path.join(OUTPUT_DIR, ".mplconfig"))
os.environ.setdefault("GRAPHVIZ_DOT", "/usr/bin/dot")

required = ["pm4py", "pandas", "numpy", "matplotlib", "pyyaml", "tabulate", "networkx"]
missing = []
for pkg in required:
    try:
        importlib.import_module(pkg)
    except Exception:
        missing.append(pkg)

if missing:
    print("Missing packages:", missing)
    # Uncomment to install into the current Python environment.
    # subprocess.check_call([sys.executable, "-m", "pip", "install", *missing])
else:
    print("All required packages are available.")
