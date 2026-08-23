import re

with open('README.md', 'r') as f:
    content = f.read()

old_install = """## 💻 Instalasi

Kami sangat merekomendasikan kompilasi langsung dari sumber (`source`) agar Anda mendapatkan **fitur-fitur mutakhir secara *real-time*** (seperti *Multiplayer Sessions*, *Vector RAG*, *Mouse Support*, dan Arsitektur Agen Terbaru):

```bash
curl -fsSL https://raw.githubusercontent.com/NanoMindExplorer/codeforge/main/install.sh | CODEFORGE_VERSION=source sh
```

*(Catatan: Dibutuhkan Go 1.25+ di sistem Anda)*

### Instalasi Rilis Stabil (*Pre-compiled Binary*)
Jika Anda hanya menginginkan *binary* tanpa proses kompilasi:
```bash
curl -fsSL https://raw.githubusercontent.com/NanoMindExplorer/codeforge/main/install.sh | sh
```"""

new_install = """## 💻 Instalasi (One-Liner)

Pasang rilis stabil CodeForge ke sistem Anda dalam sekejap menggunakan satu baris perintah berikut. Ini akan mengunduh versi biner (binary) terbaru tanpa memerlukan kompilasi:

```bash
curl -fsSL https://raw.githubusercontent.com/NanoMindExplorer/codeforge/main/install.sh | sh
```

### Alternatif Instalasi dari Sumber (*Source*)
Jika Anda ingin berpartisipasi mencoba fitur eksperimental yang belum dirilis secara stabil, Anda dapat memaksa instalasi untuk mengompilasi dari *branch* utama secara *real-time* (membutuhkan Go 1.25+):
```bash
curl -fsSL https://raw.githubusercontent.com/NanoMindExplorer/codeforge/main/install.sh | CODEFORGE_VERSION=source sh
```"""

content = content.replace(old_install, new_install)

with open('README.md', 'w') as f:
    f.write(content)
