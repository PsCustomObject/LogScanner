package main

import (
	"flag"
	"fmt"
	"io"
	"logscanner/internal/logentry"
	"logscanner/internal/output"
	"logscanner/internal/scan"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var version = "0.1.0-dev"

func main() { os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	command := "scan"
	if len(args) > 0 && (args[0] == "scan" || args[0] == "summary") {
		command, args = args[0], args[1:]
	}
	if len(args) > 0 && args[0] == "version" {
		fmt.Fprintln(stdout, "logscanner", version)
		return 0
	}
	f := flag.NewFlagSet("logscanner", flag.ContinueOnError)
	f.SetOutput(stderr)
	var o scan.Options
	var pattern, since, until, format string
	var strict bool
	f.BoolVar(&o.Recursive, "recursive", false, "recurse into directories")
	f.StringVar(&o.Include, "include", "*.log", "directory filename glob (explicit files bypass it)")
	f.StringVar(&o.Level, "level", "", "exact severity: INFO, WARNING, ERROR, UNSPECIFIED")
	f.StringVar(&o.MinLevel, "min-level", "", "minimum severity: INFO, WARNING, ERROR")
	f.StringVar(&o.Contains, "contains", "", "literal message substring")
	f.BoolVar(&o.IgnoreCase, "ignore-case", false, "case-insensitive text and regex matching")
	f.StringVar(&pattern, "regex", "", "Go regular expression matched against message")
	f.StringVar(&since, "since", "", "inclusive local time: YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS")
	f.StringVar(&until, "until", "", "inclusive local time: YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS")
	f.StringVar(&format, "output", "text", "text or jsonl")
	f.BoolVar(&strict, "strict", false, "return failure if any line is malformed")
	f.Usage = func() {
		fmt.Fprintln(stderr, "Usage: logscanner [scan|summary] [flags] <file|directory|->...\nFlags must precede paths. Use - to read stdin.")
		f.PrintDefaults()
	}
	if err := f.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	fail := func(err error) int { fmt.Fprintln(stderr, "logscanner:", err); return 2 }
	if f.NArg() == 0 {
		f.Usage()
		return 2
	}
	if format != "text" && format != "jsonl" {
		return fail(fmt.Errorf("invalid output %q", format))
	}
	o.Level, o.MinLevel = strings.ToUpper(o.Level), strings.ToUpper(o.MinLevel)
	if o.Level != "" && o.Level != "UNSPECIFIED" && logentry.Rank(o.Level) == 0 {
		return fail(fmt.Errorf("invalid level %q", o.Level))
	}
	if o.MinLevel != "" && logentry.Rank(o.MinLevel) == 0 {
		return fail(fmt.Errorf("invalid minimum level %q", o.MinLevel))
	}
	if o.Level != "" && o.MinLevel != "" {
		return fail(fmt.Errorf("use either --level or --min-level"))
	}
	if _, err := filepath.Match(o.Include, "test.log"); err != nil {
		return fail(err)
	}
	parseTime := func(value string) (time.Time, error) {
		if value == "" {
			return time.Time{}, nil
		}
		for _, layout := range []string{"2006-01-02T15:04:05", "2006-01-02"} {
			if t, err := time.Parse(layout, value); err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("invalid local time %q; use YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS (no timezone)", value)
	}
	var err error
	if o.Since, err = parseTime(since); err != nil {
		return fail(err)
	}
	if o.Until, err = parseTime(until); err != nil {
		return fail(err)
	}
	if !o.Since.IsZero() && !o.Until.IsZero() && o.Since.After(o.Until) {
		return fail(fmt.Errorf("since must not be after until"))
	}
	if pattern != "" {
		if o.IgnoreCase {
			pattern = "(?i)" + pattern
		}
		o.Regex, err = regexp.Compile(pattern)
		if err != nil {
			return fail(err)
		}
	}
	files, err := scan.Discover(f.Args(), o)
	if err != nil {
		return fail(err)
	}
	// Cap individual diagnostics while retaining exact counts for large damaged files.
	diagnostics := 0
	diagnostic := func(err error) {
		diagnostics++
		if diagnostics <= 10 {
			fmt.Fprintln(stderr, err)
		}
	}
	s, err := scan.Run(files, stdin, o, func(e logentry.Entry) error {
		if command == "summary" {
			return nil
		}
		return output.Entry(stdout, format, e)
	}, diagnostic)
	if diagnostics > 10 {
		fmt.Fprintf(stderr, "%d further diagnostics suppressed\n", diagnostics-10)
	}
	if s.Malformed > 0 {
		fmt.Fprintf(stderr, "Skipped %d malformed lines\n", s.Malformed)
	}
	if err != nil {
		return fail(err)
	}
	if command == "summary" {
		if err := output.Summary(stdout, format, s); err != nil {
			return fail(err)
		}
	}
	if s.ReadErrors > 0 || (strict && s.Malformed > 0) {
		return 2
	}
	if s.Matches == 0 {
		return 1
	}
	return 0
}
