import runpy
import sys


def _usage() -> None:
    print("Usage:")
    print("  python -m pm_assist run-script <path> [args...]")
    print("  python -m pm_assist run-module <module> [args...]")


def main() -> int:
    if len(sys.argv) < 3:
        _usage()
        return 2
    command = sys.argv[1]
    target = sys.argv[2]
    sys.argv = [target] + sys.argv[3:]
    if command == "run-script":
        runpy.run_path(target, run_name="__main__")
        return 0
    if command == "run-module":
        runpy.run_module(target, run_name="__main__")
        return 0
    _usage()
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
