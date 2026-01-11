#!/usr/bin/env python3
"""Shared utilities for the process mining CLI workflow."""

import argparse
import hashlib
import json
import logging
import os
import sys
from datetime import datetime
from typing import Any, Dict, List, Optional, Tuple


class ExitCodes:
    SCHEMA_ERROR = 10
    TIMESTAMP_ERROR = 11
    MISSING_VALUES_ERROR = 12
    RUNTIME_ERROR = 20


def setup_logging(verbosity: int) -> None:
    """Configure logging output."""
    level = logging.WARNING
    if verbosity == 1:
        level = logging.INFO
    elif verbosity >= 2:
        level = logging.DEBUG
    logging.basicConfig(
        level=level,
        format="%(asctime)s %(levelname)s %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )


def load_config(config_path: Optional[str]) -> Dict[str, Any]:
    """Load configuration from JSON or YAML if available."""
    if not config_path:
        return {}
    if not os.path.exists(config_path):
        raise FileNotFoundError(f"Config file not found: {config_path}")
    ext = os.path.splitext(config_path)[1].lower()
    with open(config_path, "r", encoding="utf-8") as handle:
        if ext in (".yaml", ".yml"):
            try:
                import yaml  # type: ignore
            except ImportError as exc:
                raise RuntimeError("PyYAML is required for YAML configs. Install pyyaml.") from exc
            return yaml.safe_load(handle) or {}
        return json.load(handle)


def merge_config(args: argparse.Namespace, config: Dict[str, Any]) -> Dict[str, Any]:
    """Merge CLI args over config, returning a flat dictionary."""
    merged = dict(config)
    for key, value in vars(args).items():
        if value is not None:
            merged[key] = value
    return merged


def ensure_output_dir(path: str) -> None:
    os.makedirs(path, exist_ok=True)


def save_json(payload: Dict[str, Any], output_path: str) -> None:
    with open(output_path, "w", encoding="utf-8") as handle:
        json.dump(payload, handle, indent=2, sort_keys=True)


def load_json(path: str, default: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
    if not os.path.isfile(path):
        return default or {}
    with open(path, "r", encoding="utf-8") as handle:
        return json.load(handle)


def save_text(lines: List[str], output_path: str) -> None:
    with open(output_path, "w", encoding="utf-8") as handle:
        for line in lines:
            handle.write(line.rstrip("\n") + "\n")


def write_manifest(output_dir: str, params: Dict[str, Any], artifacts: Dict[str, str]) -> None:
    """Write a simple manifest describing run parameters and artifacts."""
    manifest = {
        "generated_at": datetime.utcnow().isoformat() + "Z",
        "parameters": params,
        "artifacts": artifacts,
    }
    save_json(manifest, os.path.join(output_dir, "manifest.json"))


def file_hash(path: str) -> str:
    hasher = hashlib.sha256()
    with open(path, "rb") as handle:
        for chunk in iter(lambda: handle.read(8192), b""):
            hasher.update(chunk)
    return hasher.hexdigest()


def ensure_stage_dir(output_root: str, stage_name: str) -> str:
    stage_dir = os.path.join(output_root, stage_name)
    os.makedirs(stage_dir, exist_ok=True)
    return stage_dir


def stage_state_path(stage_dir: str) -> str:
    return os.path.join(stage_dir, "stage_state.json")


def read_stage_state(stage_dir: str) -> Dict[str, Any]:
    return load_json(stage_state_path(stage_dir), default={})


def write_stage_state(stage_dir: str, payload: Dict[str, Any]) -> None:
    save_json(payload, stage_state_path(stage_dir))


def record_stage_failure(stage_dir: str, message: str, next_steps: List[str], attempt_limit: int = 2) -> int:
    state = read_stage_state(stage_dir)
    failures = int(state.get("failure_count", 0)) + 1
    state.update({
        "status": "failed",
        "failure_count": failures,
        "last_error": message,
        "next_steps": next_steps,
        "updated_at": datetime.utcnow().isoformat() + "Z",
    })
    write_stage_state(stage_dir, state)
    if failures >= attempt_limit:
        print("Repeated failure detected. Stop and review the next steps below:")
        for step in next_steps:
            print(f"- {step}")
    return failures


def infer_format_from_path(path: str) -> Optional[str]:
    ext = os.path.splitext(path)[1].lower()
    if ext == ".csv":
        return "csv"
    if ext == ".xes":
        return "xes"
    return None


def ensure_notebook(output_root: str, revision: str, notebook_name: str, title: str,
                    context_lines: Optional[list] = None, code_lines: Optional[list] = None) -> str:
    notebooks_dir = os.path.join(output_root, "notebooks", revision)
    os.makedirs(notebooks_dir, exist_ok=True)
    notebook_path = os.path.join(notebooks_dir, notebook_name)
    context_lines = context_lines or []
    code_lines = code_lines or []
    notebook = {
        "cells": [
            {
                "cell_type": "markdown",
                "metadata": {},
                "source": [line if line.endswith("\n") else line + "\n" for line in [f"# {title}"] + context_lines],
            },
            {
                "cell_type": "code",
                "execution_count": None,
                "metadata": {},
                "outputs": [],
                "source": [line if line.endswith("\n") else line + "\n" for line in code_lines],
            },
        ],
        "metadata": {
            "kernelspec": {"display_name": "Python 3", "language": "python", "name": "python3"},
            "language_info": {"name": "python"},
        },
        "nbformat": 4,
        "nbformat_minor": 5,
    }
    with open(notebook_path, "w", encoding="utf-8") as handle:
        json.dump(notebook, handle, indent=2)
    return notebook_path


def write_stage_manifest(stage_dir: str,
                         params: Dict[str, Any],
                         artifacts: Dict[str, str],
                         revision: str,
                         notebook_path: Optional[str] = None,
                         notes: Optional[str] = None) -> str:
    manifest = {
        "generated_at": datetime.utcnow().isoformat() + "Z",
        "revision": revision,
        "parameters": params,
        "artifacts": artifacts,
        "notes": notes or "",
    }
    if notebook_path:
        manifest["notebook"] = {
            "path": notebook_path,
            "sha256": file_hash(notebook_path),
        }
    manifest_path = os.path.join(stage_dir, "manifest.json")
    save_json(manifest, manifest_path)
    return manifest_path


def require_file(path: str) -> None:
    if not os.path.isfile(path):
        raise FileNotFoundError(f"Input file not found: {path}")


def parse_list(value: Optional[object]) -> Optional[list]:
    if value is None:
        return None
    if isinstance(value, list):
        return value
    if isinstance(value, str):
        items = [item.strip() for item in value.split(",")]
        return [item for item in items if item]
    return [value]


def exit_with_error(message: str, code: int = 1) -> None:
    logging.error(message)
    sys.exit(code)


def validate_csv_columns(df, required: Tuple[str, ...]) -> None:
    missing = [col for col in required if col not in df.columns]
    if missing:
        raise ValueError(f"Missing required columns: {', '.join(missing)}")
