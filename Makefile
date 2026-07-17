.PHONY: test test-race build vet fmt fmt-check install-hooks check-version bump \
	release-dry release-notes release-gate smoke-matrix dogfood dogfood-offline \
	dogfood-help termux-meta coverage govulncheck ci

VERSION := $(shell tr -d '[:space:]' < VERSION 2>/dev/null || echo 0.0.0)

test:
	GOSUMDB=off go test ./...

# Q0.1 — race detector on critical packages
test-race:
	bash scripts/test-race.sh

# Q0.2 — coverage profile + floor (scripts/coverage-floor.txt)
coverage:
	bash scripts/coverage-check.sh

vet:
	GOSUMDB=off go vet ./...

# Format all Go sources (run before commit if hooks not installed)
fmt:
	gofmt -w .

# Fail if any file needs gofmt (CI / release-gate)
fmt-check:
	bash scripts/gofmt-check.sh

# Enable repo pre-commit hook (gofmt staged *.go)
install-hooks:
	bash scripts/install-hooks.sh

build:
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.ProjectVersion=$(VERSION)" -o codeforge ./cmd/codeforge/

check-version:
	bash scripts/check-version.sh

# make bump V=1.9.0
bump:
	@test -n "$(V)" || (echo "Usage: make bump V=X.Y.Z" && exit 1)
	bash scripts/bump-version.sh $(V)

# Local goreleaser snapshot (no publish)
release-dry:
	goreleaser release --snapshot --clean --skip=publish

# Print CHANGELOG section + commits for VERSION
release-notes:
	bash scripts/release-notes.sh $(VERSION)

# W4 automated release gate (+ Q0 coverage floor)
release-gate:
	bash scripts/release-gate.sh

# Batch F terminal env smoke
smoke-matrix:
	bash scripts/smoke-matrix.sh

# Emit termux-packages metadata
termux-meta:
	bash contrib/termux/package.sh

# Field + automated dogfood evidence → docs/dogfood/RESULTS.md
dogfood:
	bash scripts/dogfood-run.sh

# Q0.3 — offline only (no live API)
dogfood-offline:
	DOGFOOD_LIVE=0 bash scripts/dogfood-run.sh

# Q0.5 — vulnerability scan (warn by default)
govulncheck:
	bash scripts/govulncheck.sh

dogfood-help:
	@echo "Dogfood program: docs/dogfood/PROGRAM.md (10 working days)"
	@echo "Run evidence:    make dogfood   (DOGFOOD_LIVE=0 to skip live API)"
	@echo "Offline CI:      make dogfood-offline"
	@echo "Results:         docs/dogfood/RESULTS.md"
	@echo "Daily template:  docs/dogfood/TEMPLATE.md"
	@echo "Checklist:       docs/DOGFOOD.md"
	@echo "Scorecard:       docs/dogfood/SCORECARD.md"
	@echo "Audit roadmap:   docs/AUDIT_AND_ROADMAP.md"
	@echo "Batches:         BATCH_BC / BATCH_DE / BATCH_F"
	@echo "Release gate:    docs/RELEASE_GATE.md · make release-gate"

# Local full gate used by maintainers (mirrors CI core)
ci: check-version fmt-check vet coverage build
	@echo "CI local gate OK (v$(VERSION))"
