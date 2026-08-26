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

### 8.1 Flow Otentikasi Login & Inisialisasi UI LMS

Sistem LMS mengadopsi arsitektur otentikasi **Zero-Trust Security & Stateful Session Management** yang dikendalikan oleh **Parameter Global Sistem**. 

```
┌──────────┐               ┌──────────────────┐               ┌─────────────────────┐
│ Browser  │               │ Backend API      │               │ PostgreSQL DB       │
│ (React)  │               │ (Go / Gin)       │               │ (lms_sch)           │
└────┬─────┘               └────────┬─────────┘               └──────────┬──────────┘
     │                              │                                    │
     │ 1. POST /api/karisma/login   │                                    │
     ├─────────────────────────────►│                                    │
     │ (username, password)         │ 2. Cek Password Dual-Verify        │
     │                              ├───────────────────────────────────►│
     │                              │ (Bcrypt / Plain-text fallback)     │
     │                              │                                    │
     │                              │ 3. Simpan Sesi (lms_sch.sessions)  │
     │                              ├───────────────────────────────────►│
     │                              │                                    │
     │                              │ 4. Baca Parameter APP_TOKEN        │
     │                              ├───────────────────────────────────►│ (default: 'ewa_token')
     │                              │                                    │
     │ 5. Response HTTP 200 OK      │                                    │
     │◄─────────────────────────────┤                                    │
     │ Set-Cookie: ewa_token=HASH   │                                    │
     │ (HttpOnly, Secure/Dinamis)   │                                    │
     │                              │                                    │
     │ 6. POST /api/karisma/verify  │                                    │
     ├─────────────────────────────►│                                    │
     │ withCredentials: true        │ 7. Verifikasi Token Sesi           │
     │                              ├───────────────────────────────────►│ (Cek User & RoleID > 0)
     │                              │                                    │
     │ 8. Response User Session     │                                    │
     │◄─────────────────────────────┤                                    │
     │ Simpan LocalStorage: ewa_user│                                    │
     │                              │                                    │
     │ 9. GET /api/user-info/:user  │                                    │
     ├─────────────────────────────►│                                    │
     │ (misal: /api/user-info/nur)  │ 10. Query Role & Menus Navigasi    │
     │                              ├───────────────────────────────────►│ (Admin / RoleID 1 -> ALL)
     │ 11. Response Menu & Role     │                                    │
     │◄─────────────────────────────┤                                    │
     │ Render Sidebar & Admin UI    │                                    │
```

---

#### Detail Tahapan Otentikasi & Inisialisasi UI:

##### Tahap 1: Login Request (`POST /api/karisma/login`) & Keamanan Token (`APP_TOKEN`)
1. Pengguna memasukkan `username` dan `password` pada form login di frontend.
2. Backend menerima request dan melakukan dual-verification password (Bcrypt + plain-text fallback). Jika password plain-text cocok, sistem otomatis meng-upgrade hash password ke Bcrypt di database `lms_sch.users`.
3. Setelah otentikasi sukses, backend membuat string token 64-karakter acak (`generateSessionToken()`) dan mencatatnya ke dalam tabel `lms_sch.sessions`.
4. **Pengaturan Parameter `APP_TOKEN` (Nama Cookie Token)**:
   - Backend membaca parameter `APP_TOKEN` dari tabel `lms_sch.global_parameters` (default: **`ewa_token`**).
   - Backend menetapkan token ke dalam **HttpOnly Cookie** (`Set-Cookie: ewa_token=<HASH>; HttpOnly; SameSite=Lax`).
   - Flag `secure` disesuaikan secara dinamis (`c.Request.TLS != nil` atau header `X-Forwarded-Proto == https`) agar cookie dapat diterima dengan mulus di lingkungan HTTP maupun HTTPS.
5. **Keamanan Ekstra (XSS Protection)**:
   - Token otentikasi **SAMA SEKALI TIDAK DITAMPILKAN / DISIMPAN DI LOCAL STORAGE ATAU CACHE BROWSER**.
   - Hal ini melindungi token dari bahaya pencurian lewat skrip berbahaya (*Cross-Site Scripting / XSS*).

##### Tahap 2: Verifikasi Sesi (`POST /api/karisma/verify`) & Storage User (`APP_USER`)
1. Saat aplikasi React pertama kali dimuat (*mount*) atau setelah login berhasil, frontend memanggil endpoint `/api/karisma/verify` dengan `withCredentials: true`.
2. Backend mengambil token via fungsi `getTokenFromRequest(c)`:
   - Membaca HttpOnly Cookie dengan nama parameter `APP_TOKEN` (`ewa_token`), fallback ke `karisma_token`, atau header `Authorization: Bearer <token>`.
3. Backend memverifikasi bahwa sesi di `lms_sch.sessions` masih aktif (`is_active = true`) dan terhubung ke `lms_sch.users`.
4. **Pencegahan Security Hole (Strict Zero-Fallback Role)**:
   - Jika `user.RoleID <= 0` atau tidak terdaftar, backend membatalkan sesi dan menolak dengan **`HTTP 403 Forbidden`**. Tidak ada role default yang diberikan.
5. **Pengaturan Parameter `APP_USER` (Nama LocalStorage User)**:
   - Setelah verifikasi berhasil, frontend membaca parameter `APP_USER` dari `global_parameters` (default: **`ewa_user`**).
   - Data objek profil pengguna (nama, username, member_no, role_id) disimpan di `localStorage.setItem('ewa_user', JSON.stringify(user))` hanya untuk keperluan render nama dan avatar UI.

##### Tahap 3: Resolusi Menus Navigasi & Rendering UI (`GET /api/user-info/:userKey`)
1. Frontend memanggil `/api/user-info/${currentUser.username}` menggunakan `username` unik (misal: `/api/user-info/nur`) untuk menghindari benturan record ID data load test.
2. Backend handler `GetUserInfo` memprioritaskan kueri `LOWER(username) = LOWER('nur')`.
3. **Hak Akses Superuser Admin (`RoleID == 1`)**:
   - Jika pengguna memiliki `role_id == 1` (Admin), backend secara otomatis mengambil **SELURUH MENU SYSTEM** dari tabel `lms_sch.menus` tanpa bergantung pada relasi `lms_sch.role_menus`.
4. **Struktur Pertahanan Menu Frontend (3-Tier Fail-Safe)**:
   - Frontend me-render daftar menu sidebar (`visibleMenus`) dengan logika 3 lapis:
     - **Tier 1**: `userInfo.menus` hasil query backend API.
     - **Tier 2**: `referenceData.menus` dari database LMS.
     - **Tier 3**: System default fallback menus (*Dashboard, Pengajuan Pinjaman, Daftar Pinjaman, Approval Pinjaman, Potong Gaji (HRD), Pelunasan Manual, Produk Pinjaman, Pengaturan Parameter, Data Master*).
5. **Garansi Menu Data Master**:
   - Setiap pengguna ber-role Admin (`role_id: 1` atau `role_name: 'admin'`) — *termasuk user `nur`* — dijamin **100% di-grant menu Data Master (`path: 'master'`)** pada sidebar navigasi.
6. UI LMS terbuka seketika dengan menampilkan **Mode: Admin**, nama pengguna di top header, serta seluruh daftar menu navigasi secara lengkap.

---

### 8.2 Siklus Pengajuan Pinjaman (Origination)

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

### 8.4 Kalkulasi Cicilan, Biaya Admin & Credit Limit

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

