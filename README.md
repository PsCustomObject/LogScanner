# LogScanner

**Find the entries that matter in your PowerShell automation logs.**

LogScanner is a small Go CLI for the log format produced by IT-ToolBox's `New-LogEntry` function. Scan a file or a directory tree, filter by severity, message, or time, and get readable results with their source file and line number.

```text
[10/05/2026 09:14:34 AM] - [ERROR]: Connection failed | File: logs/application.log | Line: 3
```

Built with the Go standard library. No third-party dependencies, background services, or database. Compiled binaries run without PowerShell, .NET, or Go installed.

## Why this exists

`New-LogEntry` has been used for years to record PowerShell automation activity. Finding failures across those logs meant opening files in a text editor and searching manually. LogScanner turns that workflow into a repeatable command, while keeping each result traceable to the original file.

The scope is intentional: understand one log format well, stream the input, and make the output useful both at the terminal and in scripts.

## Features

- Scan individual files, multiple paths, directories, or standard input.
- Recurse through directories with a configurable filename pattern.
- Filter by exact or minimum severity, literal text, regular expressions, and timestamp bounds.
- Print matching entries with explicit file and line labels, or emit JSON Lines.
- Summarize matching entries by severity.
- Report malformed lines and read failures separately from results.
- Read incrementally, including lines larger than 64 KiB.
- Build native executables for Windows, Linux, and macOS.

## Quick start

Want to try it first? The repository includes [mock logs](examples/README.md) generated with `New-LogEntry`:

```sh
go run ./cmd/logscanner scan --recursive --level error ./examples/logs
```

### Download a binary

Download the executable for your operating system and CPU architecture. No runtime installation is required.

**Availability:** these links become active once the first release is published. Until then, build from source below.

