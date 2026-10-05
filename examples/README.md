# Try LogScanner with mock logs

These synthetic logs were generated with IT-ToolBox's `New-LogEntry` function while testing LogScanner. They contain no production data and are included unchanged from that test run.

From the repository root:

```sh
go run ./cmd/logscanner summary --recursive ./examples/logs
go run ./cmd/logscanner scan --recursive --level error ./examples/logs
go run ./cmd/logscanner scan --recursive --contains timeout --ignore-case ./examples/logs
go run ./cmd/logscanner scan --recursive --output jsonl ./examples/logs
```

Expected recursive summary: **2 files, 12 entries, 4 INFO, 2 WARNING, 5 ERROR, 1 UNSPECIFIED**, with no malformed lines or read errors.

Without `--recursive`, only the main file's 9 entries are scanned. The error filter returns 5 matches; case-insensitive `timeout` matching returns 2.

`logs/application.log` includes a multiline error recorded as separate prefixed lines, message punctuation, and an untagged entry. `logs/nested/worker.log` exercises recursive discovery and case-insensitive matching.
