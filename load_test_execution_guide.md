# Panduan Detil Langkah-demi-Langkah (Step-by-Step Execution Guide)

## Load & Stress Testing — Loan Management System (LMS) Kopkara

---

## Document Information

- **Target System**: Backend Go (Gin Framework) + PostgreSQL Database (`lms_sch` schema)
- **Tools**: Grafana k6, Go pprof, PostgreSQL `pg_stat_statements`
- **Tujuan**: Menjalankan pengujian beban (Load Test) dan pengujian batas kemampuan (Stress Test) secara terstruktur pada OS **Linux (Ubuntu/WSL)** maupun **Windows**.

---

## Quick Map: Alur Eksekusi

```
┌────────────────────────┐      ┌────────────────────────┐      ┌────────────────────────┐
│  FASE 1: PERSIAPAN     │─────►│  FASE 2: SEEDING DATA  │─────►│  FASE 3: RUN BACKEND   │
│  - Install k6          │      │  - Run seed_load       │      │  - Run Go Backend      │
│  - Config DB Extension │      │  - Verify Data Count   │      │  - Enable pprof        │
└────────────────────────┘      └────────────────────────┘      └────────────────────────┘
                                                                            │
                                                                            ▼
┌────────────────────────┐      ┌────────────────────────┐      ┌────────────────────────┐
│  FASE 6: ANALYSIS      │◄─────│  FASE 5: STRESS TEST   │◄─────│  FASE 4: LOAD TEST     │
│  - Go pprof Profile    │      │  - Up to 5,000 VUs     │      │  - Smoke Test (1 VU)   │
│  - DB Slow Query Log   │      │  - Find Breaking Point │      │  - Run Standard 500 VUs│
└────────────────────────┘      └────────────────────────┘      └────────────────────────┘
```

---

## FASE 1: Persiapan Environment & Tools

### Langkah 1.1: Install Grafana k6 (Load Generator CLI)

> 💡 **Catatan Arsitektur k6**:
>
> - `k6` adalah sebuah **CLI Tool** (bukan service background daemon). `k6` hanya berjalan ketika dipanggil melalui command terminal (`k6 run ...`).
> - **Mengapa Menggunakan Script JS?**: k6 menggunakan engine **`Goja`** (JavaScript runtime murni yang ditulis dalam Golang). Skrip ditulis dalam JS demi kemudahan scripting tanpa perlu mengompilasi ulang binary, tetapi eksekusinya 100% diproses oleh **Go Goroutines** dengan performa native Golang.

#### A. Di Linux Ubuntu / WSL Terminal (Bash):

Jika muncul error `Command 'k6' not found`, jalankan perintah berikut di terminal Linux:

```bash
# 1. Tambahkan GPG Key dan Repository Resmi Grafana k6
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list

# 2. Update paket & Install k6
sudo apt-get update
sudo apt-get install -y k6

# Alternatif Instan via Snap:
# sudo snap install k6
```

#### B. Di Windows Command Prompt (CMD):

```cmd
winget install k6 --source winget
# atau via Chocolatey
choco install k6
```

#### C. Verifikasi Instalasi k6 (Linux & Windows):

```bash
k6 version
```

_Output sukses yang diharapkan_: `k6 v0.45.0 (go1.20.4, ...)`

---

### Langkah 1.2: Install Native Prometheus & Grafana (Tanpa Docker)

#### A. Di Linux Ubuntu / WSL Terminal (Bash):

1. **Install & Start Prometheus**:

   ```bash
   sudo apt update && sudo apt install -y prometheus
   sudo service prometheus start
   sudo service prometheus status
   # Dashboard: http://localhost:9090
   ```

2. **Install & Start Grafana**:

   ```bash
   sudo apt-get install -y apt-transport-https software-properties-common wget
   sudo mkdir -p /etc/apt/keyrings/
   wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor | sudo tee /etc/apt/keyrings/grafana.gpg > /dev/null
   echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://apt.grafana.com stable main" | sudo tee /etc/apt/sources.list.d/grafana.list

   sudo apt-get update && sudo apt-get install -y grafana
   sudo service grafana-server start
   sudo service grafana-server status
   # Dashboard: http://localhost:3000 (admin/admin)
   ```

