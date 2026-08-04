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

### Daftar Tabel

#### Master Data
| Tabel | Deskripsi | Kunci Utama |
|---|---|---|
| `lms_sch.global_parameters` | Parameter konfigurasi sistem (folder, mode scan, dll) | `id` |
| `lms_sch.loan_products` | Produk pinjaman (APM, reguler, dll) | `id` |
| `lms_sch.employee_categories` | Kategori karyawan & limit pinjaman | `category_code` |
| `lms_sch.employees` | Data karyawan dari Karisma/HRD | `employee_id` |
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
| `gin.LoggerWithFormatter` | Custom log format `[LMS-API-LOG]` |
| `TRACE_LEVEL=3` | Log SQL detail via GORM |

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

#### Pengajuan Pinjaman
| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET` | `/api/applications` | Semua pengajuan |
| `POST` | `/api/applications/simulate` | Simulasi cicilan |
| `POST` | `/api/applications` | Submit pengajuan baru |
| `POST` | `/api/applications/:id/approve` | Setujui/tolak pengajuan |
| `POST` | `/api/applications/:id/disburse` | Cairkan pinjaman |
| `GET` | `/api/applications/:id/trackings` | Riwayat status |

#### Payroll & Billing
| Method | Endpoint | Deskripsi |
|---|---|---|
| `GET` | `/api/payroll/schedules` | Jadwal angsuran aktif |
| `POST` | `/api/payroll/export` | Generate CSV tagihan untuk HRD-Adira |
| `POST` | `/api/payroll/import` | Import & proses hasil potongan dari HRD-Adira |
| `POST` | `/api/payroll/manual-repayment` | Pelunasan manual |
| `GET` | `/api/payroll/deductions` | Log history potongan |

---

## 7. Frontend — React SPA

### Deskripsi
SPA berbasis React/Vite. Seluruh state dan UI dikelola dalam satu file `App.jsx` (~3700 baris).

### Menu Utama
| Menu | Deskripsi |
|---|---|
| Dashboard | Ringkasan status pinjaman dan anggota |
| Pengajuan Pinjaman | Form simulasi dan submit |
| Persetujuan | Approval workflow |
| Payroll & Billing | Export/Import CSV, monitor jadwal |
| Pelunasan Manual | Form pelunasan di luar payroll |
| Master Data | Kelola anggota, karyawan, produk, parameter |

### State Management
- Semua state dengan React Hooks (`useState`, `useRef`)
- `memberLoadingRef` — lock untuk mencegah concurrent duplicate API call
- Debounce 300ms pada pencarian anggota
- Dropdown anggota: **lazy load on-demand** (tidak load saat buka halaman)

---

## 8. Alur Bisnis Utama

Bagian ini menjelaskan setiap proses bisnis dalam LMS secara naratif, dari sudut pandang operasional maupun teknis.

---

### 8.1 Siklus Pengajuan Pinjaman (Origination)

Status pengajuan mengikuti alur berikut secara berurutan:

| Status | Keterangan |
|---|---|
| `SUBMITTED` | Pengajuan baru masuk, menunggu validasi |
| `PENDING` | Menunggu review admin |
| `REVIEWED` | Sedang diproses oleh admin/komite |
| `APPROVED` | Disetujui, menunggu pencairan |
| `REJECTED` | Ditolak, proses selesai |
| `REVISION_REQUIRED` | Perlu revisi dari pemohon |
| `DISBURSED` | Dana sudah dicairkan, pinjaman aktif |
| `CLOSED` | Semua angsuran lunas |

```
SUBMITTED → PENDING → REVIEWED → APPROVED → DISBURSED → CLOSED
                              ↘ REJECTED
                              ↘ REVISION_REQUIRED
