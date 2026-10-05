package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	input := "[10/05/2026 09:00:00 AM] - [ERROR]: Timeout\ninvalid\n"
	for _, tc := range []struct {
		args []string
		code int
	}{
		{[]string{"scan", "--contains", "timeout", "--ignore-case", "--output", "jsonl", "-"}, 0},
		{[]string{"scan", "--strict", "-"}, 2},
		{[]string{"scan", "--level", "info", "-"}, 1},
		{[]string{"summary", "--output", "jsonl", "-"}, 0},
		{[]string{"--since", "2026-10-06", "--until", "2026-10-05", "-"}, 2},
		{[]string{"--regex", "[", "-"}, 2},
	} {
		var out, diagnostics bytes.Buffer
		if code := run(tc.args, strings.NewReader(input), &out, &diagnostics); code != tc.code {
			t.Fatalf("%v: code %d want %d: %s", tc.args, code, tc.code, diagnostics.String())
		}
		if strings.Contains(strings.Join(tc.args, " "), "jsonl") && !json.Valid(bytes.TrimSpace(out.Bytes())) {
			t.Fatalf("invalid JSON: %s", out.String())
		}
	}
}
