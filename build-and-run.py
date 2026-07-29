#!/usr/bin/env python3

from __future__ import annotations

import os
import shutil
import subprocess
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parent
WEB_DIR = ROOT / "web"
APP_NAME = "bbs-go.exe" if os.name == "nt" else "bbs-go"
APP_PATH = ROOT / APP_NAME


def fail(message: str) -> int:
    print(f"\nError: {message}", file=sys.stderr)
    return 1


def require_command(name: str, install_hint: str) -> str:
    executable = shutil.which(name)
    if executable is None:
        raise RuntimeError(f"{name} was not found. {install_hint}")
    return executable


def command_for(executable: str, *args: str) -> list[str]:
    return [executable, *args]


def run(command: list[str], cwd: Path) -> None:
    print(f"$ {' '.join(command)}", flush=True)
    use_shell = os.name == "nt" and Path(command[0]).suffix.lower() in {
        ".bat",
        ".cmd",
    }
    args: list[str] | str = subprocess.list2cmdline(command) if use_shell else command
    subprocess.run(args, cwd=cwd, check=True, shell=use_shell)


def process_names() -> set[str]:
    if os.name == "nt":
        result = subprocess.run(
            ["tasklist", "/FO", "CSV", "/NH"],
            capture_output=True,
            text=True,
            check=False,
        )
        return {
            line.split('","', 1)[0].strip('"').lower()
            for line in result.stdout.splitlines()
            if line.strip()
        }

    proc = Path("/proc")
    if proc.is_dir():
        names: set[str] = set()
        for entry in proc.iterdir():
            if not entry.name.isdigit():
                continue
            try:
                names.add((entry / "comm").read_text().strip().lower())
            except (OSError, UnicodeError):
                continue
        return names

    result = subprocess.run(
        ["ps", "-A", "-o", "comm="],
        capture_output=True,
        text=True,
        check=False,
    )
    return {
        Path(line.strip()).name.lower()
        for line in result.stdout.splitlines()
        if line.strip()
    }


def main() -> int:
    print("=" * 40)
    print("         BBS-GO Docker build and run")
    print("=" * 40)

    # Docker Compose is the default runtime so local startup uses the same
    # MySQL-backed deployment as the release configuration. Use --native for
    # the legacy Go/SQLite development process.
    if "--native" not in sys.argv[1:]:
        # Set proxy for Docker containers to reach the host machine.
        proxy_url = "http://host.docker.internal:7897"
        os.environ.setdefault("HTTP_PROXY", proxy_url)
        os.environ.setdefault("HTTPS_PROXY", proxy_url)
        try:
            docker = require_command("docker", "Install Docker Desktop and start its engine.")
            run([docker, "compose", "up", "-d", "--build"], ROOT)
        except (OSError, RuntimeError, subprocess.CalledProcessError) as error:
            return fail(f"Docker Compose startup failed: {error}")
        print("\nBBS-GO is running at http://127.0.0.1:3000")
        print("Use `docker compose logs -f` to view logs and `docker compose down` to stop it.")
        return 0

    if not (ROOT / "go.mod").is_file():
        return fail("the script is not located in the bbs-go project root")
    if not (WEB_DIR / "package.json").is_file():
        return fail(f"frontend project not found: {WEB_DIR}")

    try:
        go = require_command("go", "Install Go and add it to PATH.")
        corepack = require_command(
            "corepack", "Install a current Node.js release with Corepack."
        )

        if APP_NAME.lower() in process_names():
            return fail(f"{APP_NAME} is already running; stop it before rebuilding")

        if not (WEB_DIR / "node_modules").is_dir():
            print("\n[prepare] Installing frontend dependencies...")
            run(
                command_for(corepack, "pnpm", "install", "--frozen-lockfile"),
                WEB_DIR,
            )

        print("\n[1/3] Building frontend SPA...")
        run(command_for(corepack, "pnpm", "build:spa"), WEB_DIR)

        print("\n[2/3] Building Go backend...")
        run(
            [
                go,
                "build",
                "-trimpath",
                "-ldflags",
                "-s -w",
                "-o",
                str(APP_PATH),
                ".",
            ],
            ROOT,
        )
    except (OSError, RuntimeError, subprocess.CalledProcessError) as error:
        return fail(f"build failed: {error}")

    print("\n[3/3] Starting service...")
    print("Press Ctrl+C to stop.\n")
    try:
        result = subprocess.run([str(APP_PATH)], cwd=ROOT, check=False)
    except KeyboardInterrupt:
        print("\nService stopped.")
        return 130
    except OSError as error:
        return fail(f"failed to start service: {error}")

    if result.returncode == 0:
        print("\nService stopped.")
    else:
        print(f"\nService exited with code {result.returncode}.", file=sys.stderr)
    return result.returncode


if __name__ == "__main__":
    raise SystemExit(main())