**Credit Limit & Validasi Multi-Loan:**
```
credit_limit = evaluasi(LOAN_LIMIT_FORMULA)
Jika credit_limit > employee_categories.max_limit → credit_limit = max_limit

Total Exposure Baru = employees.total_loan + Pengajuan Baru (Pokok + Bunga)
Jika Total Exposure Baru > credit_limit → TOLAK
Sisa Limit Tersedia = max(0, credit_limit - employees.total_loan)
```
> **Catatan Multi-Loan:** Field `total_loan` pada `lms_sch.employees` menyimpan total kewajiban (Pokok + Bunga) dari seluruh pinjaman aktif dan pengajuan gantung (`SUBMITTED`/`APPROVED`).

---

### 8.5 Contoh Simulasi Nyata

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

### 8.6 Validasi Submit & Approval Pinjaman

Seluruh proses pengajuan hingga persetujuan pinjaman dikontrol secara ketat oleh sistem validasi multi-layer pada backend API:

#### 1. Validasi Submit Pinjaman (`POST /api/applications`)
Sebelum record pengajuan disimpan ke database dengan status `SUBMITTED`, backend mengeksekusi urutan validasi berikut:

1. **Validasi Keaktifan Anggota & Karyawan**:
   - Pemohon harus terdaftar di `lms_sch.members` dan `lms_sch.employees`.
   - `is_member` harus bernilai `true` (kecuali akun manajemen Admin/HRD yang selalu diizinkan).
2. **Validasi Periode Tanggal Pengajuan (`LOAN_START_PERIOD` & `LOAN_END_PERIOD`)**:
   - Sistem mengambil tanggal hari submit (`DAY` = 1..31).
   - `DAY` harus memenuhi: `LOAN_START_PERIOD <= DAY <= LOAN_END_PERIOD` (contoh: tanggal 1 s/d 15).
   - Jika diluar periode, pengajuan ditolak dengan pesan: *"Pengajuan hanya diizinkan antara tanggal X sampai Y"*.
3. **Validasi Tenor Maksimum (`LOAN_MAX_TENOR`)**:
   - Tenor yang diajukan tidak boleh melebihi `min(loan_products.max_tenor_months, LOAN_MAX_TENOR)`.
4. **Validasi Credit Limit Dinamis & Multi-Loan (`LOAN_LIMIT_FORMULA` & `employees.total_loan`)**:
   - Credit limit dihitung secara *real-time* menggunakan expression engine `govaluate` dengan formula dari `LOAN_LIMIT_FORMULA` (variabel: `DAY`, `SALARY`, `REQUESTED_AMOUNT`, `TENOR`).
   - **Multi-Loan Credit Limit Engine**:
     - Sistem mengambil nilai kewajiban eksisting `employees.total_loan` ($	ext{Pokok} + 	ext{Bunga}$) dari tabel `lms_sch.employees`.
     - Menghitung total kewajiban pengajuan baru: $N_{	ext{total}} = 	ext{Requested Amount} + 	ext{Total Bunga Tenor}$.
     - Akumulasi total kewajiban: $	ext{Total Exposure} = 	ext{employees.total\_loan} + N_{	ext{total}}$.
     - Jika $	ext{Total Exposure} > 	ext{credit\_limit}$, pengajuan **DITOLAK** dengan pesan presisi dalam format Rupiah:
       > *"Pengajuan ditolak. Total kewajiban pinjaman Anda (Rp X) + pengajuan baru (Rp Y) = Rp Z melebihi Credit Limit (Rp CL). Sisa limit yang dapat diajukan (Pokok + Bunga): Rp Available"*
     - Jika $	ext{Total Exposure} \le 	ext{credit\_limit}$, pengajuan diterima dan `employees.total_loan` langsung diperbarui secara komulatif.
5. **Kalkulasi Biaya Admin (`LOAN_ADMIN_FORMULA` & `LOAN_MIN_ADMIN_FEE`)**:
   - Biaya admin dihitung otomatis melalui formula `LOAN_ADMIN_FORMULA`.
   - Jika hasil kalkulasi di bawah `LOAN_MIN_ADMIN_FEE` (floor value), maka biaya admin ditetapkan sebesar `LOAN_MIN_ADMIN_FEE`.

---

#### 2. Validasi & Audit Trail Approval Pinjaman (`POST /api/applications/:id/approve` / `reject` / `revision`)
Admin atau Komite Kredit memproses pengajuan berstatus `PENDING`/`REVIEWED` dengan aturan validasi:

1. **Otorisasi Role RBAC (`RBACRoleMenuMiddleware`)**:
   - Hanya pengguna dengan `role_id = 1` (Admin) atau `role_id = 3` (HRD/Approval) yang diizinkan mengakses endpoint persetujuan.
2. **Validasi Nominal Approval (`approved_amount`)**:
   - `approved_amount` yang disetujui tidak boleh bernilai `0` atau negatif.
   - `approved_amount` tidak boleh melebihi `requested_amount` pemohon maupun `credit_limit` anggota.
3. **Perekaman Audit Trail & Durasi SLA (`lms_sch.loan_trackings`)**:
   - Setiap perubahan status pengajuan mencatat log histori ke `lms_sch.loan_trackings`:
     - `application_no`: Nomor pengajuan.
     - `from_status` & `to_status`: Perubahan status (misal: `PENDING` → `APPROVED`).
     - `action_by` / `updated_user`: ID & Username petugas yang menyetujui.
     - `sla_duration`: Selisih waktu (dalam detik/menit) sejak status sebelumnya dibuat hingga aksi approval dilakukan.
     - `ip_address` & `user_agent`: Data geolokasi jaringan & browser petugas.

---

#### 3. Validasi Pencairan & Pembentukan Kontrak (`POST /api/applications/:id/disburse`)
Setelah pengajuan berstatus `APPROVED`, pencairan dana dikonfirmasi oleh petugas. Backend secara atomis (*atomic transaction*) membentuk 3 entitas data:

1. **`loan_contracts`**: Kontrak legal pinjaman berisi `contract_no`, `approved_amount`, `interest_rate`, `monthly_installment`, `admin_fee`, dan tanggal efektif.
2. **`loans`**: Record pinjaman aktif dengan `outstanding_amount = approved_amount` dan status `ACTIVE`.
3. **`loan_schedules`**: Matriks jadwal angsuran bulanan sebanyak tenor yang disetujui.
   - Tanggal jatuh tempo (`due_date`) setiap angsuran dihitung otomatis berdasarkan parameter `LOAN_DUEDATE` (default: tanggal 25 setiap bulannya).

---

### 8.8 Pencairan Pinjaman (Disbursement)

**Narasi Proses:**

Setelah pengajuan berstatus `APPROVED`, petugas melakukan pencairan (`POST /api/applications/:id/disburse`). Sistem otomatis membuat tiga record sekaligus:

1. **`loan_contracts`** — kontrak formal pinjaman (jumlah, tenor, cicilan, suku bunga, tanggal kontrak)
2. **`loans`** — record pinjaman aktif. `outstanding_amount` = `approved_amount` (saldo penuh)
3. **`loan_schedules`** — jadwal cicilan per bulan dari tenor. Tanggal jatuh tempo (`due_date`) ditentukan oleh parameter `LOAN_DUEDATE` (default tanggal 25)

---

### 8.9 Siklus Penagihan Payroll Bulanan

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



---

### 8.7 Spesifikasi Teknis Validasi Credit Limit Multi-Loan (Tabel Employees)

#### 1. Pokok-Pokok Ketentuan
1. **Lokasi Field `total_loan`**:  
   Ditambahkan pada tabel **`lms_sch.employees`** (berdampingan dengan field `salary` dan `employee_id`).
2. **Definisi `total_loan`**:  
   `total_loan` = **Pokok + Bunga** (Total Nilai Kewajiban Angsuran Pinjaman Eksisting).
3. **Formula Validasi Credit Limit**:  
   $$\text{Total Loan Kumulatif} = \text{employees.total\_loan} + \text{Pengajuan Baru (Pokok + Bunga)} \le \text{Credit Limit}$$