```

**Narasi Proses:**

Siklus dimulai ketika seorang **anggota koperasi mengajukan pinjaman** melalui UI LMS. Petugas atau anggota mengisi form dengan memilih produk pinjaman, memasukkan jumlah yang diminta, dan tenor (bulan). Sebelum pengajuan disimpan, sistem otomatis **menjalankan simulasi** untuk menghitung besaran cicilan bulanan, biaya admin, dan batas kredit.

---

### 8.2 Syarat Pengajuan Pinjaman

Setelah form disubmit (`POST /api/applications`), sistem memvalidasi beberapa syarat secara otomatis sebelum pengajuan diterima:

**1. Harus Terdaftar sebagai Anggota Koperasi**
Pemohon harus terdaftar di tabel `lms_sch.members` dengan `employee_id` yang valid. Jika tidak ditemukan, pengajuan langsung ditolak dengan pesan *"member not found"*.

**2. Harus Terdaftar sebagai Karyawan Aktif**
Data anggota harus terhubung ke tabel `lms_sch.employees` yang berisi informasi kategori dan gaji. Ini diperlukan untuk menghitung credit limit dan eligibility.

**3. Tanggal Pengajuan dalam Periode yang Diizinkan**
Pengajuan hanya boleh dilakukan antara tanggal yang dikonfigurasi di `global_parameters`:
- `LOAN_START_PERIOD` — tanggal mulai (misalnya: `1`)
- `LOAN_END_PERIOD` — tanggal akhir (misalnya: `15`)

Jika pengajuan dilakukan di luar rentang tanggal tersebut, sistem menolak dengan pesan *"pengajuan hanya diizinkan antara tanggal X sampai Y"*.

**4. Tenor Tidak Melebihi Batas Maksimum**
Sistem mengambil batas tenor dari dua sumber: `loan_products.max_tenor_months` dan parameter `LOAN_MAX_TENOR` di `global_parameters`. Nilai **terkecil** dari keduanya yang dipakai sebagai batas.

**5. Jumlah Pinjaman Tidak Melebihi Credit Limit**
Credit limit dihitung berdasarkan formula yang dikonfigurasi di `LOAN_LIMIT_FORMULA` (lihat sub-bab 8.3). Jika jumlah yang diminta melebihi limit, sistem menolak dengan pesan beserta nilai limit maksimum.

---

### 8.3 Kalkulasi Cicilan, Biaya Admin & Credit Limit

Semua formula kalkulasi **dapat dikonfigurasi** melalui tabel `lms_sch.global_parameters` — tidak perlu mengubah kode program.

#### Parameter Global yang Mengontrol Kalkulasi

| Key Name | Contoh Nilai | Fungsi |
|---|---|---|
| `LOAN_ADMIN_FORMULA` | `REQUESTED_AMOUNT * 0.01` | Formula biaya administrasi |
| `LOAN_MIN_ADMIN_FEE` | `30000` | Biaya admin minimum (floor) |
| `LOAN_LIMIT_FORMULA` | `(DAY/30) * SALARY * 0.5` | Formula credit limit |
| `LOAN_MAX_TENOR` | `36` | Maksimum tenor (bulan) |
| `LOAN_START_PERIOD` | `1` | Tanggal mulai periode pengajuan |
| `LOAN_END_PERIOD` | `15` | Tanggal akhir periode pengajuan |
| `LOAN_DUEDATE` | `25` | Tanggal jatuh tempo cicilan tiap bulan |

#### Variabel yang Tersedia dalam Formula

Formula mendukung ekspresi matematis menggunakan library **govaluate** dengan variabel:

| Variabel | Nilai |
|---|---|
| `REQUESTED_AMOUNT` | Jumlah pinjaman yang diminta |
| `SALARY` | Gaji bersih karyawan |
| `TENOR` | Tenor dalam bulan |
| `DAY` | Tanggal hari ini (1–31) |

#### Rumus Kalkulasi

**Biaya Administrasi (Admin Fee):**
```
admin_fee = evaluasi(LOAN_ADMIN_FORMULA)
Jika admin_fee < LOAN_MIN_ADMIN_FEE → admin_fee = LOAN_MIN_ADMIN_FEE
```

Contoh: `LOAN_ADMIN_FORMULA = "REQUESTED_AMOUNT * 0.01"`, `LOAN_MIN_ADMIN_FEE = "30000"`
- Pinjaman Rp 2.000.000 → admin_fee = 20.000 → karena < 30.000, dipakai 30.000
- Pinjaman Rp 10.000.000 → admin_fee = 100.000

**Cicilan Bulanan (Flat Rate):**
```
principal_per_month = approved_amount / tenor
interest_per_month  = approved_amount × (interest_rate / 100)
total_installment   = principal_per_month + interest_per_month
```
> **Catatan:** `interest_rate` diambil dari `lms_sch.loan_products.interest_rate` per bulan (sudah dalam satuan persen per bulan, bukan per tahun).

**Total Biaya Pinjaman:**
```
total_loan_cost = (total_installment × tenor) + admin_fee
disbursement_amount = approved_amount − admin_fee
```
> Dana yang benar-benar diterima anggota adalah `disbursement_amount` (setelah dipotong admin fee).

**Credit Limit:**
```
credit_limit = evaluasi(LOAN_LIMIT_FORMULA)
Jika credit_limit > employee_categories.max_limit → credit_limit = max_limit
Jika requested_amount > credit_limit → TOLAK
```

---

### 8.4 Contoh Simulasi Nyata

#### Pengaruh Tanggal Pengajuan terhadap Credit Limit

Variabel `DAY` (tanggal hari pengajuan, nilai 1–31) tersedia dalam formula dan **dapat mempengaruhi credit limit**. Ini memungkinkan konfigurasi limit yang berbeda tergantung kapan dalam sebulan anggota mengajukan pinjaman.

Contoh formula yang menggunakan `DAY`:
```
LOAN_LIMIT_FORMULA = "(DAY/30) * SALARY * 0.5"
```
Artinya: credit limit dihitung secara proporsional berdasarkan tanggal pengajuan dibandingkan dengan jumlah hari dalam satu bulan (asumsi 30 hari), dikalikan gaji dan rasio 50%.

---

#### Skenario A — Pengajuan Awal Bulan (Tanggal 4)

**Data pengajuan:**
- Tanggal submit: **4 Agustus 2026** (DAY = 4)
- Jumlah pinjaman: Rp 600.000
- Tenor: 1 bulan, bunga produk: 1% per bulan
- Gaji karyawan: Rp 10.000.000

**Parameter global aktif:**
| Parameter | Nilai |
|---|---|
| `LOAN_START_PERIOD` | `1` |
| `LOAN_END_PERIOD` | `15` |
| `LOAN_LIMIT_FORMULA` | `(DAY/30) * SALARY * 0.5` |
| `LOAN_ADMIN_FORMULA` | `REQUESTED_AMOUNT * 0.01` |
| `LOAN_MIN_ADMIN_FEE` | `30000` |
| `LOAN_DUEDATE` | `25` |

**Kalkulasi:**
```
[Cek Tanggal]
DAY = 4, LOAN_START_PERIOD = 1, LOAN_END_PERIOD = 15
1 ≤ 4 ≤ 15 → ✓ Tanggal pengajuan valid

