#!/usr/bin/env python3
"""Export pipeline artifacts into a zip and optional manifest."""

import argparse
import os
import sys
import zipfile

COMMON_DIR = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "..", "..", "pm-99-utils-and-standards", "scripts")
)
if COMMON_DIR not in sys.path:
    sys.path.insert(0, COMMON_DIR)

from common import exit_with_error, save_json


def parse_arguments() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Export artifacts.")
    parser.add_argument("--output", default="output", help="Directory containing artifacts.")
    parser.add_argument("--zip", dest="zip_path", help="Optional zip output path.")
    return parser.parse_args()


def main() -> None:
    args = parse_arguments()
    if not os.path.isdir(args.output):
        exit_with_error(f"Output directory not found: {args.output}")

    files = []
    for root, _, filenames in os.walk(args.output):
        for name in filenames:
            path = os.path.join(root, name)
            files.append(os.path.relpath(path, args.output))

    save_json({"artifacts": sorted(files)}, os.path.join(args.output, "artifact_index.json"))

    if args.zip_path:
        with zipfile.ZipFile(args.zip_path, "w", zipfile.ZIP_DEFLATED) as zf:
            for rel_path in files:
                abs_path = os.path.join(args.output, rel_path)
                zf.write(abs_path, arcname=rel_path)


if __name__ == "__main__":
    main()
