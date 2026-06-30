# Install CodeForge TUI

## Termux (Android)

    pkg install -y golang git
    git clone https://github.com/NanoMindExplorer/codeforge-tui.git
    cd codeforge-tui
    go mod tidy
    CGO_ENABLED=0 go build -ldflags="-s -w" -o codeforge ./cmd/codeforge/
    cp codeforge $PREFIX/bin/

    echo "export GEMINI_API_KEY=AIzaSy..." >> ~/.bashrc
    source ~/.bashrc

    codeforge

## Ubuntu/Debian

    sudo apt install -y golang-go git
    git clone https://github.com/NanoMindExplorer/codeforge-tui.git
    cd codeforge-tui
    go mod tidy
    go build -ldflags="-s -w" -o codeforge ./cmd/codeforge/
    sudo mv codeforge /usr/local/bin/

    export GEMINI_API_KEY="AIzaSy..."
    codeforge

## Get Free Gemini API Key

1. Visit https://aistudio.google.com/apikey
2. Sign in with Google
3. Click Create API Key
4. Copy API key (format: AIzaSy...)

Free tier: 250 requests/day (gemini-2.5-flash)

## Key Bindings

| Key | Action |
|-----|--------|
| i | INSERT mode |
| Esc | NORMAL mode |
| : | Command palette |
| / | Slash commands |
| Enter | Send message |
| q | Quit |

## Slash Commands

- /help - Show help
- /about - NanoMind info
- /cost - Token stats
- /commit - Auto-commit
- /quit - Exit

Created by NanoMind - 2026
