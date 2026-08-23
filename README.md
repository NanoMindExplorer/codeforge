# CodeForge

![Version](https://img.shields.io/badge/version-v1.9.3-blue)


> *Terminal AI coding companion — open, modular, vendor-neutral — and it feels like the future.*

CodeForge adalah antarmuka terminal (TUI) berbasis AI mutakhir untuk perekayasa perangkat lunak. Dirancang dengan filosofi IDE modern dan terintegrasi langsung dengan GitHub, Git, serta penyedia LLM pilihan Anda (Gemini, Grok, Claude, OpenAI, Ollama).

![CodeForge Demo](./docs/assets/demo.png) *(Ilustrasi UI - Anda dapat mencoba langsung di terminal Anda)*

---

## 🚀 Fitur Utama

- **Zero-Friction Workflow:** `Ctrl+K` command palette, panel sidebar, status bar cerdas, dan omnibox.
- **Agen Terisolasi (Sandboxed):** Eksekusi *tool* CLI dan modifikasi kode dengan persetujuan manusia.
- **Integrasi GitHub Native:** Tarik PR, pantau CI/Action *run*, dan buat *commit* langsung dari TUI.
- **Arsitektur Modular:** Mendukung ekstensi eksternal via MCP (Model Context Protocol).

---

## 💻 Instalasi

Anda dapat memasang CodeForge menggunakan skrip *one-liner* berikut. Kami merekomendasikan kompilasi langsung dari sumber utama (`main branch`) agar Anda mendapatkan **semua fitur mutakhir terbaru** (seperti *Multiplayer Sessions*, *Vector RAG*, dan *Mouse Support*):

```bash
curl -fsSL https://raw.githubusercontent.com/NanoMindExplorer/codeforge/main/install.sh | CODEFORGE_VERSION=source sh
```

*(Catatan: Anda membutuhkan Go 1.25+ yang terinstal di sistem Anda untuk kompilasi dari sumber)*

### Instalasi Rilis Stabil
Jika Anda hanya ingin mengunduh *binary* stabil tanpa kompilasi (namun mungkin tertinggal beberapa pembaruan):
```bash
curl -fsSL https://raw.githubusercontent.com/NanoMindExplorer/codeforge/main/install.sh | sh
```

### Alternatif Instalasi (Dari Sumber)
Pastikan Go 1.25+ terinstal:
```bash
git clone https://github.com/NanoMindExplorer/codeforge.git
cd codeforge
make build
sudo mv codeforge /usr/local/bin/
```

### Termux (Android)
```bash
pkg install -y golang git
git clone https://github.com/NanoMindExplorer/codeforge.git
cd codeforge
bash contrib/termux/build.sh
```

Verifikasi instalasi berhasil:
```bash
codeforge version
```

---

## 🔑 Konfigurasi API & GitHub

Agar CodeForge dapat berfungsi, Anda membutuhkan **salah satu** API Key LLM.

**1. Set API Key (Contoh menggunakan Gemini yang memiliki Free Tier):**
```bash
export GEMINI_API_KEY="AIzaSy..."
```
*(Opsional: `XAI_API_KEY`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, atau biarkan kosong jika memakai `ollama` lokal).*

**2. Autentikasi GitHub (Opsional tapi sangat direkomendasikan):**
```bash
gh auth login
```

---

## 🎯 Cara Penggunaan (Quick Start)

Masuk ke folder proyek/repositori Git Anda dan jalankan CodeForge:

```bash
cd /path/to/my-project
codeforge
```

### Siklus Kerja Harian
Di dalam antarmuka CodeForge, Anda cukup **mengetikkan apa yang Anda inginkan** di *prompt*. Beberapa perintah (*slash commands*) penting yang akan sering Anda pakai:

- `/act [tugas]` : Menginstruksikan AI untuk membaca *codebase*, menyusun rencana perbaikan, dan langsung mengeksekusi perubahan file. 
  - *Contoh: `/act perbaiki fungsi register agar menggunakan regex validasi email.`*
- `@` : Tekan '@' untuk memanggil jendela *file picker* interaktif (menyisipkan file ke memori AI).
- `/gh pr` : Mengelola / membuat Pull Request ke GitHub.
- `Ctrl+K` : Membuka *Command Palette* bergaya VS Code untuk fitur tingkat lanjut.
- `/mode [build|design|yolo]` : Beralih dari mode eksekusi aman (*build* - menunda file sampai Anda me-review) ke mode bebas hambatan (*yolo*).

---

## 📖 Dokumentasi Lengkap & Wiki

Untuk referensi lanjutan, silakan kunjungi [CodeForge Wiki](./docs/wiki/) yang berisi:

- 📚 **[User Guide & GitHub Integration](./docs/wiki/User-Guide.md)** - Penjelasan komprehensif seluruh antarmuka dan integrasi.
- ⚙️ **[Reference & Configuration](./docs/wiki/Reference.md)** - Daftar *keybindings* (pintasan keyboard), *slash commands*, CLI *flags*, dan variabel *environment*.
- 🛠 **[Development, Architecture & Troubleshooting](./docs/wiki/Development.md)** - Arsitektur, kontribusi kode, penyelesaian masalah, dan distribusi sistem.

---
**Lisensi:** Apache 2.0 (Created by NanoMind — 2026)