#### C. Menghubungkan k6 ke Grafana Dashboard Visual (Grafik Real-Time):

1. **Jalankan k6 dengan Output Prometheus**:
   ```bash
   cd /mnt/d/Data_NK/Project5/LMS
   K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9090/api/v1/write k6 run --out experimental-prometheus-rw load_tests/lms_load_test.js
   ```
2. **Import Dashboard Resmi k6 di Grafana**:
   - Buka Grafana di browser: `http://localhost:3000` (Login: `admin` / `admin`).
   - Klik menu **Dashboards** -> **New** -> **Import**.
   - Masukkan ID Dashboard Resmi k6: **`18030`** (atau **`10660`**) -> Klik **Load**.
   - Pilih Data Source: **Prometheus** -> Klik **Import**.
   - 🎉 _Dashboard interaktif berisi grafik real-time RPS, Latency p95/p99, Error Code Status 200/400/500, CPU/RAM, dan VUs Ramp-up langsung tampil secara visual!_

---

### Langkah 1.3: Konfigurasi PostgreSQL `pg_stat_statements` (Database Slow Query Tracker)

1. **Buka file `postgresql.conf`**:
   - **Linux / WSL**: `/etc/postgresql/<versi>/main/postgresql.conf`
   - **Windows**: `C:\Program Files\PostgreSQL\<versi>\data\postgresql.conf`
   - 💡 _Tips_: Cek lokasi pasti dengan SQL: `SHOW config_file;`

2. **Edit `postgresql.conf`**:

   ```ini
   # JIKA SUDAH ADA EXTENSION LAIN (misal: timescaledb untuk GPS):
   # Cukup tambahkan koma ( , ) seperti berikut:
   shared_preload_libraries = 'timescaledb, pg_stat_statements'

   pg_stat_statements.track = all
   ```

3. **Restart Service PostgreSQL**:
   - **Linux / WSL**: `sudo service postgresql restart`
   - **Windows**: Restart via `services.msc` atau PostgreSQL Service.

4. **Aktifkan Extension di Database LMS**:
   Buka pgAdmin / DBeaver / psql di database `lms_db`, eksekusi SQL:
   ```sql
   CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
   ```

---

## FASE 2: Data Seeding (Menyiapkan Data Dummy Skala Besar)

Sebelum pengujian beban dimulai, database staging wajib diisi data dummy minimal 10,000 anggota dan 15,000 pinjaman.

#### 📊 Rincian Data yang Di-populate oleh `seed_load.go` untuk Testing 500 VUs:

- **Master Departments**: 3 record (`DEPT01`, `DEPT02`, `DEPT03`).
- **Master Roles**: 2 record (`role_id: 1` Admin, `role_id: 2` Anggota).
- **Master Categories & Statuses**: `PERM`, `CONT`, `ACTIVE`, `RESIGNED`, `INACTIVE`.
- **Master Loan Products**: 3 record (`id: 1` Multiguna, `id: 2` Pendidikan, `id: 3` Darurat).
- **Master Karyawan (Employees)**: **10,000 record** (`employee_id`: 100001 - 110000).
- **Master Anggota (Members)**: **10,000 record** (`member_no`: 200001 - 210000).
- **Pengajuan Pinjaman (Loan Applications)**: **15,000 record** (`application_no`: 300001 - 315000).
- **Pinjaman Aktif (Loans)**: **15,000 record** (`loan_no`: 400001 - 415000).

### Langkah 2.1: Menjalankan Seeder Script Go

#### A. Di Linux / WSL Terminal (Bash):

```bash
cd /mnt/d/Data_NK/Project5/LMS/backend
go run seed_load.go
```

#### B. Di Windows Command Prompt (CMD):

```cmd
cd d:\Data_NK\Project5\LMS\backend
go run seed_load.go
```

