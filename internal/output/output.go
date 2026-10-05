package output

import (
	"encoding/json"
	"fmt"
	"io"

	"logscanner/internal/logentry"
	"logscanner/internal/scan"
)

func Entry(w io.Writer, format string, e logentry.Entry) error {
	if format == "jsonl" {
		return json.NewEncoder(w).Encode(e)
	}
	_, err := fmt.Fprintf(w, "%s | File: %s | Line: %d\n", e.Raw, e.File, e.Line)
	return err
}

func Summary(w io.Writer, format string, s scan.Stats) error {
	if format == "jsonl" {
		return json.NewEncoder(w).Encode(s)
	}
	_, err := fmt.Fprintf(w, "Files: %d\nParsed entries: %d\nMatches: %d\nMalformed lines: %d\nRead errors: %d\n", s.Files, s.Entries, s.Matches, s.Malformed, s.ReadErrors)
	if err != nil {
		return err
	}
	for _, level := range []string{"INFO", "WARNING", "ERROR", "UNSPECIFIED"} {
		if _, err := fmt.Fprintf(w, "%s: %d\n", level, s.Levels[level]); err != nil {
			return err
		}
	}
	return nil
}