| Platform | Intel / AMD (amd64) | ARM64 |
| --- | --- | --- |
| Windows | [Download .exe](https://github.com/PsCustomObject/LogScanner/releases/latest/download/logscanner-windows-amd64.exe) | [Download .exe](https://github.com/PsCustomObject/LogScanner/releases/latest/download/logscanner-windows-arm64.exe) |
| Linux | [Download](https://github.com/PsCustomObject/LogScanner/releases/latest/download/logscanner-linux-amd64) | [Download](https://github.com/PsCustomObject/LogScanner/releases/latest/download/logscanner-linux-arm64) |
| macOS | [Download](https://github.com/PsCustomObject/LogScanner/releases/latest/download/logscanner-macos-amd64) | [Download for Apple Silicon](https://github.com/PsCustomObject/LogScanner/releases/latest/download/logscanner-macos-arm64) |

[All releases](https://github.com/PsCustomObject/LogScanner/releases) · [SHA-256 checksums](https://github.com/PsCustomObject/LogScanner/releases/latest/download/SHA256SUMS.txt)

On Linux or macOS, rename the downloaded binary to `logscanner`, make it executable, and run it:

```sh
chmod +x logscanner
./logscanner --help
```

On Windows, rename it to `logscanner.exe` and run `.\logscanner.exe --help`. Put the executable in a directory on your `PATH` to use it from anywhere.

macOS binaries are currently unsigned and not notarized; macOS may require approval in Privacy & Security before they can run.

### Build from source

Building from source requires **Go 1.23 or newer**. From the repository root:

```sh
go build -o bin/logscanner ./cmd/logscanner
./bin/logscanner scan --recursive --level error ./logs
./bin/logscanner summary --recursive ./logs
```

On Windows, build and run with PowerShell:

```powershell
go build -o bin/logscanner.exe ./cmd/logscanner
.\bin\logscanner.exe scan --recursive --level error .\logs
```

For development, run directly from source:

```sh
go run ./cmd/logscanner scan --recursive --level error ./logs
```

The following examples use `logscanner`, assuming the binary is on your `PATH`. Otherwise, substitute `./bin/logscanner` or `.\bin\logscanner.exe`.

## Usage

```text
logscanner [scan|summary] [flags] <file|directory|->...
```

Place flags **before** paths. `scan` is the default command. Use `-` to read standard input.

### Find errors across a directory tree

```sh
logscanner scan --recursive --level error ./logs
```

### Find warnings and errors containing a phrase

```sh
logscanner scan --min-level warning --contains "timeout" --ignore-case application.log
```

### Restrict a scan to a time range

```sh
logscanner scan --since 2026-10-05T09:00:00 --until 2026-10-05T17:00:00 application.log
```

### Match several messages and export JSON Lines

```sh
logscanner scan --regex "timeout|connection refused" --ignore-case --output jsonl application.log
```

### Scan selected filenames or multiple locations

```sh
logscanner scan --recursive --include "worker-*.log" ./logs
logscanner scan --level error application.log ./archive
```

Quote patterns so the shell does not expand them. `--include` applies to directory discovery; explicitly supplied files are always scanned. The default pattern is `*.log`.

### Read from a pipeline

```sh
cat application.log | logscanner scan --level error -
```

### Summarize a scan

```sh
logscanner summary --recursive ./logs
logscanner summary --recursive --level error --output jsonl ./logs
```

Summary severity counts describe entries that match the supplied filters. Parsed-entry counts include all successfully parsed entries.

### Show help or version

```sh
logscanner --help
logscanner version
```

## Options

All supplied filters must match. Text filters inspect the message, excluding its timestamp and severity prefix.

| Option | Description |
| --- | --- |
| `--recursive` | Include subdirectories |
| `--include PATTERN` | Directory filename glob; defaults to `*.log` |
| `--level LEVEL` | Exact `INFO`, `WARNING`, `ERROR`, or `UNSPECIFIED` |
| `--min-level LEVEL` | `INFO`, `WARNING`, or `ERROR` and above |
| `--contains TEXT` | Literal message substring |
| `--regex PATTERN` | Go regular expression matched against the message |
| `--ignore-case` | Case-insensitive substring and regex matching |
| `--since TIME` | Inclusive lower timestamp bound |
| `--until TIME` | Inclusive upper timestamp bound |
| `--output FORMAT` | `text` (default) or `jsonl` |
| `--strict` | Return a failure code if any line is malformed |

Severity arguments are case insensitive. Use either `--level` or `--min-level`. Minimum severity filters exclude untagged entries.

Time bounds accept `YYYY-MM-DD` or `YYYY-MM-DDTHH:MM:SS`. A date alone means midnight, including for `--until`. To include an entire day, use an upper bound such as `2026-10-05T23:59:59`.

## Supported log format

LogScanner understands timestamped lines emitted by the current `New-LogEntry` formatter:

```text
[10/05/2026 09:14:32 AM] - [INFO]: Starting processing
[10/05/2026 09:14:33 AM] - [WARNING]: Retry required
[10/05/2026 09:14:34 AM] - [ERROR]: Connection failed
[10/05/2026 09:14:35 AM] - Message written with -NoTag
```

- Timestamp format: `MM/dd/yyyy hh:mm:ss AM/PM`.
- Tagged severities: `INFO`, `WARNING`, and `ERROR`.
- Untagged entries are represented as `UNSPECIFIED`.
- UTF-8 input is supported, with or without an initial UTF-8 BOM.
- LF and CRLF line endings are supported, including a final line without a newline.
- Empty messages are valid; blank physical lines are malformed.

The logger prefixes each physical line of a multiline message separately. LogScanner treats each as an independent entry because the original grouping is not recorded in the file.

An untagged message beginning with an exact recognized severity prefix is indistinguishable from a tagged entry and is parsed as tagged.

### Timestamp and encoding limitations

Source timestamps contain no timezone. Filters compare recorded wall-clock values without conversion, and timezone-bearing filter arguments are rejected. Logs from different timezones cannot be automatically aligned.

UTF-16, legacy encodings, localized date separators or AM/PM markers, and historical format variants are not currently supported. Compatibility with older logs should be checked against representative samples.

## Output

### Text

Each matching entry occupies one output line: original log text, file path, and one-based source line number.

```text
[10/05/2026 09:14:34 AM] - [ERROR]: Connection failed | File: logs/application.log | Line: 3
```

### JSON Lines

Each match is emitted as an independent JSON object:

```json
{"timestamp":"10/05/2026 09:14:34 AM","level":"ERROR","message":"Connection failed","file":"logs/application.log","line":3,"raw":"[10/05/2026 09:14:34 AM] - [ERROR]: Connection failed"}
```

The timestamp preserves the source format and does not imply UTC. The `raw` field retains the original log text, excluding the line ending and initial encoding BOM.

### Summary

For the four example entries above, an unfiltered summary is:

```text
Files: 1
Parsed entries: 4
Matches: 4
Malformed lines: 0
Read errors: 0
INFO: 1
WARNING: 1
ERROR: 1
UNSPECIFIED: 1
```

With `--output jsonl`, `summary` emits one object containing the same counts.

## Errors and exit codes

Results go to stdout; diagnostics go to stderr. Malformed lines are skipped and counted. At most ten individual diagnostics are printed, with a suppression count for additional diagnostics.

`--strict` continues processing but returns a failure code if any malformed lines were found. Discovery failures stop the scan. File open or read failures encountered during scanning are reported, and processing continues with other files.

| Exit code | Meaning |
| --- | --- |
| `0` | At least one match, with no fatal errors |
| `1` | No matches, with no fatal errors |
| `2` | Invalid arguments, discovery/read/output failure, or malformed lines with `--strict` |

Help and version commands return `0`. These are the compiled CLI's exit codes; `go run` wraps nonzero program exits, so use the binary when scripting against exit codes.

## Design

```text
Files / directories / stdin
            |
       Streaming reader
            |
       Log entry parser
            |
         Filters
            |
   Text / JSON Lines / summary
```

The implementation separates CLI behavior, parsing, scanning, and presentation:

| Package | Responsibility |
| --- | --- |
| `cmd/logscanner` | Arguments, command dispatch, and exit codes |
| `internal/logentry` | Entry model, timestamp parsing, and severity handling |
| `internal/scan` | File discovery, incremental reads, and filtering |
| `internal/output` | Text, JSON Lines, and summary rendering |

Files are processed sequentially. Memory use scales with the largest physical line rather than total input size. There is no fixed 64 KiB line limit.

Directory traversal is lexical; operand order is preserved. Results retain source order rather than being globally sorted by timestamp. Overlapping paths are deduplicated by absolute path. Symlink aliases are not resolved, and directory symlinks encountered during traversal are not followed.

The current scope is local, uncompressed log files. Live tailing, compressed input, and additional log formats are outside this version's scope.

## Development

```sh
go test ./...
go vet ./...
python3 scripts/smoke-test.py
```

Tests cover parsing, invalid timestamps, severity and text filters, inclusive date boundaries, directory recursion and deduplication, UTF-8 BOM handling, long lines, final lines without newlines, JSON output, and CLI exit behavior.

GitHub Actions runs tests, static checks, and a native executable smoke test on Windows, Linux, and macOS using stable Go. A separate Linux job verifies Go 1.23 compatibility. The smoke test requires Python 3 and checks the bundled mock logs, readable output, JSON output, and actual process exit codes. On Windows, use `python` instead of `python3`.

### Cross-platform builds

From a Unix shell:

```sh
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/logscanner-windows-amd64.exe ./cmd/logscanner
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/logscanner-linux-amd64 ./cmd/logscanner
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/logscanner-macos-amd64 ./cmd/logscanner
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/logscanner-macos-arm64 ./cmd/logscanner
```

Each operating system and CPU architecture needs its own binary. Cross-compilation verifies a build; execution should be checked on the target platform before distributing releases.

### Release builds

The release script builds all six Windows, Linux, and macOS amd64/arm64 binaries with an embedded version and generates `dist/SHA256SUMS.txt`. It requires Go and Python 3 and should be run from the repository root:

```sh
sh scripts/build-release.sh v0.1.0
```

Pushing a version tag such as `v0.1.0` triggers the GitHub Actions release workflow. It runs tests and static checks, builds the binaries, and uploads them to a **draft release** for review. Publish the draft to activate the README download links. Stable asset filenames keep the links pointing to the latest published release.

Release builds wait for the full cross-platform CI suite to pass. The MIT license is also attached to the release.

## License

Licensed under the [MIT License](LICENSE).