#### 2. Matriks Lifecycle Perubahan State `employees.total_loan`
| Event / Aksi | Perubahan `total_loan` di `lms_sch.employees` | Keterangan |
|---|---|---|
| **Submit Pinjaman Baru** | `total_loan += (Pokok + Bunga)` | Mengunci limit sejak status `SUBMITTED` / `PENDING` |
| **Penolakan (`REJECTED`) / Batal** | `total_loan -= (Pokok + Bunga)` | Melepas kuncian limit jika pengajuan ditolak/dibatalkan |
| **Pembayaran Angsuran (Payroll / Manual)** | `total_loan -= Total Angsuran Dibayar` | Mengurangi `total_loan` sebesar nominal angsuran terbayar |
| **Pelunasan Penuh (`CLOSED`)** | Resync otomatis ke sisa kewajiban | Pinjaman lunas, limit kembali pulih sepenuhnya |

#### 3. DDL Migration Script (`backend/migrations/add_total_loan_to_employees.sql`)
```sql
-- Tambah kolom total_loan pada tabel lms_sch.employees
ALTER TABLE lms_sch.employees 
ADD COLUMN IF NOT EXISTS total_loan NUMERIC(15, 2) DEFAULT 0.00;

-- Update nilai awal total_loan berdasarkan saldo pinjaman aktif + pengajuan gantung
UPDATE lms_sch.employees e
SET total_loan = COALESCE((
    SELECT SUM(l.outstanding_amount) 
    FROM lms_sch.loans l 
    JOIN lms_sch.members m ON l.member_no = m.member_no
    WHERE m.employee_id = e.employee_id AND l.status = 'ACTIVE'
), 0) + COALESCE((
    SELECT SUM(la.requested_amount * (1 + (la.tenor * 0.01))) 
    FROM lms_sch.loan_applications la 
    JOIN lms_sch.members m ON la.member_no = m.member_no
    WHERE m.employee_id = e.employee_id AND la.status IN ('SUBMITTED', 'PENDING', 'REVIEWED', 'APPROVED')
), 0);
```

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

| Variable | Nilai Contoh | Deskripsi |
|---|---|---|
| `PORT` | `8086` | Port API server |
| `APP_ENV` | `development` | Environment mode |
| `DB_HOST` | `localhost` | Host PostgreSQL |
| `DB_PORT` | `5433` | Port PostgreSQL |
| `DB_USER` | `admin_lms` | Username database |
| `DB_PASSWORD` | *(kosong jika pakai enkripsi)* | Password plain-text (fallback) |
| `DB_PASSWORD_ENCRYPTED` | `D9M/DJaw9Hz...` | Password database terenkripsi AES-256 |
| `DB_NAME` | `lms_db` | Nama database |
| `JWT_SECRET` | `LMS_K0pK4r4_S3cr3t...` | Kunci enkripsi/dekripsi password DB |
| `ALLOWED_ORIGINS` | `http://lims.local:3000,https://localhost` | Daftar origin terpilih yang diizinkan CORS |
| `HIGH_PRIVILEGE_ROLES` | `1,3,admin,hrd` | Daftar ID/Nama role yang memiliki hak akses seluruh pinjaman |
| `KARISMA_API_URL` | `http://localhost:8086` | URL simulator Karisma |
| `TRACE_LEVEL` | `3` | SQL log level (0=off, 1=warn, 3=detail) |

> **Catatan Keamanan:** Jika `DB_PASSWORD_ENCRYPTED` terisi, sistem akan otomatis mendekripsi menggunakan `JWT_SECRET`. Jika kosong, sistem akan menggunakan `DB_PASSWORD` biasa sebagai *fallback*.

---

### 11.1 Membuat JWT_SECRET

`JWT_SECRET` adalah kunci rahasia yang digunakan untuk **mengenkripsi dan mendekripsi password database**. Kunci ini harus:
- **Panjang** minimal 32 karakter
- **Acak** dan tidak bisa ditebak
- **Konsisten** — kunci yang dipakai enkripsi harus sama dengan yang dipakai dekripsi
- **Rahasia** — jangan di-*commit* ke Git atau dibagikan sembarangan

#### Cara Generate JWT_SECRET

**Pilihan 1 — OpenSSL di terminal WSL/Linux (Recommended):**
```bash
openssl rand -hex 32
# Contoh output: a3f9c2d81b7e4506c0e2f3a8d1b9c7e4a5f6d2b8c1e0f3a7b4d9c6e2f1a0b5d8
```

**Pilihan 2 — Base64 random:**
```bash
openssl rand -base64 48
# Contoh output: X9kLm3pQr7sBvN2wA1jH5tYu8eZcF4dG6oIqK0nM+WvR=
```

**Pilihan 3 — Random string alphanumeric:**
```bash
cat /dev/urandom | tr -dc 'a-zA-Z0-9!@#$%^&*' | head -c 48
```

#### Workflow Lengkap Setup Enkripsi DB Password

```
# Langkah 1: Generate JWT_SECRET dan tambahkan ke .env
openssl rand -hex 32
# → Salin hasilnya sebagai nilai JWT_SECRET di backend/.env

# Langkah 2: Generate encrypted password dari password DB asli
go run generate_db_password.go "PasswordDBAnda"
# → Salin output (string panjang) sebagai nilai DB_PASSWORD_ENCRYPTED di backend/.env

# Langkah 3: Kosongkan DB_PASSWORD di .env (opsional, untuk keamanan)
# DB_PASSWORD=

# Langkah 4: Restart LMS — log akan tampil:
# "Database: Menggunakan password terenkripsi (DB_PASSWORD_ENCRYPTED)."
```

> **⚠️ PENTING:** Simpan nilai `JWT_SECRET` di tempat yang aman (misalnya *password manager*). Jika hilang, Anda **tidak bisa mendekripsi** password DB yang sudah terenkripsi dan harus mengulang proses enkripsi dari awal.

---

### 11.2 Konfigurasi CORS (ALLOWED_ORIGINS) & Security Proxies

Variabel `ALLOWED_ORIGINS` di file `backend/.env` digunakan untuk membatasi domain/origin mana saja yang diizinkan mengakses API LMS Backend melalui Cross-Origin Resource Sharing (CORS).

#### Contoh Konfigurasi di `backend/.env`:
```env
ALLOWED_ORIGINS=http://lims.local:3000,http://lims.local:8082,http://lims.local:8087,https://lims.local,http://localhost,https://localhost,capacitor://localhost
```

#### Cara Kerja di Backend (`main.go`):
1. **Membaca Allowed Origins**: Middleware membaca string `ALLOWED_ORIGINS` dari `.env` dan memisahkan setiap origin berdasarkan tanda koma (`,`).
2. **Dynamic Origin Matching**:
   - Jika `Origin` pemanggil ada dalam daftar (misal `https://localhost:3005` atau `https://lims.local`), backend merespons dengan header:
     - `Access-Control-Allow-Origin: <origin_pemanggil>`
     - `Vary: Origin`
     - `Access-Control-Allow-Credentials: true`
   - Jika request berasal dari domain luar yang tidak terdaftar, header `Access-Control-Allow-Origin` tidak dikirim, sehingga browser akan **memblokir** request tersebut secara otomatis.
   - Jika `ALLOWED_ORIGINS` tidak diisi (kosong), sistem fallback mengizinkan `*` untuk kemudahan pengujian di lingkungan lokal/development.
3. **Keamanan Trusted Proxies**:
   - Backend memanggil `_ = r.SetTrustedProxies(nil)` untuk menghilangkan peringatan `[WARNING] You trusted all proxies` dari framework Gin secara aman ketika tidak berjalan di belakang reverse proxy khusus.