_Output Sukses yang Diharapkan:_

```text
[SEEDER] Connecting to Database: localhost:5433/lms_db ...
[SEEDER] Database connected successfully!
[SEEDER] Starting Seeding Process: 10000 Employees/Members & 15000 Loans...
[SEEDER] Step 1: Seeding Parent Master Tables (Departments, Roles, Categories, Statuses)...
[SEEDER] Step 1 Complete: All Parent Master Tables successfully seeded.
[SEEDER] Step 2: Seeding Employees & Members in batches...
[SEEDER] Step 2 Complete: Successfully seeded 10000 Employees & Members.
[SEEDER] Step 3: Seeding Loan Applications & Schedules in batches...
[SEEDER SUCCESS] Completed All Seeding in 2.15s!
[SEEDER] Test Environment is Ready for Load Testing.
```

### Langkah 2.2: Verifikasi Jumlah Data di Database (Linux & Windows)

Jalankan query ini di PostgreSQL GUI (pgAdmin/DBeaver/psql):

```sql
SELECT
    (SELECT COUNT(*) FROM lms_sch.members) AS total_members,
    (SELECT COUNT(*) FROM lms_sch.loans) AS total_loans,
    (SELECT COUNT(*) FROM lms_sch.loan_schedules) AS total_schedules;
```

_Pastikan `total_members` >= 10,000 dan `total_loans` >= 15,000._

### Langkah 2.3: Perintah DML Cleanup / Reset Data Dummy (Opsional)

Jika pengujian selesai atau Anda ingin membersihkan ulang data dummy hasil seeder:

#### A. Menjalankan via Go CLI Script:

- **Linux / WSL**:
  ```bash
  cd /mnt/d/Data_NK/Project5/LMS/backend
  go run clean_seed.go
  ```
- **Windows**:
  ```cmd
  cd d:\Data_NK\Project5\LMS\backend
  go run clean_seed.go
  ```

#### B. Menjalankan via DML SQL (Tersimpan di [`backend/cleanup_load_test_data.sql`](file:///d:/Data_NK/Project5/LMS/backend/cleanup_load_test_data.sql)):

```sql
-- 1. Hapus Loan Schedules dummy
DELETE FROM lms_sch.loan_schedules WHERE loan_no BETWEEN 400001 AND 415000;

-- 2. Hapus Loans dummy
DELETE FROM lms_sch.loans WHERE loan_no BETWEEN 400001 AND 415000;

-- 3. Hapus Loan Trackings & Contracts dummy
DELETE FROM lms_sch.loan_trackings WHERE application_no BETWEEN 300001 AND 315000;
DELETE FROM lms_sch.loan_contracts WHERE application_no BETWEEN 300001 AND 315000;

-- 4. Hapus Loan Applications dummy
DELETE FROM lms_sch.loan_applications WHERE application_no BETWEEN 300001 AND 315000;

-- 5. Hapus Members dummy
DELETE FROM lms_sch.members WHERE member_no BETWEEN 200001 AND 210000;

-- 6. Hapus Employees dummy
DELETE FROM lms_sch.employees WHERE employee_id BETWEEN 100001 AND 110000;

-- 7. Hapus Loan Products dummy
DELETE FROM lms_sch.loan_products WHERE id IN (1, 2, 3);

-- 8. Hapus Master Parent Tables dummy
DELETE FROM lms_sch.departments WHERE deptno IN ('DEPT01', 'DEPT02', 'DEPT03');
DELETE FROM lms_sch.roles WHERE role_id IN (1, 2);
DELETE FROM lms_sch.employee_categories WHERE category_code IN ('PERM', 'CONT');
DELETE FROM lms_sch.employee_statuses WHERE status_code IN ('ACTIVE', 'RESIGNED');
DELETE FROM lms_sch.kopkara_statuses WHERE status_code IN ('ACTIVE', 'INACTIVE');
```

---

## FASE 3: Menjalankan Backend Go dengan Profiling (`pprof`)

### Langkah 3.1: Aktifkan Option Profiling & Run Backend