[Credit Limit]
credit_limit = (4 / 30) × 10.000.000 × 0.5 = 666.666,67
requested_amount (600.000) ≤ credit_limit (666.666,67) → ✓ Lolos

[Cicilan]
principal_per_month = 600.000 / 1          = 600.000
interest_per_month  = 600.000 × 1%          = 6.000
total_installment   = 600.000 + 6.000        = 606.000

[Admin Fee]
admin_fee = 600.000 × 1% = 6.000 (< min 30.000 → pakai 30.000)

[Total]
total_loan_cost     = (606.000 × 1) + 30.000   = 636.000
disbursement_amount = 600.000 - 30.000         = 570.000

[Jadwal Cicilan]
Due date pertama: 25 September 2026 (LOAN_DUEDATE = 25, +1 bulan dari pencairan)
```

**Hasil Simulasi — Skenario A:**
| Komponen | Nilai |
|---|---|
| Tanggal Pengajuan | 4 Agustus 2026 (DAY=4) |
| Jumlah Pinjaman | Rp 600.000 |
| Credit Limit | Rp 666.666,67 |
| Dana Diterima | Rp 570.000 |
| Biaya Admin | Rp 30.000 |
| Tenor | 1 bulan |
| Cicilan Bulanan | Rp 606.000 |
| Total yang Dibayar | Rp 636.000 |
| Suku Bunga | 1% / bulan |
| Due Date Pertama | 25 September 2026 |

---

#### Skenario B — Pengajuan Akhir Periode (Tanggal 15)

**Data pengajuan:**
- Tanggal submit: **15 Agustus 2026** (DAY = 15)
- Jumlah pinjaman: Rp 600.000 (sama)

**Kalkulasi credit limit:**
```
[Cek Tanggal]
DAY = 15, LOAN_START_PERIOD = 1, LOAN_END_PERIOD = 15
1 ≤ 15 ≤ 15 → ✓ Tanggal pengajuan valid (batas akhir)

