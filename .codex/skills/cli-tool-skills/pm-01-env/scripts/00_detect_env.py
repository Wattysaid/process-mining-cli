#!/usr/bin/env python3
"""Detect OS, shell, and tooling to surface environment issues early."""

import argparse
import json
import logging
import os
import platform
import shutil
import sys
from typing import Dict, List

COMMON_DIR = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "..", "..", "pm-99-utils-and-standards", "scripts")
)
if COMMON_DIR not in sys.path:
    sys.path.insert(0, COMMON_DIR)

from common import ensure_notebook, ensure_stage_dir, save_json, setup_logging, write_stage_manifest


def detect_wsl() -> bool:
    if os.environ.get("WSL_INTEROP") or os.environ.get("WSL_DISTRO_NAME"):
        return True
    try:
        with open("/proc/version", "r", encoding="utf-8") as handle:
            return "microsoft" in handle.read().lower()
    except OSError:
        return False


def detect_shell() -> Dict[str, str]:
    env = os.environ
    shell = env.get("SHELL") or env.get("ComSpec") or ""
    terminal = env.get("TERM_PROGRAM") or env.get("TERM") or ""
    shell_family = ""
    if os.name == "nt":
        if env.get("PSModulePath") or env.get("POWERSHELL_DISTRIBUTION_CHANNEL"):
            shell_family = "powershell"
        elif env.get("ComSpec", "").lower().endswith("cmd.exe"):
            shell_family = "cmd"
    else:
        shell_family = os.path.basename(shell)
    return {
        "shell": shell,
        "shell_family": shell_family,
        "terminal": terminal,
        "windows_terminal": "WT_SESSION" in env,
        "term_program": env.get("TERM_PROGRAM", ""),
        "term": env.get("TERM", ""),
        "colorterm": env.get("COLORTERM", ""),
    }


def detect_tools(tool_names: List[str]) -> Dict[str, str]:
    tools = {}
    for name in tool_names:
        path = shutil.which(name) or ""
        tools[name] = path
    return tools


def build_warnings(info: Dict[str, object]) -> List[str]:
    warnings: List[str] = []
    tools = info.get("tools", {})
    python_path = tools.get("python") or tools.get("python3")
    pip_path = tools.get("pip") or tools.get("pip3")
    if not python_path:
        warnings.append("Python is not on PATH (python/python3 not found).")
    if not pip_path:
        warnings.append("pip is not on PATH (pip/pip3 not found).")
    if sys.version_info < (3, 9):
        warnings.append(f"Python version is {sys.version_info.major}.{sys.version_info.minor}; 3.9+ recommended.")
    if info.get("is_wsl"):
        warnings.append("WSL detected; ensure data paths use /mnt/... and Linux tooling is installed.")
    return warnings


def main() -> None:
    parser = argparse.ArgumentParser(description="Detect environment and tooling details.")
    parser.add_argument("--output", default="output", help="Output root directory.")
    parser.add_argument("--notebook-revision", default="R1.00", help="Notebook revision label.")
    parser.add_argument("--pretty", action="store_true", help="Print JSON to stdout.")
    args = parser.parse_args()

    setup_logging(0)
    stage_dir = ensure_stage_dir(args.output, "stage_00_detect_env")

    os_info = {
        "system": platform.system(),
        "release": platform.release(),
        "version": platform.version(),
        "machine": platform.machine(),
        "platform": platform.platform(),
    }
    python_info = {
        "executable": sys.executable,
        "version": platform.python_version(),
        "implementation": platform.python_implementation(),
    }
    shell_info = detect_shell()
    tools = detect_tools(
        [
            "python",
            "python3",
            "pip",
            "pip3",
            "conda",
            "mamba",
            "micromamba",
            "poetry",
            "pipx",
            "uv",
            "git",
        ]
    )
    info = {
        "os": os_info,
        "python": python_info,
        "shell": shell_info,
        "tools": tools,
        "is_wsl": detect_wsl(),
        "wsl_distro": os.environ.get("WSL_DISTRO_NAME", ""),
        "cwd": os.getcwd(),
    }
    warnings = build_warnings(info)
    info["warnings"] = warnings

    json_path = os.path.join(stage_dir, "detect_env.json")
    log_path = os.path.join(stage_dir, "detect_env.log")
    save_json(info, json_path)
    with open(log_path, "w", encoding="utf-8") as handle:
        if warnings:
            handle.write("Warnings:\n")
            for warning in warnings:
                handle.write(f"- {warning}\n")
        else:
            handle.write("No obvious environment issues detected.\n")

    notebook_context = [
        "",
        f"- OS: {os_info['system']} {os_info['release']}",
        f"- Python: {python_info['version']} ({python_info['executable']})",
        f"- Shell: {shell_info.get('shell_family') or shell_info.get('shell')}",
        f"- WSL: {info['is_wsl']}",
        "Warnings:",
    ]
    if warnings:
        notebook_context.extend([f"- {warning}" for warning in warnings])
    else:
        notebook_context.append("- None")
    notebook_path = ensure_notebook(
        args.output,
        args.notebook_revision,
        "00_detect_env.ipynb",
        "Environment Detection",
        context_lines=notebook_context,
        code_lines=["# Review detect_env.json for full details."],
    )

    write_stage_manifest(
        stage_dir,
        {"output": args.output, "notebook_revision": args.notebook_revision},
        {"detect_env_json": json_path, "detect_env_log": log_path},
        args.notebook_revision,
        notebook_path=notebook_path,
    )

    if args.pretty:
        print(json.dumps(info, indent=2))
    else:
        logging.info("Environment detection complete: %s", json_path)


if __name__ == "__main__":
    main()
