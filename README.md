# CodeForge 🚀

![Version](https://img.shields.io/badge/version-v1.9.3-blue)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
![Platform](https://img.shields.io/badge/platform-Linux%20|%20macOS%20|%20Termux-lightgrey)

> *Terminal AI coding companion — open, modular, vendor-neutral — and it feels like the future.*

CodeForge adalah antarmuka terminal (TUI) berbasis kecerdasan buatan (*Agentic AI*) mutakhir untuk perekayasa perangkat lunak. Dirancang dengan filosofi IDE modern dan terintegrasi langsung dengan GitHub, Git, serta penyedia LLM pilihan Anda (Gemini, Grok, Claude, OpenAI, Ollama). 

Melalui pembaruan UI/UX arsitektur terbaru, CodeForge kini sepenuhnya dikendalikan oleh **Natural Language Intent**. Tidak perlu menghafal puluhan perintah yang membingungkan. Cukup tuliskan apa yang Anda butuhkan, dan agen AI kami akan membaca, merencanakan, dan menulis kode untuk Anda.

![CodeForge Demo](./docs/assets/demo.png) *(Ilustrasi UI - Anda dapat mencoba langsung di terminal Anda)*

---

## ✨ Arsitektur UI/UX & Workflow Mutakhir (Terbaru!)

1. **Zero-Friction Onboarding (UI Interaktif):** Saat pertama kali dijalankan, Anda tidak akan disambut oleh layar kosong atau kewajiban mengetik instruksi CLI yang rumit. TUI CodeForge akan menampilkan menu navigasi visual (mendukung *mouse* & *keyboard*) untuk memilih API provider (Gemini, Grok, Ollama, dll) dengan aman.
2. **"The Golden Dozen" Slash Commands:** Mengurangi beban memori Anda (*Cognitive Overload*). Dari yang sebelumnya memiliki 50+ perintah (`/read`, `/grep`, `/theme`), kini sistem disederhanakan menjadi perintah tingkat orkestrasi seperti `/act`, `/plan`, `/gh`, dan `/session`. 
3. **Intent-Driven Workflow:** Apakah Anda ingin membaca file, mencari letak *bug*, atau mengeksekusi *script*? Cukup ucapkan secara natural di chat atau gunakan sebutan **`@nama_file`**. Agen otonom akan memanggil *tools* miliknya secara otomatis.
4. **Command Palette (Ctrl+K):** Seluruh manajemen kosmetik (Tema, Mode Vim, tata letak) kini tersentralisasi pada Command Palette bergaya VS Code.

---

## 💻 Instalasi

Kami sangat merekomendasikan kompilasi langsung dari sumber (`source`) agar Anda mendapatkan **fitur-fitur mutakhir secara *real-time*** (seperti *Multiplayer Sessions*, *Vector RAG*, *Mouse Support*, dan Arsitektur Agen Terbaru):

```bash
curl -fsSL https://raw.githubusercontent.com/NanoMindExplorer/codeforge/main/install.sh | CODEFORGE_VERSION=source sh
```

*(Catatan: Dibutuhkan Go 1.25+ di sistem Anda)*

### Instalasi Rilis Stabil (*Pre-compiled Binary*)
Jika Anda hanya menginginkan *binary* tanpa proses kompilasi:
```bash
curl -fsSL https://raw.githubusercontent.com/NanoMindExplorer/codeforge/main/install.sh | sh
```

### Termux (Android)
CodeForge bekerja dengan sangat luar biasa dan optimal di atas OS Android (Termux):
```bash
pkg install -y golang git
git clone https://github.com/NanoMindExplorer/codeforge.git
cd codeforge
bash contrib/termux/build.sh
```

---

## 🔑 Autentikasi Mudah

CodeForge dirancang untuk independen secara vendor (*Vendor-Neutral*). Anda dapat menggunakan model AI pilihan Anda.
Saat Anda menjalankan perintah `codeforge` di terminal untuk pertama kalinya, antarmuka *Setup UI* akan membimbing Anda.

Namun, jika Anda lebih suka menggunakan konfigurasi Environment Variable klasik:
```bash
# Contoh untuk Gemini (Tersedia Free-Tier)
export GEMINI_API_KEY="AIzaSy..."

# Contoh untuk Grok (xAI)
export XAI_API_KEY="xai-..."
```

*(Jangan lupa lakukan `gh auth login` agar agen CodeForge Anda bisa membuat Pull Request dan Issue di GitHub).*

---

## 🎯 Cara Penggunaan (Quick Start)

Masuk ke folder repositori proyek Anda dan panggil CodeForge:

```bash
cd /path/to/my-project
codeforge
```

### Navigasi Utama Harian
- **Ketik Natural:** *"Tolong periksa mengapa autentikasi Oauth2 saya gagal, berikan perbaikannya."*
- **Sebut Konteks:** Tekan `@` untuk menyisipkan direktori, file, atau tautan ke dalam ingatan AI.
- **`/act` (Action Mode):** Memerintahkan agen untuk langsung memodifikasi *codebase* secara otonom.
- **`/plan` (Architect Mode):** Memerintahkan agen untuk melakukan riset mendalam, membaca file di latar belakang, dan memberikan cetak biru (*blueprint*) sebelum memodifikasi kode.
- **`Ctrl+K` :** Akses kilat ke *Command Palette* (Ganti Tema, Preferensi, dll).
- **`/git` & `/gh` :** Integrasi *version control* terpusat (contoh: `/git commit Update fitur`, `/gh pr`).

---

## 📖 Dokumentasi Lanjutan (Wiki)

CodeForge dikembangkan dengan filosofi terbuka. Seluruh dokumentasi arsitektur, panduan *tools*, hingga panduan berkontribusi ada di dalam repositori ini:

- 📚 **[Dokumentasi Pembangunan (BUILD.md)](./docs/wiki/BUILD.md)** - Panduan cara merakit dan mengembangkan CodeForge dari kode sumber.
- ⚙️ **[Model Ancaman & Keamanan (THREAT_MODEL.md)](./docs/THREAT_MODEL.md)** - Penjelasan tentang sistem *Sandbox* agen.
- 🤖 **[Sistem Subagen (SUBAGENTS.md)](./docs/SUBAGENTS.md)** - Bagaimana agen CodeForge mendelegasikan tugas asinkron.

---
**Lisensi:** Apache 2.0 (Created by NanoMind — 2026)