[Credit Limit]
credit_limit = (15 / 30) × 10.000.000 × 0.5 = 2.500.000
requested_amount (600.000) ≤ credit_limit (2.500.000) → ✓ Lolos
```

> Skenario C — Tanggal 16 Agustus: `DAY = 16 > LOAN_END_PERIOD (15)` → **Pengajuan DITOLAK otomatis** dengan pesan *"pengajuan hanya diizinkan antara tanggal 1 sampai 15"*.

---

### 8.5 Proses Approval

**Narasi Proses:**

Admin membuka menu **Persetujuan** dan melihat daftar pengajuan berstatus `PENDING` atau `REVIEWED`. Untuk setiap pengajuan, admin dapat:

- **Menyetujui** (`APPROVED`): Memasukkan jumlah yang disetujui (bisa berbeda dari yang diminta) dan catatan wajib. Sistem mengupdate status `loan_applications` dan mencatat ke `loan_trackings`.
- **Menolak** (`REJECTED`): Memasukkan alasan penolakan. Status berubah menjadi `REJECTED` dan tidak bisa diproses lebih lanjut.
- **Minta Revisi** (`REVISION_REQUIRED`): Mengembalikan ke pemohon untuk perbaikan data.

Setiap aksi approval menyimpan informasi ke `loan_trackings`: siapa yang melakukan (`updated_user`), kapan (`action_date`), durasi SLA dari status sebelumnya (`sla_duration`), IP address, dan user agent browser.

---

### 8.6 Pencairan Pinjaman (Disbursement)

**Narasi Proses:**

Setelah pengajuan berstatus `APPROVED`, petugas melakukan pencairan (`POST /api/applications/:id/disburse`). Sistem otomatis membuat tiga record sekaligus:

1. **`loan_contracts`** — kontrak formal pinjaman (jumlah, tenor, cicilan, suku bunga, tanggal kontrak)
2. **`loans`** — record pinjaman aktif. `outstanding_amount` = `approved_amount` (saldo penuh)
3. **`loan_schedules`** — jadwal cicilan per bulan dari tenor. Tanggal jatuh tempo (`due_date`) ditentukan oleh parameter `LOAN_DUEDATE` (default tanggal 25)

---

### 8.7 Siklus Penagihan Payroll Bulanan

**Narasi Proses:**

Setiap bulan, LMS berinteraksi dengan sistem HRD Adira (Karisma) melalui pertukaran file CSV dalam dua tahap.

**Tahap 1 — Export Tagihan (LMS → Adira):**
Petugas membuka menu Payroll & Billing dan klik **Export Billing**. LMS men-generate file CSV berisi semua angsuran `UNPAID`/`PARTIAL` yang jatuh tempo periode tersebut. File disimpan otomatis ke `Billing/Export/` dan bisa diunduh dari UI.

**Format file export (`ADIRA_PAYROLL_KOPKARA_OUTGOING_YYYYMM.csv`):**

| Kolom | Isi | Contoh |
|---|---|---|
| `NIK_ADIRA` | Nomor rekening bank anggota | `3171012345670001` |
| `EMPLOYEE_ID` | ID karyawan di Adira | `1001` |
| `LOAN_NO` | Nomor pinjaman LMS | `1785812731738` |
| `NAMA_KARYAWAN` | Nama lengkap | `Fairuz Agha` |
| `DEPT_NO` | Kode departemen | `IT-01` |
| `KODE_POTONGAN` | Kode tetap | `POT_KOPKARA` |
| `NAMA_POTONGAN` | Keterangan tetap | `Potongan Angsuran Kopkara` |
| `PERIODE` | Bulan tagihan | `2026-08` |
| `NOMINAL_TAGIHAN` | Jumlah cicilan yang harus dipotong | `1580000` |
| `NOMINAL_TERPOTONG` | Dikosongkan (diisi Adira) | *(kosong)* |
| `STATUS_POTONGAN` | Dikosongkan (diisi Adira) | *(kosong)* |
| `KETERANGAN` | Dikosongkan (diisi Adira) | *(kosong)* |
| `NO_REFERENSI` | ID unik referensi | `LMS-PAY-202608-1001` |

**Contoh isi file export:**
```csv
NIK_ADIRA,EMPLOYEE_ID,LOAN_NO,NAMA_KARYAWAN,DEPT_NO,KODE_POTONGAN,NAMA_POTONGAN,PERIODE,NOMINAL_TAGIHAN,NOMINAL_TERPOTONG,STATUS_POTONGAN,KETERANGAN,NO_REFERENSI
3171012345670001,1001,1785812731738,Fairuz Agha,IT-01,POT_KOPKARA,Potongan Angsuran Kopkara,2026-08,1580000.00,,,, LMS-PAY-202608-1001
3171012345670002,1002,1785812731799,Budi Santoso,HR-02,POT_KOPKARA,Potongan Angsuran Kopkara,2026-08,850000.00,,,,LMS-PAY-202608-1002
3171012345670003,1003,1785812731812,Siti Aminah,FIN-01,POT_KOPKARA,Potongan Angsuran Kopkara,2026-08,1200000.00,,,,LMS-PAY-202608-1003
```

**Tahap 2 — Import Hasil Potongan (Adira → LMS):**
Adira mengisi kolom `NOMINAL_TERPOTONG`, `STATUS_POTONGAN`, dan `KETERANGAN` pada file yang sama, lalu mengirimkan kembali ke Kopkara. Petugas menaruh file di `Billing/Import/` dan klik **Import** di UI.

**Format file import (dikembalikan oleh Adira):**

| Kolom | Isi | Contoh |
|---|---|---|
| `NIK_ADIRA` | Nomor rekening / ID karyawan | `1112345` |
| `EMPLOYEE_ID` | ID karyawan di Adira | `110101` |
| `LOAN_NO` | Nomor pinjaman (jika ada) | `1785558548726` |
| `NAMA_KARYAWAN` | Nama karyawan | `Eko` |
| `DEPT_NO` | Kode departemen | `FIN` |
| `KODE_POTONGAN` | Kode potongan | `POT_KOPKARA` |
| `NAMA_POTONGAN` | Nama potongan | `Potongan Angsuran Kopkara` |
| `PERIODE` | Periode tagihan | `2026-08` |
| `NOMINAL_TAGIHAN` | Tagihan asal dari LMS | `609000.00` |
| `NOMINAL_TERPOTONG` | Aktual yang terpotong dari gaji | `650000.00` |
| `STATUS_POTONGAN` | Hasil potongan | `SUCCESS` / `FAILED` |
| `KETERANGAN` | Keterangan dari Adira | `Potongan Gaji Diproses` |
| `NO_REFERENSI` | Referensi pembayaran | `LMS-PAY-202608-110101` |

**Contoh isi file import:**
```csv
NIK_ADIRA,EMPLOYEE_ID,LOAN_NO,NAMA_KARYAWAN,DEPT_NO,KODE_POTONGAN,NAMA_POTONGAN,PERIODE,NOMINAL_TAGIHAN,NOMINAL_TERPOTONG,STATUS_POTONGAN,KETERANGAN,NO_REFERENSI
1112345,110101,1785558548726,Eko,FIN,POT_KOPKARA,Potongan Angsuran Kopkara,2026-08,609000.00,650000.00,SUCCESS,Potongan Gaji Diproses,LMS-PAY-202608-110101
1112345,110101,1785558573396,Eko,FIN,POT_KOPKARA,Potongan Angsuran Kopkara,2026-08,307500.00,650000.00,SUCCESS,Potongan Gaji Diproses,LMS-PAY-202608-110101
1112345,110101,,Budi,HR,POT_KOPKARA,Potongan Angsuran Kopkara,2026-08,850000.00,0.00,FAILED,Gaji tidak cukup,LMS-PAY-202608-1002
```

---

### 8.8 Pembayaran Cicilan — Direct Branch vs Waterfall Branch

**Narasi Proses:**

Saat memproses setiap baris CSV import, LMS menerapkan dua strategi:

**Direct Branch** (ada `LOAN_NO` di CSV):
LMS langsung menemukan schedule yang sesuai berdasarkan `loan_no`. Ini jalur paling akurat karena pembayaran tepat ke pinjaman yang dimaksud.

**Waterfall Branch** (tidak ada `LOAN_NO`):
LMS mencari semua cicilan aktif (`UNPAID`/`PARTIAL`) milik `EMPLOYEE_ID` tersebut, diurutkan dari due_date terlama. Pembayaran didistribusikan bertahap: cicilan terlama dilunasi dulu, sisanya ke cicilan berikutnya, hingga dana habis. Strategi ini memastikan tidak ada tunggakan yang terlewat.

---

### 8.9 Kalkulasi Update Outstanding & Penutupan Otomatis

**Narasi Proses:**

Setiap pembayaran diproses dalam satu query UPDATE yang efisien:

```sql
UPDATE lms_sch.loans
SET outstanding_amount = GREATEST(0, outstanding_amount - [jumlah_bayar]),
    status = CASE
        WHEN (SELECT COUNT(*) FROM lms_sch.loan_schedules
              WHERE loan_no = [id] AND status != 'PAID') = 0
        THEN 'CLOSED'
        ELSE status
    END,
    updated_at = NOW(), updated_user = [user]
