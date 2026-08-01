# Panduan Mengunggah Proyek LIMS ke GitHub

Dokumen ini berisi panduan langkah demi langkah untuk mengamankan *source code* aplikasi LIMS Anda ke platform **GitHub**.

> [!IMPORTANT]
> Sangat penting untuk **TIDAK** mengunggah file yang berisi kata sandi, *secret key*, atau konfigurasi kredensial produksi (seperti file `.env` produksi) ke GitHub publik.

## 1. Persiapan File `.gitignore`

Sebelum melakukan inisialisasi Git, kita harus memastikan file dan direktori yang tidak penting (atau rahasia) tidak ikut terunggah. Buatlah file bernama `.gitignore` di direktori utama (`d:\Data_NK\Project5\AI\LIM_System_Linux_OK`) dan pastikan isinya mencakup hal-hal berikut:

```text
# Node & Frontend
node_modules/
dist/
build/
.npm/

# Backend (Go)
/backend/main
/backend/*.exe
/backend/tmp/

# Environment & Secrets
.env
.env.production
.env.local

# IDE & OS Files
.vscode/
.idea/
.DS_Store
Thumbs.db
```

## 2. Inisialisasi Repositori Git Lokal

Buka Terminal/PowerShell Anda dan arahkan ke direktori utama proyek LIMS, lalu jalankan perintah berikut secara berurutan:

```bash
# 1. Pindah ke direktori proyek (jika belum)
cd /mnt/d/Data_NK/Project5/AI/LIM_System_Linux_OK

# 2. Inisialisasi Git
git init

# 3. Tambahkan semua file ke dalam "staging area" (kecuali yang ada di .gitignore)
git add .

# 4. Simpan perubahan pertama Anda (Commit)
git commit -m "Initial commit: LIMS System backend and frontend"

# 5. Ubah nama branch utama menjadi 'main' (standar modern)
git branch -M main
```

## 3. Membuat Repositori di GitHub

