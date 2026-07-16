package github

import "testing"

func TestParseChecksOutput(t *testing.T) {
	raw := `test	pass	1s	https://example
lint	fail	2s	https://example
build	pending	0	https://example`
	cs := ParseChecksOutput(raw)
	if cs.Passed < 1 || cs.Failed < 1 || cs.Pending < 1 {
		t.Fatalf("%+v", cs)
	}
	if cs.AllGreen {
		t.Fatal("should not be all green")
	}
}

func TestParseChecksAllPass(t *testing.T) {
	cs := ParseChecksOutput("ci\tpass\t1s\n")
	if !cs.AllGreen {
		t.Fatalf("%+v", cs)
	}
}

func TestFormatCheckStatus(t *testing.T) {
	s := FormatCheckStatus(CheckStatus{AllGreen: true, Summary: "ok", Passed: 1})
	if s == "" {
		t.Fatal("empty")
	}
}
