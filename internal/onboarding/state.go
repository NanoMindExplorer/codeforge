// Package onboarding tracks first-run setup and key source helpers (W2).
package onboarding

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// State is persisted at ~/.codeforge/onboarding.json.
type State struct {
	Completed bool      `json:"completed"`
	Skipped   bool      `json:"skipped"`
	Provider  string    `json:"provider,omitempty"`
	Model     string    `json:"model,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// Dir returns ~/.codeforge.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codeforge"), nil
}

// Path is the onboarding state file.
func Path() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "onboarding.json"), nil
}

// Load reads state; missing file → zero value (not completed).
func Load() (State, error) {
	p, err := Path()
	if err != nil {
		return State{}, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return State{}, nil
		}
		return State{}, err
	}
	var s State
	if err := json.Unmarshal(b, &s); err != nil {
		return State{}, err
	}
	return s, nil
}

// Save writes state (creates ~/.codeforge if needed).
func Save(s State) error {
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	s.UpdatedAt = time.Now().UTC()
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}

// MarkCompleted records a successful setup.
func MarkCompleted(provider, model string) error {
	return Save(State{Completed: true, Provider: provider, Model: model})
}

// MarkSkipped records user skip (--skip-wizard / decline).
func MarkSkipped() error {
	return Save(State{Skipped: true})
}

// NeedsWizard reports whether the CLI first-run wizard should run.
// False when skip flag is set, any API key is present, or user previously skipped.
func NeedsWizard(skipFlag bool) bool {
	if skipFlag {
		return false
	}
	if HasAnyAPIKey() {
		return false
	}
	st, err := Load()
	if err == nil && st.Skipped {
		return false
	}
	return true
}
