# Dokumentasi Pembangunan (Build & Development) CodeForge

Dokumen ini ditujukan bagi para pengembang (developer) dan kontributor yang ingin mengompilasi CodeForge dari kode sumber, serta memahami siklus pengembangan (development lifecycle) dalam repositori ini.

## Persyaratan Sistem (Prerequisites)

Untuk membangun CodeForge, sistem Anda membutuhkan perangkat berikut:
- **Go (Golang)** versi `1.25` atau yang lebih baru.
- **Git** untuk manajemen repositori.
- **Make** (Opsional, namun direkomendasikan untuk menjalankan skrip build otomatis).

## 1. Kompilasi Lokal (Local Build)

### Langkah Cepat (Quick Start)
Cara paling mudah untuk membangun dan menguji CodeForge di lingkungan lokal:

```bash
# 1. Kloning repositori
git clone https://github.com/NanoMindExplorer/codeforge.git
cd codeforge

# 2. Unduh dependensi
go mod download

# 3. Kompilasi Binari
go build -o codeforge ./cmd/codeforge

# 4. Instalasi ke sistem (Opsional)
sudo mv codeforge /usr/local/bin/
```

### Menggunakan Makefile
Proyek ini dilengkapi dengan `Makefile` untuk mempermudah rutinitas pengembangan:

- `make build` : Menghasilkan binari `codeforge` di dalam folder proyek.
- `make fmt` : Merapikan format kode sumber menggunakan `gofmt`.
- `make test` : Menjalankan seluruh pengujian unit (*unit tests*) untuk memastikan tidak ada fitur yang rusak.
- `make install-hooks` : Memasang Git hooks lokal untuk mencegah komit yang tidak lolos *formatting*.

## 2. Pengujian (Testing) & Jaminan Mutu (QA)

CodeForge menggunakan sistem GitHub Actions CI yang sangat ketat. Sebelum Anda membuat *Pull Request* (PR), pastikan Anda melakukan hal berikut:

1. **Uji Fungsionalitas:**
   ```bash
   go test ./...
   ```
2. **Uji Kondisi Balapan (Race Condition):**
   *(Sangat penting karena CodeForge memiliki arsitektur subagen dan tugas latar belakang asinkron yang kompleks)*
   ```bash
   go test -race ./...
   ```
3. **Pengecekan Gaya Kode (Linting):**
   Pastikan tidak ada teguran dari `gofmt`.
   ```bash
   gofmt -w .
   ```

## 3. Kompilasi Khusus Lingkungan

### Termux (Android)
CodeForge dirancang secara natif agar sangat ramah bagi perangkat *mobile* (Android/Termux). Terdapat skrip kompilasi khusus yang dirancang untuk mengatasi limitasi *linker* pada sistem file Android:

```bash
bash contrib/termux/build.sh
```

### Skrip Instalasi Pihak Ketiga (One-Liner)
Sistem penyebaran otomatis CodeForge mengandalkan skrip `install.sh` di *root* direktori. 
Anda dapat menguji perombakan instalasi lokal dengan *flag* simulasi:

```bash
cat install.sh | CODEFORGE_VERSION=source sh
```
Ini akan secara dinamis menyalin dan mengompilasi dari *branch* lokal Anda.

## 4. Arsitektur Mutakhir CodeForge (UI/UX)
Harap perhatikan konvensi desain yang wajib dipatuhi:

### 4.1. Filosofi TUI & Desain Visual
- **Desain Minimalis & Mutakhir:** CodeForge mengadopsi antarmuka estetika generasi baru (mirip Cursor/Zed). Jangan gunakan garis tebal atau blok kotak *full-width* untuk pesan pengguna. Gunakan indentasi transparan, *badge* teks untuk status (seperti `● You`, `✧ CodeForge`, `◇ Tool`), dan komponen *pill-style floating input*.
- **Natural Language Intent:** Jangan menambah *Slash Command* baru kecuali untuk orkestrasi arsitektur tinggi. Ikuti filosofi **"The Golden Dozen"**. Fungsionalitas agen harus berbasis *Natural Language Intent*.

### 4.2. Performa & Manajemen Latensi (Zero-Latency Rendering)
TUI BubbleTea di CodeForge dibangun dengan sistem rendering tingkat tinggi untuk memastikan **latensi 0 (nol)** saat *scrolling* menggunakan *mouse* atau saat mengetik cepat:
- **Jangan memanggil `lipgloss.Width()` atau `lipgloss.NewStyle().Render()` di dalam iterasi *hot-loop*** (seperti perulangan baris *scrollbar* atau pembuatan *frame*). Kalkulasi ANSI ini sangat membebani CPU. Gunakan gaya (*styles*) yang telah dikompilasi sebelumnya (*pre-compiled*).
- **Sistem `CachedVisual` & `CachedBody`:** Seluruh blok *markdown*, penyorotan sintaks (*syntax highlighting*), dan padding spasi harus **dikalkulasi sekali di awal dan disimpan ke dalam *cache*** (`internal/tui/blocks/render.go`). Jangan pernah me-render blok berulang kali pada *frame* yang sama atau mengabaikan *cache* untuk blok yang *collapsed* (tersembunyi).
- **Asynchronous Non-Blocking:** Pastikan interaksi bebas dari pemblokiran I/O (*I/O Blocking*). Selalu gunakan fungsi asinkron (seperti `internal/bgtask`) untuk menghindari pembekuan UI (*freeze*).

### 4.3. Keamanan & Izin
- Jika ada *tool* atau izin akses (*permissions*) baru yang diimplementasikan, Anda **wajib** memperbarui perlindungan dan model kebijakan di `internal/sandbox`.
