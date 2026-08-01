# Proposal: Loan Management System (LMS) Koperasi Karyawan

## Pendahuluan
Loan Management System (LMS) adalah istilah industri yang paling standar dan akurat. LMS mencakup seluruh *lifecycle* pinjaman, mulai dari pengajuan (*loan origination*), kalkulasi bunga/margin, penetapan jadwal angsuran, pencairan (*disbursement*), pemotongan gaji (*payroll deduction*), hingga pelunasan dan manajemen tunggakan.

Berikut adalah rancangan **Data Flow Diagram (DFD)** dan **Detail Proses** untuk **Loan Management System (LMS) Koperasi Karyawan**. Rancangan ini disesuaikan dengan karakteristik khas koperasi karyawan, seperti integrasi ke sistem HRD/Payroll untuk verifikasi gaji dan skema pemotongan otomatis.

## 1. System Architecture
* **System architecture**: MVC
* **Backend**: Golang
* **Frontend**: React JS
* **Database**: PostgreSQL
* **UI**: Sama dengan LIMS. Background warna biru. Logo, background dan nama system dikonfigurasi di global parameter.

## 2. DFD Level 0 (Context Diagram)
Context Diagram menggambarkan batasan sistem dan interaksi LMS dengan entitas luar (*External Entities*).

```text
               |   Tim Payroll /   |
               |   HRD Perusahaan  |
               +---------+---------+
                         |
      (Rincian Potongan Gaji / Status Karyawan)
                         |
                         v
+--------------+    +-------------------+    +--------------+
|              |--->|                   |--->|              |
|   Anggota    |    |  LOAN MANAGEMENT  |    |  Pengurus /  |
|  (Karyawan)  |<---|   SYSTEM (LMS)    |<---| Komite Kredit|
|              |    |     KOPERASI      |    |   Koperasi   |
+--------------+    +-------------------+    +--------------+
```

### Entitas Luar & Aliran Data:
1. **Anggota (Karyawan)**:
   * **Input**: Form Pengajuan Pinjaman, Dokumen Pendukung, Konfirmasi Pelunasan Cepat.
   * **Output**: Informasi Limit Kredit, Status Persetujuan, Jadwal Angsuran, Bukti Pencairan.
2. **Pengurus / Komite Kredit Koperasi**:
   * **Input**: Persetujuan/Penolakan Pinjaman, Pengaturan Parameter Bunga/Plafon.
   * **Output**: Laporan Pengajuan, Laporan Analisis Kredit, Laporan Portofolio Pinjaman (NPL/Daftar Tunggakan).
3. **Tim Payroll / HRD Perusahaan**:
   * **Input**: Verifikasi Status Karyawan & Masa Kerja, Konfirmasi Potongan Gaji Bulanan.
   * **Output**: File/Daftar Tagihan Potongan Gaji Karyawan (*Payroll Deduction File*).

## 3. Detail Proses LMS
Berikut adalah rincian logika proses dari setiap tahapan pada DFD Level 1:

### Proses 1.0: Login Existing System Karisma
* **Deskripsi**: Credentials Anggota
* **Aturan Bisnis (Business Rules)**: User input username dan password aplikasi Karisma
* **Output**: User bisa akses Sistem Karisma.

### Proses 2.0: Buka Sub-menu LMS
* **Deskripsi**: User buka sub-menu LMS untuk akses cash-advance
* **Aturan Bisnis (Business Rules)**:
  * User klik tombol sub-menu LMS
  * LMS akan melakukan pengecekan token dan username di cache atau local storage
  * Jika:
    * Session ketemu: sistem LMS akan melakukan validasi ke API Karisma dan Karisma akan meresponse expiration session dan username.
    * LMS akan mengecek username dari cache dan Karisma.
    * Session tidak ketemu atau/dan expired atau/dan username tidak sama maka LMS akan memberikan notifikasi session tidak valid.
    * LMS akan melakukan pengecekan eligible anggota Kopkara di database berdasarkan username.
    * Jika bukan anggota Kopkara maka LMS akan menampilkan notifikasi "Anda bukan anggota Kopkara".
* **Output**: User bisa akses LMS.