WHERE loan_no = [id]
```

Query ini sekaligus mengurangi outstanding dan menutup pinjaman jika semua cicilan lunas — tanpa dua roundtrip ke database. Setelah `CLOSED`, `loan_applications` juga diupdate untuk konsistensi.

---

### 8.10 Rekonsiliasi & Adjustment Payroll

Rekonsiliasi adalah proses **verifikasi dan pencocokan** antara tagihan yang dikirim ke Adira dengan hasil potongan yang diterima kembali. Proses ini dilakukan setiap bulan setelah import selesai.

#### Status Rekonsiliasi Per Periode
| Status | Deskripsi |
|---|---|
| `OPEN` | Rekonsiliasi periode tersebut masih berjalan / belum ditutup |
| `CLOSED` | Rekonsiliasi telah resmi ditutup dan ditandatangani |

#### Alur Rekonsiliasi
```
Import CSV selesai
    ↓
Review hasil di UI (bandingkan tagihan vs realisasi)
    ↓
Identifikasi selisih (FAILED, OVERPAYMENT, UNDERPAYMENT)
    ↓
Input Adjustment jika ada koreksi
    ↓
Close Rekonsiliasi (tanda tangan 3 pihak)
    ↓ Status: CLOSED
```

#### Adjustment (Penyesuaian)

Adjustment diperlukan ketika ada **selisih antara tagihan dan realisasi**, misalnya:
- **Karyawan tidak dipotong** (gaji tidak cukup / resign) → `FAILED_CORRECTION`
- **Dipotong lebih dari tagihan** (overpayment) → `OVERPAYMENT_REFUND` atau `OVERPAYMENT_OFFSET`

Petugas mengisi form Adjustment dengan:
- `ref_no` — referensi dari payroll_deductions atau no. referensi CSV
- `loan_no` — nomor pinjaman yang terpengaruh
- `period` — periode bulan
- `adjustment_type` — jenis koreksi (`FAILED_CORRECTION` / `OVERPAYMENT_REFUND` / `OVERPAYMENT_OFFSET`)
- `original_amount` — tagihan asal
- `deducted_amount` — yang benar-benar terpotong
- `adjusted_amount` — selisih yang perlu dikoreksi (otomatis dihitung sistem)
- `notes` — keterangan

**Efek Adjustment ke Database:**
1. Insert record ke `payroll_adjustments`
2. Update status di `payroll_deductions` menjadi `ADJUSTED`
3. Jika ada `loan_no`: update `loan_schedules` — set `remaining_installment = 0`, status `ADJUSTED`

#### Reset Rekonsiliasi

Jika terjadi kesalahan import, admin dapat me-**reset** seluruh data rekonsiliasi periode tertentu (`POST /payroll/reset-reconciliation?period=YYYY-MM`):
- Semua `loan_schedules` periode tersebut dikembalikan ke `UNPAID`
- Semua record `payroll_deductions`, `payroll_adjustments` periode tersebut dihapus
- Log import dihapus
- Status rekonsiliasi dikembalikan ke `OPEN`

> ⚠️ **Perhatian:** Reset rekonsiliasi tidak dapat dibatalkan. Seluruh data pembayaran periode tersebut akan hilang.

#### Penutupan Rekonsiliasi (Close)

Setelah semua data diverifikasi dan adjustment selesai, petugas menutup rekonsiliasi (`POST /payroll/close-reconciliation`) dengan mengisi:
- Nama penandatangan dari HRD Adira (`hrd_signatory`)
- Nama penandatangan dari Finance/Keuangan (`finance_signatory`)
- Nama penandatangan dari Kopkara (`kopkara_signatory`)
- Catatan penutupan (`closing_notes`)

Data penutupan tersimpan di `payroll_reconciliation_closings`. Status berubah menjadi `CLOSED` dan periode tersebut tidak bisa diimport ulang kecuali di-reset.

---

### 8.11 Pelunasan Manual

**Narasi Proses:**

Pelunasan manual digunakan untuk kasus di luar siklus payroll:
- Karyawan **resign** dan membayar via transfer bank
- Pelunasan dari **uang pesangon**
- **Kompensasi simpanan** koperasi

Petugas mencari anggota (lazy load pencarian teks), memilih pinjaman, mengisi nominal dan jenis pembayaran:

| Jenis Pembayaran | Kode |
|---|---|
| Transfer bank | `TRANSFER_BANK` |
| Potong pesangon | `POTONG_PESANGON` |
| Kompensasi simpanan | `KOMPENSASI_SIMPANAN` |

**Dua mode pembayaran:**
- **Sebagian**: Bayar satu cicilan tertua yang belum lunas
- **Full Settlement** (`is_full_settlement=true`): Semua cicilan `UNPAID`/`PARTIAL` langsung `PAID`, pinjaman otomatis `CLOSED`

Setiap pelunasan manual dicatat ke `payroll_deductions` dengan `status = 'SUCCESS_MANUAL'` sebagai jejak audit.

---

### 8.12 Status Loan Schedule
| Status | Deskripsi |
|---|---|
| `UNPAID` | Belum ada pembayaran untuk cicilan ini |
| `PARTIAL` | Ada pembayaran tetapi belum mencapai total cicilan |
| `PAID` | Cicilan terbayar penuh |
| `ADJUSTED` | Cicilan dikoreksi via proses adjustment rekonsiliasi |

### 8.13 Status Loan
| Status | Deskripsi |
|---|---|
| `ACTIVE` | Pinjaman berjalan, masih ada cicilan yang belum lunas |
| `CLOSED` | Semua cicilan lunas — ditutup otomatis atau via pelunasan manual |



---

## 9. Integrasi Billing HRD-Adira

### 9.1 Export (LMS → HRD-Adira)

```
Admin klik "Export Billing"
    ↓ POST /api/payroll/export