---

### 11.3 Konfigurasi High-Privilege Roles & Proteksi Kerahasiaan Data Pinjaman

Demi menjaga kerahasiaan data keuangan anggota, sistem LMS membedakan hak akses daftar pinjaman (*loan applications*) antara **User Biasa (Anggota)** dan **User High-Privilege (Admin / HRD)**.

#### Parameter Konfigurasi:
1. **Aturan di `backend/.env`**:
   ```env
   HIGH_PRIVILEGE_ROLES=1,3,admin,hrd
   ```
2. **Aturan di `lms_sch.global_parameters` (Opsional)**:
   Key Name: `HIGH_PRIVILEGE_ROLES`  
   Key Value: `1,3,admin,hrd`

#### Aturan Akses & Workflow UI:
1. **Initial Load (Pembukaan Tab)**:
   - Data pinjaman **TIDAK di-load otomatis** saat tab "Daftar Pinjaman" pertama kali dibuka.
   - Tabel menampilkan instruksi: `🔍 Silakan pilih periode & klik tombol "Cari Pinjaman" untuk menampilkan data.`
2. **Filter Periode (Bulan & Tahun)**:
   - Disediakan dropdown **Bulan** (01-12) dan **Tahun** (2024-2028) agar pengguna dapat melihat data pinjaman pada periode mana saja secara terstruktur.
3. **User Biasa (Role `Anggota` / ID 2)**:
   - Text field pencarian **pre-filled otomatis dengan Nomor Anggota / Employee ID pengguna yang sedang login** dan statusnya **Disabled / Greyed out** (tidak bisa diedit).
   - Saat pengguna memilih periode dan meng-klik **"Cari Pinjaman"**, backend secara ketat memfilter data pinjaman hanya milik anggota tersebut (`member_no = current_user_employee_id`).
4. **User High-Privilege (Role `Admin` ID 1 / `HRD` ID 3 / Sesuai Konfigurasi)**:
   - Text field pencarian **Enabled** dan dapat diketik secara bebas.
   - **Jika text field kosong** lalu diklik "Cari Pinjaman" → Menampilkan **seluruh pinjaman** semua anggota pada periode terpilih.
   - **Jika text field diisi** (misal `10104`) lalu diklik "Cari Pinjaman" → Menampilkan pinjaman spesifik milik anggota tersebut pada periode terpilih.
5. **Fallback Notifikasi Tabel Partisi**:
   - Jika tabel partisi periode tidak ada atau data kosong, UI menampilkan notifikasi: `⚠️ Data pinjaman tidak ditemukan (Periode YYYY-MM)`.
6. **Kolom Employee ID**:
   - Tabel "Daftar Pengajuan & Pinjaman" mencantumkan kolom **Employee ID** di samping kolom No. Pengajuan untuk kejelasan identitas pemilik pinjaman.
7. **Modal Riwayat Status Pengajuan Pinjaman (Tracking Modal)**:
   - Kolom **Employee ID** ditambahkan tepat sebelum kolom **Nama**.
   - Nama kolom **User** diperbarui/diubah menjadi **Nama** untuk kejelasan identitas pengguna/petugas yang melakukan aksi.
   - Ukuran lebar modal diperluas hingga `1200px` dan kolom **Catatan** diatur dengan batas maksimal 2 baris (*line clamp*) agar tampilan ringkas dan rapi.
8. **Filter Status Tab Approval Pengajuan Pinjaman**:
   - Badge `STATUS AKTIF` pada box filter Verifikasi Administratif diubah secara resmi menjadi **`SUBMITTED`**.
   - Pemanggilan API dari tab Approval dikirim dengan parameter `&status=SUBMITTED` sehingga kueri SQL yang dieksekusi ke PostgreSQL secara ketat memfilter:  
     `SELECT * FROM "lms_sch"."loan_applications_YYYYMM" WHERE status = 'SUBMITTED' ORDER BY created_at desc`
9. **Filter Status Tab Pencairan Dana (Disbursement Treasury)**:
   - Pemanggilan API dari tab Pencairan Dana dikirim dengan parameter `&status=APPROVED` sehingga kueri SQL yang dieksekusi ke PostgreSQL secara ketat memfilter:  
     `SELECT * FROM "lms_sch"."loan_applications_YYYYMM" WHERE status IN ('APPROVED') ORDER BY created_at desc`
10. **Optimasi Performa Pengambilan Data Anggota (`/api/members/all`)**:
    - Pemanggilan API `/api/members/all` dihilangkan pada perpindahan ke tab Approval dan Pencairan Dana sehingga tidak menjalankan kueri berat `SELECT m.member_no, COALESCE(e.employee_id...) FROM lms_sch.members LEFT JOIN lms_sch.employees`.