#### A. Di Linux / WSL Terminal (Bash):

```bash
cd /mnt/d/Data_NK/Project5/LMS/backend

# Opsi TRACE_LEVEL Backend:
# TRACE_LEVEL=0 -> Logging OFF (Maksimal Throughput)
# TRACE_LEVEL=1 -> HTTP Request Log Only [LMS-API-LOG] (Format angka murni ms %.3f tanpa SQL)
# TRACE_LEVEL=2 -> HTTP Log + SQL SELECT Only
# TRACE_LEVEL=3 -> HTTP Log + ALL SQL Queries (SELECT, INSERT, UPDATE, DELETE)
export ENABLE_PPROF=true
export TRACE_LEVEL=1
export PORT=8086

# Atau jalankan inline sekaligus:
# ENABLE_PPROF=true TRACE_LEVEL=0 PORT=8086 go run main.go

# go run main.go
ENABLE_PPROF=true TRACE_LEVEL=1 go run main.go > backend_activity.log 2>&1 &
```

#### B. Di Windows Command Prompt (CMD):

```cmd
cd d:\Data_NK\Project5\LMS\backend

set ENABLE_PPROF=true
set TRACE_LEVEL=0
set PORT=8086
go run main.go
```

_Output Terminal Backend:_

```text
2026/08/17 11:20:00 [PPROF] Starting Go profiler server on http://localhost:6060/debug/pprof
Starting HTTPS server on port 8086 using cert ./sertifikat/lims.local+2.pem and key ./sertifikat/lims.local+2-key.pem...
[GIN-debug] Listening and serving HTTPS on :8086
```

### Langkah 3.2: Verifikasi Endpoint Backend & Profiler

- **Cek Web Profiler Go**: Buka browser: **`http://localhost:6060/debug/pprof/`**  
  _(Pastikan muncul daftar profiler Go seperti `allocs`, `goroutine`, `heap`, `profile`)._
- **Cek Endpoint Health Check Backend (HTTPS)**: Buka browser: **`https://localhost:8086/api/health`**  
  _(💡 Catatan: Backend LMS berjalan menggunakan sertifikat SSL HTTPS. Jika diakses via HTTP `http://...`, browser akan mengembalikan `HTTP ERROR 400`)._

---

## FASE 4: Eksekusi Load Test (Beban Operasional Standar)

### Langkah 4.1: Dry-Run / Smoke Test (Uji Validasi Skrip 5 VUs)

#### A. Di Linux / WSL Terminal (Bash):

```bash
cd /mnt/d/Data_NK/Project5/LMS
#k6 run --vus 5 --duration 15s load_tests/lms_load_test.js
k6 run --vus 500 load_tests/lms_load_test.js
```

#### B. Di Windows Command Prompt (CMD):

```cmd
cd d:\Data_NK\Project5\LMS
k6 run --vus 5 --duration 15s load_tests/lms_load_test.js
```

### Langkah 4.2: Menjalankan Standard Load Test (500 Virtual Users)

#### A. Di Linux / WSL Terminal (Bash):

```bash
cd /mnt/d/Data_NK/Project5/LMS
k6 run --summary-export=load_test_summary.json load_tests/lms_load_test.js
```

#### B. Di Windows Command Prompt (CMD):

```cmd
cd d:\Data_NK\Project5\LMS
k6 run --summary-export=load_test_summary.json load_tests/lms_load_test.js
```

---

## FASE 5: Eksekusi Stress Test & Breaking Point Test

### Langkah 5.1: Menjalankan Ramp-Up Stress Test (Hingga 5,000 VUs)

#### A. Di Linux / WSL Terminal (Bash):

```bash
cd /mnt/d/Data_NK/Project5/LMS
k6 run --stage 30s:500,1m:1000,2m:3000,1m:5000,30s:0 load_tests/lms_load_test.js
```

#### B. Di Windows Command Prompt (CMD):

```cmd
cd d:\Data_NK\Project5\LMS
k6 run --stage 30s:500,1m:1000,2m:3000,1m:5000,30s:0 load_tests/lms_load_test.js
```