Baca SCAN_DUEDATE_BILLING dari parameter cache
    ↓
Query loan_schedules UNPAID/PARTIAL sesuai cutoff
    ↓
Generate: ADIRA_PAYROLL_KOPKARA_OUTGOING_YYYYMM.csv
    ↓
Simpan ke FOLDER_BILL_EXPORT (dari cache parameter)
    ↓
Return CSV content ke frontend
```

**Format CSV Export:**
```
NIK_ADIRA, EMPLOYEE_ID, LOAN_NO, NAMA_KARYAWAN, DEPT_NO,
KODE_POTONGAN, NAMA_POTONGAN, PERIODE, NOMINAL_TAGIHAN,
NOMINAL_TERPOTONG, STATUS_POTONGAN, KETERANGAN, NO_REFERENSI
```

**Mode Scan (`SCAN_DUEDATE_BILLING`):**
| Nilai | Perilaku |
|---|---|
| `PERIOD` *(default)* | `ls.period <= cutoff_period` |
| `DUEDATE` | `ls.due_date <= cutoff_date` |
| Lainnya | Kombinasi OR keduanya |

### 9.2 Import (HRD-Adira → LMS)

```
Admin pilih file dari FOLDER_BILL_IMPORT
    ↓ POST /api/payroll/import
Cek duplikasi file (loan_payroll_import_logs)
    ↓
