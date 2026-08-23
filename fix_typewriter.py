import re

with open('internal/tui/chat.go', 'r') as f:
    content = f.read()

bad_block = """		if c.typewriterOn && len(c.typewriterQ) > 0 {
			for i := 0; i < 2 && len(c.typewriterQ) > 0; i++ {
				c.store.AddSystem(c.typewriterQ[0])
				c.typewriterQ = c.typewriterQ[1:]
			}
			if len(c.typewriterQ) == 0 {
				c.typewriterOn = false
			}
		}"""

content = content.replace(bad_block, "")

with open('internal/tui/chat.go', 'w') as f:
    f.write(content)
