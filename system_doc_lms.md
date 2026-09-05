# Dokumentasi Sistem LMS (Loan Management System)

**Koperasi Kopkara — Internal Loan & Payroll Deduction System**
**Versi:** 1.0 · **Terakhir diperbarui:** Agustus 2026

---

## Daftar Isi

1. [Gambaran Umum](#1-gambaran-umum)
2. [`A`r`s`i`t`e`k`t`u`r` `S`i`s`t`e`m`]`(`#`2`-`a`r`s`i`t`e`k`t`u`r`-`s`i`s`t`e`m`)`
`3`.` `[`S`t`a`c`k` `T`e`k`n`o`l`o`g`i`]`(`#`3`-`s`t`a`c`k`-`t`e`k`n`o`l`o`g`i`)
3. [Struktur Direktori](#4-struktur-direktori)
4. [Database](#5-database)
5. [Backend — Go/Gin API](#6-backend--gogin-api)
6. [Frontend — React SPA](#7-frontend--react-spa)
7. [Alur Bisnis Utama](#8-alur-bisnis-utama)
8. [Integrasi Billing HRD-Adira](#9-integrasi-billing-hrd-adira)
9. [Cache Strategy](#10-cache-strategy)
10. [Konfigurasi & Environment](#11-konfigurasi--environment)
11. [Folder Billing](#12-folder-billing)
12. [Global Parameters Kunci](#13-global-parameters-kunci)
13. [Mobile Self-Service EWA & Autentikasi Security](#14-mobile-self-service-ewa--autentikasi-security)
    - 14.1 [Arsitektur Pendaftaran Self-Service (4-Factor Matching)](#141-arsitektur-pendaftaran-self-service-4-factor-matching)
    - 14.2 [Parameter Keamanan Dinamis](#142-parameter-keamanan-dinamis-lmsschglobal_parameters)
    - 14.3 [Script DDL SQL Standalone](#143-script-ddl-sql-standalone-alter_table_employees_and_userssql)
    - 14.4 [Build Android APK (Capacitor & Gradle)](#144-build-android-apk-capacitor--gradle)
    - 14.5 [Spesifikasi Alur Pendaftaran Resmi 6-Step](#145-spesifikasi-alur-pendaftaran-resmi-6-step-postapikarismaregister)
    - 14.6 [Hasil Testing & Verifikasi End-to-End](#146-hasil-testing--verifikasi-end-to-end-verified-test-results)
    - 14.7 [Format Dashboard & Pemindahan Field Rekening Bank](#147-format-dashboard--pemindahan-field-rekening-bank)
    - 14.8 [Tampilan Single-Screen Compact Dashboard & Penyelesaian Menu EWA](#148-tampilan-single-screen-compact-dashboard--penyelesaian-menu-ewa)
    - 14.9 [Refinement Tabel Daftar Pinjaman & Navigasi Paginasi 26 Data](#149-refinement-tabel-daftar-pinjaman--navigasi-paginasi-26-data)
    - 14.10 [Refinement UI Login & Modal Registrasi EWA Mobile](#1410-refinement-ui-login--modal-registrasi-ewa-mobile)
    - 14.11 [Proteksi Menu Password, Notifikasi & Fitur Ganti Password/PIN](#1411-proteksi-menu-password-notifikasi--fitur-ganti-passwordpin)
    - 14.12 [Refinement Verifikasi Password/PIN Form Pengajuan](#1412-refinement-verifikasi-passwordpin-form-pengajuan)
    - 14.13 [Perbaikan AuthMiddleware Cookie Public Endpoints & Eye Icon Universal](#1413-perbaikan-authmiddleware-cookie-public-endpoints--eye-icon-universal)
    - 14.14 [Deteksi Otomatis & Adaptasi Dinamis Teks Label Ganti Password vs Ganti PIN](#1414-deteksi-otomatis--adaptasi-dinamis-teks-label-ganti-password-vs-ganti-pin)
    - 14.15 [Deteksi Username Custom Non-Numeric (User `nur` Role Admin)](#1415-deteksi-username-custom-non-numeric-user-nur-role-admin)
    - 14.16 [Restriksi Platform: Ganti PIN (Web & Mobile) vs Ganti Password (Khusus Web-Apps)](#1416-restriksi-platform-ganti-pin-web--mobile-vs-ganti-password-khusus-web-apps)
    - 14.34 [Pembaruan Ekspor Laporan PDF, Format Tanggal/Mata Uang, Eliminasi Notifikasi DB & Proteksi Registrasi](#1434-pembaruan-ekspor-laporan-pdf-format-tanggalmata-uang-eliminasi-notifikasi-db--proteksi-registrasi)

---

## 1. Gambaran Umum

LMS (Loan Management System) adalah sistem internal Koperasi **Kopkara** untuk mengelola siklus penuh pinjaman karyawan PT. Adira Finance dan perusahaan yang berkaitan. Sistem ini mencakup:

- **Pengajuan & Persetujuan Pinjaman** — origination dari anggota/karyawan hingga pencairan
- **Penjadwalan Angsuran** — pembuatan loan schedule otomatis berdasarkan tenor dan suku bunga
- **Billing & Rekonsiliasi Payroll** — export tagihan ke HRD-Adira (Karisma) dan import hasil potongan gaji
- **Pelunasan Manual** — penanganan pelunasan di luar siklus payroll (resign, pesangon, transfer bank)
- **Master Data** — pengelolaan anggota, karyawan, produk pinjaman, dan parameter sistem

---

## 2. Arsitektur Sistem

```
+-----------------------------------------------------+
|                  FRONTEND (React)                   |
|          Single Page Application (SPA)              |
|            Port: Vite Dev Server (5173)             |
+---------------------+-------------------------------+
                      | HTTP REST (axios)
                      v
+-----------------------------------------------------+
|              BACKEND (Go / Gin)                     |
|                  Port: 8086                         |
|                                                     |
|  +-----------------------------------------------+  |
|  |  In-Memory RAM Cache                          |  |
|  |  +- ParameterCache (global_parameters)        |  |
|  +-----------------------------------------------+  |
|                                                     |
|  Handlers -> Usecases -> Repositories -> GORM       |
+---------------------+-------------------------------+
                      | GORM / Raw SQL
                      v
+-----------------------------------------------------+
|           PostgreSQL Database                       |
|           lms_db / Schema: lms_sch                  |
|           Port: 5433                                |
+-----------------------------------------------------+
```

### Alur Request Standar

```
Browser → GET/POST /api/* → CORSMiddleware → LoggerMiddleware
        → Handler → UseCase → Repository → Cache / DB
        → JSON Response
```

---

## 3. Stack Teknologi

| Komponen           | Teknologi           | Versi  |
| ------------------ | ------------------- | ------ |
| Backend Language   | Go                  | 1.23.1 |
| Backend Framework  | Gin                 | 1.9.1  |
| ORM                | GORM                | 1.25.7 |
| Database Driver    | gorm/postgres (pgx) | 1.5.7  |
| Expression Engine  | govaluate           | 3.0.0  |
| Config             | godotenv            | 1.5.1  |
| Database           | PostgreSQL          | —      |
| Frontend Framework | React (Vite)        | —      |
| HTTP Client (FE)   | axios               | —      |

---

## 4. Struktur Direktori

```
D:\Data_NK\Project5\LMS\
├── backend/                    # Go API Server
│   ├── main.go                 # Entry point, route definitions, inline handlers
│   ├── .env                    # Environment variables
│   ├── go.mod / go.sum         # Go module dependencies
│   │
│   ├── cache/
│   │   └── parameter_cache.go  # In-memory RAM cache khusus global_parameters
│   │
│   ├── config/
│   │   └── database.go         # PostgreSQL connection (GORM)
│   │
│   ├── models/
│   │   ├── base.go             # BaseModel, MasterBaseModel (timestamps, soft-delete)
│   │   ├── master.go           # Master data models (GlobalParameter, Employee, Member, dll)
│   │   └── transaction.go      # Transaction models (LoanApplication, Loan, LoanSchedule, dll)
│   │
│   ├── repositories/
│   │   ├── application_repository.go
│   │   ├── parameter_repository.go   # Includes in-memory cache (map) untuk global_parameters
│   │   └── product_repository.go     # Query real-time langsung dari database PostgreSQL
│   │
│   ├── usecases/
│   │   ├── application_usecase.go    # Logika bisnis: eligibility, simulasi, submit, approve, disburse
│   │   ├── parameter_usecase.go
│   │   └── product_usecase.go
│   │
│   └── handlers/
│       ├── application_handler.go
│       ├── master_data_handler.go    # Generic CRUD handler untuk tabel master
│       ├── parameter_handler.go
│       └── product_handler.go
├── frontend/                   # React SPA
│   └── src/
│       ├── App.jsx             # Komponen utama, semua state & UI (~3700 baris)
│       ├── App.css
│       ├── index.css
│       └── main.jsx
├── Billing/                    # Folder file CSV billing
│   ├── Export/                 # Output: file tagihan untuk HRD-Adira
│   ├── Import/                 # Input: file hasil potongan dari HRD-Adira
│   └── BCK/                    # Backup file yang sudah diproses
├── system_doc_lms.md           # Dokumentasi sistem (file ini)
└── simulator/                  # Simulator Karisma (testing)
```

---

## 5. Database

### Koneksi

- **Host:** localhost
- **Port:** 5433
- **Database:** `lms_db`
- **Schema:** `lms_sch`
- **User:** `admin_lms`
- **Table Prefix:** semua tabel di-prefix `lms_sch.` via GORM `NamingStrategy`

### Create User `lms_app` & Grant Privileges

```sql
-- 1. Buat user (role) dengan password yang sama seperti yang dipakai di .env
CREATE ROLE lms_app WITH LOGIN PASSWORD '*********';
-- 2. Izinkan koneksi ke database
GRANT CONNECT ON DATABASE lms_db TO lms_app;
-- 3. Izinkan penggunaan schema lms_sch
GRANT USAGE ON SCHEMA lms_sch TO lms_app;
-- 4. Beri hak akses pada semua tabel yang sudah ada
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA lms_sch TO lms_app;
-- 5. Beri hak akses pada semua sequence (auto‑increment)
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA lms_sch TO lms_app;
-- 6. Pastikan tabel/sequence yang dibuat di masa depan otomatis memiliki hak yang sama
ALTER DEFAULT PRIVILEGES IN SCHEMA lms_sch
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO lms_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA lms_sch
  GRANT USAGE, SELECT ON SEQUENCES TO lms_app;
```

> **Catatan:** Pastikan password di atas sama dengan nilai `DB_PASSWORD` (atau gunakan `DB_PASSWORD_ENCRYPTED` bersama `JWT_SECRET`). Setelah menjalankan perintah ini, restart aplikasi LMS untuk memastikan user dapat mengakses schema `lms_sch`.

### Daftar Tabel

#### Master Data

| Tabel                         | Deskripsi                                                                                     | Kunci Utama        |
| ----------------------------- | --------------------------------------------------------------------------------------------- | ------------------ |
| `lms_sch.global_parameters`   | Parameter konfigurasi sistem (folder, mode scan, dll)                                         | `id`               |
| `lms_sch.loan_products`       | Produk pinjaman (APM, reguler, dll)                                                           | `id`               |
| `lms_sch.employee_categories` | Kategori karyawan & limit pinjaman                                                            | `category_code`    |
| `lms_sch.employees`           | Data karyawan dari Karisma/HRD (termasuk `total_loan` untuk validasi credit limit multi-loan) | `employee_id`      |
| `lms_sch.members`             | Anggota koperasi (mapping employee → member)                                                  | `member_no`        |
| `lms_sch.employee_statuses`   | Referensi status karyawan                                                                     | `status_code`      |
| `lms_sch.departments`         | Referensi departemen                                                                          | `deptno`           |
| `lms_sch.kopkara_statuses`    | Referensi status anggota koperasi                                                             | `status_code`      |
| `lms_sch.admin_fees`          | Biaya administrasi pinjaman                                                                   | `id`               |
| `lms_sch.roles`               | Role pengguna sistem                                                                          | `role_id`          |
| `lms_sch.menus`               | Definisi menu/navigasi                                                                        | `menu_id`          |
| `lms_sch.role_menus`          | Mapping role ke menu (RBAC)                                                                   | `role_id, menu_id` |

#### Transaksi

| Tabel                              | Deskripsi                                        | Kunci Utama      |
| ---------------------------------- | ------------------------------------------------ | ---------------- |
| `lms_sch.loan_applications`        | Pengajuan pinjaman (origination)                 | `application_no` |
| `lms_sch.loan_trackings`           | Riwayat perubahan status pengajuan (audit trail) | `id`             |
| `lms_sch.loan_contracts`           | Kontrak pinjaman setelah disetujui               | `contract_no`    |
| `lms_sch.loans`                    | Pinjaman aktif (post-disbursement)               | `loan_no`        |
| `lms_sch.loan_schedules`           | Jadwal angsuran bulanan                          | `id`             |
| `lms_sch.payroll_deductions`       | Log potongan payroll per baris CSV               | `id`             |
| `lms_sch.payroll_adjustments`      | Penyesuaian manual potongan payroll              | `id`             |
| `lms_sch.loan_payroll_import_logs` | Log summary file import                          | `id`             |

### Relasi Antar Tabel

```
employees (employee_id)
    └── members (employee_id → employee_id)
            └── loan_applications (member_no)
                    ├── loan_trackings (application_no)
                    ├── loan_contracts (application_no)
                    └── loans (application_no)
                            ├── loan_schedules (loan_no)
                            ├── payroll_deductions (loan_no)
                            └── payroll_adjustments (loan_no)
```

### Soft Delete

Semua tabel master menggunakan kolom `deleted_at` untuk soft-delete. Query normal selalu menambahkan filter `WHERE deleted_at IS NULL`.

### Index Penting

| Tabel            | Kolom            | Alasan                               |
| ---------------- | ---------------- | ------------------------------------ |
| `loan_schedules` | `period`         | Performa query billing export/import |
| `members`        | `employee_id`    | Join dengan tabel employees          |
| `loan_trackings` | `application_no` | Filter riwayat per pengajuan         |
| `loan_trackings` | `user_id`        | Filter riwayat per user              |

---

## 6. Backend — Go/Gin API

### Entry Point: `main.go`

File utama berisi:

1. Load `.env` dan inisialisasi database
2. Dependency injection: `repo → usecase → handler`
3. **WarmCache**: load `loan_products` ke memori saat startup
4. Definisi semua route dan middleware
5. Banyak handler inline (billing export/import, payroll, dll)

### Middleware

| Middleware                | Fungsi                                                                                 |
| ------------------------- | -------------------------------------------------------------------------------------- |
| `CORSMiddleware`          | Allow-all CORS untuk development                                                       |
| `gin.LoggerWithFormatter` | Custom log format `[LMS-API-LOG]` (Output latency berupa angka murni milidetik `%.3f`) |
| `TRACE_LEVEL=0`           | All Logging OFF (Maksimal Throughput)                                                  |
| `TRACE_LEVEL=1`           | HTTP Request Log Only (`[LMS-API-LOG]`, tanpa SQL log)                                 |
| `TRACE_LEVEL=2`           | HTTP Log + SQL `SELECT` Queries Only                                                   |
| `TRACE_LEVEL=3`           | HTTP Log + ALL SQL Queries (`SELECT`, `INSERT`, `UPDATE`, `DELETE`)                    |

### Layer Arsitektur

```
Handler  →  terima HTTP request, validasi, panggil usecase, return JSON
UseCase  →  logika bisnis (eligibility check, kalkulasi, approval workflow)
Repository → akses database (GORM / raw SQL), manage cache
Cache    ->  in-memory KHUSUS untuk global_parameters (ParameterCache)
```

### Daftar API Endpoint

#### Health & Autentikasi

| Method | Endpoint                      | Deskripsi                                                                                       |
| ------ | ----------------------------- | ----------------------------------------------------------------------------------------------- |
| `GET`  | `/api/health`                 | Health check backend                                                                            |
| `POST` | `/api/karisma/login`          | Login (admin / admin123)                                                                        |
| `GET`  | `/api/user-info/:employee_id` | Info user berdasarkan employee_id                                                               |
| `GET`  | `/api/dashboard/summary`      | Summary metriks real-time (Available Limit, Total Hutang, Pinjaman Aktif, & 5 Pinjaman Terbaru) |

##### Kueri & Mesin Kalkulasi Real-Time Dashboard:

1. **Plafon Kredit Maksimum (Credit Limit)**:  
   Mengkalkulasi secara presisi menggunakan **Rumus Resmi LMS** dari Cache `paramRepo.FindByKey("LOAN_LIMIT_FORMULA")` (`(DAY/30) * SALARY * 0.5`) dan dibatasi oleh Batas Maksimum Kategori (`c.max_limit` pada `lms_sch.employee_categories`):
   - **Anggota (Individu)**: `calculateLMSCreditLimitFromCache(employee_id)` $\rightarrow$ Mengevaluasi `LOAN_LIMIT_FORMULA` dari cache dengan parameter `SALARY` dan `DAY`, dibatasi oleh `category.max_limit`.
   - **Admin / HRD (Global)**: Membaca parameter `GLOBAL_CREDIT_LIMIT` langsung dari in-memory cache `paramRepo.FindByKey("GLOBAL_CREDIT_LIMIT")` tanpa kueri SQL berulang-ulang.
   - **Available Credit Limit**: $\text{Credit Limit Terhitung} - \text{Total Sisa Hutang}$.
   - **Logging**: Setiap pemanggilan endpoint ini mencetak log resmi pada console backend:  
     `[DASHBOARD-SUMMARY] UserID: 10101 | Role: 'admin' | HighPriv: true | Formula: 'GLOBAL_CREDIT_LIMIT (CACHE)' | CreditLimit (CL): Rp 5000000000.00 | TotalDebt: Rp 0.00 | AvailableLimit: Rp 5000000000.00`
   - **Sinkronisasi Reaktif Initial Load**: `useEffect` React dikonfigurasi mereaksi perubahan `[activeTab, currentUser, userInfo, roleId, realRoleName]` sehingga data real-time langsung di-fetch otomatis begitu autentikasi user siap tanpa perlu refresh manual.
2. **Total Pinjaman Aktif (Count)**:  
   `SELECT COUNT(*) FROM lms_sch.loan_applications WHERE status IN ('DISBURSED', 'APPROVED')`
3. **Total Sisa Hutang Real-Time (Termasuk Angsuran / Pembayaran Parsial)**:  
   `SELECT COALESCE(SUM(GREATEST(0, s.principal - COALESCE(s.amount_paid, 0))), 0) FROM lms_sch.loan_schedules s JOIN lms_sch.loans l ON s.loan_no = l.loan_no WHERE l.status IN ('DISBURSED', 'APPROVED', 'ACTIVE') AND s.status != 'PAID'`
4. **5 Pinjaman Terbaru**:  
   `SELECT application_no, member_no, product_id, submission_date, requested_amount, tenor, status FROM lms_sch.loan_applications ORDER BY created_at DESC LIMIT 5`

#### Master Data

| Method   | Endpoint                 | Deskripsi                        |
| -------- | ------------------------ | -------------------------------- |
| `GET`    | `/api/master/:table`     | Get semua data tabel master      |
| `POST`   | `/api/master/:table`     | Insert/update data master        |
| `DELETE` | `/api/master/:table/:id` | Soft-delete data master          |
| `GET`    | `/api/products`          | Semua produk **(dari cache)**    |
| `GET`    | `/api/products/:id`      | Produk by ID **(dari cache)**    |
| `POST`   | `/api/products`          | Simpan produk + invalidate cache |
| `PUT`    | `/api/products/:id`      | Update produk + invalidate cache |
| `DELETE` | `/api/products/:id`      | Hapus produk + invalidate cache  |
| `GET`    | `/api/parameters`        | Semua parameter **(dari cache)** |
| `POST`   | `/api/parameters`        | Simpan parameter + refresh cache |

#### Anggota

| Method | Endpoint           | Deskripsi                             |
| ------ | ------------------ | ------------------------------------- |
| `GET`  | `/api/members`     | Paginated search (`?q=&page=&limit=`) |
| `GET`  | `/api/members/all` | Semua anggota tanpa pagination        |

#### Pengajuan Pinjaman & Payroll

| Method | Endpoint                         | Deskripsi                                            |
| ------ | -------------------------------- | ---------------------------------------------------- |
| `GET`  | `/api/applications`              | Semua pengajuan                                      |
| `POST` | `/api/applications`              | Submit pengajuan baru                                |
| `POST` | `/api/applications/:id/approve`  | Setujui/tolak pengajuan                              |
| `POST` | `/api/applications/:id/disburse` | Cairkan pinjaman                                     |
| `POST` | `/api/payroll/export`            | Generate CSV tagihan untuk HRD-Adira                 |
| `POST` | `/api/payroll/import`            | Import & proses hasil potongan dari HRD-Adira        |
| `POST` | `/api/change-password`           | Penggantian Password Pengurus Web LMS                |
| `POST` | `/api/change-pin`                | Penggantian PIN 6-Digit Anggota Mobile EWA           |
| `POST` | `/api/verify-menu-password`      | Verifikasi Password/PIN untuk akses menu terproteksi |

---

## 7. Frontend — React SPA

- Single-page application React Vite di `frontend/src/App.jsx`.
- Modal dynamic Password vs PIN auto-switch & level platform enforcement (Capacitor detection).

---

## 8. Origination & Multi-Loan Credit Limit Engine

- **Origination**: Pengajuan pinjaman mandiri / via Web/Mobile.
- **Credit Limit & Multi-Loan Engine**: `employees.total_loan` dihitung kumulatif ($Pokok + Bunga$) untuk memvalidasi ketersediaan limit sebelum submit pengajuan baru.

---

## 9. Integrasi Billing HRD-Adira & Export/Import

- Export CSV billing ke Adira via `POST /api/payroll/export`.
- Import CSV hasil potongan gaji Adira via `POST /api/payroll/import` (Waterfall & Direct distribution).

---

## 10. Cache Strategy (`cache/parameter_cache.go` & `repositories/parameter_repository.go`)

- **ParameterCache (`cache/parameter_cache.go`)**: **SATU-SATUNYA** in-memory RAM cache di LMS. Memuat seluruh data `lms_sch.global_parameters` ke RAM memori saat backend startup (`cache.ParameterCache.Init`). Pencarian parameter menggunakan O(1) map lookup 0 SQL Query.
- **Pembersihan Cache Non-Global Parameters**: Seluruh caching untuk entitas selain `global_parameters` (`loan_products`, `employees`, `members`, `users`, dan `employee_categories`) telah **dihapus total / dinonaktifkan**. Seluruh entitas tersebut kini dibaca secara real-time langsung dari database PostgreSQL untuk menjamin konsistensi data 100%.

---

## 11. Konfigurasi & Environment (`backend/.env`)

```env
PORT=8086
DB_HOST=localhost
DB_PORT=5433
DB_USER=admin_lms
DB_NAME=lms_db
JWT_SECRET=LMS_K0pK4r4_S3cr3t...
ALLOWED_ORIGINS=http://localhost:3000,https://localhost,capacitor://localhost
HIGH_PRIVILEGE_ROLES=1,3,admin,hrd
```

---

## 12. Folder Billing & Configuration

| Folder            | Parameter Key            | Default                 | Fungsi                                       |
| ----------------- | ------------------------ | ----------------------- | -------------------------------------------- |
| `Billing/Export/` | `FOLDER_BILL_EXPORT`     | `D:\...\Billing\Export` | Output file billing tagihan ke Adira         |
| `Billing/Import/` | `FOLDER_BILL_IMPORT`     | `D:\...\Billing\Import` | Input file billing hasil potongan dari Adira |
| `Billing/BCK/`    | `FOLDER_BILL_IMPORT_BCK` | `D:\...\Billing\BCK`    | Backup file diproses                         |

---

## 13. Global Parameters Kunci

| Parameter                   | Default            | Deskripsi                                                                        |
| --------------------------- | ------------------ | -------------------------------------------------------------------------------- |
| `BILL_FILE_EXPORT_FORMAT`   | `xlsx`             | Format file billing export ke HRD Adira (`xlsx` / `xls` / `csv`)                 |
| `BILL_FILE_IMPORT_FORMAT`   | `xlsx`             | Format file billing import dari HRD Adira (`xlsx` / `xls` / `csv`)               |
| `LOAN_START_PERIOD`         | `1`                | Tanggal awal periode diizinkan pengajuan pinjaman                                |
| `LOAN_END_PERIOD`           | `31`               | Tanggal akhir periode diizinkan pengajuan pinjaman                               |
| `LOAN_APPROVAL_AUTOMATIC`   | `true`             | Mode approval otomatis pengajuan pinjaman                                        |
| `LOAN_DISBURSE_AUTOMATIC`   | `true`             | Mode pencairan otomatis                                                          |
| `LOAN_DUEDATE`              | `25`               | Tanggal jatuh tempo angsuran setiap bulannya                                     |
| `LOAN_MAX_TENOR`            | `12`               | Batas maksimum tenor (bulan) secara global                                       |
| `APPROVAL_MODE`             | `MANUAL`           | Mode persetujuan pinjaman (`MANUAL`: review Admin, `AUTO`: persetujuan otomatis) |
| `PAGINATION_LIMIT`          | `5`                | Jumlah baris data per halaman untuk query tabel master/transaksi                 |
| `LOG_LOAN_TRANSACTION_PATH` | `./logs/loans.log` | Path dan nama file log transaksi seluruh siklus pinjaman                         |

---

## 14. Mobile Self-Service EWA & Autentikasi Security (Selesai ✅)

Sistem LMS telah dilengkapi dengan modul **Mobile Self-Service Registration**, **Autentikasi PIN 6-Digit**, **Dynamic Global Parameters Security**, **Integrasi WhatsApp OTP**, serta **Proteksi Hak Akses Penggantian Password/PIN**.

### 14.1 Arsitektur Pendaftaran Self-Service (4-Factor Matching)

1. **Input 4-Factor Matching**: No. KTP (16-Digit NIK), Employee ID (NIP), Nama Karyawan, No. Handphone / WhatsApp.
2. **Validasi Match ke `lms_sch.employees`**: Backend mencocokkan data pendaftar secara presisi dengan data master HRD.
3. **Verifikasi WhatsApp OTP**: OTP dikirim via Fonnte WA Gateway.
4. **Setup PIN 6-Digit**: Hash Bcrypt disimpan ke `lms_sch.users`.

### 14.2 Parameter Keamanan Dinamis (`lms_sch.global_parameters`)

| Parameter Key                  | Default Value | Deskripsi Fungsi Keamanan                               |
| ------------------------------ | ------------- | ------------------------------------------------------- |
| `PIN_MAX_FAILED_ATTEMPTS`      | `3`           | Batas maksimum salah PIN sebelum akun terkunci          |
| `PIN_LOCKOUT_DURATION_MINUTES` | `15`          | Durasi kuncian akun (menit) jika salah PIN 3x           |
| `PIN_IDLE_TIMEOUT_MINUTES`     | `3`           | Waktu idle (menit) sebelum Mobile App terkunci otomatis |
| `PASSWORD_MAX_FAILED_ATTEMPTS` | `3`           | Batas maksimum salah password Web App                   |
| `PASSWORD_ROTATION_DAYS`       | `90`          | Batas usia wajib rotasi password (hari)                 |
| `PASSWORD_MIN_LENGTH`          | `9`           | Panjang minimum password Web App                        |
| `PWD_MIN_LOWERCASE`            | `1`           | Minimal huruf kecil (a-z)                               |
| `PWD_MIN_UPPERCASE`            | `1`           | Minimal huruf besar (A-Z)                               |
| `PWD_MIN_NUMERIC`              | `1`           | Minimal angka (0-9)                                     |
| `PWD_MIN_SPECIAL`              | `1`           | Minimal karakter spesial (`!@#$%...`)                   |

### 14.3 Script DDL SQL Standalone (`lms_sch.menus` & `global_parameters`)

```sql
ALTER TABLE lms_sch.menus ADD COLUMN IF NOT EXISTS is_password BOOLEAN DEFAULT FALSE;
ALTER TABLE lms_sch.menus ADD COLUMN IF NOT EXISTS notification_type INT DEFAULT 0;
```

### 14.4 Build Android APK (Capacitor & Gradle)

- **Lokasi File APK Output**: `D:\Data_NK\Project5\LMS\app-debug.apk` (~4.47 MB)
- **Lokasi Build Android Project**: `D:\Data_NK\Project5\LMS\frontend\android`
- **Perbaikan Form Submit Modal Pendaftaran (`Un-nested Modal Form`)**: Modal diletakkan di luar tag `<form>` utama untuk menjamin submit mandiri.

### 14.5 Spesifikasi Alur Pendaftaran Resmi 6-Step (`POST /api/karisma/register`)

1. **STEP 1 — Cek Data Karyawan (`employees`)**: Query ke `lms_sch.employees`.
2. **STEP 2 — Penolakan NIK Tidak Ditemukan**: Return `HTTP 400 Bad Request`.
3. **STEP 3 — Pencocokan 4-Faktor Data HRD**: NIK, Phone, KTP, Nama.
4. **STEP 4 — Penolakan Mismatch HRD**: Return `HTTP 400 Bad Request`.
5. **STEP 5 — Pencarian `member_no` Koperasi**: Sync dengan `lms_sch.members`.
6. **STEP 6 — Insert User, Random PIN & Notifikasi WA**: Bcrypt PIN ke `users` & kirim notifikasi Fonnte WA.

### 14.6 Hasil Testing & Verifikasi End-to-End

- Teruji 100% pada endpoint `POST /api/karisma/register` & `POST /api/karisma/login` (`HTTP 200 OK`).

### 14.7 Format Dashboard & Pemindahan Field Rekening Bank

1. **Redesain 3 Card Ringkasan**: Available Limit, Total Hutang, Pinjaman Aktif.
2. **Kartu Profile 8 Field**: Memindahkan kolom bank dari `members` ke `lms_sch.employees`.

### 14.8 Tampilan Single-Screen Compact Dashboard & Layout HP (Selesai ✅)

1. **Sidebar Auto-Collapse**: Disembunyikan di HP (`max-width: 768px`) untuk memberikan 100% full width.
2. **Judul Form & SLA Tracking**: Judul form diubah menjadi **`📝 Pengajuan Pinjaman`** dan SLA Tracking Modal berposisi di dead-center layar.

### 14.9 Refinement Tabel Daftar Pinjaman & Navigasi Paginasi 26 Data (Selesai ✅)

1. **Pembersihan Kolom**: Kolom `Employee ID` & `Catatan HRD` dihilangkan.
2. **Urutan 7 Kolom Baru**: Tanggal Submit, Nominal, Status, Tenor, No. Pengajuan, Tanggal Approval, Aksi.
3. **Navigasi Paginasi 26 Data**: Backend query un-paginated (`limit = 0`) saat limit tidak dikirim, frontend membagi 26 data ke dalam 6 halaman paginasi (`PAGINATION_LIMIT = 5`) dengan navigasi `Prev` / `Next` berjalan 100% aktif.

### 14.10 Refinement UI Login & Modal Registrasi EWA Mobile (Selesai ✅)

1. **UI Login**: Tombol "Ganti Password" diganti menjadi **`Ganti PIN`**, layout link bawah rata kiri `Forgot?` dan rata kanan `Register`.
2. **Modal Registrasi EWA**: Header registrasi tanpa icon HP, placeholder phone `Contoh: 081...`, label NIK KTP 2 baris (`NIK KTP\n(16-DIGIT) *`).

### 14.11 Proteksi Menu Password, Notifikasi & Fitur Ganti Password/PIN (Selesai ✅)

1. **DDL Kolom Menu**: `is_password` (BOOLEAN) dan `notification_type` (INT).
2. **Proteksi Akses Menu**: Akses menu/aksi yang diproteksi `is_password = true` memerlukan verifikasi Password/PIN via `/api/verify-menu-password`.

### 14.12 Refinement Verifikasi Password/PIN Form Pengajuan (Selesai ✅)

1. **Trigger Verifikasi**: Pop-up verifikasi muncul setelah user mengklik tombol **"Kirim Pengajuan"**.
2. **Pembersihan Header & Eye Toggle**: Header modal bersih `Pengajuan Pinjaman`, teks `Masukkan Password / PIN Anda`, & Eye Toggle (_show/hide_).

### 14.13 Perbaikan AuthMiddleware Cookie Public Endpoints & Dual Route Handler (Selesai ✅)

1. **Public Skip List**: Menambahkan `/api/change-password`, `/api/change-pin`, dan `/api/verify-menu-password` ke public skip list `AuthMiddleware()` di `backend/main.go`.
2. **Dual Route Handler**: Mengasosiasikan `changePasswordHandler` dan `changePinHandler` ke kedua router group (`/api/...` dan `/api/karisma/...`) untuk mencegah HTTP 404 Route Mismatch saat dipanggil oleh Web SPA maupun Mobile APK.
3. **Eye Toggle Universal**: Icon mata (_show/hide_) terpasang pada seluruh field PIN/Password aplikasi.

### 14.14 Deteksi Otomatis & Adaptasi Dinamis Teks Label Ganti Password vs Ganti PIN (Selesai ✅)

1. **Auto-Switch Mode**: Mengetik username `admin` atau `hrd` otomatis beralih ke mode `🔑 Ganti Password`.
2. **Adaptasi Dinamis**: Header (`🔑 Ganti Password LMS` vs `🔐 Ganti PIN Mobile EWA`), label field, placeholder, 5/4 syarat parameter keamanan, dan tombol submit (`Simpan Password` vs `Simpan PIN`) beradaptasi 100% dinamis.

### 14.15 Deteksi Username Custom Non-Numeric (User `nur` Role Admin) (Selesai ✅)

1. **Auto-Detect Username Non-Numeric**: Username custom non-numeric (`nur`, `budi`, `superadmin`) otomatis terdeteksi sebagai mode Password.
2. **Role-Based Default**: Sesi login Admin/HRD otomatis mengaktifkan mode Password sejak pertama kali modal dibuka.

### 14.16 Restriksi Platform: Ganti PIN (Web & Mobile) vs Ganti Password (Khusus Web-Apps) (Selesai ✅)

1. **Ganti PIN**: Dapat dilakukan di **Web Application** (Desktop) maupun **Mobile Apps** (Android APK).
2. **Ganti Password**: **Hanya dapat dilakukan di Web Application** (Desktop LMS). Di Mobile Apps, percobaan Ganti Password dibatasi dengan pesan informatif:
   `⚠️ Penggantian Password Pengurus hanya dapat dilakukan melalui Web Application (Desktop LMS). Di Mobile App Anda hanya dapat melakukan Ganti PIN.`

---

### 14.17 Refinement Master User Accounts, Enforce Password Rules & Force Pwd Change (Selesai ✅)

1. **Pembersihan Kolom `updated_user` pada Master Users Table**:
   - Kolom `updated_user` telah di-filter out dari header dan baris tabel Master Users sehingga baris tabel tidak lagi bergeser/terlihat tidak rapi ketika data `updated_user` kosong.
2. **Enforcement 5 Syarat Parameter Keamanan Password pada Master Users**:
   - Pembuatan user baru (_Tambah Data User_) dan pengubahan password (_Edit Data User_) wajib memenuhi 5 parameter keamanan global (`PWD_MIN_LENGTH`, `PWD_MIN_LOWERCASE`, `PWD_MIN_UPPERCASE`, `PWD_MIN_NUMERIC`, `PWD_MIN_SPECIAL`).
   - Modal Master Users dilengkapi dengan **Kotak Checklist Syarat Keamanan Password** dan validasi ganda (Frontend & Backend Handler `MasterDataHandler.Save`).
3. **Fitur Mandatory Password Rotation (`force_pwd_change`)**:
   - Menambahkan field `force_pwd_change` (BOOLEAN DEFAULT FALSE) pada tabel `lms_sch.users` dan checkbox `Force Pwd Change` pada form Master Users.
   - Apabila `force_pwd_change = true`, pengguna yang login akan secara otomatis diarahkan ke modal **🔑 Ganti Password (Wajib)** dan diwajibkan memperbarui password sebelum dapat beraktivitas di dalam sistem.
   - Setelah ganti password berhasil dilakukan, `force_pwd_change` secara otomatis di-reset menjadi `false`.
4. **Script DDL SQL Standalone (`lms_sch.users`)**:
   ```sql
   ALTER TABLE lms_sch.users ADD COLUMN IF NOT EXISTS force_pwd_change BOOLEAN DEFAULT FALSE;
   ```

---

### 14.18 Fix Flow Mandatory Ganti Password, Optimasi Query Username, & Pembersihan Panic Route (Selesai ✅)

1. **Perbaikan Login Flow Forced Password Change (`force_pwd_change`)**:
   - Respon `/api/karisma/verify` kini secara konsisten menyertakan status `"force_pwd_change": user.ForcePwdChange`.
   - Komponen `<ChangePinModal>` kini di-mount di level utama aplikasi (`app-container`). Saat user dengan `force_pwd_change = true` berhasil login, sistem langsung mengunci tampilan dengan modal **🔑 Ganti Password (Wajib)** dan memblokir navigasi LMS sampai password baru berhasil disimpan.
   - Setelah password baru berhasil diperbarui, backend otomatis mengubah `force_pwd_change = false` di DB dan sesi pengguna diperbarui secara realtime.
2. **Optimasi Query Lookup Username Non-Numeric**:
   - Pencarian akun berbasis username string non-numeric (seperti `nur`, `admin`, `hrd`) kini langsung mengeksekusi query bersih tanpa variasi `phone_number`:
     ```sql
     SELECT * FROM "lms_sch"."users"
     WHERE (username = 'nur' OR (member_no IS NOT NULL AND member_no = 0))
       AND deleted_at IS NULL ORDER BY "users"."id" LIMIT 1;
     ```
   - Apabila username tidak ditemukan di database, sistem memberikan notifikasi yang konsisten:
     `Username atau Password salah! Sisa percobaan login: 4 kali.`
3. **Eliminasi Panic Duplicate Route Handler**:
   - Memindahkan definisi endpoint `/api/karisma/logout` dari dalam fungsi handler `/login` ke tingkat utama router registration.
   - Menghilangkan crash panic `handlers are already registered for path '/api/karisma/logout'` saat login gagal atau logout dipanggil.

---

### 14.19 Dukungan Dinamis Format File Billing Export & Import (`xlsx`, `xls`, & `csv`) (Selesai ✅)

1. **Dua Parameter Keuangan Global Baru**:
   - `BILL_FILE_EXPORT_FORMAT`: Menentukan format file tagihan yang digenerate untuk HRD Adira (`xlsx` / `xls` / `csv`). Default: `'xlsx'`.
   - `BILL_FILE_IMPORT_FORMAT`: Menentukan format file hasil rekonsiliasi yang diimpor dari HRD Adira (`xlsx` / `xls` / `csv`). Default: `'xlsx'`.
2. **Dynamic Export Generation (Backend `main.go`)**:
   - Jika `BILL_FILE_EXPORT_FORMAT = 'xlsx'` atau `'xls'`, backend membuat spreadsheet Excel menggunakan pustaka `excelize/v2` lengkap dengan styling header dan kolom data terstruktur (`ADIRA_PAYROLL_KOPKARA_OUTGOING_YYYYMM.xlsx`).
   - Jika `BILL_FILE_EXPORT_FORMAT = 'csv'`, backend membuat file text `.csv` (`ADIRA_PAYROLL_KOPKARA_OUTGOING_YYYYMM.csv`).
3. **Dynamic Import Parser (Frontend `App.jsx`)**:
   - Frontend mendukung pengunggahan dan pembacaan otomatis ketiga jenis file (`.xlsx`, `.xls`, `.csv`).
   - File Excel (.xlsx, .xls) dibaca via `FileReader.readAsArrayBuffer` + `XLSX.read`, sedangkan file CSV dibaca via `FileReader.readAsText`.
   - Label tombol dan `accept` filter input pada UI Rekonsiliasi secara otomatis menyesuaikan nilai parameter (`📥 Export File Payroll (.xlsx)` / `📤 Import Result Rekonsiliasi (.xlsx)`).
4. **Dynamic Auto-Backup & File Relocation (Backend `main.go`)**:
   - Setelah proses pemrosesan import selesai, backend memeriksa ekstensi file asli dan parameter `BILL_FILE_IMPORT_FORMAT`.
   - File `.xlsx`, `.xls`, atau `.csv` yang telah diproses otomatis dipindahkan dari `FOLDER_BILL_IMPORT` ke `FOLDER_BILL_IMPORT_BCK` / `FOLDER_BILL_EXPORT_BCK` dengan ekstensi yang sesuai (misal: `NAME_PROCESSED_YYYYMMDD_HHMMSS.xlsx`).
   - Pesan konfirmasi dan log sistem mengonfirmasi tipe file yang dipindahkan secara presisi (`📦 File XLSX [...] telah otomatis dipindahkan ke folder Backup`).
5. **Standarisasi Kolom Identitas Karyawan (`EMPLOYEE_ID` & `MEMBER_NO`)**:
   - Kolom 1 diubah nama dari `NIK_ADIRA` menjadi `EMPLOYEE_ID` dengan nilainya diisi dari `employees.employee_id` atau `members.employee_id` (misal: `123456100001`).
   - Kolom 2 diubah nama dari `EMPLOYEE_ID` menjadi `MEMBER_NO` dengan nilainya diisi dari `members.member_no` (misal: `200001`).
   - Urutan header resmi file export/import billing: `EMPLOYEE_ID`, `MEMBER_NO`, `LOAN_NO`, `NAMA_KARYAWAN`, `DEPT_NO`, `KODE_POTONGAN`, `NAMA_POTONGAN`, `PERIODE`, `NOMINAL_TAGIHAN`, `NOMINAL_TERPOTONG`, `STATUS_POTONGAN`, `KETERANGAN`, `NO_REFERENSI`.
6. **Optimasi Pembacaan Parameter Export via In-Memory Cache (Zero SQL Overhead)**:
   - Handler `/api/payroll/export` kini sepenuhnya menggunakan `paramRepo.FindByKey()` (in-memory RLock cache) untuk parameter `SCAN_DUEDATE_BILLING`, `BILL_FILE_EXPORT_FORMAT`, dan `FOLDER_BILL_EXPORT`.
   - Mengeliminasi query SQL `SELECT key_value FROM lms_sch.global_parameters` saat mengeksport file billing sehingga mempercepat respon sistem.
7. **Pembaruan Query Backend `/payroll/schedules` & Laporan Rekonsiliasi**:
   - Query backend `/api/payroll/schedules` diperbarui: kolom `nik` kini mengambil `COALESCE(CAST(e.employee_id AS VARCHAR), CAST(m.employee_id AS VARCHAR), CAST(l.member_no AS VARCHAR))`, menggantikan pengambilannya dari nomor rekening bank.
   - Pada tabel UI layar dan cetak PDF laporan rekonsiliasi:
     - Header `NIK Adira` diubah menjadi `NIK` dengan nilainya `employees.employee_id` / `members.employee_id`.
     - Header `Anggota / Karyawan` diubah menjadi `Anggota` dengan nilainya `members.member_no` (misal: `200001`).
   - Tabel `RIWAYAT ADJUSTMENT SELISIH PAYROLL` difilter ketat hanya untuk transaksi yang sedang ditampilkan, dan disembunyikan jika tidak ada adjustment terkait.

### 14.21 Optimasi Alur Login & Eliminasi Query Berulang (Performance Tuning)

1. **Throttling Update Tabel Sessions (`POST /api/karisma/verify`)**:
   - Eksekusi `UPDATE lms_sch.sessions SET last_activity_at = NOW()` kini diberi pengondisian waktu (`time.Since(session.LastActivityAt) > 5*time.Minute`).
   - Mengeliminasi query `UPDATE` berulang ke tabel `sessions` pada setiap verifikasi API / perpindahan layar.
2. **Lazy Loading Reference Data Master (Frontend `App.jsx`)**:
   - Panggilan `fetchReferenceData()` (9 request master parallel: `departments`, `employee-categories`, `employee-statuses`, `kopkara-statuses`, `employees`, `members`, `roles`, `menus`, `role-menus`) telah dihapus dari fungsi verifikasi sesi login (`verifySession()`).
   - Data master kini **hanya di-load secara lazily** saat pengguna secara aktif membuka menu Master Data (`activeTab === 'master'`).
3. **Eliminasi Panggilan API Ganda `GET /api/dashboard/summary`**:
   - Matriks dependensi `useEffect` pada `App.jsx` disederhanakan menjadi `[activeTab, currentUser?.username]`.
4. **Inisialisasi `ParameterCache` RAM Memori (`main.go`)**:
   - `cache.ParameterCache.Init(config.DB)` dipanggil secara otomatis pada saat server startup. Semua panggilan `ParameterCache.Get()` membaca langsung dari RAM memori (0 SQL query).
5. **Perbaikan State `form.member_no` Pengajuan Pinjaman (`App.jsx`)**:
   - Nilai default hardcoded `member_no: 1001` pada state `form` telah dihapus dan disinkronkan secara dinamis dari pengguna yang sedang login (misal: `200001` / `100001`).

### 14.22 Standarisasi Keamanan Password & PIN LIMS (Standard Compliance)

1. **`PWD_ROTATION_DAYS`** _(Default: 90 Hari)_:
   - Batas waktu (hari) password harus diperbarui. Jika `time.Since(lastChanged) > PWD_ROTATION_DAYS`, sistem secara otomatis mengaktifkan `force_pwd_change = true` pada pengguna (`main.go:949`).
2. **`MAX_PASSWORD_ATTEMPTS`** _(Default: 3 Kali)_:
   - Batas maksimal percobaan input password yang salah berturut-turut. Respon HTTP 401 menginfokan sisa percobaan login yang tersisa (`main.go:895`).
3. **`LOGIN_LOCKOUT_MINUTES`** _(Default: 15 Menit)_:
   - Durasi waktu (menit) akun dikunci sementara (`user.LockedUntil`) akibat salah password berturut-turut melebihi `MAX_PASSWORD_ATTEMPTS` (`main.go:858`).
4. **`DEFAULT_IDLE_TIMEOUT_MINUTES`** _(Default: 30 Menit)_:
   - Batas waktu idle sistem (menit) di frontend sebelum sesi pengguna di-logout secara otomatis karena tidak ada aktivitas (`App.jsx:1573`).
5. **`PIN_IDLE_TIMEOUT_MINUTES`** _(Default: 15 Menit)_:
   - Batas waktu idle di latar belakang (menit) sebelum modal Re-Autentikasi PIN 6-digit terkunci secara otomatis (`App.jsx:1573`).
6. **Konsolidasi & Eliminasi Parameter Duplikat**:
   - `PASSWORD_ROTATION_DAYS` $\rightarrow$ di-konsolidasikan ke **`PWD_ROTATION_DAYS`**.
   - `PASSWORD_MAX_FAILED_ATTEMPTS` / `LOGIN_MAX_ATTEMPTS` $\rightarrow$ di-konsolidasikan ke **`MAX_PASSWORD_ATTEMPTS`**.
   - `PASSWORD_MIN_LENGTH` $\rightarrow$ di-konsolidasikan ke **`PWD_MIN_LENGTH`**.

### 14.23 Optimasi Kueri & RAM Parameter Cache (Ultra-Fast Performance)

1. **Khusus `global_parameters` di In-Memory RAM Cache (`cache/parameter_cache.go`)**:
   - Hanya tabel `lms_sch.global_parameters` yang di-load ke RAM cache (`ParameterCache`). Panggilan `ParameterCache.Get()` membaca langsung dari RAM memori (**0 SQL Query**).
   - Seluruh caching selain `global_parameters` (`loan_products`, `users`, `members`, `employees`, `employee_categories`) telah **dihapus total / dinonaktifkan**. Seluruh entitas tersebut kini dibaca secara real-time langsung dari database PostgreSQL untuk menjamin konsistensi data 100%.
2. **Optimasi Dashboard Total Hutang Tanpa JOIN Schedule (<3ms)**:
   - Kueri `totalHutang` pada `GET /api/dashboard/summary` membaca langsung dari `lms_sch.employees` (`total_loan`) atau `lms_sch.loan_applications`, mengeliminasi pemindaian berat (_JOIN_) ke tabel `loan_schedules` dan `loans`.
3. **Komponen Visual Card Available Limit & Total Hutang UI Pengajuan Pinjaman**:
   - Menampilkan kartu ringkas _Available Limit_ & _Total Hutang_ di atas form Pengajuan Pinjaman dengan font nominal diperbesar (`1.25rem`) berwarna hijau dan merah.
4. **Real-Time Refresh Available Limit & Total Hutang Setelah Submit**:
   - `doSubmitApplication` di `App.jsx` secara otomatis memanggil `fetchDashboardSummary()` sesaat setelah pengajuan pinjaman berhasil dikirim untuk meng-update batas limit dan total hutang real-time.

### 14.24 Filter Soft-Delete `role_menus` & Konsistensi Auth Modal Menu

1. **Penyaringan Soft-Delete `lms_sch.role_menus` (`master_data_handler.go:971`)**:
   - Kueri SQL pemuatan menu berdasarkan role menyertakan syarat `AND lms_sch.role_menus.deleted_at IS NULL`, memastikan menu yang di-soft delete tidak lagi muncul di sidebar UI.
2. **Penambahan Field `is_password` pada Response `GetUserPermissions`**:
   - Kolom `is_password` disertakan dalam kueri SELECT `GetUserPermissions` di backend (`master_data_handler.go:963`).
3. **Helper `isPasswordRequired` & Eliminasi Bug Intermiten Modal Auth**:
   - Navigasi menu dan submit aplikasi mendahulukan `userInfo.menus` dengan helper `isPasswordRequired(m)`, menjamin modal verifikasi Password/PIN **100% selalu muncul secara konsisten** saat `is_password = true`.

### 14.25 Perbaikan Soft-Delete `lms_sch.global_parameters` & Tag GORM `MasterBaseModel`

1. **Eliminasi Restriksi Tag `<-:create` pada `DeletedAt` (`base.go:53`)**:
   - Tag GORM `<-:create` pada struct `MasterBaseModel.DeletedAt` membatasi izin penulisan kolom `deleted_at` hanya pada saat `INSERT` (Create).
   - Tag tersebut kini telah dihapus (`DeletedAt gorm.DeletedAt`), sehingga GORM dapat menuliskan timestamp `deleted_at = NOW()` secara sempurna pada operasi soft-delete.
2. **Pembaruan Handler `DeleteMasterData` untuk Tabel `parameters` (`master_data_handler.go:816`)**:
   - Penghapusan parameter global kini secara eksplisit mengisikan `deleted_at = NOW()`, `updated_at = NOW()`, dan `deleted_user = currentUser`.
   - Menjalankan sinkronisasi penghapusan langsung ke RAM memori `cache.ParameterCache.Delete(key)` agar parameter yang di-soft delete langsung terlepas dari memori RAM server.

### 14.26 Optimasi Vite Build Chunking & Eliminasi Warning (`vite.config.js`)

1. **Pemisahan Vendor Bundle (Code Splitting `manualChunks(id)` Function)**:
   - Menggunakan fungsi `manualChunks(id)` pada `rollupOptions.output` yang kompatibel dengan Bundler Vite 8 (Rolldown Engine) untuk memilah modul pihak ketiga:
     - `vendor-react`: `react`, `react-dom` (~140 kB)
     - `vendor-icons`: `lucide-react` (~80 kB)
     - `vendor-axios`: `axios` (~30 kB)
   - Meningkatkan kecepatan pemuatan halaman dan efisiensi browser caching.
2. **Penyesuaian Threshold Chunk Size Warning (`chunkSizeWarningLimit: 1600`)**:
   - Mengatur `chunkSizeWarningLimit: 1600` untuk mengeliminasi peringatan bawaan Vite `(!) Some chunks are larger than 500 kB after minification`.

### 14.27 Implementasi Tabel Master Bank (`lms_sch.banks`) & Alter Skema `employees`

1. **Pembuatan Tabel `lms_sch.banks` (`create_banks_and_alter_employees.sql`)**:
   - Dibuat tabel `lms_sch.banks` dengan kolom: `bank_code` (VARCHAR(15) PK), `bank_name` (VARCHAR(60)), `bank_code_provider` (VARCHAR(15)), serta kolom audit soft delete (`created_at`, `created_user`, `updated_at`, `updated_user`, `deleted_at`, `deleted_user`).
2. **Restrukturisasi Skema `lms_sch.employees`**:
   - Kolom `bank_name` dan `bank_account_name` dihapus dari tabel `employees`.
   - Kolom `bank_code` (VARCHAR(15)) ditambahkan dengan relasi Foreign Key `fk_employees_bank` mengacu pada `lms_sch.banks(bank_code)` (`ON UPDATE CASCADE ON DELETE RESTRICT`).
3. **Backend Struct, Handlers & UseCases (`models/master.go`, `master_data_handler.go`, `application_usecase.go`)**:
   - Dibuat struct `models.Bank` embedding `MasterBaseModel`.
   - Diperbarui struct `models.Employee` (`BankCode` menggantikan `BankName`, `BankAccountName` dihapus).
   - Penyesuaian `application_usecase.go` (`DisburseRequest` & auto-disbursement) dan `GetUserInfo` response menggunakan `emp.BankCode`.
   - Menambahkan handler CRUD lengkap `/api/master/banks` (`GetAll`, `GetOne`, `Save`, `Delete`) dengan dukungan soft delete.
4. **Pembaruan Frontend Master Data UI (`App.jsx`)**:
   - Menambahkan Tab **Master Bank** pada Menu Master Data.
   - Form Master Karyawan kini menampilkan **Dropdown Bank** (`bank_code`) yang bersumber dari tabel `lms_sch.banks` dan mengeliminasi field `bank_account_name`.
5. **Skrip DML Registrasi Menu Bank (`seed_bank_menu.sql`)**:
   - Dibuat file DML [`backend/migrations/seed_bank_menu.sql`](file:///d:/Data_NK/Project5/LMS/backend/migrations/seed_bank_menu.sql) untuk mendaftarkan menu `Master Bank` (`menu_id: 912`, `path: master-banks`) ke `lms_sch.menus` dan memberikan otorisasi ke `lms_sch.role_menus` untuk Admin (`role_id: 1`).
6. **Perbaikan UI Form Reset, Username Audit & Presisi Kolom Tabel (`App.jsx` & `master_data_handler.go`)**:
   - Reset state `masterForm` & `paramForm` saat klik "+ Tambah Data" dan pasca simpan sukses sehingga data lama tidak muncul kembali di modal.
   - Mengisikan `created_user` & `updated_user` pada handler `Save` backend secara otomatis dari username login aktif (`c.GetString("username")`).
   - Menghapus kolom `deleted_user` dari tabel master aktif dan memberlakukan `masterTableKeys` tetap untuk semua baris sehingga posisi kolom **Aksi (Edit/Hapus) 100% Fix & Presisi** tanpa pergeseran sel.

### 14.28 Implementasi Modul Sub-Ledger Kas & Bank, Rekening Operasional, dan Partisi 2026-2027

1. **Skrip Migration DDL/DML (`create_cashbank_tables.sql`)**:
   - Dibuat file SQL [`backend/migrations/create_cashbank_tables.sql`](file:///d:/Data_NK/Project5/LMS/backend/migrations/create_cashbank_tables.sql) yang mencakup:
     - Tabel `lms_sch.cashbank_accounts` (Master Rekening Operasional Kas & Bank Koperasi dengan `initial_balance` dan `current_balance`).
     - Tabel `lms_sch.transaction_types` (`type_code CHAR(5)` PK) + Seed Initial Data (`DISB`, `RPMN`, `RPIP`, `ADJI`, `ADJO`, `FEEB`).
     - Tabel Induk `lms_sch.cashbank_transactions` yang di-partisi bulanan (`PARTITION BY RANGE (transaction_date)`).
     - 24 Tabel Partisi Bulanan `cashbank_transactions_YYYYMM` untuk tahun 2026 dan 2027 serta 1 partisi _Default_.
     - Registrasi menu `Master Rekening Bank` (`menu_id: 913`), `Master Tipe Transaksi` (`menu_id: 914`), dan `Transaksi Kas Bank` (`menu_id: 915`).
2. **Model Struct & Handlers Go Backend (`models/master.go` & `master_data_handler.go`)**:
   - Menambahkan struct `CashBankAccount`, `TransactionType`, dan `CashBankTransaction`.
   - Menambahkan handler CRUD lengkap `/api/master/cashbank-accounts`, `/api/master/transaction-types`, dan `/api/master/cashbank-transactions` dengan dukungan audit soft delete.
3. **Frontend React Master UI (`App.jsx`)**:
   - Menambahkan submenu dan form pengelolaan Master Rekening Bank Operasional, Master Tipe Transaksi, dan Transaksi Kas Bank pada Data Master.

### 14.29 Mekanisme Kontrol Saldo Two-Tier, Atomic Rollback Payment Gateway & Sumber Rekening Operasional

1. **DML Registrasi Menu `cashbank_accounts` (`seed_cashbank_accounts_menu.sql`)**:
   - Berkas DML terpisah [`backend/migrations/seed_cashbank_accounts_menu.sql`](file:///d:/Data_NK/Project5/LMS/backend/migrations/seed_cashbank_accounts_menu.sql) untuk mendaftarkan menu `Master Rekening Bank` (`menu_id: 913`), `Master Tipe Transaksi` (`menu_id: 914`), dan `Transaksi Kas Bank` (`menu_id: 915`) tanpa mengikutsertakan dummy data.
2. **Kontrol Saldo Two-Tier & Mekanisme Rollback Transaksi (Integration dengan Midtrans Iris)**:
   - **Fase 1 (Pre-Validation)**: LMS memeriksa saldo `current_balance` pada `lms_sch.cashbank_accounts`. Jika kurang, transaksi ditolak langsung sebelum API Midtrans dipanggil.
   - **Fase 2 (Execution)**: LMS memanggil API Disbursement Midtrans. Midtrans memeriksa saldo deposit riil Koperasi.
   - **Fase 3 (Atomic Rollback / Commit)**:
     - **Jika Gagal di Midtrans**: Seluruh DB Transaction di LMS di-**ROLLBACK TOTAL**. Saldo `cashbank_accounts` TIDAK dipotong, status pinjaman TIDAK diubah ke `DISBURSED`, dan record `cashbank_transactions` BATAL/DIBATALKAN.
     - **Jika Sukses di Midtrans**: LMS melakukan **COMMIT**. Saldo `cashbank_accounts` dipotong, status pinjaman menjadi `DISBURSED`, dan mutasi `cashbank_transactions` dicatat.
3. **Mekanisme Penentuan Rekening Sumber Operasional & Unique Constraint Index**:
   - Pencairan pinjaman memotong **Rekening Sumber Operasional Koperasi** (`cashbank_accounts.account_number` milik Koperasi), bukan rekening anggota.
   - Anggota menerima dana pada **Rekening Tujuan Anggota** (`employees.bank_account_no` / `disbursement_bank_account_no`).
   - Sistem menentukan Rekening Sumber Operasional melalui parameter default (`DEFAULT_DISBURSEMENT_ACCOUNT_ID`) atau pilihan saat persetujuan pencairan.
   - Kolom `cashbank_accounts.account_number` menggunakan **Partial Unique Index** (`CREATE UNIQUE INDEX uq_cashbank_account_number_active ON lms_sch.cashbank_accounts(account_number) WHERE deleted_at IS NULL;`) untuk menjamin nomor rekening 100% unik tanpa duplikasi pada data aktif.
4. **Pengikatan Otomatis pada Modul Disbursement (`recordCashBankDisbursement`)**:
   - Fungsi `recordCashBankDisbursement` telah diintegrasikan secara otomatis pada `DisburseApplication` ([`backend/usecases/application_usecase.go`](file:///d:/Data_NK/Project5/LMS/backend/usecases/application_usecase.go#L855)).
   - Setiap kali pinjaman berhasil di-disburse (baik auto-disburse maupun manual disburse), sistem secara otomatis:
     - Memotong saldo berjalan `current_balance` pada `lms_sch.cashbank_accounts` sesuai rekening parameter `DEFAULT_DISBURSEMENT_ACCOUNT_ID`.
     - Membentuk record mutasi transaksi baru di `lms_sch.cashbank_transactions` (`type_code: 'DISB'`, `direction: 'OUT'`) yang langsung masuk ke tabel partisi bulanan `cashbank_transactions_YYYYMM`.
5. **Integrasi Seluruh Alur Transaksi ke Sub-Ledger Kas & Bank (`recordCashBankTransactionHelper`)**:
   - **Pelunasan Manual (`/api/payroll/manual-repayment`)**: Secara otomatis mencatat mutasi kas masuk (`type_code: 'RPMN'`, `direction: 'IN'`) dan menambah saldo `current_balance`.
   - **Import File Payroll HRD (`/api/payroll/import`)**: Memproses pencatatan **Opsi A (Detail Mutasi Granular per Karyawan)**. Setiap baris angsuran karyawan yang sukses dipotong dibuatkan 1 record mutasi tersendiri (`type_code: 'RPIP'`, `direction: 'IN'`) lengkap dengan `member_no`, `employee_id`, `reference_no` (`PAY-LoanNo-Period`), dan deskripsi nama karyawan. Saldo `current_balance` ter-update secara sequential.
   - **Adjustment Kas & Bank (`/api/payroll/adjust`)**: Secara otomatis mencatat mutasi koreksi penyesuaian kas masuk/keluar (`type_code: 'ADJI'`/`'ADJO'`, `direction: 'IN'`/`'OUT'`) dan menyesuaikan saldo `current_balance`.
   - **UI & Laporan Audit Kas & Bank**: Menambahkan menu **Transaksi & Mutasi Kas Bank** (`menu_id: 915`) dan **Laporan Audit Kas & Bank** (`menu_id: 916`) pada Data Master UI dengan tampilan badge visual arah kas (`IN`/`OUT`) dan visualisasi saldo berjalan.

### 14.30 Skrip DML Registrasi Menu Transaksi & Mutasi Kas Bank (`seed_cashbank_transactions_menu.sql`)

1. **Skrip DML Registrasi Menu Transaksi & Mutasi Kas Bank (`seed_cashbank_transactions_menu.sql`)**:
   - Dibuat file DML terpisah [`backend/migrations/seed_cashbank_transactions_menu.sql`](file:///d:/Data_NK/Project5/LMS/backend/migrations/seed_cashbank_transactions_menu.sql) untuk mendaftarkan menu `Transaksi & Mutasi Kas Bank` (`menu_id: 915`, `path: master-cashbank-transactions`) dan `Laporan Audit Kas & Bank` (`menu_id: 916`, `path: report-cashbank-transactions`) ke `lms_sch.menus` dan memberikan otorisasi ke `lms_sch.role_menus` untuk Admin (`role_id: 1`) tanpa menyertakan dummy data seeding.

### 14.31 Pembersihan Referensi Kolom `bank_account_name` di Backend & Frontend

1. **Penyebab Error**:
   - Kolom `bank_account_name` sebelumnya telah di-drop dari tabel `lms_sch.employees`, namun pada beberapa query SQL di `backend/main.go` (termasuk pencarian anggota `GET /api/members`, tagihan payroll, dan laporan adjustment) masih terdapat referensi `COALESCE(e.name, e.bank_account_name, ...)`.
2. **Pembaruan Backend (`main.go`)**:
   - Seluruh klausa SQL pada `main.go` telah dibersihkan dari `e.bank_account_name` dan disesuaikan menjadi `COALESCE(e.name, CONCAT('Anggota #', m.member_no))`. Query pencarian anggota pada Form Pelunasan Manual kini berjalan 100% lancar tanpa error `SQLSTATE 42703`.
3. **Pembaruan Frontend (`App.jsx`)**:
   - Menghapus referensi `bank_account_name` pada Form Pelunasan Manual dan dropdown seleksi anggota.

### 14.32 Panduan Audit Transaksi & Rekonsiliasi Kas & Bank (`lms_sch.cashbank_transactions`)

1. **Tujuan & Standar Audit Keuangan**:
   - Sistem LMS Kopkara EWA menerapkan standar pencatatan _Financial Audit Trail_ yang menjamin keterlacakan (_traceability_) 100% dari seluruh arus kas masuk (`IN`) dan arus kas keluar (`OUT`).
   - Setiap pergerakan uang yang terjadi di LMS secara otomatis terhubung ke Rekening Operasional Kas/Bank Koperasi (`lms_sch.cashbank_accounts`) dan tersimpan permanen pada tabel partisi bulanan `lms_sch.cashbank_transactions_YYYYMM`.
2. **Skema Pencatatan Audit Granular (Opsi A)**:
   - **Pencairan Pinjaman (`DISB` / `OUT`)**: Dicatat terpisah per aplikasi pinjaman yang dicairkan. Menyertakan `member_no`, `employee_id`, nominal pencairan bersih, dan nomor referensi aplikasi pinjaman.
   - **Pelunasan Manual (`RPMN` / `IN`)**: Dicatat per transaksi pelunasan mandiri/pesangon/kompensasi. Menyertakan nomor bukti transfer bank, `member_no`, `employee_id`, dan nominal angsuran.
   - **Import File Payroll HRD (`RPIP` / `IN`)**: Menggunakan **Opsi A (Detail Mutasi Granular per Karyawan)**. Setiap baris karyawan dalam file Excel/CSV yang sukses dipotong dibuatkan 1 record mutasi tersendiri. Menyertakan `member_no`, `employee_id`, `reference_no` (`PAY-[loan_no]-[period]`), NIK, serta nama karyawan.
   - **Adjustment / Koreksi Kas (`ADJI` / `ADJO`)**: Dicatat per transaksi koreksi (koreksi terima atau refund keluar) lengkap dengan catatan alasan adjustment dan nomor referensi.
3. **Formulasi Matematis Saldo Berjalan (_Sequential Running Balance_)**:
   - Saldo awal (`balance_before`) dan saldo akhir (`balance_after`) dihitung secara berurutan per transaksi dan langsung memperbarui saldo `current_balance` pada `lms_sch.cashbank_accounts` menggunakan _Direct SQL UPDATE_:
     \[ \text{BalanceAfter} = \begin{cases} \text{BalanceBefore} + \text{Amount} & \text{jika Direction = 'IN'} \\ \text{BalanceBefore} - \text{Amount} & \text{jika Direction = 'OUT'} \end{cases} \]
4. **Karakteristik Kolom Audit pada Table `lms_sch.cashbank_transactions`**:
   | Nama Kolom | Tipe Data | Deskripsi Audit |
   |---|---|---|
   | `transaction_id` | `BIGINT` | Primary Key unik id mutasi transaksi. |
   | `transaction_no` | `VARCHAR(30)` | Kode transaksi unik format `CB-YYYYMM-XXXXXX`. |
   | `transaction_date` | `TIMESTAMP` | Waktu persis mutasi dilakukan (digunakan untuk Partitioning Range). |
   | `bank_code` | `VARCHAR(15)` | Kode bank rekening operasional (misal: `BCA`). |
   | `bank_account_no` | `VARCHAR(30)` | Nomor rekening operasional Koperasi (`6815004314`). |
   | `type_code` | `CHAR(5)` | Kode tipe transaksi (`DISB`, `RPMN`, `RPIP`, `ADJI`, `ADJO`). |
   | `direction` | `CHAR(3)` | Arah arus kas (`IN` = Kas Masuk, `OUT` = Kas Keluar). |
   | `amount` | `NUMERIC(15,2)` | Nominal mutasi transaksi dalam Rupiah. |
   | `balance_before` | `NUMERIC(15,2)` | Saldo kas/bank sebelum transaksi terjadi. |
   | `balance_after` | `NUMERIC(15,2)` | Saldo kas/bank sesudah transaksi terjadi. |
   | `reference_type` | `VARCHAR(50)` | Jenis referensi (`LOAN_APPLICATION`, `REPAYMENT_MANUAL`, `PAYROLL_IMPORT`, `ADJUSTMENT`). |
   | `reference_no` | `VARCHAR(100)` | Nomor referensi unik asal transaksi. |
   | `member_no` | `BIGINT` | Nomor Anggota Koperasi yang terkait. |
   | `employee_id` | `BIGINT` | ID Karyawan yang terkait. |
   | `description` | `TEXT` | Rincian narasi audit transaksi. |
5. **Akses UI & Fitur Laporan Audit Kas & Bank**:
   - Menu **Transaksi & Mutasi Kas Bank** (`menu_id: 915`, path: `master-cashbank-transactions`).
   - Menu **Laporan Audit Kas & Bank** (`menu_id: 916`, path: `report-cashbank-transactions`).
   - Dilengkapi visualisasi badge arah arus kas (`⬇️ IN (Masuk)` dan `⬆️ OUT (Keluar)`), badge tipe transaksi, serta pencarian/filter cepat berdasarkan `member_no`, `employee_id`, atau nomor transaksi.

### 14.33 Panduan Konfigurasi Engine Notifikasi Email (Gmail/SMTP) & WhatsApp

1. **Modul Engine Notifikasi (`services/notification_service.go`)**:
   - Menangani pengiriman notifikasi terpadu Email (via SMTP Direct TLS/SSL) dan WhatsApp (via Fonnte Gateway / Meta WA Cloud API).
   - Pengiriman dilakukan secara **Asinkron (Background Goroutine)** sehingga respon API Backend tetap super cepat (< 50ms).
2. **Daftar Parameter Global Notifikasi di Database (`lms_sch.global_parameters`)**:
   | Nama Parameter | Nilai Default | Deskripsi |
   |---|---|---|
   | `SMTP_HOST` | `smtp.gmail.com` | Host server SMTP Email. |
   | `SMTP_PORT` | `587` | Port server SMTP (587 STARTTLS, 465 SSL, 25 Standard). |
   | `SMTP_USERNAME` | `your-email@gmail.com` | Username/Email pengirim SMTP. |
   | `SMTP_PASSWORD` | `xxxx xxxx xxxx xxxx` | App Password Gmail (16 Karakter) / Password SMTP. |
   | `SMTP_FROM_NAME` | `Kopkara LMS EWA System` | Nama pengirim yang tampil di inbox penerima. |
   | `SMTP_FROM_EMAIL` | `your-email@gmail.com` | Alamat email reply-to pengirim. |
   | `FONNTE_TOKEN` | _(Opsional)_ | Token API Fonnte WhatsApp Gateway. |
   | `NOTIFICATION_ENABLED_CHANNELS` | `EMAIL,WA` | Channel notifikasi aktif global (`EMAIL,WA`, `EMAIL`, `WA`, `OFF`). |
   | `NOTIF_EVENT_DISBURSEMENT` | `EMAIL,WA` | Channel notifikasi event Pencairan Pinjaman. |
   | `NOTIF_EVENT_REPAYMENT` | `EMAIL,WA` | Channel notifikasi event Pelunasan Angsuran. |
   | `NOTIF_EVENT_APPROVAL` | `EMAIL,WA` | Channel notifikasi event Persetujuan/Penolakan Pinjaman. |
3. **Cara Konfigurasi Gmail SMTP (Google Mail)**:
   - **Langkah 1**: Aktifkan **2-Step Verification** pada akun Google Koperasi.
   - **Langkah 2**: Masuk ke [Google App Passwords](https://myaccount.google.com/apppasswords), buat App Password baru untuk "LMS Kopkara", salin kode **16 karakter** (misal: `abcd efgh ijkl mnop`).
   - **Langkah 3**: Buka menu **Pengaturan Parameter Global** di LMS UI, lalu update:
     - `SMTP_HOST` = `smtp.gmail.com`
     - `SMTP_PORT` = `587`
     - `SMTP_USERNAME` = `emailkoperasi@gmail.com`
     - `SMTP_PASSWORD` = `abcdefghijklmnop` (tanpa spasi)
4. **Cara Mengganti dari Gmail SMTP ke Provider SMTP Lainnya**:
   - Untuk mengganti dari Gmail ke Mailtrap, Office 365, Amazon SES, SendGrid, atau cPanel Webmail Koperasi, cukup ubah parameter berikut di menu **Pengaturan Parameter Global** UI tanpa ubah kode program:
     - **Microsoft Office 365 / Outlook**:
       - `SMTP_HOST`: `smtp.office365.com`
       - `SMTP_PORT`: `587`
       - `SMTP_USERNAME`: `admin@kopkara.co.id`
       - `SMTP_PASSWORD`: `password_office365`
     - **Amazon SES (AWS)**:
       - `SMTP_HOST`: `email-smtp.us-east-1.amazonaws.com`
       - `SMTP_PORT`: `587`
       - `SMTP_USERNAME`: `[AWS_SMTP_USERNAME]`
       - `SMTP_PASSWORD`: `[AWS_SMTP_PASSWORD]`
     - **Mailtrap (Testing & Staging)**:
       - `SMTP_HOST`: `sandbox.smtp.mailtrap.io`
       - `SMTP_PORT`: `2525`
       - `SMTP_USERNAME`: `[MAILTRAP_USER]`
       - `SMTP_PASSWORD`: `[MAILTRAP_PASS]`
     - **cPanel / Corporate Webmail Koperasi**:
       - `SMTP_HOST`: `mail.kopkara.co.id`
       - `SMTP_PORT`: `465` (SSL) atau `587` (TLS)
       - `SMTP_USERNAME`: `notifikasi@kopkara.co.id`
       - `SMTP_PASSWORD`: `password_email_koperasi`
5. **Pengaturan Notifikasi Berdasarkan Kolom `lms_sch.menus.notification_type`**:
   - Pengiriman notifikasi dikontrol secara presisi berdasarkan setting kolom `notification_type` pada tabel `lms_sch.menus` (dapat diubah via UI **Master Menu Navigation**):
     - **`0`**: Tidak kirim notifikasi (OFF).
     - **`1`**: Kirim Email saja.
     - **`2`**: Kirim WhatsApp saja.
     - **`3`**: Kirim Email & WhatsApp.
   - _Catatan:_ Opsi notifikasi `4` dan `5` (In-App LMS & Lonceng 🔔) serta tabel `lms_sch.user_notifications` telah **dihapus secara total** untuk mengeliminasi beban polling database dan mengoptimalkan kecepatan sistem.
6. **Dukungan Fleksibilitas File `.env` & Endpoint Diagnostic (`/api/test-email`)**:
   - Backend secara otomatis mendukung variasi nama kunci di file `.env` (`SMTP_USER` / `SMTP_USERNAME` dan `SMTP_PASS` / `SMTP_PASSWORD`), serta mengutamakan nilai `.env` di atas database fallback.
   - **Endpoint Uji Diagnosa Instan**: Sediakan URL diagnosa `GET /api/test-email?to=penerima@domain.com` untuk melakukan pengujian pengiriman email secara sinkron 1-detik dengan pengembalian log status & hint solusi error secara real-time.
   - **Transparansi Perhitungan Finansial Notifikasi Pencairan**: Pesan Notifikasi (Email & WA) menampilkan rincian lengkap:
     - **Nominal Pengajuan Pinjaman (Kotor)**: Misal `Rp 99.000`
     - **Biaya Administrasi (Dipotong)**: Misal `- Rp 30.000`
     - **Nominal Net Transfer (Pencairan)**: Misal `Rp 69.000` (dana bersih yang benar-benar ditransfer ke rekening bank anggota).

### 14.34 Pembaruan Ekspor Laporan PDF, Format Tanggal/Mata Uang, Eliminasi Notifikasi DB & Proteksi Registrasi (Selesai ✅)

1. **Fitur Tombol Cetak PDF Laporan (`📄 Cetak PDF`)**:
   - Menambahkan tombol `📄 Cetak PDF` (warna biru `#0284c7`) pada kontrol header tabel Data Master & Laporan UI ([`App.jsx`](file:///d:/Data_NK/Project5/LMS/frontend/src/App.jsx#L5046)).
   - Memicu fungsi `window.print()` untuk mengeksport/mencetak dokumen laporan PDF atau menyimpan sebagai PDF secara langsung di browser.
2. **Pemformatan Tanggal & Saldo Laporan Mutasi Kas Bank**:
   - Kolom tanggal transaksi (`transaction_date`, `created_at`, `updated_at`) otomatis diformat dari string ISO (misal: `2026-08-31T09:26:23.789766Z`) menjadi format bersih **`dd-mmm-yyyy hh:mm:ss`** (contoh: **`31-Agt-2026 09:26:23`**).
   - Kolom saldo dan nominal (`balance_before`, `balance_after`, `amount`, `salary`, `max_limit`) otomatis diformat menjadi mata uang IDR persis seperti nominal amount (contoh: **`Rp 102.546.000`** & **`Rp 102.476.000`**).
3. **Penghapusan Total Tabel `lms_sch.user_notifications` & `lms_sch.user_device_tokens`**:
   - Mengeliminasi total SQL `INSERT INTO lms_sch.user_notifications` dan interval polling 15 detik di frontend untuk menghilangkan beban CPU/RAM database overhead.
   - Endpoint `/api/notifications`, `/api/device-token`, dan `/api/test-fcm` diubah menjadi no-op statis (< 0.1ms).
   - Tabel `lms_sch.user_notifications` dan `lms_sch.user_device_tokens` secara resmi dapat di-DROP permanen dari PostgreSQL (`DROP TABLE IF EXISTS ... CASCADE`).
4. **Optimasi Credit Limit & Total Hutang Real-Time (`lms_sch.employees`)**:
   - Dashboard summary dan validasi pengajuan membaca data `salary` dan `total_loan` secara real-time langsung dari `lms_sch.employees` tanpa query JOIN berulang (< 1ms).
   - Eliminasi nilai fallback `15000000` menjadi `COALESCE(c.max_limit, 0)` langsung dari `lms_sch.employee_categories`.
   - Format nominal pada Card Dashboard & Form Pengajuan ditampilkan tanpa desimal (`Math.floor`) dan tanpa prefix `"Rp "` agar tampil sebaris horizontal secara rapi (contoh: `938.333`).
5. **Guarded Loan Application Query**:
   - Panggilan `GET /api/applications` tidak dipanggil saat berpindah ke tab Pengajuan Pinjaman (`activeTab === 'pengajuan'`) via guard shield pada `fetchApplications()` dan `verifySession()`.
   - `SimulateApplication` tetap menghitung dan mengembalikan rincian breakdown (Pokok, Bunga, Biaya Admin) secara real-time meskipun nominal melebihi sisa limit, sedangkan penolakan ketat over-limit dieksekusi saat tombol **Kirim Pengajuan** diklik.
6. **Case-Insensitive Match & Proteksi Registrasi User Ganda**:
   - Pencocokan nama pendaftar dengan data HRD menggunakan algoritma case-insensitive (`strings.EqualFold` & `strings.ToLower`), contoh: `"Nur"` vs `"nur"` dianggap **MATCH 100%**.
   - Jika username/HP/Member No sudah ada di database `lms_sch.users`, sistem mengembalikan error HTTP 400 Bad Request: `"Registrasi gagal karena user tsb sudah melakukan registrasi"` tanpa meng-update record user di database.

---

## 15. PANDUAN BUILD APK MOBILE APP ANDROID (CAPACITOR + GRADLE CLI)

Dokumen panduan lengkap pembuatan file **.APK Standalone** HP Android **100% via Linux Command Line (Ubuntu / Debian / WSL CLI)** tanpa perlu membuka Android Studio GUI, sangat ringan dan hemat RAM (laptop 8GB RAM).

### 15.1 Ringkasan Perintah Build Rutin Sehari-hari (Daily Build)

Jika sudah pernah melakukan setup awal dan ingin membuat file `.apk` baru setelah ada perubahan kode UI/React:

```bash
cd /mnt/d/Data_NK/Project5/LMS/frontend
# 1. Build aset web React terbaru
./node_modules/.bin/vite build
# 2. Sinkronkan aset terbaru ke folder android
npx cap sync android
# 3. Pindah folder & kompilasi file APK baru
cd android
# Jika di Linux / WSL CLI (Ubuntu/Debian):
./gradlew assembleDebug
# Jika di Windows Command Prompt (cmd.exe):
gradlew.bat assembleDebug
# Jika di Windows PowerShell:
.\gradlew.bat assembleDebug
```

📍 **Path Hasil Output File APK**:  
`/mnt/d/Data_NK/Project5/LMS/frontend/android/app/build/outputs/apk/debug/app-debug.apk`  
_(Atau `d:\Data_NK\Project5\LMS\frontend\android\app\build\outputs\apk\debug\app-debug.apk` di Windows)_

---

### 15.2 Konfigurasi Firewall Windows Server (1x Setup)

Jalankan sekali di PowerShell (Run as Administrator) agar HP Android lokal dapat berkomunikasi dengan server backend laptop:

```powershell
New-NetFirewallRule -DisplayName "LMS Kopkara-EWA (Ports 8086 & 3005)" -Direction Inbound -LocalPort 8086,3005 -Protocol TCP -Action Allow
```

---

### 15.3 Detail Alur Kompilasi & Frekuensi Eksekusi

| Langkah       | Perintah                                                                                   | Frekuensi              | Keterangan                                   |
| ------------- | ------------------------------------------------------------------------------------------ | ---------------------- | -------------------------------------------- |
| **Firewall**  | `New-NetFirewallRule ... -Action Allow`                                                    | **1x (Admin Windows)** | Membuka Port Inbound 8086 & 3005 di Firewall |
| **Langkah 1** | `npm install @capacitor/core @capacitor/cli @capacitor/android`                            | **1x (Setup Awal)**    | Installing dependensi Capacitor              |
| **Langkah 2** | `npx cap init "Kopkara Mobile EWA" "com.kopkara.ewa.app" --web-dir dist`                   | **1x (Setup Awal)**    | Inisialisasi Nama App & Package ID           |
| **Langkah 3** | `./node_modules/.bin/vite build`                                                           | **Setiap Build Baru**  | Generasi berkas HTML/JS/CSS React di `dist`  |
| **Langkah 4** | `npx cap add android`                                                                      | **1x (Setup Awal)**    | Generasi platform native Android             |
| **Langkah 5** | `npx cap sync android`                                                                     | **Setiap Build Baru**  | Menyalin aset `dist` ke folder `/android`    |
| **Kompilasi** | `cd android && ./gradlew assembleDebug`<br>_(atau `gradlew.bat assembleDebug` di Windows)_ | **Setiap Build Baru**  | Kompilasi file **`app-debug.apk`**           |

---

### 15.4 Setup & Pengujian di HP Android

1. Kirim file **`app-debug.apk`** ke HP Android (via WhatsApp/USB/Drive).
2. Install file `.apk` di HP Android.
3. Buka aplikasi **Kopkara Mobile EWA**.
4. Klik ikon **⚙️ Pengaturan Server API** di pojok kanan atas:
   - Masukkan IP laptop lokal Anda (contoh: `192.168.0.103:8086`).
   - Klik **Tes Koneksi** $\rightarrow$ **Simpan & Terapkan**.
5. Aplikasi Mobile EWA di HP Android siap digunakan dan terhubung 100% Real ke backend LMS.

---

### 15.5 Catatan Log Output Gradle (`BUILD SUCCESSFUL` & Warning)

1. **`BUILD SUCCESSFUL in 1m 47s`**:
   - Status kompilasi **100% SUKSES**! File APK **`app-debug.apk`** telah berhasil dibuat dan siap di-install di HP Android.
2. **`WARNING: Using flatDir should be avoided...`**:
   - Peringatan standar bawaan Capacitor Android saat membaca dependensi plugin Cordova/Capacitor lokal. Peringatan ini **normal dan aman diabaikan**.
3. **Penyebab & Solusi Error `Build-tool 35.0.0 is missing AAPT` / `Build Tools revision 35.0.0 is corrupted`**:
   - **Penyebab**: Menunjuk `sdk.dir` pada `local.properties` ke folder Windows Android SDK (`C:\Users\...\Android\Sdk`) saat menjalankan Gradle dari terminal Linux WSL. Gradle versi Linux mencari biner ELF Linux (`aapt`), sedangkan folder Windows SDK hanya berisi biner Windows (`aapt.exe`).
   - **Solusi**: Pada kompilasi via Linux WSL, biarkan baris `sdk.dir` di-comment out (`# sdk.dir=...`) pada file `frontend/android/local.properties`. Gradle Linux akan secara otomatis menggunakan environment Android SDK Linux dan kompilasi akan **100% SUKSES (`BUILD SUCCESSFUL`)**.
