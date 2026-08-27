# Dokumentasi Sistem LMS (Loan Management System)
**Koperasi Kopkara — Internal Loan & Payroll Deduction System**
**Versi:** 1.0 · **Terakhir diperbarui:** Agustus 2026

---

## Daftar Isi
1. [Gambaran Umum](#1-gambaran-umum)
2. [Arsitektur Sistem](#2-arsitektur-sistem)
3. [Stack Teknologi](#3-stack-teknologi)
4. [Struktur Direktori](#4-struktur-direktori)
5. [Database](#5-database)
6. [Backend — Go/Gin API](#6-backend--gogin-api)
7. [Frontend — React SPA](#7-frontend--react-spa)
8. [Alur Bisnis Utama](#8-alur-bisnis-utama)
9. [Integrasi Billing HRD-Adira](#9-integrasi-billing-hrd-adira)
10. [Cache Strategy](#10-cache-strategy)
11. [Konfigurasi & Environment](#11-konfigurasi--environment)
12. [Folder Billing](#12-folder-billing)
13. [Global Parameters Kunci](#13-global-parameters-kunci)
14. [Mobile Self-Service EWA & Autentikasi Security](#14-mobile-self-service-ewa--autentikasi-security)
    - 14.1 [Arsitektur Pendaftaran Self-Service (4-Factor Matching)](#141-arsitektur-pendaftaran-self-service-4-factor-matching)
    - 14.2 [Parameter Keamanan Dinamis](#142-parameter-keamanan-dinamis-lmsschglobal_parameters)
    - 14.3 [Script DDL SQL Standalone](#143-script-ddl-sql-standalone-alter_table_employees_and_userssql)
    - 14.4 [Build Android APK (Capacitor & Gradle)](#144-build-android-apk-capacitor--gradle)
    - 14.5 [Spesifikasi Alur Pendaftaran Resmi 6-Step](#145-spesifikasi-alur-pendaftaran-resmi-6-step-postapikarismaregister)
    - 14.6 [Hasil Testing & Verifikasi End-to-End](#146-hasil-testing--verifikasi-end-to-end-verified-test-results)
    - 14.7 [Format Dashboard Shopee & Pemindahan Field Rekening Bank](#147-format-dashboard-shopee--pemindahan-field-rekening-bank)
    - 14.8 [Tampilan Single-Screen Compact Dashboard & Penyelesaian Menu EWA](#148-tampilan-single-screen-compact-dashboard--penyelesaian-menu-ewa)
    - 14.9 [Refinement Tabel Daftar Pinjaman & Navigasi Paginasi 26 Data](#149-refinement-tabel-daftar-pinjaman--navigasi-paginasi-26-data)
    - 14.10 [Refinement UI Login & Modal Registrasi EWA Mobile](#1410-refinement-ui-login--modal-registrasi-ewa-mobile)
    - 14.11 [Proteksi Menu Password, Notifikasi & Fitur Ganti Password/PIN](#1411-proteksi-menu-password-notifikasi--fitur-ganti-passwordpin)
    - 14.12 [Refinement Verifikasi Password/PIN Form Pengajuan](#1412-refinement-verifikasi-passwordpin-form-pengajuan)
    - 14.13 [Perbaikan AuthMiddleware Cookie Public Endpoints & Eye Icon Universal](#1413-perbaikan-authmiddleware-cookie-public-endpoints--eye-icon-universal)
    - 14.14 [Deteksi Otomatis & Adaptasi Dinamis Teks Label Ganti Password vs Ganti PIN](#1414-deteksi-otomatis--adaptasi-dinamis-teks-label-ganti-password-vs-ganti-pin)
    - 14.15 [Deteksi Username Custom Non-Numeric (User `nur` Role Admin)](#1415-deteksi-username-custom-non-numeric-user-nur-role-admin)
    - 14.16 [Restriksi Platform: Ganti PIN (Web & Mobile) vs Ganti Password (Khusus Web-Apps)](#1416-restriksi-platform-ganti-pin-web--mobile-vs-ganti-password-khusus-web-apps)

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
┌─────────────────────────────────────────────────────┐
│                  FRONTEND (React)                    │
│          Single Page Application (SPA)               │
│            Port: Vite Dev Server (5173)              │
└─────────────────────┬───────────────────────────────┘
                      │ HTTP REST (axios)
                      ▼
┌─────────────────────────────────────────────────────┐
│              BACKEND (Go / Gin)                      │
│                  Port: 8086                          │
│                                                      │
│  ┌──────────────────────────────────────────────┐   │
│  │  In-Memory Cache                             │   │
│  │  ├─ ProductCache  (loan_products)            │   │
│  │  └─ ParameterCache (global_parameters)       │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
│  Handlers → Usecases → Repositories → GORM          │
└─────────────────────┬───────────────────────────────┘
                      │ GORM / Raw SQL
                      ▼
┌─────────────────────────────────────────────────────┐
│           PostgreSQL Database                        │
│           lms_db / Schema: lms_sch                   │
│           Port: 5433                                 │
└─────────────────────────────────────────────────────┘
```

### Alur Request Standar
```
Browser → GET/POST /api/* → CORSMiddleware → LoggerMiddleware
        → Handler → UseCase → Repository → Cache / DB
        → JSON Response
```

---

## 3. Stack Teknologi

| Komponen | Teknologi | Versi |
|---|---|---|
| Backend Language | Go | 1.23.1 |
| Backend Framework | Gin | 1.9.1 |
| ORM | GORM | 1.25.7 |
| Database Driver | gorm/postgres (pgx) | 1.5.7 |
| Expression Engine | govaluate | 3.0.0 |
| Config | godotenv | 1.5.1 |
| Database | PostgreSQL | — |
| Frontend Framework | React (Vite) | — |
| HTTP Client (FE) | axios | — |

---

## 4. Struktur Direktori

```
D:\Data_NK\Project5\LMS\
│
├── backend/                    # Go API Server
│   ├── main.go                 # Entry point, route definitions, inline handlers
│   ├── .env                    # Environment variables
│   ├── go.mod / go.sum         # Go module dependencies
│   │
│   ├── cache/
│   │   └── product_cache.go    # In-memory cache untuk loan_products
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
│   │   └── product_repository.go     # Includes WarmCache + in-memory cache
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
│
├── frontend/                   # React SPA
│   └── src/
│       ├── App.jsx             # Komponen utama, semua state & UI (~3700 baris)
│       ├── App.css
│       ├── index.css
│       └── main.jsx
│
├── Billing/                    # Folder file CSV billing
│   ├── Export/                 # Output: file tagihan untuk HRD-Adira
│   ├── Import/                 # Input: file hasil potongan dari HRD-Adira
│   └── BCK/                    # Backup file yang sudah diproses
│
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
CREATE ROLE lms_app WITH LOGIN PASSWORD 'Nkl@130200';

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
| Tabel | Deskripsi | Kunci Utama |
|---|---|---|
| `lms_sch.global_parameters` | Parameter konfigurasi sistem (folder, mode scan, dll) | `id` |
| `lms_sch.loan_products` | Produk pinjaman (APM, reguler, dll) | `id` |
| `lms_sch.employee_categories` | Kategori karyawan & limit pinjaman | `category_code` |
| `lms_sch.employees` | Data karyawan dari Karisma/HRD (termasuk `total_loan` untuk validasi credit limit multi-loan) | `employee_id` |
| `lms_sch.members` | Anggota koperasi (mapping employee → member) | `member_no` |
| `lms_sch.employee_statuses` | Referensi status karyawan | `status_code` |
| `lms_sch.departments` | Referensi departemen | `deptno` |
| `lms_sch.kopkara_statuses` | Referensi status anggota koperasi | `status_code` |
| `lms_sch.admin_fees` | Biaya administrasi pinjaman | `id` |
| `lms_sch.roles` | Role pengguna sistem | `role_id` |
| `lms_sch.menus` | Definisi menu/navigasi | `menu_id` |
| `lms_sch.role_menus` | Mapping role ke menu (RBAC) | `role_id, menu_id` |

#### Transaksi
| Tabel | Deskripsi | Kunci Utama |
|---|---|---|
| `lms_sch.loan_applications` | Pengajuan pinjaman (origination) | `application_no` |
| `lms_sch.loan_trackings` | Riwayat perubahan status pengajuan (audit trail) | `id` |
| `lms_sch.loan_contracts` | Kontrak pinjaman setelah disetujui | `contract_no` |
| `lms_sch.loans` | Pinjaman aktif (post-disbursement) | `loan_no` |
| `lms_sch.loan_schedules` | Jadwal angsuran bulanan | `id` |
| `lms_sch.payroll_deductions` | Log potongan payroll per baris CSV | `id` |
| `lms_sch.payroll_adjustments` | Penyesuaian manual potongan payroll | `id` |
| `lms_sch.loan_payroll_import_logs` | Log summary file import | `id` |

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
| Tabel | Kolom | Alasan |
|---|---|---|
| `loan_schedules` | `period` | Performa query billing export/import |
| `members` | `employee_id` | Join dengan tabel employees |
| `loan_trackings` | `application_no` | Filter riwayat per pengajuan |
| `loan_trackings` | `user_id` | Filter riwayat per user |

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
| Middleware | Fungsi |
|---|---|
| `CORSMiddleware` | Allow-all CORS untuk development |
| `gin.LoggerWithFormatter` | Custom log format `[LMS-API-LOG]` (Output latency berupa angka murni milidetik `%.3f`) |
| `TRACE_LEVEL=0` | All Logging OFF (Maksimal Throughput) |
| `TRACE_LEVEL=1` | HTTP Request Log Only (`[LMS-API-LOG]`, tanpa SQL log) |
| `TRACE_LEVEL=2` | HTTP Log + SQL `SELECT` Queries Only |
| `TRACE_LEVEL=3` | HTTP Log + ALL SQL Queries (`SELECT`, `INSERT`, `UPDATE`, `DELETE`) |

### Layer Arsitektur
```
Handler  →  terima HTTP request, validasi, panggil usecase, return JSON
UseCase  →  logika bisnis (eligibility check, kalkulasi, approval workflow)
Repository → akses database (GORM / raw SQL), manage cache
Cache    →  in-memory untuk loan_products dan global_parameters
```

### Daftar API Endpoint

#### Health & Autentikasi
| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET` | `/api/health` | Health check backend |
| `POST` | `/api/karisma/login` | Login (admin / admin123) |
| `GET` | `/api/user-info/:employee_id` | Info user berdasarkan employee_id |
| `GET` | `/api/dashboard/summary` | Summary metriks real-time (Available Limit, Total Hutang, Pinjaman Aktif, & 5 Pinjaman Terbaru) |

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
| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET` | `/api/master/:table` | Get semua data tabel master |
| `POST` | `/api/master/:table` | Insert/update data master |
| `DELETE` | `/api/master/:table/:id` | Soft-delete data master |
| `GET` | `/api/products` | Semua produk **(dari cache)** |
| `GET` | `/api/products/:id` | Produk by ID **(dari cache)** |
| `POST` | `/api/products` | Simpan produk + invalidate cache |
| `PUT` | `/api/products/:id` | Update produk + invalidate cache |
| `DELETE` | `/api/products/:id` | Hapus produk + invalidate cache |
| `GET` | `/api/parameters` | Semua parameter **(dari cache)** |
| `POST` | `/api/parameters` | Simpan parameter + refresh cache |

#### Anggota
| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET` | `/api/members` | Paginated search (`?q=&page=&limit=`) |
| `GET` | `/api/members/all` | Semua anggota tanpa pagination |

#### Pengajuan Pinjaman & Payroll
| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET` | `/api/applications` | Semua pengajuan |
| `POST` | `/api/applications` | Submit pengajuan baru |
| `POST` | `/api/applications/:id/approve` | Setujui/tolak pengajuan |
| `POST` | `/api/applications/:id/disburse` | Cairkan pinjaman |
| `POST` | `/api/payroll/export` | Generate CSV tagihan untuk HRD-Adira |
| `POST` | `/api/payroll/import` | Import & proses hasil potongan dari HRD-Adira |
| `POST` | `/api/change-password` | Penggantian Password Pengurus Web LMS |
| `POST` | `/api/change-pin` | Penggantian PIN 6-Digit Anggota Mobile EWA |
| `POST` | `/api/verify-menu-password` | Verifikasi Password/PIN untuk akses menu terproteksi |

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

## 10. Cache Strategy (`cache/product_cache.go` & `repositories/parameter_repository.go`)
- **ProductCache**: Global singleton thread-safe `sync.RWMutex`, warm-up saat backend startup.
- **ParameterCache**: Global parameter O(1) map lookup.

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

| Folder | Parameter Key | Default | Fungsi |
|---|---|---|---|
| `Billing/Export/` | `FOLDER_BILL_EXPORT` | `D:\...\Billing\Export` | Output file billing tagihan ke Adira |
| `Billing/Import/` | `FOLDER_BILL_IMPORT` | `D:\...\Billing\Import` | Input file billing hasil potongan dari Adira |
| `Billing/BCK/` | `FOLDER_BILL_IMPORT_BCK` | `D:\...\Billing\BCK` | Backup file diproses |

---

## 13. Global Parameters Kunci

| Parameter | Default | Deskripsi |
|---|---|---|
| `BILL_FILE_EXPORT_FORMAT` | `xlsx` | Format file billing export ke HRD Adira (`xlsx` / `csv`) |
| `BILL_FILE_IMPORT_FORMAT` | `xlsx` | Format file billing import dari HRD Adira (`xlsx` / `csv`) |
| `LOAN_START_PERIOD` | `1` | Tanggal awal periode diizinkan pengajuan pinjaman |
| `LOAN_END_PERIOD` | `31` | Tanggal akhir periode diizinkan pengajuan pinjaman |
| `LOAN_APPROVAL_AUTOMATIC` | `true` | Mode approval otomatis pengajuan pinjaman |
| `LOAN_DISBURSE_AUTOMATIC` | `true` | Mode pencairan otomatis |
| `LOAN_DUEDATE` | `25` | Tanggal jatuh tempo angsuran setiap bulannya |
| `LOAN_MAX_TENOR` | `12` | Batas maksimum tenor (bulan) secara global |
| `APPROVAL_MODE` | `MANUAL` | Mode persetujuan pinjaman (`MANUAL`: review Admin, `AUTO`: persetujuan otomatis) |
| `PAGINATION_LIMIT` | `5` | Jumlah baris data per halaman untuk query tabel master/transaksi |
| `LOG_LOAN_TRANSACTION_PATH` | `./logs/loans.log` | Path dan nama file log transaksi seluruh siklus pinjaman |

---

## 14. Mobile Self-Service EWA & Autentikasi Security (Selesai ✅)

Sistem LMS telah dilengkapi dengan modul **Mobile Self-Service Registration**, **Autentikasi PIN 6-Digit**, **Dynamic Global Parameters Security**, **Integrasi WhatsApp OTP**, serta **Proteksi Hak Akses Penggantian Password/PIN**.

### 14.1 Arsitektur Pendaftaran Self-Service (4-Factor Matching)
1. **Input 4-Factor Matching**: No. KTP (16-Digit NIK), Employee ID (NIP), Nama Karyawan, No. Handphone / WhatsApp.
2. **Validasi Match ke `lms_sch.employees`**: Backend mencocokkan data pendaftar secara presisi dengan data master HRD.
3. **Verifikasi WhatsApp OTP**: OTP dikirim via Fonnte WA Gateway.
4. **Setup PIN 6-Digit**: Hash Bcrypt disimpan ke `lms_sch.users`.

### 14.2 Parameter Keamanan Dinamis (`lms_sch.global_parameters`)
| Parameter Key | Default Value | Deskripsi Fungsi Keamanan |
|---|---|---|
| `PIN_MAX_FAILED_ATTEMPTS` | `3` | Batas maksimum salah PIN sebelum akun terkunci |
| `PIN_LOCKOUT_DURATION_MINUTES` | `15` | Durasi kuncian akun (menit) jika salah PIN 3x |
| `PIN_IDLE_TIMEOUT_MINUTES` | `3` | Waktu idle (menit) sebelum Mobile App terkunci otomatis |
| `PASSWORD_MAX_FAILED_ATTEMPTS` | `3` | Batas maksimum salah password Web App |
| `PASSWORD_ROTATION_DAYS` | `90` | Batas usia wajib rotasi password (hari) |
| `PASSWORD_MIN_LENGTH` | `9` | Panjang minimum password Web App |
| `PWD_MIN_LOWERCASE` | `1` | Minimal huruf kecil (a-z) |
| `PWD_MIN_UPPERCASE` | `1` | Minimal huruf besar (A-Z) |
| `PWD_MIN_NUMERIC` | `1` | Minimal angka (0-9) |
| `PWD_MIN_SPECIAL` | `1` | Minimal karakter spesial (`!@#$%...`) |

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

### 14.7 Format Dashboard Shopee & Pemindahan Field Rekening Bank
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
2. **Pembersihan Header & Eye Toggle**: Header modal bersih `Pengajuan Pinjaman`, teks `Masukkan Password / PIN Anda`, & Eye Toggle (*show/hide*).

### 14.13 Perbaikan AuthMiddleware Cookie Public Endpoints & Dual Route Handler (Selesai ✅)
1. **Public Skip List**: Menambahkan `/api/change-password`, `/api/change-pin`, dan `/api/verify-menu-password` ke public skip list `AuthMiddleware()` di `backend/main.go`.
2. **Dual Route Handler**: Mengasosiasikan `changePasswordHandler` dan `changePinHandler` ke kedua router group (`/api/...` dan `/api/karisma/...`) untuk mencegah HTTP 404 Route Mismatch saat dipanggil oleh Web SPA maupun Mobile APK.
3. **Eye Toggle Universal**: Icon mata (*show/hide*) terpasang pada seluruh field PIN/Password aplikasi.

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
   - Pembuatan user baru (*Tambah Data User*) dan pengubahan password (*Edit Data User*) wajib memenuhi 5 parameter keamanan global (`PWD_MIN_LENGTH`, `PWD_MIN_LOWERCASE`, `PWD_MIN_UPPERCASE`, `PWD_MIN_NUMERIC`, `PWD_MIN_SPECIAL`).
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

### 14.19 Dukungan Dinamis Format File Billing Export & Import (`xlsx` & `csv`) (Selesai ✅)
1. **Dua Parameter Keuangan Global Baru**:
   - `BILL_FILE_EXPORT_FORMAT`: Menentukan format file tagihan yang digenerate untuk HRD Adira (`xlsx` / `csv`). Default: `'xlsx'`.
   - `BILL_FILE_IMPORT_FORMAT`: Menentukan format file hasil rekonsiliasi yang diimpor dari HRD Adira (`xlsx` / `csv`). Default: `'xlsx'`.
2. **Dynamic Export Generation (Backend `main.go`)**:
   - Jika `BILL_FILE_EXPORT_FORMAT = 'xlsx'`, backend membuat spreadsheet `.xlsx` menggunakan pustaka `excelize/v2` lengkap dengan styling header dan kolom data terstruktur (`ADIRA_PAYROLL_KOPKARA_OUTGOING_YYYYMM.xlsx`).
   - Jika `BILL_FILE_EXPORT_FORMAT = 'csv'`, backend membuat file text `.csv` (`ADIRA_PAYROLL_KOPKARA_OUTGOING_YYYYMM.csv`).
3. **Dynamic Import Parser (Frontend `App.jsx`)**:
   - Frontend mendukung pengunggahan dan pembacaan otomatis kedua jenis file (`.xlsx`, `.xls`, `.csv`).
   - File Excel dibaca via `FileReader.readAsArrayBuffer` + `XLSX.read`, sedangkan file CSV dibaca via `FileReader.readAsText`.
   - Label tombol dan `accept` filter input pada UI Rekonsiliasi secara otomatis menyesuaikan nilai parameter (`📥 Export File Payroll (.xlsx)` / `📤 Import Result Rekonsiliasi (.xlsx)`).
