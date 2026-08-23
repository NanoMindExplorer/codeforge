with open('cmd/codeforge/main.go', 'r') as f:
    content = f.read()

bad_block = """	if skipWizard {
		_ = onboarding.MarkSkipped()
		// reload config after wizard may have written default_provider
		if c2, err := config.Load(); err == nil && c2 != nil {
			cfg = c2
		}
	}"""

good_block = """	if skipWizard {
		_ = onboarding.MarkSkipped()
	}"""

content = content.replace(bad_block, good_block)

with open('cmd/codeforge/main.go', 'w') as f:
    f.write(content)