1. Buka browser dan login ke akun [GitHub](https://github.com/).
2. Klik tombol **New** (atau ikon **+** di pojok kanan atas -> **New repository**).
3. Isi kolom **Repository name** (misalnya: `lims-system`).
4. Atur visibilitas ke **Private** (Sangat disarankan, karena ini kode aplikasi berpemilik/perusahaan).
5. **JANGAN** centang "Add a README file" atau "Add .gitignore" (karena kita sudah punya kode lokal).
6. Klik tombol **Create repository**.

## 4. Menghubungkan Lokal ke GitHub

Setelah repositori dibuat, GitHub akan menampilkan halaman panduan. Salin perintah pada bagian *"…or push an existing repository from the command line"* dan jalankan di terminal Anda.

Biasanya perintahnya terlihat seperti ini:

```bash
# Tambahkan URL GitHub sebagai "remote origin"
git remote add origin https://github.com/UsernameAnda/lims-system.git

# Unggah kode Anda ke GitHub
git push -u origin main
```

> [!TIP]
> Jika Anda belum pernah login Git di terminal, GitHub mungkin akan meminta Anda untuk memasukkan *Username* dan *Personal Access Token* (atau otentikasi melalui browser).

## 5. Sinkronisasi Perubahan (Pengganti Rsync)

Banyak pengguna terbiasa menggunakan `rsync` atau FTP/SFTP untuk memperbarui kode dari lokal ke server. Namun, saat menggunakan GitHub, kita menggunakan mekanisme **Version Control** bawaan Git.

Alurnya berubah dari:
`Laptop (rsync) ---> VPS`
Menjadi:
`Laptop (git push) ---> GitHub ---> VPS (git pull)`

Setiap kali Anda selesai melakukan perubahan atau menambahkan fitur baru pada LIMS di komputer lokal Anda, lakukan 3 langkah wajib ini untuk "mensinkronisasi" (mengunggah) perubahan tersebut ke GitHub:

```bash
# 1. Tambahkan SEMUA file yang baru diubah agar siap disimpan
git add .

# 2. Bungkus perubahan tersebut dan beri "Catatan/Pesan" (Wajib)
# Pesan ini akan muncul di GitHub, jadi buatlah sejelas mungkin.
git commit -m "feat: Menambahkan menu Laporan Keuangan dan Tombol Tutup"

# 3. Sinkronisasikan (Unggah) bungkusan tersebut ke GitHub
git push
```

> [!TIP]
> Jika Anda mengalami *error* saat `git push`, biasanya karena ada perubahan di GitHub yang belum ada di laptop Anda. Cukup jalankan `git pull` terlebih dahulu untuk menyamakan data, lalu ulangi `git push`.

### Mengatasi Error GH013: GitHub Push Protection (Rahasia Bocor)
Jika saat melakukan `git push` Anda mendapatkan *error* berwarna merah dengan pesan **`GH013: Repository rule violations found`** atau **`Push cannot contain secrets`**, ini berarti fitur keamanan otomatis GitHub memblokir proses unggah Anda karena mendeteksi adanya *file* rahasia (seperti `.env`, API Key, atau password) yang tidak sengaja terbawa di dalam *commit* Anda.

**Cara Memperbaikinya:**
1. Keluarkan *file* yang terdeteksi dari Git tanpa menghapus *file* aslinya dari komputer Anda:
   ```bash
   git rm --cached "path/ke/file/rahasia" 
   # (Ganti dengan path file yang ditunjukkan pada log error GitHub, contoh: "backend/.env")
   ```
2. Perbarui *commit* terakhir Anda agar bersih dari *file* tersebut:
   ```bash
   git commit --amend --no-edit
   ```
3. Lakukan *push* kembali:
   ```bash
   git push
   ```
> [!IMPORTANT]
> Pastikan *file* rahasia tersebut (misalnya `.env`) sudah terdaftar di dalam `.gitignore` agar kejadian ini tidak terulang di masa depan.

## 6. Mengambil Perubahan dari GitHub ke VPS (Opsional)
Jika VPS Anda sudah disiapkan dengan Git, Anda tidak perlu lagi menggunakan `rsync` untuk memindahkan *source code*. Cukup masuk ke VPS melalui SSH, arahkan ke folder proyek, dan jalankan:
```bash
git pull origin main
```
Ini akan otomatis menarik perubahan terbaru dari GitHub ke VPS Anda secara efisien.

## 7. Melihat Riwayat Perubahan (History/Log)

Salah satu keunggulan utama Git adalah kemampuannya melacak sejarah. Untuk melihat siapa yang mengubah kode, kapan, dan apa pesan perubahannya:

**Melalui Terminal/Command Line:**
```bash
# Melihat log lengkap (tekan tombol 'q' untuk keluar dari log)
git log

# Melihat log versi ringkas (hanya menampilkan ID Unik dan pesan)
git log --oneline

# Jika Anda sedang berada di versi masa lalu dan ingin melihat SELURUH log (termasuk masa depan)
git log --all --oneline --graph
```

**Melalui Website GitHub (Paling Mudah):**
Cukup buka halaman *repository* Anda di GitHub, lalu klik tulisan **"Commits"** (biasanya di pojok kanan atas daftar file, ada tulisan seperti *"45 Commits"*). Di sana Anda bisa melihat seluruh sejarah perubahan secara visual, bahkan melihat baris kode mana saja yang ditambah/dihapus (berwarna hijau/merah).

## 8. Menarik/Mengembalikan ke Versi Tertentu (Rollback)

Jika Anda melakukan kesalahan pada pembaruan (push) hari ini dan ingin mengembalikan VPS atau lokal Anda ke versi *kemarin* atau versi tertentu, ikuti langkah ini:

1. Cari ID Unik (*Commit Hash*) dari versi yang ingin Anda tuju menggunakan `git log --oneline` (atau lihat ID-nya di web GitHub). ID ini berupa 7 karakter acak, contoh: `a1b2c3d`.
2. Gunakan perintah berikut di terminal (contoh kita ingin kembali ke versi `a1b2c3d`):

```bash
# Cara 1: Sekadar "melihat-lihat" versi lama tanpa menghapus versi baru
git checkout a1b2c3d
```

> [!WARNING]
> **PENTING TENTANG `git checkout`:**
> Ketika Anda menjalankan perintah di atas, *source code* di folder Anda (di VSCode/Notepad) **akan benar-benar berubah tertimpa menjadi kode versi lama saat itu**. Git memasuki mode *'detached HEAD'*.
> Namun jangan panik! Kode baru Anda tidak hilang. Untuk mengembalikan *source code* Anda menjadi versi terbaru (masa kini) kembali, cukup ketik:
> `git checkout main`

```bash
# Cara 2: MENGEMBALIKAN SECARA PERMANEN (Hati-hati!)
# Ini akan menghapus semua perubahan setelah versi a1b2c3d dan memaksa kode kembali seperti saat itu
git reset --hard a1b2c3d
```

> [!CAUTION]
> Perintah `git reset --hard` bersifat destruktif. Jika Anda belum yakin, sangat disarankan menggunakan `git checkout [id_commit]` (Cara 1) untuk melihat-lihat terlebih dahulu.

## Referensi Tambahan
- Jika Anda ingin Github Actions otomatis melakukan *build* atau *deploy* ke VPS setiap kali Anda melakukan `git push`, kita bisa membuat skrip CI/CD di `.github/workflows/deploy.yml`. Beri tahu saya jika Anda membutuhkannya!
