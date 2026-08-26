# Panduan Build APK Mobile App EWA Kopkara (100% Linux CLI - Hemat RAM 8GB)

Dokumen ini berisi panduan pembuatan file **.APK Standalone** untuk HP Android **100% via Linux Command Line (Ubuntu / Debian / WSL CLI)** tanpa perlu membuka Android Studio GUI, sehingga sangat ringan dan hemat RAM (laptop 8GB RAM).

---

## ⚡ Ringkasan Perintah Build Rutin Sehari-hari (Daily Build)

Jika Anda sudah pernah melakukan setup awal dan ingin membuat file `.apk` baru setelah melakukan perubahan kode UI/React, Anda **hanya perlu menjalankan 3 langkah ringkas berikut**:

```bash
cd /mnt/d/Data_NK/Project5/LMS/frontend

# 1. Build aset web React terbaru
./node_modules/.bin/vite build

# 2. Sinkronkan aset terbaru ke folder android
npx cap sync android

# 3. Pindah folder & kompilasi file APK baru
cd android
./gradlew assembleDebug
```

📍 **Path File APK Hasil Build**:  
`/mnt/d/Data_NK/Project5/LMS/frontend/android/app/build/outputs/apk/debug/app-debug.apk`

---

## 🛡️ Konfigurasi Windows Firewall (Wajib Diset 1x di Laptop Windows Server)

Agar HP Android Anda di jaringan Wi-Fi lokal dapat terhubung ke server Laptop Anda, jalankan perintah PowerShell ini **sekali saja** di PowerShell (Run as Administrator):

```powershell
New-NetFirewallRule -DisplayName "LMS Kopkara-EWA (Ports 8086 & 3005)" -Direction Inbound -LocalPort 8086,3005 -Protocol TCP -Action Allow
```

---

## 📋 Alur Lengkap Setup & Pembeda Frekuensi Eksekusi Perintah

| Langkah | Perintah | Frekuensi Eksekusi | Keterangan |
|---|---|---|---|
| **Firewall** | `New-NetFirewallRule ... -Action Allow` | **1x Saja (Admin Windows)** | Membuka Port Inbound 8086 & 3005 di Windows Firewall |
| **Langkah 1** | `npm install @capacitor/core @capacitor/cli @capacitor/android` | **1x Saja (Setup Awal)** | Meng-install dependensi library Capacitor |
| **Langkah 2** | `npx cap init "Kopkara Mobile EWA" "com.kopkara.ewa.app" --web-dir dist` | **1x Saja (Setup Awal)** | Inisialisasi Nama Aplikasi & Package ID |
| **Langkah 3** | `./node_modules/.bin/vite build` | **Setiap Build Baru** | Memperbarui berkas React HTML/JS/CSS di `dist` |
| **Langkah 4** | `npx cap add android` | **1x Saja (Setup Awal)** | Generasi folder native proyek Android (`/android`) |
| **Langkah 5** | `npx cap sync android` | **Setiap Build Baru** | Menyalin aset `dist` terbaru ke folder Android |
| **Kompilasi** | `cd android && ./gradlew assembleDebug` | **Setiap Build Baru** | Kompilasi file **`app-debug.apk`** standalone |

---

## 🛠️ Detail Langkah Setup Pertama Kali (First-Time Setup Only)

Buka terminal Linux Anda dan navigasi ke direktori `frontend`:

```bash
cd /mnt/d/Data_NK/Project5/LMS/frontend

# 1. Install dependensi Capacitor
npm install @capacitor/core @capacitor/cli @capacitor/android

# 2. Inisialisasi Proyek Mobile App (Hanya 1x)
npx cap init "Kopkara Mobile EWA" "com.kopkara.ewa.app" --web-dir dist

# 3. Build Aset Frontend React
./node_modules/.bin/vite build

# 4. Tambahkan Platform Android (Hanya 1x)
npx cap add android

# 5. Sinkronkan Aset Web ke Proyek Android
npx cap sync android

# 6. Kompilasi APK via Gradle CLI
cd android
chmod +x gradlew
./gradlew assembleDebug
```

---

## 📲 Pengujian & Konfigurasi Server di HP Android

1. Kirimkan file **`app-debug.apk`** ke HP Android Anda (via WhatsApp, Google Drive, atau USB).
2. Install file `.apk` tersebut di HP Android.
3. Buka aplikasi **Kopkara Mobile EWA** di HP Android Anda.
4. Klik ikon **⚙️ Pengaturan Server API** pada pojok kanan atas aplikasi:
   - Pilih **📶 Wi-Fi Lokal**: Masukkan IP laptop/Linux Anda (misal `192.168.0.103` : `8086`).
   - Atau pilih **🌐 Internet**: Masukkan URL server VPS / Ngrok Anda.
   - Klik **Tes Koneksi** $\rightarrow$ **Simpan & Terapkan**.
5. Aplikasi Mobile EWA di HP Anda kini telah terhubung **100% Real** dengan Backend LMS!
