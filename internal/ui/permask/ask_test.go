package permask

import (
	"strings"
	"testing"

	"github.com/codeforge/tui/internal/theme"
)

func TestClassifyRisk(t *testing.T) {
	if ClassifyRisk("run_command", `{"command":"ls"}`, false) != RiskMedium {
		t.Fatal("shell medium")
	}
	if ClassifyRisk("run_command", `{"command":"sudo rm -rf /"}`, true) != RiskHigh {
		t.Fatal("dangerous high")
	}
	if ClassifyRisk("read_file", `{"path":"a.go"}`, false) != RiskLow {
		t.Fatal("read low")
	}
	if ClassifyRisk("write_file", `{"path":"a.go"}`, false) != RiskMedium {
		t.Fatal("write medium")
	}
}

func TestFormatCommandExtractsShell(t *testing.T) {
	in := `{"command":"go test ./... -count=1","timeout_sec":60}`
	out := FormatCommand("run_command", in)
	if !strings.Contains(out, "go test ./...") {
		t.Fatal(out)
	}
	if !strings.Contains(out, "command:") {
		t.Fatal(out)
	}
}

func TestViewShowsRiskBadgeAndFullCommand(t *testing.T) {
	theme.Set(theme.Aurora())
	theme.SetMotion(false)
	m := New()
	cmd := "echo hello && go test ./internal/tool -count=1"
	m.Open("run_command", `{"command":"`+cmd+`"}`, "Shell requires approval", false)
	view := m.View()
	if !strings.Contains(view, "MEDIUM") && !strings.Contains(view, "HIGH") {
		// risk badge text
		if !strings.Contains(view, "RISK") {
			t.Fatal("missing risk badge:\n", view)
		}
	}
	if !strings.Contains(view, "go test") {
		t.Fatal("full command missing:\n", view)
	}
	if !strings.Contains(view, "Permission") {
		t.Fatal(view)
	}
}

func TestDangerousCannotAlways(t *testing.T) {
	m := New()
	m.Open("run_command", `{"command":"rm -rf /"}`, "dangerous", true)
	m.Yes(true)
	if m.Always {
		t.Fatal("must not remember dangerous always-allow")
	}
	if !m.Allow {
		t.Fatal("allow once ok")
	}
}