### Proses 3.1: Dashboard Credit Limit, Total Hutang & Sisa Credit Limit
* **Deskripsi**:
  * Sistem menghitung limit maksimal pinjaman yang bisa diajukan anggota secara otomatis sebelum pengajuan dibuat.
  * Sistem menghitung total hutang dan sisa Credit Limit.
* **Aturan Bisnis (Business Rules)**:
  * Plafon pinjaman berbasis masa kerja dan/atau kelipatan simpanan (misal: Max $3 \times \text{Total Simpanan}$).
  * Debt Service Ratio (DSR): Total angsuran (termasuk angsuran eksisting) tidak boleh melebihi **40%–50% dari gaji pokok** karyawan.
  * Persentase limit pinjaman (plafon) disimpan di table `global_parameters`.
* **Output**: Informasi Available Credit Limit, total hutang dan sisa Credit Limit pada dashboard anggota.

### Proses 3.2: Input Nominal & Lihat Simulasi
* **Deskripsi**: User menginput nominal pinjaman dan tenor.
* **Aturan Bisnis (Business Rules)**:
  * Rumus biaya administrasi dan minimal biaya administrasi disimpan di table `global_parameters`.
  * Jika perhitungan biaya administrasi < minimal biaya administrasi maka biaya administrasi adalah biaya minimal administrasi.
  * Rate perhitungan bunga diambil dari table product.

### Proses 3.2 (Lanjutan): Konfirmasi Pengajuan (Loan Origination)
* **Deskripsi**: Anggota memilih nominal, tenor (jangka waktu), dan jenis pinjaman (Reguler, Darurat, atau Khusus).
* **Langkah Kerja**:
  1. Input nominal & tenor. Maksimal tenor disimpan di `global_parameters`.
  2. Validasi Periode Pengajuan: Hanya diizinkan pada tanggal 1 – 15. Periode Pengajuan dimasukan dalam `global_parameters`.
  3. Ambil Data Salary & Kategori: Menarik data gaji dan level kategori karyawan.
  4. Hitung Maksimum Pinjaman: Maksimal x% dari Salary & batas plafon kategori. Maksimum salary dimasukan dalam `global_parameters` dan batas plafon kategori disimpan di table `category_employees`.
  5. Hitung Prorata: Berdasarkan rumus `Limit = (Tanggal Pengajuan / 30) × x% × Nominal Salary`. Contoh pengajuan tanggal 3 dan misal x:50% = 3/30 × 50% × salary. Rumus tsb dimasukan dalam `global_parameters`.
  6. Sistem merender simulasi angsuran (Pokok + Bunga Flat/Anuitas/Efektif).
  7. Pembentukan draf akad/perjanjian pinjaman digital.

#### Definisi Produk
| Parameter | Ketentuan Awal | Status |
| --- | --- | --- |
| Nama Produk | Advance Salary | Confirmed |
| Jenis Produk | Pinjaman koperasi jangka pendek | Confirmed |
| Pengguna | Karyawan Adira yang menjadi anggota Kopkara | Confirmed |
| Jangka Waktu | 1 bulan | Confirmed |
| Periode Pengajuan | Tanggal 1-15 setiap bulan | Confirmed |
| Pembayaran | Potong gaji | Confirmed |
| Maksimum Dasar | Sampai dengan 50% dari salary | Confirmed - detail perlu validasi |
| Limit Kategori | Ditentukan berdasarkan kategori karyawan | Pending parameter |
| Prorata | Berdasarkan tanggal pengajuan | Pending formula |
| Biaya Administrasi | Configurable dan memiliki nilai minimum | Pending formula |
| Kanal Akses | Submenu pada aplikasi Karisma | Confirmed |
| Target Go-Live | Akhir tahun 2026 | Target |

#### Informasi, Eligibility, dan Simulasi
| ID | Requirement |
| --- | --- |
| FR-07 | Sistem menampilkan status eligibility pengguna beserta alasan apabila tidak memenuhi syarat. |
| FR-08 | Sistem menampilkan periode pengajuan yang berlaku. |
| FR-09 | Sistem mengambil atau menerima data salary sebagai dasar perhitungan. |
| FR-10 | Sistem mengambil kategori karyawan dan parameter limit yang sesuai. |
| FR-11 | Sistem menghitung maksimum pinjaman secara real-time. |
| FR-12 | Sistem menyediakan simulasi nominal pinjaman, biaya administrasi, nilai bersih, dan nilai potong gaji. |
| FR-13 | Sistem menampilkan rincian formula/perhitungan secara transparan kepada pengguna. |

