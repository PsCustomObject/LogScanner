package scan

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"logscanner/internal/logentry"
)

type Options struct {
	Recursive                 bool
	Include                   string
	Level, MinLevel, Contains string
	IgnoreCase                bool
	Regex                     *regexp.Regexp
	Since, Until              time.Time
}

type Stats struct {
	Files      int            `json:"files"`
	Entries    int            `json:"entries"`
	Matches    int            `json:"matches"`
	Malformed  int            `json:"malformed"`
	ReadErrors int            `json:"read_errors"`
	Levels     map[string]int `json:"levels"`
}

// Discover preserves operand order and sorts directory results through WalkDir.
// Overlapping operands are deduplicated; directory symlinks are not followed.
func Discover(paths []string, o Options) ([]string, error) {
	var files []string
	seen := map[string]bool{}
	add := func(path string) error {
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if !seen[abs] {
			seen[abs] = true
			files = append(files, path)
		}
		return nil
	}
	for _, path := range paths {
		if path == "-" {
			if !seen[path] {
				seen[path] = true
				files = append(files, path)
			}
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if info.Mode().IsRegular() {
			if err := add(path); err != nil {
				return nil, err
			}
			continue
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("%s: not a regular file or directory", path)
		}
		err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if p != path && !o.Recursive {
					return filepath.SkipDir
				}
				return nil
			}
			if !d.Type().IsRegular() {
				return nil
			}
			match, err := filepath.Match(o.Include, d.Name())
			if err != nil {
				return err
			}
			if match {
				return add(p)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func (o Options) matches(e logentry.Entry) bool {
	if o.Level != "" && e.Level != o.Level {
		return false
	}
	if o.MinLevel != "" && logentry.Rank(e.Level) < logentry.Rank(o.MinLevel) {
		return false
	}
	if !o.Since.IsZero() && e.Time.Before(o.Since) {
		return false
	}
	if !o.Until.IsZero() && e.Time.After(o.Until) {
		return false
	}
	message, contains := e.Message, o.Contains
	if o.IgnoreCase {
		message = strings.ToLower(message)
		contains = strings.ToLower(contains)
	}
	if !strings.Contains(message, contains) {
		return false
	}
	return o.Regex == nil || o.Regex.MatchString(e.Message)
}

func Run(files []string, stdin io.Reader, o Options, emit func(logentry.Entry) error, diagnostic func(error)) (Stats, error) {
	s := Stats{Levels: map[string]int{"INFO": 0, "WARNING": 0, "ERROR": 0, "UNSPECIFIED": 0}}
	for _, file := range files {
		var r io.Reader = stdin
		var f *os.File
		if file != "-" {
			var err error
			f, err = os.Open(file)
			if err != nil {
				s.ReadErrors++
				diagnostic(err)
				continue
			}
			r = f
		}
		s.Files++
		reader := bufio.NewReader(r)
		for line := 1; ; line++ {
			raw, err := reader.ReadString('\n')
			if len(raw) > 0 {
				raw = strings.TrimSuffix(strings.TrimSuffix(raw, "\n"), "\r")
				if line == 1 {
					raw = strings.TrimPrefix(raw, "\ufeff")
				}
				e, parseErr := logentry.Parse(raw)
				if parseErr != nil {
					s.Malformed++
					diagnostic(fmt.Errorf("%s:%d: %w", file, line, parseErr))
				} else {
					e.File, e.Line = file, line
					s.Entries++
					if o.matches(e) {
						s.Matches++
						s.Levels[e.Level]++
						if err := emit(e); err != nil {
							if f != nil {
								f.Close()
							}
							return s, err
						}
					}
				}
			}
			if err != nil {
				if err != io.EOF {
					s.ReadErrors++
					diagnostic(fmt.Errorf("%s: %w", file, err))
				}
				break
			}
		}
		if f != nil {
			f.Close()
		}
	}
	return s, nil
}