Per baris CSV:
  ├─ Ada loan_no → Direct Branch
  │   UPDATE loan_schedules
  │   UPDATE loans (outstanding + CASE CLOSED) — 1 query
  │   Jika CLOSED → UPDATE loan_applications
  │
  └─ Tidak ada loan_no → Waterfall Branch
      Cari schedules member_no (UNPAID/PARTIAL, oldest first)
      Distribusi pembayaran waterfall
      UPDATE loans (outstanding + CASE CLOSED) — 1 query
      Jika CLOSED → UPDATE loan_applications
    ↓
Backup file ke FOLDER_BILL_IMPORT_BCK (rename + timestamp)
    ↓
INSERT loan_payroll_import_logs (summary)
    ↓
Return: total records, success, failed
```

### 9.3 Format CSV Import (dari HRD-Adira)
```
REF_NO, EMPLOYEE_ID, LOAN_NO, PERIOD, NOMINAL_ORIGINAL,
NOMINAL_DEDUCTED, STATUS, KETERANGAN
```

---

## 10. Cache Strategy

### 10.1 Product Cache (`cache/product_cache.go`)

| Aspek | Detail |
|---|---|
| Singleton | `cache.ProductCache` — global |
| Thread-safety | `sync.RWMutex` |
| Warm-up | `WarmCache()` saat backend startup |
| Read | `FindAll()`, `FindByID()` → cache first |
| Invalidate | Otomatis saat `Create()`, `Save()`, `Delete()` |
| Fallback | Cache miss → query DB → populate cache |

```
Startup   → WarmCache() → DB → ProductCache.Set()
GET req   → ProductCache.Get() → return data (0 DB query)
Write req → DB operation → ProductCache.Invalidate()
Next GET  → cache miss → DB reload → ProductCache.Set()
```

### 10.2 Parameter Cache (`repositories/parameter_repository.go`)

| Aspek | Detail |
|---|---|
| Load | `refreshCache()` saat `NewParameterRepository()` (startup) |
| Storage | Slice + `map[string]GlobalParameter` (O(1) lookup) |
| Read | `FindByKey(key)` → map lookup, no DB query |
| Refresh | Otomatis setelah `Create()`, `Update()`, `Delete()` |

---

## 11. Konfigurasi & Environment

File: `backend/.env`

| Variable | Nilai | Deskripsi |
|---|---|---|
| `PORT` | `8086` | Port API server |
| `APP_ENV` | `development` | Environment mode |
| `DB_HOST` | `localhost` | Host PostgreSQL |
| `DB_PORT` | `5433` | Port PostgreSQL |
| `DB_USER` | `admin_lms` | Username database |
| `DB_PASSWORD` | `Nkl@130200` | Password database |
| `DB_NAME` | `lms_db` | Nama database |
| `KARISMA_API_URL` | `http://localhost:8086` | URL simulator Karisma |
| `TRACE_LEVEL` | `3` | SQL log level (0=off, 1=warn, 3=detail) |

