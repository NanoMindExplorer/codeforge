import re

with open('internal/tui/onboard.go', 'r') as f:
    content = f.read()

# Fix the ascii art string
bad_ascii = """	asciiArt := ` ____          _      _____                    
|  _ \        | |    |  ___|                   
| |_) | _   _ | |__  | |_  ___  _ __  __ _  ___ 
|  _ < | | | || '_ \ |  _|/ _ \| '__|/ _` |/ _ \
| |_) || |_| || |_) || | | (_) | |  | (_| |  __/
|____/  \__,_||_.__/ \_|  \___/|_|   \__, |\___|
                                      __/ |     
                                     |___/      `"""

good_ascii = '	asciiArt := " ____          _      _____\\n" +\n' + \
'		"    / ___|___   __| | ___|  ___|__  _ __ __ _  ___\\n" +\n' + \
'		"   | |   / _ \\\\ / _` |/ _ \\\\ |_ / _ \\\\| \\'__/ _` |/ _ \\\\\\n" +\n' + \
'		"   | |__| (_) | (_| |  __/  _| (_) | | | (_| |  __/\\n" +\n' + \
'		"    \\\\____\\\\___/ \\\\__,_|\\\\___|_|  \\\\___/|_|  \\\\__, |\\\\___|\\n" +\n' + \
'		"                                        |___/"'

content = content.replace(bad_ascii, good_ascii)

with open('internal/tui/onboard.go', 'w') as f:
    f.write(content)
