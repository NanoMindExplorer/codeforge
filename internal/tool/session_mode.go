package tool

import "strings"

// SessionMode is the Grok-style session cycle (Shift+Tab):
// BUILD (staged writes) → DESIGN (plan-only) → YOLO (always approve) → BUILD.
type SessionMode int

const (
	// SessionBuild stages file writes for review (legacy "Plan" write mode).
	SessionBuild SessionMode = iota
	// SessionDesign is read-only design phase; only plan.md may be written.
	SessionDesign
	// SessionYolo applies writes immediately (legacy "Act" / always-approve).
	SessionYolo
)

// Label returns the footer badge text.
func (m SessionMode) Label() string {
	switch m {
	case SessionDesign:
		return "DESIGN"
	case SessionYolo:
		return "YOLO"
	default:
		return "BUILD"
	}
}

// Description is a short user-facing explanation.
func (m SessionMode) Description() string {
	switch m {
	case SessionDesign:
		return "Design plan only — explores codebase, writes plan.md; no other file edits"
	case SessionYolo:
		return "Always-approve — writes apply to disk immediately"
	default:
		return "Build — file writes staged for review before apply"
	}
}

// WriteMode maps session mode onto StagedWriter write gating.
func (m SessionMode) WriteMode() WriteMode {
	switch m {
	case SessionDesign:
		return ModeDesign
	case SessionYolo:
		return ModeAct
	default:
		return ModePlan
	}
}

// Next cycles BUILD → DESIGN → YOLO → BUILD.
func (m SessionMode) Next() SessionMode {
	switch m {
	case SessionBuild:
		return SessionDesign
	case SessionDesign:
		return SessionYolo
	default:
		return SessionBuild
	}
}

// ParseSessionMode accepts design/plan, build, yolo/act, etc.
func ParseSessionMode(s string) (SessionMode, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "design", "plan", "design-plan":
		return SessionDesign, true
	case "build", "normal", "default":
		return SessionBuild, true
	case "yolo", "act", "always", "always-approve", "always_approve":
		return SessionYolo, true
	default:
		return SessionBuild, false
	}
}