11. **Pembaruan & Optimasi Tab Pelunasan Manual (`manual-repayment`)**:
    - **Penghapusan Tabel Audit Log Pelunasan & Query `GET /api/payroll/deductions`**: Card Riwayat Audit Log (`lms_sch.payroll_deductions`) serta pemanggilan API `GET /api/payroll/deductions` pada `useEffect` tab *manual-repayment* dihapus sepenuhnya untuk membebaskan database dari *full table scan & sorting* PostgreSQL (`SELECT pd.id, pd.loan_no... FROM lms_sch.payroll_deductions... ORDER BY pd.id DESC`).
    - **Form Lebih Besar & Lebar**: Container form dibuat menjadi layout terpusat yang luas (`maxWidth: 900px`) dengan tampilan visual yang lebih bersih dan nyaman digunakan.
    - **Tombol Filter Eksplisit `🔍 Cari`**: Fitur pencarian anggota menggunakan tombol **`🔍 Cari`** (atau tekan `Enter`), sehingga proses ketik `employee_id` tidak lagi memicu kueri berulang (0 kueri saat pengetikan).
    - **Penjelasan Box Pink (Early Full Settlement)**: Diberikan penjelasan eksplisit bahwa opsi centang `🔥 Pelunasan Lunas Sekaligus (Lunas Total / Early Full Settlement)` berguna untuk melunasi seluruh sisa hutang pinjaman secara permanen (misal: Karyawan Resign) hingga status menjadi **`CLOSED`**.
    - **Auto-Generate Kuitansi Bukti Pembayaran & Nama File PDF**: Setelah tombol **`💳 Proses`** diklik dan pelunasan berhasil diproses, LMS secara otomatis menerbitkan Modal **Kuitansi Bukti Pelunasan Pinjaman**. Saat mengeklik tombol **`🖨️ Cetak Kuitansi`**, judul dokumen secara otomatis disesuaikan menjadi **`Kopkara LMS - Pelunasan Manual`** sehingga nama file PDF bawaan saat disimpan berubah menjadi **`Kopkara LMS - Pelunasan Manual.pdf`**.
    - **Eliminasi Kueri `payroll_adjustments` & `UNION ALL` (`/api/payroll/schedules`)**: Kueri `payroll_adjustments` serta sub-kueri `UNION ALL` dihapus sepenuhnya dari endpoint `/api/payroll/schedules`. Endpoint kini 100% hanya mengeksekusi 1 kueri tunggal ke `lms_sch.loan_schedules` dengan filter `WHERE ls.status != 'PAID' AND ls.status != 'CLOSED'`, sehingga tidak ada kueri berlebih ke tabel adjustment maupun deductions.
    - **Eliminasi Auto-Fetch `loan_schedules` pada Tab Potong Gaji (`payroll`)**: Pemanggilan otomatis `fetchPayrollSchedules()` pada `useEffect` tab *payroll* (Potong Gaji) dihapus sepenuhnya, sehingga saat membuka form Potong Gaji tidak ada kueri SQL berlebih yang mengeksekusi tabel `loan_schedules`.
    - **Eliminasi Auto-Fetch `loan_applications` Pasca-Import CSV**: Pemanggilan `fetchApplications()` pasca-eksekusi impor CSV hasil rekonsiliasi gaji dihilangkan dari callback `App.jsx`, sehingga PostgreSQL tidak lagi mengeksekusi kueri `SELECT * FROM "lms_sch"."loan_applications_YYYYMM" ORDER BY created_at desc` saat proses impor CSV selesai.
    - **Eliminasi Kueri `payroll_reconciliation_closings` Pasca-Adjustment**: Pemanggilan `fetchReconciliationStatus()` pasca-eksekusi simpan adjustment (`/api/payroll/adjust`) dihilangkan dari `App.jsx`, sehingga PostgreSQL tidak lagi mengeksekusi kueri `SELECT ... FROM lms_sch.payroll_reconciliation_closings WHERE period = '2026-08' LIMIT 1` saat adjustment selesai disimpan.
    - **Pengurutan Laporan Pengajuan (`loan_applications`) Berdasarkan Employee ID / Member No**: Kueri pengajuan pinjaman pada `application_repository.go` diubah urutannya menjadi `ORDER BY member_no ASC, created_at DESC`, sehingga laporan dan daftar pengajuan pinjaman diurutkan secara berurutan sesuai NIK/ID Karyawan.
    - **Modul Resmi Laporan Pengajuan Pinjaman (`report-loan-applications`) & Layout Cetak PDF**: Modul UI lengkap untuk *Laporan Pengajuan Pinjaman* resmi diaktifkan di frontend (`App.jsx`). Dilengkapi penyesuaian layout cetak profesional:
      1. Header cetak dengan Logo Gambar yang dibaca secara dinamis dari `/frontend/public/<nama file>` berdasarkan parameter `LMS_Title` (default: `kopkara.jfif`).
      2. Judul utama laporan (H1) **`Laporan loan periode [NamaBulan] [Tahun]`** diletakkan rapat kiri (*left-aligned*) berdampingan di sebelah kanan logo, menggantikan judul default sebelumnya.
      3. Format Tanggal Cetak lengkap beserta jam: **`Tanggal Cetak: 7 Agustus 2026 20:18:10`** (format: `DD MMMM YYYY HH:mm:ss`).
      4. Pemeliharaan penomoran halaman resmi browser (*"Halaman X dari Y"*) dan judul file default PDF (**`Laporan loan periode [NamaBulan] [Tahun].pdf`**) dengan margin `@page { margin: 10mm; }`.
      5. Garis pembatas tebal (`3.5px solid #0A2540`) di bawah header cetak.
      6. Menyembunyikan header navigasi, card metrics summary, dan filter kontrol saat dicetak (`@media print` / `.no-print`).
      7. Penambahan kolom nomor urut **`No.`** (1, 2, 3...) di sebelah kiri sebelum kolom *No. Pengajuan*.
      8. Penghapusan kolom *Catatan* dari tabel detail.
    - **Database Index `application_no` (`lms_sch.loan_trackings`)**: Tag GORM `gorm:"column:application_no;index"` dikonfigurasi pada `LoanTracking` struct. Seluruh eksekusi DDL & Seeding programmatic dari backend startup telah dihapus sepenuhnya sesuai arahan user.
    - **Penamaan Tombol**: Tombol eksekusi resmi diubah namanya menjadi **`💳 Proses`**.

### 11.4 Implementasi Audit Performa & Keamanan (LIMS Standard Compliance)
- **Eksekusi Middleware GZIP Compression (`GzipMiddleware`)**: Diterapkan pada `backend/main.go` menggunakan modul kompresi native standard library `compress/gzip` untuk mengompres seluruh payload JSON & static asset. Penggunaan kuota jaringan & latensi berkurang hingga 70%-80%.
- **Standarisasi GORM Soft-Delete (`gorm.DeletedAt`)**: `MasterBaseModel` di `backend/models/base.go` dikonfigurasi menggunakan tipe `gorm.DeletedAt` agar seluruh operasi penghapusan data master & transaksi diubah otomatis menjadi soft-delete (`updated_at` & `deleted_at`), mencegah hilangnya jejak audit (*Anti-Repudiation*).
- **Integrasi Otomatis Parameter `PAGINATION_LIMIT`**: `application_usecase.go` kini secara otomatis membaca nilai parameter global `PAGINATION_LIMIT` (default `5`) dari DB cache ketika parameter `limit` tidak dikirim oleh client. PostgreSQL mengeksekusi klausa SQL `LIMIT 5 OFFSET 0` dan tabel *Daftar Pinjaman* di frontend menampilkan kontrol halaman (halaman 1 dari N, Prev/Next) persis 5 record per halaman.
- **Proteksi Autentikasi API (`AuthMiddleware`)**: Seluruh endpoint API internal di bawah grup `/api` (seperti `/api/payroll/reconciliation-status`, `/api/applications`, dll) kini dilindungi `AuthMiddleware()`. Akses langsung via browser/cURL tanpa header `Authorization: Bearer <token>` otomatis ditolak dengan status **`401 Unauthorized`**.
- **Isolasi Token HttpOnly Cookie (`Secure Token Storage`)**: Token autentikasi saat login kini diset oleh server Go dalam bentuk **HttpOnly & Secure Cookie (`karisma_token`)** (`Set-Cookie: karisma_token=...; HttpOnly; Secure; SameSite=Lax`) dan tidak lagi disimpan di `localStorage` browser. Skrip JavaScript browser sama sekali tidak dapat mengakses token ini, menjamin kekebalan 100% dari potensi pencurian token via skrip jahat (XSS).
- **Pembatasan Laju Request (`RateLimitMiddleware` & Parameter Availability)**: Terpasang `RateLimitMiddleware` pada backend Go yang membaca parameter `RATE_LIMIT_GENERAL_RPM` (default `60` request/menit per IP) dan `RATE_LIMIT_HEAVY_ENDPOINTS` untuk membatasi endpoint transaksi berat (`/api/applications`, `/api/payroll/reconcile`, `/api/karisma/login`, dll). Jika laju dipaksakan melebihi ambang batas, backend otomatis menolak request dengan status **`HTTP 429 Too Many Requests`**.
- **Otorisasi Otomatis Berbasis Tabel Database (`RBACRoleMenuMiddleware`)**: Backend Go mengeksekusi verifikasi hak akses dinamis ke tabel `lms_sch.role_menus` & `lms_sch.menus` sebelum mengeksekusi perintah API. Jika pengguna (misal Role ID 2 / Anggota) memanggil API transaksi/master yang tidak terdaftar di `lms_sch.role_menus`, backend otomatis menolak request dengan respon **`403 Forbidden`**.

---

## 12. Folder Billing

| Folder | Parameter Key | Default | Fungsi |
|---|---|---|---|
| `Billing/Export/` | `FOLDER_BILL_EXPORT` | `D:\...\Billing\Export` | Output CSV tagihan ke Adira |
| `Billing/Import/` | `FOLDER_BILL_IMPORT` | `D:\...\Billing\Import` | Input CSV hasil potongan dari Adira |
| `Billing/BCK/` | `FOLDER_BILL_IMPORT_BCK` | `D:\...\Billing\BCK` | Backup file diproses |

> **Catatan:** Backend otomatis konversi path `D:\...` → `/mnt/d/...` saat berjalan di WSL/Linux.

---



---

## 13. Global Parameters Kunci

