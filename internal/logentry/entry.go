// Package logentry parses the line format emitted by IT-ToolBox New-LogEntry.
package logentry

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const Layout = "01/02/2006 03:04:05 PM"

type Entry struct {
	Timestamp string    `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	File      string    `json:"file"`
	Line      int       `json:"line"`
	Raw       string    `json:"raw"`
	Time      time.Time `json:"-"`
}

func Parse(raw string) (Entry, error) {
	e := Entry{Raw: raw, Level: "UNSPECIFIED"}
	if !utf8.ValidString(raw) || strings.ContainsRune(raw, '\x00') {
		return e, fmt.Errorf("invalid UTF-8 or unsupported encoding")
	}
	if !strings.HasPrefix(raw, "[") {
		return e, fmt.Errorf("missing timestamp prefix")
	}
	end := strings.Index(raw, "] - ")
	if end < 0 {
		return e, fmt.Errorf("missing timestamp separator")
	}
	e.Timestamp = raw[1:end]
	t, err := time.Parse(Layout, e.Timestamp)
	if err != nil {
		return e, fmt.Errorf("invalid timestamp %q", e.Timestamp)
	}
	e.Time = t // UTC is a neutral container for the recorded wall-clock value.
	e.Message = raw[end+4:]
	for _, level := range []string{"INFO", "WARNING", "ERROR"} {
		prefix := "[" + level + "]: "
		if strings.HasPrefix(e.Message, prefix) {
			e.Level = level
			e.Message = strings.TrimPrefix(e.Message, prefix)
			break
		}
	}
	return e, nil
}

func Rank(level string) int {
	switch level {
	case "INFO":
		return 1
	case "WARNING":
		return 2
	case "ERROR":
		return 3
	}
	return 0
}
