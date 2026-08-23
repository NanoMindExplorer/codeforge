import re

with open('internal/onboarding/welcome.go', 'r') as f:
    content = f.read()

bad_func = """func StatusCard(cfg *config.Config, activeName, activeModel string, healthy bool) string {
	var b strings.Builder
	// Compact ASCII (5 lines) + byline
	b.WriteString(BrandASCII())
	b.WriteString("\\n  ")
	b.WriteString(BrandByline)
	b.WriteString("\\n")
	b.WriteString(strings.Repeat("─", 48))
	b.WriteByte('\\n')

	present := PresentCloudKeys()
	res := ResolveActive(cfg)

	if !healthy {
		b.WriteString("Status  ⚠  No API key yet\\n")
		b.WriteString("  /setup gemini <AIza…>   free tier\\n")
		b.WriteString("  /setup grok xai-…       Grok 4.5\\n")
		b.WriteString("  export XAI_API_KEY=…    then restart\\n")
	} else {
		b.WriteString(fmt.Sprintf("Status  ✓  %s", activeName))
		if activeModel != "" {
			b.WriteString(" · " + activeModel)
		}
		b.WriteByte('\\n')
		if res.Reason != "" {
			b.WriteString("  why  " + res.Reason + "\\n")
		}
		if src, _ := KeySource(activeName); src != "" {
			b.WriteString("  key  " + src + "\\n")
		}
		if len(present) > 1 {
			var names []string
			for _, p := range present {
				if p.Name != normalizeName(activeName) {
					names = append(names, p.Name)
				}
			}
			if len(names) > 0 {
				b.WriteString(fmt.Sprintf("  also %s  → /provider\\n", strings.Join(names, ", ")))
			}
		}
	}
	b.WriteString(strings.Repeat("─", 48))
	b.WriteString("\\nShift+Tab modes · /help · /setup · /doctor\\n")
	return b.String()
}"""

good_func = """func StatusCard(cfg *config.Config, activeName, activeModel string, healthy bool) string {
	if !healthy {
		return "⚡ CodeForge — No AI provider configured.\\nPress Ctrl+K for Command Palette to setup."
	}
	
	var meta []string
	if activeModel != "" {
		meta = append(meta, activeModel)
	}
	if src, _ := KeySource(activeName); src != "" {
		meta = append(meta, src)
	}

	return fmt.Sprintf("🚀 CodeForge Ready · %s\\n%s\\n\\n💡 Tip: Press Ctrl+K for Palette, @ to attach files, or just type a request.", 
		activeName, strings.Join(meta, " · "))
}"""

content = content.replace(bad_func, good_func)

with open('internal/onboarding/welcome.go', 'w') as f:
    f.write(content)