#### Pengajuan Pinjaman
| ID | Requirement |
| --- | --- |
| FR-14 | Pengguna dapat memilih atau memasukkan nominal pinjaman selama tidak melebihi limit. |
| FR-15 | Sistem menolak nominal yang melebihi limit dan memberikan pesan yang jelas. |
| FR-16 | Sistem menampilkan syarat dan ketentuan sebelum pengguna mengonfirmasi pengajuan. |
| FR-17 | Sistem mencatat tanggal/waktu pengajuan dan menghasilkan nomor pengajuan. |
| FR-18 | Sistem menyimpan nominal, biaya administrasi, nilai bersih, periode payroll, dan status pengajuan. |
| FR-19 | Sistem mencegah duplikasi pengajuan sesuai kebijakan pinjaman aktif. |

### Proses 3.3: Evaluasi Credit
* **Deskripsi**: Evaluasi pengajuan oleh Komite Kredit.
* **Langkah Kerja**:
  1. **Routing Approval**: Jika nominal Credit tertentu (misal: Pinjaman Darurat < Rp 2 Juta), sistem dapat melakukan *Auto-Approval*. Jika di atas threshold, masuk ke *queue* Pengurus.
  2. Nilai credit threshold disimpan di `global_parameters`.
  3. Pengurus memeriksa kelayakan, riwayat pembayaran masa lalu (*repayment history*), dan sisa masa kontrak kerja.
  4. Pengurus memberikan keputusan: **Approve, Reject, Revisi** (dengan alasan).
  5. Alasan reject dimasukan dalam table misal status kontrak, dsb.
  6. Jika keputusan Revisi, maka kembali ke proses 3.1.

#### Approval dan Keputusan
| ID | Requirement |
| --- | --- |
| FR-20 | Sistem mendukung straight-through approval apabila seluruh syarat terpenuhi dan kebijakan memperbolehkan. |
| FR-21 | Sistem dapat mengarahkan pengajuan ke approver apabila diperlukan. |
| FR-22 | Approver dapat menyetujui, menolak, atau meminta perbaikan. |
| FR-23 | Sistem mencatat pengguna, waktu, keputusan, dan catatan approval. |
| FR-24 | Sistem mengirimkan notifikasi status keputusan melalui kanal yang disepakati. |

### Proses 4.0: Pencairan Dana (Disbursement)
* **Deskripsi**: Eksekusi transfer dana ke rekening bank anggota setelah pengajuan disetujui.
* **Langkah Kerja**:
  1. Generate nomor kontrak/rekening pinjaman resmi.
  2. Pembentukan jadwal angsuran (*Amortization Schedule*).
  3. Pemotongan biaya administrasi/provisi/asuransi di awal (jika ada).
  4. Hitung Pencairan Bersih (Nominal Pinjaman - Biaya Admin).
  5. Pengiriman instruksi pencairan dana ke sistem pembayaran/kasir koperasi.
  6. LMS akan generate akad/perjanjian pinjaman digital.

#### Pencairan, Potong Gaji, dan Pelunasan
| ID | Requirement |
| --- | --- |
| FR-25 | Sistem mencatat nilai pinjaman yang disetujui dan nilai pencairan bersih. |
| FR-26 | Sistem menentukan periode potong gaji sesuai business rule. |
| FR-27 | Sistem membentuk data potong gaji untuk dikirim atau di-upload ke Adira. |
| FR-28 | Sistem mencatat hasil pencairan dan status pengiriman data payroll. |
| FR-29 | Sistem menerima atau mengimpor hasil payroll dari Adira. |
| FR-30 | Sistem memperbarui status pembayaran, outstanding, dan pelunasan. |
| FR-31 | Sistem menangani transaksi gagal potong sesuai mekanisme retry/exception yang disepakati. |

