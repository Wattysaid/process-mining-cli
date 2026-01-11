#!/usr/bin/env python3
"""Validate environment dependencies for process mining."""

import json
import logging
import os
import subprocess
import sys
import venv
from typing import Dict, List, Optional, Tuple

COMMON_DIR = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "..", "..", "pm-99-utils-and-standards", "scripts")
)
if COMMON_DIR not in sys.path:
    sys.path.insert(0, COMMON_DIR)

from common import ensure_notebook, ensure_stage_dir, save_json, setup_logging, write_stage_manifest

REQUIREMENTS_DEFAULT = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "..", "..", "pm-99-utils-and-standards", "requirements.txt")
)


def python_path(venv_dir: str) -> str:
    if os.name == "nt":
        return os.path.join(venv_dir, "Scripts", "python.exe")
    return os.path.join(venv_dir, "bin", "python")


def load_detect_env(path: Optional[str]) -> Dict[str, object]:
    if not path:
        return {}
    if not os.path.isfile(path):
        return {}
    with open(path, "r", encoding="utf-8") as handle:
        return json.load(handle)


def resolve_detected_python(detect_env: Dict[str, object]) -> Optional[str]:
    python_info = detect_env.get("python")
    if isinstance(python_info, dict):
        candidate = python_info.get("executable")
        if isinstance(candidate, str) and candidate and os.path.isfile(candidate):
            return candidate
    return None


def ensure_virtualenv(
    venv_dir: str,
    requirements_path: str,
    upgrade_pip: bool,
    python_exec: Optional[str],
) -> None:
    if not os.path.isfile(requirements_path):
        raise FileNotFoundError(f"Requirements file not found: {requirements_path}")
    if not os.path.isdir(venv_dir) or not os.path.isfile(python_path(venv_dir)):
        if python_exec:
            subprocess.check_call([python_exec, "-m", "venv", venv_dir])
        else:
            builder = venv.EnvBuilder(with_pip=True)
            builder.create(venv_dir)
    pip_exec = [python_path(venv_dir), "-m", "pip"]
    if upgrade_pip:
        subprocess.check_call(pip_exec + ["install", "--upgrade", "pip"])
    subprocess.check_call(pip_exec + ["install", "-r", requirements_path])


def validate_venv_exists(venv_dir: str) -> None:
    if not os.path.isdir(venv_dir) or not os.path.isfile(python_path(venv_dir)):
        raise FileNotFoundError(f"Virtualenv not found or incomplete at {venv_dir}.")


def check_packages_with_python(python_exec: str, packages: List[str]) -> Tuple[List[str], Dict[str, str]]:
    payload = json.dumps(packages)
    script = (
        "import importlib, json\n"
        f"packages = json.loads({payload!r})\n"
        "versions = {}\n"
        "missing = []\n"
        "for pkg in packages:\n"
        "    try:\n"
        "        module = importlib.import_module(pkg)\n"
        "        versions[pkg] = getattr(module, '__version__', 'unknown')\n"
        "    except Exception:\n"
        "        missing.append(pkg)\n"
        "print(json.dumps({'missing': missing, 'versions': versions}))\n"
    )
    output = subprocess.check_output([python_exec, "-c", script], text=True)
    result = json.loads(output.strip())
    return result.get("missing", []), result.get("versions", {})


def read_previous_failures(stage_dir: str) -> int:
    path = os.path.join(stage_dir, "validate_env.json")
    if not os.path.isfile(path):
        return 0
    try:
        with open(path, "r", encoding="utf-8") as handle:
            payload = json.load(handle)
        if payload.get("status") == "missing":
            return int(payload.get("failure_count", 0))
    except (ValueError, TypeError):
        return 0
    return 0


def resolve_platform_context(detect_env: Dict[str, object]) -> Dict[str, str]:
    os_name = os.name
    shell_family = ""
    is_wsl = False
    detect_os = detect_env.get("os")
    if isinstance(detect_os, dict):
        system = detect_os.get("system")
        if isinstance(system, str) and system:
            os_name = system
    shell = detect_env.get("shell")
    if isinstance(shell, dict):
        shell_family = shell.get("shell_family", "") or ""
    is_wsl = bool(detect_env.get("is_wsl"))
    return {
        "os_name": str(os_name),
        "shell_family": str(shell_family),
        "is_wsl": "true" if is_wsl else "false",
    }


