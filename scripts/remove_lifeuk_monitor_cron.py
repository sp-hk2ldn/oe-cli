#!/usr/bin/env python3

import os
import subprocess
import tempfile


BEGIN_MARKER = "# BEGIN LIFEUK_MONITOR"
END_MARKER = "# END LIFEUK_MONITOR"


def main() -> int:
    result = subprocess.run(["crontab", "-l"], capture_output=True, text=True)
    existing = result.stdout if result.returncode == 0 else ""

    lines = existing.splitlines()
    filtered = []
    inside = False
    for line in lines:
        stripped = line.strip()
        if stripped == BEGIN_MARKER:
            inside = True
            continue
        if stripped == END_MARKER:
            inside = False
            continue
        if not inside:
            filtered.append(line)

    with tempfile.NamedTemporaryFile("w", delete=False, encoding="utf-8") as handle:
        if filtered:
            handle.write("\n".join(filtered) + "\n")
        temp_path = handle.name

    try:
        subprocess.run(["crontab", temp_path], check=True)
    finally:
        os.unlink(temp_path)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
