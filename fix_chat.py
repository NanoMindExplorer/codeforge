import re

with open('internal/tui/chat.go', 'r') as f:
    content = f.read()

bad_func = """func (c *ChatModel) AddSystemMessage(text string) {
	lines := strings.Split(text, "\\n")
	// Typewriter only for small multi-line messages (≤8 lines); larger = instant.
	if theme.MotionEnabled() && len(lines) > 2 && len(lines) <= 8 {
		c.typewriterQ = append(c.typewriterQ, lines...)
		c.typewriterOn = true
		return
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		c.store.AddSystem(line)
	}
}"""

good_func = """func (c *ChatModel) AddSystemMessage(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	c.store.AddSystem(text)
}"""

content = content.replace(bad_func, good_func)

# We also need to remove the typewriterUpdate block from Update if it exists, or just leave the fields unused.
with open('internal/tui/chat.go', 'w') as f:
    f.write(content)