### Proses 5.0: Kirim/Upload Data Potong Gaji ke Adira
* **Deskripsi**: Pengelolaan tagihan bulanan dan rekonsiliasi pembayaran angsuran.
* **Langkah Kerja**:
  * **Cut-off Date**: Setiap tanggal tertentu (misal: tanggal 20), sistem menarik daftar angsuran aktif untuk bulan berjalan. Tanggal cut-off data dimasukan dalam `global_parameters`.
  * **Generate Payroll File**: Sistem mengirimkan daftar potong gaji ke bagian HRD/Payroll Perusahaan.

### Proses 6.0: Terima & Rekonsiliasi Hasil Payroll
* **Deskripsi**: Upload data payment dari HRD Adira.
* **Langkah Kerja**:
  1. Terima data balik dari Sistem Adira setelah siklus penggajian selesai.
  2. Update status pembayaran dan catat nominal yang berhasil terpotong di LMS.
  3. Catat transaksi gagal jika ada kendala pemotongan.
  4. Update sisa kewajiban/utang anggota.

### Proses 7.0: Prepayment
* **Deskripsi**: Penanganan khusus jika anggota ingin melunasi seluruh sisa pinjaman sebelum jatuh tempo (menghitung sisa pokok + penalti/diskon bunga jika ada).
* **Langkah Kerja**:
  * Sistem akan menghitung sisa pokok hutang + penalty/diskon bunga jika ada.
  * User menerima pembayaran dan input ke LMS.
  * LMS akan generate kuitansi pelunasan.

### Proses 8.0: Pelaporan
**Fungsi**:
1. Menyediakan *Dashboard & Daftar Pinjaman* untuk monitoring status pengajuan, pencairan, potong gaji, dan pelunasan.
2. Menghasilkan Laporan Pengajuan, Pencairan, Potong Gaji, dan Error/Gagal Kirim.
3. Status Pelunasan: Mengubah status pinjaman menjadi Lunas setelah sisa hutang 0.

#### Administrasi, Monitoring, dan Audit
| ID | Requirement |
| --- | --- |
| FR-32 | Administrator dapat mengelola parameter produk, periode, limit, kategori karyawan, biaya administrasi, dan minimum biaya. |
| FR-33 | Sistem menyediakan daftar pengajuan dan transaksi berdasarkan status. |
| FR-34 | Sistem menyediakan pencarian berdasarkan employee ID, nomor anggota, nomor pengajuan, atau nomor pinjaman. |
| FR-35 | Sistem menampilkan transaksi yang gagal dikirim atau gagal diproses beserta alasan kegagalan. |
| FR-36 | Sistem menyediakan audit trail untuk perubahan parameter, transaksi, keputusan, dan aktivitas integrasi. |
| FR-37 | Sistem mendukung export laporan sesuai format yang disepakati. |

## Minimal Field Tiap Table
| Entitas | Data Utama | Sumber Awal |
| --- | --- | --- |
| Employee | Employee ID, nama, status kerja, unit kerja, kategori karyawan | Adira/Karisma |
| Anggota | Nomor anggota, status anggota, tanggal aktif, status keanggotaan | Karisma/Core Kopkara |
| Salary | Periode salary, nominal salary, jenis salary, sumber data, waktu pembaruan | Adira/Upload |
| Produk | Tenor, periode pengajuan, persentase maksimum, status produk | Parameter Kopkara |
| Kategori Karyawan | Kode kategori, deskripsi, maksimum limit, eligibility | Parameter Kopkara |
| Biaya Administrasi | Jenis biaya, nominal/persentase, minimum, periode berlaku | Parameter Kopkara |
| Pengajuan | Nomor, tanggal, nominal, hasil eligibility, hasil simulasi, status | Advance Salary |
| Pinjaman | Nilai disetujui, biaya, pencairan, outstanding, status pelunasan | Advance Salary/Core Kopkara |
| Payroll | Periode potong, nominal, hasil pemotongan, tanggal proses, alasan gagal | Adira |
| Integrasi | Batch ID, waktu kirim/terima, status, jumlah record, error message | Integration Log |
| Audit | User, aktivitas, waktu, nilai sebelum/sesudah, sumber perubahan | Audit Trail |