| Parameter | Default | Deskripsi |
|---|---|---|
| `LOAN_START_PERIOD` | `1` | Tanggal awal periode diizinkan pengajuan pinjaman |
| `LOAN_END_PERIOD` | `31` | Tanggal akhir periode diizinkan pengajuan pinjaman |
| `LOAN_APPROVAL_AUTOMATIC` | `true` | Mode approval otomatis pengajuan pinjaman (`true` = otomatis, `false` = manual approval) |
| `LOAN_DISBURSE_AUTOMATIC` | `true` | Mode pencairan otomatis (`true` = otomatis terbit kontrak & jadwal, `false` = manual disburse) |
| `LOAN_DUEDATE` | `25` | Tanggal jatuh tempo angsuran setiap bulannya |
| `LOAN_DUEMONTH` | `0` | Offset bulan jatuh tempo angsuran pertama (`0` = bulan ini, `1` = bulan depan) |
| `LOAN_MAX_TENOR` | `12` | Batas maksimum tenor (bulan) secara global |
| `LOAN_LIMIT_FORMULA` | `SALARY * 0.5` | Formula perhitungan credit limit berbasis gaji/hari |
| `LOAN_ADMIN_FORMULA` | `REQUESTED_AMOUNT * 0.01` | Formula biaya administrasi pinjaman |
| `LOAN_MIN_ADMIN_FEE` | `30000` | Minimum biaya administrasi pinjaman (floor value) |
| `APPROVAL_MODE` | `MANUAL` | Mode persetujuan pinjaman (`MANUAL`: review Admin, `AUTO`: persetujuan otomatis) |
| `AUTO_APPROVAL_MAX_AMOUNT` | `0` | Maksimum nominal pinjaman untuk auto-approve (`0` = tanpa batasan nominal) |
| `DISBURSEMENT_MODE` | `MANUAL` | Mode pencairan dana (`MANUAL`: transfer Admin, `AUTO`: otomatis terbit kontrak & jadwal) |
| `PAGINATION_LIMIT` | `5` | Jumlah baris data per halaman untuk query tabel master/transaksi |
| `LOG_LOAN_TRANSACTION_PATH` | `./logs/loans.log` | Path dan nama file log transaksi seluruh siklus pinjaman |
| `LOG_WARNING_HIGH_PRIV` | `HIGH_PRIVILEGE_ACTION` | Pesan peringatan aktivitas pengguna high privilege (Admin/HRD) pada log |
| `HIGH_PRIVILEGE_ROLES` | `1,3,Admin,HRD,Administrator` | Daftar Role ID / Role Name yang dikategorikan sebagai High Privilege Users |
| `RC_SUBMIT_LOAN_SUCCESS` | `00` | Response Code transaksi pengajuan pinjaman sukses (`SUBMITTED`) |
| `RC_SUBMIT_LOAN_NON_KOPKARA` | `401` | Response Code ditolak bukan anggota Kopkara |
| `RC_SUBMIT_LOAN_NON_ADIRA` | `11` | Response Code ditolak bukan karyawan Adira |
| `RC_SUBMIT_LOAN_TENOR` | `12` | Response Code ditolak Tenor Melebihi Batas Maksimum |
| `RC_SUBMIT_LOAN_CREDIT_LIMIT` | `13` | Response Code ditolak Jumlah Pinjaman Melebihi Credit Limit |
| `RC_SUBMIT_LOAN_PERIOD` | `14` | Response Code ditolak Di Luar Periode Tanggal Pengajuan |
| `RC_SUBMIT_LOAN_OTHERS` | `99` | Response Code ditolak alasan lainnya (`OTHERS`) |

---

### 13.1 Spesifikasi Log Transaksi Pinjaman (Unified Loans Log) & Response Code (RC)

Seluruh siklus transaksi pinjaman (mulai dari **Submit**, **Approval**, **Pencairan/Disbursement**, **Import File HRD/Payroll**, **Pelunasan Manual**, hingga **Adjustment**) secara otomatis dicatat ke file log terpusat yang ditentukan oleh parameter `LOG_LOAN_TRANSACTION_PATH` (default: `./logs/loans.log`).

**Format Baris File Log (`./logs/loans.log`):**
```text
datetime, application_no, RC, employee_id, product_id, transaction_date, amount, tenor, status, created_user, role_name, high_priv_warning
```

**Penjelasan 12 Field Log:**
1. `datetime`: Timestamp presisi saat log ditulis (`YYYY-MM-DD HH:MM:SS`).
2. `application_no`: Nomor pengajuan pinjaman (`application_no`). Jika transaksi gagal (RC != `00`), bernilai `NULL`.
3. `RC`: Response Code / Result Code transaksi (`00` Sukses, `401` Non-Kopkara, `11` Non-Adira, `12` Tenor Exceed, `13` Limit Exceed, `14` Out of Period, `99` Others).
4. `employee_id`: ID karyawan / pemohon pinjaman.
5. `product_id`: ID produk pinjaman (`0` untuk pelunasan manual/payroll).
6. `transaction_date`: Tanggal efektif transaksi (`YYYY-MM-DD`).
7. `amount`: Nominal transaksi (requested_amount / approved_amount / repayment_amount).
8. `tenor`: Tenor pinjaman dalam bulan (`0` untuk pelunasan manual/payroll).
9. `status`: Status siklus transaksi (`SUBMITTED`, `APPROVED`, `DISBURSED`, `PAYROLL_RECONCILED`, `MANUAL_REPAYMENT`, `LOAN_ADJUSTMENT`, `REJECTED_*`).
10. `created_user`: Username/User ID pelaksana aksi. Khusus status `SUBMITTED` diisi user pemohon (`created_user`), sedangkan selain `SUBMITTED` diisi user pengubah/eksekutor (`updated_user`).
11. `role_name`: Nama Role pengguna yang mengeksekusi transaksi (`Administrator`, `HRD`, `Anggota`, `System Automation`).
12. `high_priv_warning`: Hanya diisi untuk status `SUBMITTED` jika role user terdaftar pada parameter `HIGH_PRIVILEGE_ROLES` (bernilai parameter `LOG_WARNING_HIGH_PRIV`). Selain `SUBMITTED` atau jika role bukan high privilege, diisi kosong (`""`).




---

### 13.2 Pengujian Validasi Backend via cURL (Perintah Real & Respon JSON)

Berikut adalah riwayat eksekusi perintah `curl` real beserta respon JSON asli dari server backend LMS:

#### 🔑 1. Login Ke Sistem (`POST /api/karisma/login`)
```bash
curl -k -c cookies.txt -X POST https://localhost:8086/api/karisma/login \
  -H "Content-Type: application/json" \
  -d "{\"username\": \"nur\", \"password\": \"Nkl@130200\"}"
```
**Respon Server:**
```json
{"status":"success","token":"820bfacb5c321dfdebd557837aa147eeba60fa03c8bd3c8185891daa62b0f171","token_name":"ewa_token"}
```

---

#### 🛑 2. Testing RC 11 — Ditolak Bukan Karyawan Adira (`REJECTED_NON_ADIRA`)
```bash
curl -k -X POST https://localhost:8086/api/applications \
  -H "Authorization: Bearer f2576101bd813faeacaf74b94f1764fd6d0c618403783f975b065ee575c17368" \
  -H "Content-Type: application/json" \
  -d "{\"member_no\": 999999, \"product_id\": 1, \"requested_amount\": 1000000, \"tenor\": 1}"
```
**Respon Server:**
```json
{"error":"ditolak bukan karyawan Adira"}
```

---

#### 🛑 3. Testing RC 12 — Ditolak Tenor Melebihi Batas Maksimum (`REJECTED_TENOR`)
```bash
curl -k -X POST https://localhost:8086/api/applications \
  -H "Authorization: Bearer f2576101bd813faeacaf74b94f1764fd6d0c618403783f975b065ee575c17368" \
  -H "Content-Type: application/json" \
  -d "{\"member_no\": 100001, \"product_id\": 1, \"requested_amount\": 1000000, \"tenor\": 24}"
```
**Respon Server:**
```json
{"error":"ditolak Tenor Melebihi Batas Maksimum (1 bulan)"}
```

