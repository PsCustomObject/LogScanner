package scan

import (
	"logscanner/internal/logentry"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestFilterBoundaries(t *testing.T) {
	e, err := logentry.Parse("[10/05/2026 09:00:00 AM] - [WARNING]: Connection Timeout")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		options Options
		want    bool
	}{
		{"inclusive", Options{Since: e.Time, Until: e.Time}, true},
		{"after", Options{Since: e.Time.Add(1)}, false},
		{"before", Options{Until: e.Time.Add(-1)}, false},
		{"exact severity", Options{Level: "ERROR"}, false},
		{"minimum severity", Options{MinLevel: "INFO"}, true},
		{"literal case", Options{Contains: "timeout"}, false},
		{"ignore case", Options{Contains: "timeout", IgnoreCase: true}, true},
		{"regex", Options{Regex: regexp.MustCompile("Connection.*Timeout")}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.options.matches(e); got != tc.want {
				t.Fatalf("matches=%v want=%v", got, tc.want)
			}
		})
	}
	untagged, _ := logentry.Parse("[10/05/2026 09:00:00 AM] - hello")
	if (Options{MinLevel: "INFO"}).matches(untagged) {
		t.Fatal("unspecified entry matched minimum severity")
	}
}

func TestStreamingAndFilters(t *testing.T) {
	long := strings.Repeat("x", 100000)
	input := "\ufeff[10/05/2026 09:00:00 AM] - [INFO]: start\r\nnoise\n[10/05/2026 09:00:01 AM] - [ERROR]: " + long
	var entries []logentry.Entry
	s, err := Run([]string{"-"}, strings.NewReader(input), Options{MinLevel: "WARNING"}, func(e logentry.Entry) error { entries = append(entries, e); return nil }, func(error) {})
	if err != nil || s.Entries != 2 || s.Malformed != 1 || s.Matches != 1 || len(entries) != 1 || entries[0].Line != 3 || entries[0].Message != long {
		t.Fatalf("unexpected result %+v %v", s, err)
	}
}

func TestDiscovery(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "nested")
	if err := os.Mkdir(nested, 0700); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{filepath.Join(dir, "a.log"), filepath.Join(dir, "b.txt"), filepath.Join(nested, "c.log")} {
		if err := os.WriteFile(p, nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, recursive := range []bool{false, true} {
		files, err := Discover([]string{dir, filepath.Join(dir, "a.log")}, Options{Recursive: recursive, Include: "*.log"})
		want := 1
		if recursive {
			want = 2
		}
		if err != nil || len(files) != want {
			t.Fatalf("files=%v err=%v", files, err)
		}
	}
}
