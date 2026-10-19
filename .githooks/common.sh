#!/usr/bin/env bash
# common.sh — shared helpers for pkt-git hooks (Linux / macOS / Windows)

# Print the project venv python executable, cross-platform.
# Layout differs: POSIX uses .venv/bin/python, Windows uses .venv/Scripts/python.exe
resolve_py() {
    if [[ -f .venv/bin/python ]]; then
        echo ".venv/bin/python"
    elif [[ -f .venv/Scripts/python.exe ]]; then
        echo ".venv/Scripts/python.exe"
    else
        echo "python3"
    fi
}