---

#### 🛑 4. Testing RC 13 — Ditolak Jumlah Pinjaman Melebihi Credit Limit (`REJECTED_CREDIT_LIMIT`)
```bash
curl -k -X POST https://localhost:8086/api/applications \
  -H "Authorization: Bearer f2576101bd813faeacaf74b94f1764fd6d0c618403783f975b065ee575c17368" \
  -H "Content-Type: application/json" \
  -d "{\"member_no\": 200001, \"product_id\": 5, \"requested_amount\": 100000000, \"tenor\": 1}"
```
**Respon Server:**
```json
{"error":"Pengajuan ditolak. Total kewajiban pinjaman Anda (Rp 1.616.000) + pengajuan baru (Rp 101.000.000) = Rp 102.616.000 melebihi Credit Limit (Rp 4.583.333). Sisa limit yang dapat diajukan (Pokok + Bunga): Rp 2.967.333"}
```

---

#### ✅ 5. Testing RC 00 — Transaksi Pengajuan Berhasil (`SUBMITTED`)
```bash
curl -k -X POST https://localhost:8086/api/applications \
  -H "Authorization: Bearer f2576101bd813faeacaf74b94f1764fd6d0c618403783f975b065ee575c17368" \
  -H "Content-Type: application/json" \
  -d "{\"member_no\": 200001, \"product_id\": 5, \"requested_amount\": 100000, \"tenor\": 1}"
```
**Respon Server:**
```json
{
  "data": {
    "application_no": "1787405426153100276",
    "member_no": "200001",
    "product_id": 5,
    "submission_date": "2026-08-22T20:30:26.153076664+07:00",
    "requested_amount": 100000,
    "tenor": 1,
    "eligibility_result": "ELIGIBLE",
    "status": "SUBMITTED",
    "principal_per_month": 100000,
    "interest_per_month": 1000,
    "admin_fee": 30000,
    "total_installment": 101000,
    "total_loan_cost": 131000,
    "interest_rate": 1,
    "credit_limit": 4583333.333333333
  },
  "message": "Application submitted successfully"
}
```

---

#### 🛑 6. Testing RC 401 — Ditolak Bukan Anggota Kopkara (`REJECTED_NON_KOPKARA`)

Untuk menguji skenario **RC 401** (pengguna/karyawan memiliki kredensial valid tetapi `kopkara_status` di tabel `lms_sch.members` bernilai `'INACTIVE'` / `'RESIGNED'` atau belum menjadi Anggota Koperasi):

**Langkah Testing:**
1. Non-aktifkan status keanggotaan sementara di PostgreSQL (PgAdmin):
```sql
UPDATE lms_sch.members SET kopkara_status = 'INACTIVE' WHERE member_no = 200001;
```

2. Jalankan perintah `curl` submit loan:
```bash
curl -k -X POST https://localhost:8086/api/applications \
  -H "Authorization: Bearer f2576101bd813faeacaf74b94f1764fd6d0c618403783f975b065ee575c17368" \
  -H "Content-Type: application/json" \
  -d "{\"member_no\": 200001, \"product_id\": 1, \"requested_amount\": 1000000, \"tenor\": 1}"
```
**Respon Server:**
```json
{"error":"ditolak bukan anggota Kopkara"}
```
**Hasil Log (`logs/submit_loan.log`):**
```text
2026-08-22 20:51:00, 401, 100001, 1, 2026-08-22, 1000000, 1, REJECTED_NON_KOPKARA
```

3. Kembalikan status keanggotaan setelah testing selesai:
```sql
UPDATE lms_sch.members SET kopkara_status = 'ACTIVE' WHERE member_no = 200001;
```


### 13.3 Cara Testing RC 401 (Unauthorized)

Untuk menguji respon **`RC 401 / HTTP Status 401 Unauthorized`**, kirimkan request **tanpa header `Authorization`** atau menggunakan **token yang salah/kadaluarsa**:

```bash
curl -k -X POST https://localhost:8086/api/applications \
  -H "Authorization: Bearer invalid-token-12345" \
  -H "Content-Type: application/json" \
  -d "{\"member_no\": 200001, \"product_id\": 5, \"requested_amount\": 100000, \"tenor\": 1}"
```

**Respon Server (RC 401):**
```json
{"error":"Unauthorized: Akses ditolak. Token autentikasi tidak ditemukan di Cookie maupun Authorization Header."}
```

---

### 13.4 Daftar Lengkap Response Code (RC) Sistem LMS

#### A. Response Code Log Transaksi Submit Loan (`LOG_LOAN_TRANSACTION_PATH`)
| Kode RC | Parameter Global | Deskripsi Status |
|---|---|---|
| `00` | `RC_SUBMIT_LOAN_SUCCESS` | Transaksi pengajuan pinjaman berhasil (`SUBMITTED`) |
| `401` | `RC_SUBMIT_LOAN_NON_KOPKARA` | Ditolak bukan anggota Kopkara (`REJECTED_NON_KOPKARA`) |
| `11` | `RC_SUBMIT_LOAN_NON_ADIRA` | Ditolak bukan karyawan Adira (`REJECTED_NON_ADIRA`) |
| `12` | `RC_SUBMIT_LOAN_TENOR` | Ditolak Tenor Melebihi Batas Maksimum (`REJECTED_TENOR`) |
| `13` | `RC_SUBMIT_LOAN_CREDIT_LIMIT` | Ditolak Jumlah Pinjaman Melebihi Credit Limit (`REJECTED_CREDIT_LIMIT`) |
| `14` | `RC_SUBMIT_LOAN_PERIOD` | Ditolak Di Luar Periode Tanggal Pengajuan (`REJECTED_PERIOD`) |
| `99` | `RC_SUBMIT_LOAN_OTHERS` | Ditolak alasan lainnya / system error (`REJECTED_OTHERS`) |

#### B. Response Code HTTP Standard (System API RC)
| Status Code | Nama Status | Keterangan & Trigger |
|---|---|---|
| `200` / `201` | **OK / Created** | Request API berhasil diproses |
| `400` | **Bad Request** | Payload JSON tidak valid atau parameter request salah format |
| `401` | **Unauthorized** | Token autentikasi/cookie tidak ditemukan, salah, atau kadaluarsa |
| `403` | **Forbidden** | Otorisasi RBAC ditolak (Role ID pengguna tidak memiliki hak akses menu) |
| `404` | **Not Found** | Endpoint API atau data record (user/loan/application) tidak ditemukan |
| `429` | **Too Many Requests** | Laju request melebihi ambang batas Rate Limiter (`RATE_LIMIT_GENERAL_RPM`) |
| `500` | **Internal Server Error** | Eror internal server atau kegagalan query database PostgreSQL |


---

## 14. Mobile Self-Service EWA & Autentikasi Security

Sistem LMS telah dilengkapi dengan modul **Mobile Self-Service Registration**, **Autentikasi PIN 6-Digit**, **Dynamic Global Parameters Security**, dan **Integrasi WhatsApp OTP**.

---

### 14.1 Arsitektur Pendaftaran Self-Service (4-Factor Matching)