def activation_command(venv_dir: str, os_name: str, shell_family: str) -> str:
    os_label = os_name.lower()
    shell = shell_family.lower()
    if os_label.startswith("win"):
        if shell in {"powershell", "pwsh"}:
            return rf".\{venv_dir}\Scripts\Activate.ps1"
        return rf"{venv_dir}\Scripts\activate.bat"
    return f"source {venv_dir}/bin/activate"


def deactivation_command() -> str:
    return "deactivate"


def manual_setup_instructions(
    venv_dir: str,
    requirements_path: str,
    output_dir: str,
    detect_env_json: str,
    os_name: str,
    shell_family: str,
) -> List[str]:
    os_label = os_name.lower()
    shell = shell_family.lower()
    lines = [
        "Manual virtualenv setup instructions:",
        "1) Create the virtual environment:",
    ]
    if os_label.startswith("win"):
        lines.append(f"   - python -m venv {venv_dir}")
    else:
        lines.append(f"   - python3 -m venv {venv_dir}")
    lines.append("2) Activate the virtual environment:")
    lines.append(f"   - {activation_command(venv_dir, os_name, shell_family)}")
    if os_label.startswith("win") and shell in {"powershell", "pwsh"}:
        lines.append("   - If activation fails, run: Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass")
    lines.append("3) Install requirements:")
    lines.append(f"   - python -m pip install -r {requirements_path}")
    lines.append("4) Re-run validation (do not use --setup-venv):")
    lines.append("   - python .codex/skills/pm-01-env/scripts/00_validate_env.py "
                 f"--output {output_dir} --venv-dir {venv_dir} "
                 f"--requirements {requirements_path} --detect-env-json {detect_env_json}")
    lines.append("5) When outputs are confirmed complete, deactivate the virtualenv:")
    lines.append(f"   - {deactivation_command()}")
    return lines


def maybe_print_manual_instructions(
    failure_count: int,
    venv_dir: str,
    requirements_path: str,
    output_dir: str,
    detect_env_json: str,
    detect_env: Dict[str, object],
) -> Optional[str]:
    if failure_count < 2:
        return None
    context = resolve_platform_context(detect_env)
    instructions = manual_setup_instructions(
        venv_dir,
        requirements_path,
        output_dir,
        detect_env_json,
        context["os_name"],
        context["shell_family"],
    )
    text = "\n".join(instructions)
    print(text)
    return text


