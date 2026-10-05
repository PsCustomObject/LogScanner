package logentry

import "testing"

func TestParse(t *testing.T) {
	for _, tc := range []struct{ raw, level, message string }{
		{"[10/05/2026 12:00:00 AM] - [INFO]: start", "INFO", "start"},
		{"[10/05/2026 12:00:00 PM] - [ERROR]: [server]: failed", "ERROR", "[server]: failed"},
		{"[10/05/2026 09:14:32 AM] - plain", "UNSPECIFIED", "plain"},
		{"[10/05/2026 09:14:32 AM] - [WARNING]: ", "WARNING", ""},
	} {
		e, err := Parse(tc.raw)
		if err != nil || e.Level != tc.level || e.Message != tc.message || e.Raw != tc.raw {
			t.Fatalf("Parse(%q) = %+v, %v", tc.raw, e, err)
		}
	}
	for _, raw := range []string{"", "noise", "[02/30/2026 09:14:32 AM] - [INFO]: x", "[10/05/2026 19:14:32 AM] - x"} {
		if _, err := Parse(raw); err == nil {
			t.Errorf("accepted invalid line %q", raw)
		}
	}
}