```
[ Pendaftaran Self-Service Mobile ]
               │
               ▼
 [ 1. Input 4-Factor Matching ]
 ├── No. KTP (16-Digit NIK)
 ├── Employee ID (NIP)
 ├── Nama Karyawan
 └── No. Handphone / Email
               │
               ▼
 [ 2. Validasi Match ke lms_sch.employees ]
 ├── Valid? -> Lanjut ke OTP
 └── Tidak Valid? -> Return MATCH_FAILED
               │
               ▼
 [ 3. Verifikasi WhatsApp OTP ]
 ├── Mode Mocking ($0) -> Static Code '123456'
 └── Mode Production   -> Meta WA Cloud API / Fonnte Gateway
               │
               ▼
 [ 4. Setup PIN 6-Digit ]
 ├── Validasi Larangan (Bukan 111111, 123456, 6-Digit NIK)
 └── Enkripsi Bcrypt Hash -> Save to lms_sch.users
               │
               ▼
 [ 5. Mobile Login (No. HP + PIN 6-Digit) ]
 ├── Lockout Counter (PIN_MAX_FAILED_ATTEMPTS = 3x)
 ├── Lockout Timeout (PIN_LOCKOUT_DURATION_MINUTES = 15m)
 └── Idle Timeout    (PIN_IDLE_TIMEOUT_MINUTES = 3m)
```

---

### 14.2 Parameter Keamanan Dinamis (`lms_sch.global_parameters`)

Seluruh parameter keamanan dibaca langsung melalui RAM Cache (`cache.ParameterCache`) dengan **0 SQL Query** pada runtime:

| Parameter Key | Default Value | Deskripsi Fungsi Keamanan |
|---|---|---|
| `PIN_MAX_FAILED_ATTEMPTS` | `3` | Batas maksimum salah PIN sebelum akun terkunci |
| `PIN_LOCKOUT_DURATION_MINUTES` | `15` | Durasi kuncian akun (menit) jika salah PIN 3x |
| `PIN_IDLE_TIMEOUT_MINUTES` | `3` | Waktu idle (menit) sebelum Mobile App terkunci otomatis |
| `PASSWORD_MAX_FAILED_ATTEMPTS` | `3` | Batas maksimum salah password Web App |
| `PASSWORD_ROTATION_DAYS` | `90` | Batas usia wajib rotasi password (hari) |
| `PASSWORD_MIN_LENGTH` | `9` | Panjang minimum password Web App |

---

### 14.3 Script DDL SQL Standalone (`alter_table_employees_and_users.sql`)

Eksekusi manual via PgAdmin:

```sql
-- 1. Penambahan Kolom pada Tabel lms_sch.employees
ALTER TABLE lms_sch.employees 
  ADD COLUMN IF NOT EXISTS no_ktp VARCHAR(20),
  ADD COLUMN IF NOT EXISTS phone_number VARCHAR(30),
  ADD COLUMN IF NOT EXISTS email VARCHAR(100);

-- 2. Penambahan Kolom pada Tabel lms_sch.users
ALTER TABLE lms_sch.users 
  ADD COLUMN IF NOT EXISTS no_ktp VARCHAR(20),
  ADD COLUMN IF NOT EXISTS phone_number VARCHAR(30),
  ADD COLUMN IF NOT EXISTS pin VARCHAR(255),
  ADD COLUMN IF NOT EXISTS failed_pin_attempts INT DEFAULT 0,
  ADD COLUMN IF NOT EXISTS pin_locked_until TIMESTAMP;

-- 3. Inserksi Parameter Keamanan Dinamis
INSERT INTO lms_sch.global_parameters (param_key, param_value, description, created_user, updated_user, created_at, updated_at)
VALUES 
  ('PIN_MAX_FAILED_ATTEMPTS', '3', 'Batas maksimum salah PIN sebelum akun terkunci', 'SYSTEM_AUTO', 'SYSTEM_AUTO', NOW(), NOW()),
  ('PIN_LOCKOUT_DURATION_MINUTES', '15', 'Durasi kuncian akun (menit) jika salah PIN 3x', 'SYSTEM_AUTO', 'SYSTEM_AUTO', NOW(), NOW()),
  ('PIN_IDLE_TIMEOUT_MINUTES', '3', 'Waktu idle (menit) sebelum Mobile App auto-lock', 'SYSTEM_AUTO', 'SYSTEM_AUTO', NOW(), NOW()),
  ('PASSWORD_MAX_FAILED_ATTEMPTS', '3', 'Batas maksimum salah password web app', 'SYSTEM_AUTO', 'SYSTEM_AUTO', NOW(), NOW()),
  ('PASSWORD_ROTATION_DAYS', '90', 'Batas rotasi password web (hari)', 'SYSTEM_AUTO', 'SYSTEM_AUTO', NOW(), NOW()),
  ('PASSWORD_MIN_LENGTH', '9', 'Panjang minimum password web app', 'SYSTEM_AUTO', 'SYSTEM_AUTO', NOW(), NOW())
ON CONFLICT (param_key) DO UPDATE 
SET param_value = EXCLUDED.param_value,
    description = EXCLUDED.description,
    updated_at = NOW();
```

---

### 14.4 End-to-End Verification Test Cases & Results

Berikut adalah hasil pengujian riil yang telah terverifikasi sukses (Verified Case Results):

#### 🟢 Test #1: Mobile Register Check (4-Factor Matching)
- **Endpoint**: `POST /api/v1/auth/mobile-register-check`
- **Request Payload**:
  ```json
  {
    "no_ktp": "234567",
    "employee_id": 100001,
    "name": "Nur Kholim",
    "phone_number": "085882500073"
  }
  ```
- **Response Server (Status 200 OK)**:
  ```json
  {
    "status": "MATCH_SUCCESS",
    "message": "Data karyawan terverifikasi valid dengan record HRD.",
    "is_registered": false,
    "has_pin": false,
    "employee": {
      "employee_id": 100001,
      "name": "Nur Kholim",
      "deptno": "DEPT01",
      "category_code": "PERM",
      "salary": 12500000,
      "employee_status": "ACTIVE"
    }
  }
  ```

#### 🟢 Test #2: Request & Verifikasi OTP
- **Endpoint Request OTP**: `POST /api/v1/auth/request-otp`
- **Endpoint Verify OTP**: `POST /api/v1/auth/verify-otp`
- **Request Payload**:
  ```json
  {
    "phone_number": "085882500073",
    "otp_code": "123456"
  }
  ```
- **Response Server (Status 200 OK)**:
  ```json
  {
    "status": "SUCCESS",
    "message": "Verifikasi OTP Berhasil!"
  }
  ```

#### 🟢 Test #3: Setup PIN 6-Digit
- **Endpoint**: `POST /api/v1/auth/setup-pin`
- **Request Payload**:
  ```json
  {
    "employee_id": 100001,
    "no_ktp": "234567",
    "phone_number": "085882500073",
    "pin": "859204"
  }
  ```
- **Response Server (Status 200 OK)**:
  ```json
  {
    "status": "SUCCESS",
    "message": "PIN 6-Digit berhasil didaftarkan! Anda kini dapat login menggunakan No. HP & PIN."
  }
  ```

#### 🟢 Test #4: Mobile Login (No. HP + PIN 6-Digit)
- **Endpoint**: `POST /api/v1/auth/mobile-login`
- **Request Payload**:
  ```json
  {
    "phone_number": "085882500073",
    "pin": "859204"
  }
  ```
- **Response Server (Status 200 OK)**:
  ```json
  {
    "status": "SUCCESS",
    "message": "Login Berhasil!",
    "idle_timeout_minutes": 3,
    "user": {
      "id": 50015,
      "username": "085882500073",
      "name": "Nur Kholim",
      "role_id": 2,
      "member_no": 200001,
      "phone_number": "085882500073"
    },
    "employee": {
      "employee_id": 100001,
      "name": "Nur Kholim",
      "deptno": "DEPT01",
      "category_code": "PERM",
      "salary": 12500000,
      "total_loan": 4603125,
      "employee_status": "ACTIVE"
    }
  }
  ```