def main() -> None:
    setup_logging(0)
    import argparse
    parser = argparse.ArgumentParser(description="Validate environment dependencies.")
    parser.add_argument("--output", default="output", help="Output root directory.")
    parser.add_argument("--notebook-revision", default="R1.00", help="Notebook revision label.")
    parser.add_argument("--setup-venv", action="store_true", help="Create a virtual environment and install requirements.")
    parser.add_argument("--venv-dir", default=".venv", help="Virtual environment directory.")
    parser.add_argument("--requirements", default=REQUIREMENTS_DEFAULT, help="Path to requirements.txt.")
    parser.add_argument(
        "--detect-env-json",
        default=os.path.join("output", "stage_00_detect_env", "detect_env.json"),
        help="Path to detect_env.json produced by 00_detect_env.py.",
    )
    parser.add_argument("--upgrade-pip", action="store_true", help="Upgrade pip before installing requirements.")
    args = parser.parse_args()
    stage_dir = ensure_stage_dir(args.output, "stage_00_validate_env")
    packages = ["pm4py", "pandas", "numpy", "matplotlib"]
    detect_env = load_detect_env(args.detect_env_json)
    detected_python = resolve_detected_python(detect_env)
    if detected_python and detected_python != sys.executable:
        logging.info("Using detected Python for venv: %s", detected_python)
    missing: List[str] = []
    versions: Dict[str, str] = {}
    venv_label = f"{args.venv_dir} ({'created' if args.setup_venv else 'existing'})"
    base_failure_count = read_previous_failures(stage_dir)
    failure_count = base_failure_count
    manual_instructions = None
    if args.setup_venv:
        try:
            ensure_virtualenv(args.venv_dir, args.requirements, args.upgrade_pip, detected_python)
            missing, versions = check_packages_with_python(python_path(args.venv_dir), packages)
        except (OSError, subprocess.CalledProcessError) as exc:
            missing = packages
            versions = {}
            failure_count += 1
            logging.error("Virtualenv setup failed: %s", exc)
            manual_instructions = maybe_print_manual_instructions(
                failure_count,
                args.venv_dir,
                args.requirements,
                args.output,
                args.detect_env_json,
                detect_env,
            )
    else:
        try:
            validate_venv_exists(args.venv_dir)
        except FileNotFoundError as exc:
            missing = packages
            versions = {}
            logging.error("%s", exc)
        else:
            missing, versions = check_packages_with_python(python_path(args.venv_dir), packages)
    if missing:
        if failure_count == base_failure_count:
            failure_count += 1
        if manual_instructions is None:
            manual_instructions = maybe_print_manual_instructions(
                failure_count,
                args.venv_dir,
                args.requirements,
                args.output,
                args.detect_env_json,
                detect_env,
            )
        message = f"Missing packages: {', '.join(missing)}"
        logging.error(message)
        with open(f"{stage_dir}/validate_env.log", "w", encoding="utf-8") as handle:
            handle.write(message + "\n")
            if manual_instructions:
                handle.write("\n" + manual_instructions + "\n")
        save_json(
            {
                "status": "missing",
                "missing": missing,
                "failure_count": failure_count,
                "manual_instructions": manual_instructions or "",
            },
            f"{stage_dir}/validate_env.json",
        )
        notebook_path = ensure_notebook(
            args.output,
            args.notebook_revision,
            "00_validate_env.ipynb",
            "Environment Validation",
            context_lines=[
                "",
                "This notebook captures the environment validation step.",
                f"- Status: missing ({', '.join(missing)})",
                f"- Virtualenv: {venv_label}",
                f"- Requirements: {args.requirements}",
                f"- Detect env: {args.detect_env_json}",
                f"- Detected python: {detected_python or 'not found'}",
            ],
            code_lines=["# Review missing packages above."],
        )
        write_stage_manifest(
            stage_dir,
            {
                "output": args.output,
                "notebook_revision": args.notebook_revision,
                "setup_venv": args.setup_venv,
                "venv_dir": args.venv_dir,
                "requirements": args.requirements,
                "detect_env_json": args.detect_env_json,
                "detected_python": detected_python or "",
                "failure_count": failure_count,
            },
            {"validate_env_json": f"{stage_dir}/validate_env.json", "validate_env_log": f"{stage_dir}/validate_env.log"},
            args.notebook_revision,
            notebook_path=notebook_path,
            notes="Dependency validation failed.",
        )
        sys.exit(1)
    for package, version in versions.items():
        logging.info("%s: %s", package, version)
    with open(f"{stage_dir}/validate_env.log", "w", encoding="utf-8") as handle:
        handle.write("Environment looks good.\n")
        for package, version in versions.items():
            handle.write(f"{package}: {version}\n")
    save_json({"status": "ok", "versions": versions}, f"{stage_dir}/validate_env.json")
    notebook_path = ensure_notebook(
        args.output,
        args.notebook_revision,
        "00_validate_env.ipynb",
        "Environment Validation",
        context_lines=[
            "",
            "This notebook captures the environment validation step.",
            f"- Virtualenv: {venv_label}",
            f"- Requirements: {args.requirements}",
            f"- Detect env: {args.detect_env_json}",
            f"- Detected python: {detected_python or 'not found'}",
            "Versions:",
        ] + [f"- {pkg}: {ver}" for pkg, ver in versions.items()],
        code_lines=["# Environment looks good."],
    )
    write_stage_manifest(
        stage_dir,
        {
            "output": args.output,
            "notebook_revision": args.notebook_revision,
            "setup_venv": args.setup_venv,
            "venv_dir": args.venv_dir,
            "requirements": args.requirements,
            "detect_env_json": args.detect_env_json,
            "detected_python": detected_python or "",
        },
        {"validate_env_json": f"{stage_dir}/validate_env.json", "validate_env_log": f"{stage_dir}/validate_env.log"},
        args.notebook_revision,
        notebook_path=notebook_path,
    )
    print("Environment looks good.")


if __name__ == "__main__":
    main()