---

## FASE 6: Diagnosa Bottleneck Real-Time & Profiling

### Langkah 6.1: Diagnosa Alokasi Memori & CPU Go Backend (`pprof`)

Buka terminal baru di **Linux / WSL** atau **Windows**:

1. **Cek Alokasi Memori Heap (Mendeteksi Memory Leak)**:

   ```bash
   go tool pprof http://localhost:6060/debug/pprof/heap
   ```

   _Ketik `top10` di dalam prompt pprof untuk melihat 10 fungsi teratas yang paling banyak memakan RAM._

2. **Cek Profil CPU (Mendeteksi Fungsi Lambat)**:

   ```bash
   go tool pprof http://localhost:6060/debug/pprof/profile?seconds=20
   ```

   _Ketik `top10` untuk melihat fungsi mana yang paling lama mengonsumsi CPU._

3. **Cek Jumlah Goroutine**: Buka browser `http://localhost:6060/debug/pprof/goroutine?debug=1`.

### Langkah 6.2: Diagnosa Slow Query PostgreSQL (Linux & Windows)

Di GUI Database (pgAdmin/DBeaver/psql), jalankan query ini:

```sql
SELECT
    substring(query, 1, 100) AS query_snippet,
    calls,
    round(total_exec_time::numeric, 2) AS total_time_ms,
    round(mean_exec_time::numeric, 2) AS avg_time_ms,
    rows
FROM pg_stat_statements
WHERE query NOT LIKE '%pg_stat_statements%'
ORDER BY total_exec_time DESC
LIMIT 5;
```

---

## FASE 7: Pelaporan & Hasil Uji Benchmark Aktual (500 VUs)

| Parameter Metrik             | Target SLA   | Hasil Uji Aktual (500 VUs)              | Status              |
| ---------------------------- | ------------ | --------------------------------------- | ------------------- |
| **Business Check Pass Rate** | `> 95.00 %`  | **100.00 %** (56,144 / 56,144 Passed)   | **PASSED (✓ 100%)** |
| **Response Time (p95)**      | `< 500 ms`   | **3.44 ms** (Super Low Latency)         | **PASSED (✓)**      |
| **Response Time (p99)**      | `< 5,000 ms` | **4.97 ms** (No Queueing Delay)         | **PASSED (✓)**      |
| **Average Response Time**    | `< 200 ms`   | **2.04 ms**                             | **PASSED (✓)**      |
| **Throughput / Speed**       | `> 200 RPS`  | **235.42 RPS** (42,108 total HTTP reqs) | **OPTIMAL**         |
| **Completed Iterations**     | `> 3,000`    | **7,018 Iterations**                    | **PASSED (✓)**      |

### 💡 Cara Memeriksa Log Pengujian & Backend:

1. **File Laporan JSON k6**: Laporan ringkasan k6 otomatis tersimpan pada [`load_test_summary.json`](file:///d:/Data_NK/Project5/LMS/load_test_summary.json).
2. **Streaming Log Backend Go**:
   - Di Linux / WSL:
     ```bash
     cd /mnt/d/Data_NK/Project5/LMS/backend
     ENABLE_PPROF=true TRACE_LEVEL=1 go run main.go > backend_activity.log 2>&1 &
     tail -f backend_activity.log
     ```
3. **Database Slow Query Log (PostgreSQL)**:
   - Jalankan SQL di pgAdmin/psql:
     ```sql
     SELECT query, calls, round(total_exec_time::numeric, 2) AS total_ms, round(mean_exec_time::numeric, 2) AS avg_ms
     FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT 10;
     ```

```go
sqlDB.SetMaxOpenConns(100)                // Membuka hingga 100 koneksi DB simultan
sqlDB.SetMaxIdleConns(25)                 // Menjaga 25 koneksi siap pakai
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

---

_Panduan ini dikonfigurasi lengkap untuk eksekusi di lingkungan Linux (Ubuntu/WSL) dan Windows._
