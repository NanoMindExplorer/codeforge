import re

with open('internal/tui/blocks/block.go', 'r') as f:
    content = f.read()
content = content.replace("CachedBody  string\\n\\tCachedVisual      []string\\n\\tCachedVisualWidth int\\n\\tCachedVisualSel   bool", "CachedBody  string")
content = content.replace("CachedBody  string", "CachedBody  string\\n\\tCachedVisual      []string\\n\\tCachedVisualWidth int\\n\\tCachedVisualSel   bool")
with open('internal/tui/blocks/block.go', 'w') as f:
    f.write(content)

