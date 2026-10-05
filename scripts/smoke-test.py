"""Build and exercise the native executable, including its real exit codes."""

import json
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]


def main():
    with tempfile.TemporaryDirectory(prefix="logscanner-smoke-") as directory:
        binary = Path(directory) / ("logscanner.exe" if os.name == "nt" else "logscanner")
        subprocess.run(
            ["go", "build", "-o", str(binary), "./cmd/logscanner"],
            cwd=ROOT, check=True,
        )

        def run(*args, code=0, input=None):
            result = subprocess.run(
                [str(binary), *args], cwd=ROOT, input=input,
                capture_output=True, text=True, encoding="utf-8",
            )
            assert result.returncode == code, (args, result.returncode, result.stderr)
            return result

        summary = json.loads(run("summary", "--recursive", "--output", "jsonl", "examples/logs").stdout)
        assert summary == {
            "files": 2, "entries": 12, "matches": 12, "malformed": 0,
            "read_errors": 0,
            "levels": {"INFO": 4, "WARNING": 2, "ERROR": 5, "UNSPECIFIED": 1},
        }, summary
        shallow = json.loads(run("summary", "--output", "jsonl", "examples/logs").stdout)
        assert shallow["files"] == 1 and shallow["matches"] == 9, shallow

        errors = run("scan", "--recursive", "--level", "error", "examples/logs")
        lines = errors.stdout.splitlines()
        assert len(lines) == 5, lines
        for line in lines:
            assert line.startswith("[10/05/2026 10:18:49 PM] - [ERROR]: "), line
            assert " | File: " in line and " | Line: " in line, line
        assert lines[0].endswith("application.log | Line: 4"), lines[0]

        matches = run("scan", "--recursive", "--contains", "timeout", "--ignore-case", "--output", "jsonl", "examples/logs")
        assert len([json.loads(line) for line in matches.stdout.splitlines()]) == 2
        run("scan", "--contains", "no-such-message", "examples/logs", code=1)
        malformed = run("scan", "--strict", "-", code=2, input="malformed\n")
        assert "Skipped 1 malformed lines" in malformed.stderr
        assert run("--help").returncode == 0
        assert run("version").stdout.startswith("logscanner ")
        print("Native CLI smoke tests passed.")


if __name__ == "__main__":
    main()