---

## 12. Folder Billing

| Folder | Parameter Key | Default | Fungsi |
|---|---|---|---|
| `Billing/Export/` | `FOLDER_BILL_EXPORT` | `D:\...\Billing\Export` | Output CSV tagihan ke Adira |
| `Billing/Import/` | `FOLDER_BILL_IMPORT` | `D:\...\Billing\Import` | Input CSV hasil potongan dari Adira |
| `Billing/BCK/` | `FOLDER_BILL_IMPORT_BCK` | `D:\...\Billing\BCK` | Backup file diproses |

> **Catatan:** Backend otomatis konversi path `D:\...` → `/mnt/d/...` saat berjalan di WSL/Linux.

---

## 13. Global Parameters Kunci

| Key Name | Contoh Nilai | Fungsi |
|---|---|---|
| `FOLDER_BILL_EXPORT` | `D:\Data_NK\Project5\LMS\Billing\Export` | Folder output export CSV |
| `FOLDER_BILL_IMPORT` | `D:\Data_NK\Project5\LMS\Billing\Import` | Folder input import CSV |
| `FOLDER_BILL_IMPORT_BCK` | `D:\Data_NK\Project5\LMS\Billing\BCK` | Backup file import |
| `FOLDER_BILL_EXPORT_BCK` | `D:\Data_NK\Project5\LMS\Billing\BCK` | Backup file export |
| `SCAN_DUEDATE_BILLING` | `PERIOD` | Mode scan export: PERIOD / DUEDATE |

---

## Lampiran: Konvensi Kode

### Naming Convention
- **Go:** `camelCase` lokal, `PascalCase` publik
- **Database:** `snake_case` kolom dan tabel
- **JSON API:** `snake_case` field request/response

### Error Handling
- Handler return `gin.H{"error": "..."}` dengan HTTP status sesuai
- DB error di-log via `log.Printf()`
- Import error dicatat ke `payroll_deductions` sebagai `status = 'FAILED'`

### Audit Trail
- Setiap write: `updated_user` (employee_id) + `updated_at` (timestamp)
- Tracking lengkap: `loan_trackings` untuk lifecycle pengajuan
- Log import: `loan_payroll_import_logs` per file CSV

---

*Dokumentasi teknis sistem LMS Kopkara. Tim IT, Agustus 2026.*
