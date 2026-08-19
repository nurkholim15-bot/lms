# Proposal Stress & Load Testing
## System: Loan Management System (LMS) — Koperasi Kopkara

---

## Document Control

| Information | Detail |
|---|---|
| **Judul Proyek** | Proposal Stress & Load Testing Loan Management System (LMS) |
| **Klien / Organisasi** | Koperasi Karyawan Kopkara (PT Adira Dinamika Multi Finance Tbk) |
| **Sistem Target** | Backend Go (Gin Framework) + PostgreSQL DB + React SPA |
| **Versi Dokumen** | 1.0 |
| **Tanggal** | Agustus 2026 |
| **Panduan Eksekusi Detil** | [`load_test_execution_guide.md`](file:///d:/Data_NK/Project5/LMS/load_test_execution_guide.md) |
| **Status** | Proposed |

---

## Executive Summary

Sistem **Loan Management System (LMS) Koperasi Kopkara** merupakan platform inti yang menangani pengajuan pinjaman (*loan origination*), kalkulasi angsuran, persetujuan komite kredit, hingga proses krusial **Billing & Rekonsiliasi Payroll Adira** yang memproses ribuan data transaksi anggota setiap bulannya. 

Seiring berkembangnya jumlah anggota dan volume pengajuan pinjaman serta pemrosesan file tagihan (*bulk billing batch*), sistem dituntut memiliki ketahanan (*resilience*), kecepatan respon (*low latency*), dan stabilitas tinggi di bawah tekanan beban puncak (*peak load*).

Proposal ini menyajikan rencana menyeluruh untuk pelaksanaan **Stress & Load Testing** pada LMS Kopkara. Pengujian ini bertujuan untuk mengukur performa baseline, mengidentifikasi *bottleneck* (pada level aplikasi Go, query PostgreSQL, in-memory cache, maupun infrastruktur server), serta memastikan sistem siap menangani lonjakan lalu lintas data secara efisien tanpa mengalami *downtime*, *data corruption*, atau kemacetan kalkulasi finansial.

---

## 1. Latar Belakang & Tujuan

### 1.1 Latar Belakang
LMS Kopkara mengintegrasikan alur bisnis keuangan yang kritikal. Lonjakan lalu lintas (*traffic spike*) dapat terjadi pada kondisi tertentu, seperti:
1. **Periode Pembukaan Pengajuan Pinjaman (Peak Loan Submission)**: Lonjakan pengguna simultan yang mengakses aplikasi untuk kalkulasi simulasi dan pengajuan pinjaman secara bersamaan.
2. **Periode Cut-Off Payroll Billing (Bulk Import/Export)**: Pemrosesan berkas CSV tagihan gaji (*Adira Payroll Karisma Integration*) berisi puluhan ribu baris data yang memerlukan kalkulasi angsuran, update status angsuran, dan pembuatan laporan rekonsiliasi.
3. **Pencarian & Pelaporan Real-Time**: Akses laporan tunggakan (NPL), portofolio pinjaman, dan audit trail oleh tim pengurus/komite kredit pada jam kerja.

Tanpa pengujian beban yang terstruktur, risiko seperti *database connection pool exhaustion*, *goroutine leak*, keterlambatan *garbage collection* (GC pause Go), atau query SQL yang lambat (*unindexed full table scan*) dapat mengganggu operasional bisnis koperasi.

### 1.2 Tujuan Pengujian
1. **Mengukur Performa Baseline**: Menentukan batas kemampuan (*capacity limits*) sistem dalam kondisi operasional normal (RPS/TPS, Latency p95/p99, CPU/Memory utilization).
2. **Menemukan Breaking Point (Stress Testing)**: Menguji batas maksimal beban di mana sistem mulai mengalami degradasi performa atau *failure*.
3. **Validasi Pemrosesan Bulk Billing**: Memastikan fitur import/export billing payroll sanggup memproses data skala besar tanpa *timeout* HTTP 504 atau *out-of-memory* (OOM).
4. **Deteksi Bottleneck & Memory Leak**: Mengidentifikasi kemacetan pada koneksi database PostgreSQL, penggunaan in-memory cache (`ProductCache`, `ParameterCache`), serta *goroutine leak* dalam runtime Go.
5. **Rekomendasi Optimization & Capacity Planning**: Memberikan masukan teknis berupa rekomendasi tuning database, ukuran connection pool, konfigurasi Nginx/Reverse Proxy, dan spesifikasi server yang optimal.

---

## 2. Ruang Lingkup Pengujian (Scope of Testing)

### 2.1 Jenis Pengujian

```
                                  LOAD / STRESS TEST STRATEGY
                                 
    [ Load Testing ]            [ Stress Testing ]           [ Spike Testing ]         [ Endurance Test ]
   500 - 1,500 VUs             2,000 - 5,000 VUs            Sudden Jump to 3k VUs        Constant 800 VUs
 (Normal/Peak Operating)       (Find Breaking Point)       (Batch Billing/Submit)       (4 - 12 Hours Test)
```

1. **Load Testing (Pengujian Beban Standar & Puncak)**
   - Menguji sistem dengan beban pengguna simultan (*Virtual Users / VUs*) yang diperkirakan pada jam sibuk (500 - 1,500 VUs) selama 30–60 menit.
   - Mengukur tingkat keberhasilan transaksi, throughput (RPS), dan waktu respon.

2. **Stress Testing (Pengujian Tekanan Batas Maksimal)**
   - Meningkatkan jumlah VUs secara bertahap hingga mencapai 2,000 - 5,000+ VUs untuk mencari *breaking point* sistem.
   - Mengamati mekanisme kecelakaan (*graceful degradation*), penanganan error (HTTP 500/503), dan waktu pemulihan (*recovery time*).

3. **Spike Testing (Pengujian Lonjakan Mendadak)**
   - Menguji lonjakan pengguna dari kondisi idle ke beban ekstrem dalam durasi singkat (misal: 100 VUs melompat ke 3,000 VUs dalam 10 detik).
   - Mensimulasikan kondisi di mana ribuan karyawan mengakses sistem bersamaan saat ada pengumuman program pinjaman baru.

4. **Endurance / Soak Testing (Pengujian Ketahanan Durasi Panjang)**
   - Mengoperasikan sistem pada beban 70–80% kapasitas maksimum selama 4 hingga 12 jam.
   - Mengidentifikasi *memory leak*, kebocoran *goroutine*, akumulasi koneksi DB gantung, atau penurunan performa bertahap.

5. **Bulk Data & Batch Processing Test (Billing Integration)**
   - Menguji performa API pengunggahan dan eksekusi parsing berkas CSV payroll Adira (10,000 - 50,000 record).

---

## 3. Skenario & User Journey Pengujian

Pengujian akan menguji endpoint REST API Go Backend (`port 8086`) dan database PostgreSQL (`lms_sch`). Skenario dibagi menjadi 5 *test suite* utama:

```
+-----------------------------------------------------------------------------------+
|                            TEST SUITE ARCHITECTURE                                |
+-----------------------------------------------------------------------------------+
| 1. Auth & Session Suite      | POST /api/v1/auth/login, Refresh Token            |
| 2. Loan Origination Suite    | GET Products, POST Calculator, POST Applications  |
| 3. Approval & Disburse Suite | POST /api/v1/loans/:id/approve, /disburse        |
| 4. Billing Import/Export     | POST /api/v1/billing/import, GET /billing/export  |
| 5. Reporting & Query Suite   | GET /api/v1/reports/npl, /loans/active, /members  |
+-----------------------------------------------------------------------------------+
```

### Skenario 1: Authentication & User Session (Bobot Load: 25%)
- **Alur**: User Login -> Menerima JWT Token -> Access Authorized Profile.
- **Endpoint**:
  - `POST /api/v1/auth/login`
  - `GET /api/v1/auth/me`
- **Target**: Menguji throughput enkripsi password (Bcrypt/Argon2) dan verifikasi token JWT.

### Skenario 2: Simulasi & Pengajuan Pinjaman / Loan Origination (Bobot Load: 40%)
- **Alur**: Anggota membuka daftar produk pinjaman -> melakukan simulasi angsuran (formula interest & admin fee via `govaluate`) -> submit form pengajuan pinjaman + upload file pendukung.
- **Endpoint**:
  - `GET /api/v1/loan-products` (Cache hit vs DB hit)
  - `POST /api/v1/loans/calculate` (Kalkulasi skema bunga/angsuran)
  - `POST /api/v1/loans/submit` (Insert transaksi ke `lms_sch.loan_applications` & `loan_schedules`)
- **Target**: Menguji kecepatan kalkulasi formula, performa in-memory cache (`ProductCache`), dan transaksi DB.

### Skenario 3: Approval & Pencairan Pinjaman (Bobot Load: 10%)
- **Alur**: Admin / Komite Kredit melakukan tinjauan pengajuan -> Approve / Reject -> Generate pencairan pinjaman.
- **Endpoint**:
  - `GET /api/v1/loans/pending`
  - `POST /api/v1/loans/:id/approve`
  - `POST /api/v1/loans/:id/disburse`
- **Target**: Menguji integritas transaksi database (ACID), pembaruan status angsuran, dan pembentukan sisa pokok pinjaman.

### Skenario 4: Pemrosesan Bulk Payroll Billing Adira (Bobot Load: 15%)
- **Alur**: Upload CSV Payroll (Adira Incoming Result) -> Parse & Matching NIK Anggota -> Allocating Potongan -> Update Status Schedule Angsuran.
- **Endpoint / Script**:
  - `POST /api/v1/billing/upload-incoming`
  - `POST /api/v1/billing/process-allocation`
  - `GET /api/v1/billing/export-outgoing`
- **Target**: Menguji performa *batch processing*, penggunaan RAM saat handling CSV besar, serta efisiensi query `UPSERT` / `BULK INSERT` GORM.

### Skenario 5: Pelaporan & Master Data Search (Bobot Load: 10%)
- **Alur**: Admin membuka dashboard -> melakukan filter data anggota -> pencarian pinjaman aktif -> export ringkasan bulanan.
- **Endpoint**:
  - `GET /api/v1/members?search=...&page=1`
  - `GET /api/v1/reports/active-loans`
  - `GET /api/v1/reports/npl-summary`
- **Target**: Menguji indexing PostgreSQL pada tabel berukuran besar (`members`, `loans`, `repayments`).

---

## 4. Key Performance Indicators (KPI) & Target Performa

Tabel berikut menentukan batas toleransi keberhasilan pengujian performa:

| Parameter Metrik | Target Normal Load | Target Peak Load | Tolerance Limit (Stress) |
|---|---|---|---|
| **Response Time (p95)** | `< 150 ms` | `< 300 ms` | `< 1,000 ms` |
| **Response Time (p99)** | `< 300 ms` | `< 600 ms` | `< 2,000 ms` |
| **Throughput (RPS)** | `> 1,000 RPS` | `> 2,500 RPS` | Depends on breaking point |
| **Error Rate (% Fail)** | `0.00 %` | `< 0.10 %` | `< 1.00 %` |
| **CPU Utilization (App)** | `< 60 %` | `< 80 %` | `< 90 %` |
| **Memory Utilization (App)** | Stable (No Leak) | Stable | No OOM Crash |
| **PostgreSQL DB CPU** | `< 50 %` | `< 75 %` | `< 85 %` |
| **DB Connection Pool** | `< 70% active` | `< 90% active` | No connection timeout |
| **Bulk Billing Upload (10k rows)** | `< 3 detik` | `< 5 detik` | `< 15 detik` |

### 💡 Penjelasan Istilah Metrik Utama (p95, p99, RPS)

1. **p95 (Percentile 95 Response Time)**
   - **Kepanjangan**: *95th Percentile Response Time*.
   - **Arti**: Nilai batas waktu respon di mana **95% dari seluruh total request** selesai diproses dalam waktu kurang dari atau sama dengan nilai tersebut. Hanya 5% request paling lambat yang melebihi p95.
   - **Fungsi**: Digunakan sebagai indikator utama kepuasan mayoritas pengguna. p95 jauh lebih realistis dibanding *Average (Rata-rata)* karena tidak terdistorsi oleh 1 atau 2 request ekstrem yang lambat.

2. **p99 (Percentile 99 Response Time)**
   - **Kepanjangan**: *99th Percentile Response Time*.
   - **Arti**: Nilai batas waktu respon di mana **99% dari seluruh request** selesai diproses. Nilai ini menggambarkan performa **1% request terburuk** (*tail latency*).
   - **Fungsi**: Membantu mendeteksi lonjakan latency yang terjadi akibat *Garbage Collection (GC) pause* pada Go runtime, antrean koneksi database PostgreSQL (*DB connection lock*), atau penundaan I/O disk.

3. **RPS (Requests Per Second)**
   - **Kepanjangan**: *Requests Per Second* (sering juga disebut *Throughput* atau *TPS / Transactions Per Second*).
   - **Arti**: Jumlah HTTP Request yang berhasil diterima, diproses, dan direspon oleh server Backend LMS dalam durasi 1 detik.
   - **Fungsi**: Mengukur daya tampung (*capacity & throughput*) aplikasi. Semakin tinggi nilai RPS yang mampu ditangani tanpa error, semakin efisien dan handal sistem LMS tersebut.

---

## 5. Metodologi & Peralatan (Tools & Infrastructure)

### 5.1 Tools yang Digunakan & Perbandingan dengan JMeter

#### ❓ Mengapa Menggunakan Grafana k6 Dibanding Apache JMeter?

| Kriteria | Grafana k6 (Dipilih) | Apache JMeter |
|---|---|---|
| **Bahasa & Scripting** | JavaScript / TypeScript (Modern, Developer-Friendly) | XML / GUI-based (`.jmx` file, rumit di-version control) |
| **Penggunaan Resource (RAM/CPU)** | **Sangat Ringan**: Ditulis dalam **Go**, mampu mensimulasikan ribuan VUs dari 1 server biasa tanpa hambatan JVM. | **Sangat Berat**: Berbasis **Java (JVM)**, membutuhkan RAM besar per Virtual User, rawan crash OOM di generator. |
| **Integrasi Observability** | Native mendukung ekspor data langsung ke **Prometheus & Grafana Dashboard**. | Membutuhkan plugin tambahan yang relatif rumit dikonfigurasi. |
| **Integrasi CI/CD** | Sangat mudah diintegrasikan ke Git / GitHub Actions / GitLab CI via CLI. | Cenderung lebih cocok dikendalikan via GUI desktop. |

#### ⚡ Arsitektur k6 Engine: Alasan Menggunakan JavaScript dengan Engine `Goja`

Meskipun skrip pengujian k6 ditulis dalam bahasa **JavaScript (`.js`)**, di balik layar Grafana k6 sendiri diciptakan **100% menggunakan bahasa GOLANG**.

Berikut adalah alasan teknis pemilihan kombinasi ini:
1. **Performa Native Golang via Engine `Goja`**:
   - k6 tidak menggunakan Node.js atau V8 engine, melainkan menggunakan **`Goja`** — yaitu JavaScript/ECMAScript 6 engine murni yang ditulis dalam bahasa **Go**.
   - Setiap Virtual User (VU) dieksekusi secara terisolasi di dalam **Go Goroutines** (bukan OS thread berat), sehingga mampu menghasilkan kecepatan penembakan HTTP native Go dengan konsumsi RAM yang sangat minim (~MB per VU).
2. **Fleksibilitas Scripting Tanpa Kompilasi Ulang**:
   - Jika skrip ditulis dalam Go murni, setiap kali ada perubahan variabel/skenario test (misal mengubah jumlah VU dari 100 ke 500), penguji harus **mengompilasi ulang binary Go** (`go build`).
   - Dengan JS pada engine Goja, skrip bersifat *interpreted/dynamic*, sehingga dapat langsung diedit dan dijalankan secara fleksibel tanpa kompilasi.
3. **Standar Universal Industri**:
   - JavaScript adalah bahasa universal yang dikuasai oleh tim QA Automation, Backend Developer, dan Frontend Developer, sehingga mempermudah kolaborasi antar tim.

#### 🛠️ Panduan Instalasi Tools Pengujian

1. **Grafana k6 (Load Generator)**
   - **Windows (Chocolatey / Winget)**:
     ```cmd
     winget install k6 --source winget
     # atau
     choco install k6
     ```
   - **Linux (Debian/Ubuntu)**:
     ```bash
     sudo gpg -k
     sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
     echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
     sudo apt-get update && sudo apt-get install k6
     ```

2. **Grafana & Prometheus (Monitoring Platform)**
   - **Via Docker Compose (Rekomendasi)**:
     ```yaml
     version: '3.8'
     services:
       prometheus:
         image: prom/prometheus
         ports:
           - "9090:9090"
         volumes:
           - ./prometheus.yml:/etc/prometheus/prometheus.yml
       grafana:
         image: grafana/grafana
         ports:
           - "3000:3000"
     ```
   - **Akses Dashboard**: Prometheus di `http://localhost:9090` dan Grafana di `http://localhost:3000`.

3. **Go pprof & Trace (Profiling Go Backend)**
   - **Apakah Perlu Modifikasi Existing Backend Go?**
     - **Untuk Load Testing Biasa (RPS / Latency Test)**: **TIDAK PERLU** modifikasi kode backend Go sama sekali. Server Go Gin dapat dites sebagaimana adanya.
     - **Untuk Deep Profiling (Memory Leak & CPU Bottleneck)**: Cukup tambahkan listener pprof di `backend/main.go` yang diset aktif hanya pada environment Staging/Testing.
   - **Cara Aktifkan di Code Go (`main.go`)**:
     ```go
     import _ "net/http/pprof"

     // Di main.go (aktifkan di staging / testing environment):
     if os.Getenv("ENABLE_PPROF") == "true" || os.Getenv("APP_ENV") == "development" {
         go func() {
             log.Println("[PPROF] Starting profiler on http://localhost:6060/debug/pprof")
             log.Println(http.ListenAndServe("localhost:6060", nil))
         }()
     }
     ```
   - **Cara Membaca Heap/CPU Profile**:
     ```bash
     go tool pprof http://localhost:6060/debug/pprof/heap
     go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
     ```

4. **PostgreSQL pg_stat_statements & pg_top (Database Monitoring)**
   - **Lokasi Folder `postgresql.conf`**:
     - **Windows**: `C:\Program Files\PostgreSQL\<versi>\data\postgresql.conf` (atau `C:\ProgramData\PostgreSQL\<versi>\data\postgresql.conf`).
     - **Linux (Ubuntu/Debian)**: `/etc/postgresql/<versi>/main/postgresql.conf`.
     - **Linux (RHEL/CentOS)**: `/var/lib/pgsql/data/postgresql.conf`.
     - 💡 *Tips Kepastian*: Jalankan SQL `SHOW config_file;` di PostgreSQL GUI/CLI untuk mengetahui lokasi pastinya di OS mana pun.
   - **Edit `postgresql.conf`**:
     ```ini
     shared_preload_libraries = 'pg_stat_statements'
     pg_stat_statements.track = all
     ```
   - **Jalankan SQL Command di DB `lms_db`**:
     ```sql
     CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
     ```
   - **Monitoring Query Slowest**:
     ```sql
     SELECT query, calls, total_exec_time, mean_exec_time, rows 
     FROM pg_stat_statements 
     ORDER BY total_exec_time DESC LIMIT 10;
     ```
   - **Install / Penggunaan `pg_top` di Windows**:
     - *Catatan*: `pg_top` merupakan utility bawaan Unix/Linux (`ncurses`).
     - **Opsi 1 (Rekomendasi Utama Windows)**: Gunakan **pgAdmin 4 (Dashboard / Server Activity Monitor)** atau **DBeaver (Dashboard Monitor)** yang menyediakan grafik CPU, Lock, dan Transaksi DB secara visual GUI.
     - **Opsi 2 (Via WSL - Windows Subsystem for Linux)**:
       ```bash
       sudo apt install pgtop
       pg_top -h localhost -p 5433 -U postgres -d lms_db
       ```
     - **Opsi 3 (Via Query SQL Activity Monitor)**:
       ```sql
       SELECT pid, usename, client_addr, state, now() - query_start AS duration, query 
       FROM pg_stat_activity 
       WHERE state != 'idle' 
       ORDER BY duration DESC;
       ```

---

```
+-----------------------------------------------------------------------------------+
|                               TEST ENVIRONMENT SETUP                              |
+-----------------------------------------------------------------------------------+
|                                                                                   |
|  ┌─────────────────────┐       HTTP / REST Load       ┌────────────────────────┐  |
|  │  Load Generator     ├─────────────────────────────►│  LMS Staging Backend   │  |
|  │  (Grafana k6 Node)  │                              │  (Go / Gin - Port 8086)│  |
|  └──────────┬──────────┘                              └───────────┬────────────┘  |
|             │ Metrik k6                                           │ DB Query      |
|             ▼                                                     ▼               |
|  ┌─────────────────────┐                              ┌────────────────────────┐  |
|  │  Grafana Dashboard  │◄─────────────────────────────┤  PostgreSQL Database   │  |
|  │  & Prometheus       │    System & DB Metrics       │  (Port 5433 / lms_sch) │  |
|  └─────────────────────┘                              └────────────────────────┘  |
+-----------------------------------------------------------------------------------+
```

### 5.2 Lingkungan Pengujian (Test Environment Requirement)
- **Prinsip Isolasi**: Pengujian **Wajib** dilakukan pada lingkungan *Staging / Pre-Production* yang dikloning identik dengan spesifikasi Production. **Dilarang keras** melakukan stress testing pada database Production aktif.
- **Ketersediaan Data Dummy (Seeding & Cleanup Tool)**:
  - Telah disediakan script automasi seeder Go di [`backend/seed_load.go`](file:///d:/Data_NK/Project5/LMS/backend/seed_load.go) yang mampu men-generate:
    - 10,000+ Master Data Karyawan & Anggota (`lms_sch.employees` & `members`).
    - 15,000+ Data Pengajuan Pinjaman, Pinjaman Aktif & Jadwal Angsuran (`lms_sch.loan_applications`, `loans`, & `loan_schedules`).
    - Cara menjalankan seeder:
      ```bash
      cd backend
      go run seed_load.go
      ```
  - Telah disediakan pula script DML Cleanup di [`backend/clean_seed.go`](file:///d:/Data_NK/Project5/LMS/backend/clean_seed.go) dan SQL script di [`backend/cleanup_load_test_data.sql`](file:///d:/Data_NK/Project5/LMS/backend/cleanup_load_test_data.sql) untuk menghapus data dummy jika pengujian selesai.

---

## 6. Tahapan Pelaksanaan & Panduan Eksekusi Detil

Proyek Stress & Load Testing LMS Kopkara dirancang berlangsung selama **3 Minggu** dengan alur kerja dan panduan eksekusi perintah terminal (*command line*) secara langsung:

```
Week 1: [Persiapan Tools & Seeding] ────► [Dry-Run k6 Script]
Week 2: [Eksekusi Load Test (500 VUs)] ──► [Eksekusi Stress Test (5,000 VUs)]
Week 3: [Diagnosa Profiling & DB]     ──► [Tuning & Final Reporting]
```

### 📋 Panduan Eksekusi Langkah-demi-Langkah (Dual OS: Linux & Windows)

#### 🔹 Langkah 1: Persiapan Environment & Database Extension
- **Linux (Ubuntu/WSL)**: `sudo apt update && sudo apt install -y k6 prometheus grafana`
- **Windows**: `winget install k6 --source winget`
- **Database Extension**: Edit `postgresql.conf` -> `shared_preload_libraries = 'timescaledb, pg_stat_statements'` -> Restart DB -> Jalankan SQL `CREATE EXTENSION IF NOT EXISTS pg_stat_statements;`

#### 🔹 Langkah 2: Menjalankan Data Seeder (10,000 Members & 15,000 Loans)
- **Linux / WSL**:
  ```bash
  cd /mnt/d/Data_NK/Project5/LMS/backend
  go run seed_load.go
  ```
- **Windows CMD**:
  ```cmd
  cd d:\Data_NK\Project5\LMS\backend
  go run seed_load.go
  ```

#### 🔹 Langkah 3: Menjalankan Backend Go dengan Profiling (`pprof`)
- **Linux / WSL**:
  ```bash
  cd /mnt/d/Data_NK/Project5/LMS/backend
  export ENABLE_PPROF=true
  export TRACE_LEVEL=0
  export PORT=8086
  go run main.go
  ```
- **Windows CMD**:
  ```cmd
  cd d:\Data_NK\Project5\LMS\backend
  set ENABLE_PPROF=true
  set TRACE_LEVEL=0
  set PORT=8086
  go run main.go
  ```
*Pastikan profiler aktif di*: `http://localhost:6060/debug/pprof/`

#### 🔹 Langkah 4: Eksekusi Load Test Standar (500 Virtual Users)
- **Linux / WSL**:
  ```bash
  cd /mnt/d/Data_NK/Project5/LMS
  k6 run --summary-export=load_test_summary.json load_tests/lms_load_test.js
  ```
- **Windows CMD**:
  ```cmd
  cd d:\Data_NK\Project5\LMS
  k6 run --summary-export=load_test_summary.json load_tests/lms_load_test.js
  ```

#### 🔹 Langkah 5: Eksekusi Stress Test & Breaking Point Test (Hingga 5,000 VUs)
- **Linux / Windows**:
  ```bash
  k6 run --stage 30s:500,1m:1000,2m:3000,1m:5000,30s:0 load_tests/lms_load_test.js
  ```

#### 🔹 Langkah 6: Diagnosa Bottleneck Real-Time
1. **Go Heap/CPU Profiling**: `go tool pprof http://localhost:6060/debug/pprof/heap`
2. **PostgreSQL Slow Queries**:
   ```sql
   SELECT query, calls, total_exec_time, mean_exec_time 
   FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT 5;
   ```

*Panduan eksekusi komprehensif dua OS tersimpan terpisah pada*: [`load_test_execution_guide.md`](file:///d:/Data_NK/Project5/LMS/load_test_execution_guide.md)

---

## 7. Deliverables (Hasil Kerja)

Di akhir kegiatan, tim penguji akan menyerahkan berkas deliverables sebagai berikut:

1. **Dokumen Laporan Akhir (Final Benchmark & Recommendation Report)**
   - Executive Summary hasil uji.
   - Grafik perbandingan performa sebelum dan sesudah optimasi (Latency, RPS, Error Rate).
   - Analisis *breaking point* sistem dan kapasitas maksimum pengguna simultan.
   - Daftar *bottleneck* yang ditemukan beserta status rekomendasi perbaikannya.
2. **Kumpulan Skrip Pengujian (k6 Test Scripts Repository)**
   - Source code skrip k6 (`.js`) modular yang reusable untuk pengujian berkala di kemudian hari (*Regression Performance Test*).
3. **Konfigurasi Monitoring & Export Dashboard**
   - File JSON dashboard Grafana untuk pemantauan performa LMS Kopkara.
4. **Petunjuk Eksekusi Mandiri (Testing Playbook & Documentation)**
   - Panduan langkah demi langkah cara menjalankan stress test secara mandiri oleh tim internal IT Kopkara.

---

## 8. Manajemen Risiko (Risk Management)

| Risiko | Tingkat Risiko | Mitigation Strategy (Pencegahan) |
|---|---|---|
| **Pencemaran Data Production** | High | Pengujian **100% dilakukan di Staging Environment** yang terpisah total secara database dan network dari Production. |
| **Staging Environment Server Crash** | Medium | Menyediakan skrip restart otomatis (*systemd / docker compose*) dan pemulihan backup database staging secara cepat. |
| **False Positive pada k6 Generator** | Low | Memastikan server pembawa skrip k6 memiliki resource CPU/RAM dan bandwidth yang cukup agar tidak menjadi bottleneck pengujian. |
| **Kebocoran Data Sensitif Anggota** | High | Seluruh data test menggunakan *anonymized / synthetic data generator* (nama, NIK, dan nomor rekening samaran). |

---

## 9. Struktur Tim & Pembagian Peran

| Peran | Tanggung Jawab Utama |
|---|---|
| **Lead Performance Test Engineer** | Merancang strategi, membuat skrip k6, memimpin eksekusi test, dan menyusun laporan akhir. |
| **Backend Go Engineer (LMS)** | Membantu analisa pprof, optimasi handler Gin/GORM, dan melakukan patch bug/bottleneck. |
| **Database Administrator (DBA) / DevOps** | Memantau PostgreSQL, menganalisis slow query, mengonfigurasi connection pool, dan setup Prometheus/Grafana. |
| **Business Analyst / System Owner Kopkara** | Validasi skenario bisnis, volume data riil, dan penentuan kriteria penerimaan (SLA). |

---

## 10. Kesimpulan & Langkah Selanjutnya

Pengujian **Stress & Load Testing** merupakan investasi krusial untuk menjamin keandalan, keamanan transaksi, dan stabilitas **Loan Management System (LMS) Koperasi Kopkara**. Dengan pelaksanaan pengujian ini, Koperasi Kopkara dapat memastikan bahwa sistem siap mendukung operasional harian serta pemrosesan *bulk billing payroll* Adira secara lancar tanpa hambatan teknis.

**Rekomendasi Langkah Selanjutnya:**
1. Review dan persetujuan proposal oleh manajemen / tim teknis Koperasi Kopkara.
2. Penentuan jadwal kick-off meeting dan alokasi server staging.
3. Inisialisasi seeding data dan persiapan tools k6.

---
*Dokumen ini disusun untuk Koperasi Karyawan Kopkara — LMS Project 2026.*
