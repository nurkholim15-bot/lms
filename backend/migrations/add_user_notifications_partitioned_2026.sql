-- =========================================================================
-- DDL POSTGRESQL: TABEL USER_NOTIFICATIONS BERPARTISI BULANAN (2026 & 2027)
-- =========================================================================
-- Mengubah / Mengkonversi lms_sch.user_notifications menjadi Tabel Berpartisi 
-- berbasis Range kolom created_at (12 Bulan 2026 & 12 Bulan 2027 + Default).
-- =========================================================================

-- 1. Drop tabel unpartitioned lama jika sudah pernah dibuat tanpa partisi
DROP TABLE IF EXISTS lms_sch.user_notifications CASCADE;

-- 2. Sequence untuk Auto-Increment Primary Key ID
CREATE SEQUENCE IF NOT EXISTS lms_sch.user_notifications_id_seq;

-- 3. Pembuatan Tabel Utama Berpartisi Range (created_at)
CREATE TABLE IF NOT EXISTS lms_sch.user_notifications (
    id              BIGINT NOT NULL DEFAULT nextval('lms_sch.user_notifications_id_seq'),
    user_id         BIGINT NOT NULL,                  -- Member No / Employee ID Penerima
    title           VARCHAR(255) NOT NULL,            -- Judul Singkat Notifikasi
    message         TEXT NOT NULL,                    -- Isi Pesan Ringkas
    category        VARCHAR(50) NOT NULL,             -- DISBURSEMENT, REPAYMENT, APPROVAL, SYSTEM
    reference_type  VARCHAR(50),                      -- LOAN_APPLICATION, PAYROLL_IMPORT, MANUAL_REPAYMENT
    reference_no    VARCHAR(100),                     -- No. Pengajuan / No. Transaksi / Loan No
    action_url      VARCHAR(255),                     -- Router Path UI (misal: 'loans?app_no=123')
    is_read         BOOLEAN DEFAULT FALSE,            -- Status Dibaca (FALSE = Unread, TRUE = Read)
    read_at         TIMESTAMP WITH TIME ZONE,         -- Waktu Dibaca
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_user    VARCHAR(100) DEFAULT 'SYSTEM',
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_user    VARCHAR(100) DEFAULT 'SYSTEM',
    deleted_at      TIMESTAMP WITH TIME ZONE,
    deleted_user    VARCHAR(100),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- 4. Pembuatan Partisi Bulanan Tahun 2026 (12 Bulan)
CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202601 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2026-01-01 00:00:00+00') TO ('2026-02-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202602 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2026-02-01 00:00:00+00') TO ('2026-03-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202603 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2026-03-01 00:00:00+00') TO ('2026-04-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202604 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2026-04-01 00:00:00+00') TO ('2026-05-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202605 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2026-05-01 00:00:00+00') TO ('2026-06-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202606 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2026-06-01 00:00:00+00') TO ('2026-07-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202607 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2026-07-01 00:00:00+00') TO ('2026-08-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202608 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2026-08-01 00:00:00+00') TO ('2026-09-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202609 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2026-09-01 00:00:00+00') TO ('2026-10-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202610 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2026-10-01 00:00:00+00') TO ('2026-11-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202611 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2026-11-01 00:00:00+00') TO ('2026-12-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202612 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2026-12-01 00:00:00+00') TO ('2027-01-01 00:00:00+00');

-- 5. Pembuatan Partisi Bulanan Tahun 2027 (12 Bulan)
CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202701 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2027-01-01 00:00:00+00') TO ('2027-02-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202702 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2027-02-01 00:00:00+00') TO ('2027-03-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202703 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2027-03-01 00:00:00+00') TO ('2027-04-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202704 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2027-04-01 00:00:00+00') TO ('2027-05-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202705 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2027-05-01 00:00:00+00') TO ('2027-06-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202706 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2027-06-01 00:00:00+00') TO ('2027-07-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202707 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2027-07-01 00:00:00+00') TO ('2027-08-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202708 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2027-08-01 00:00:00+00') TO ('2027-09-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202709 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2027-09-01 00:00:00+00') TO ('2027-10-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202710 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2027-10-01 00:00:00+00') TO ('2027-11-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202711 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2027-11-01 00:00:00+00') TO ('2027-12-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_202712 PARTITION OF lms_sch.user_notifications
    FOR VALUES FROM ('2027-12-01 00:00:00+00') TO ('2028-01-01 00:00:00+00');

-- 6. Partisi Catch-All Default (Mencegah Error jika ada data di luar range)
CREATE TABLE IF NOT EXISTS lms_sch.user_notifications_default PARTITION OF lms_sch.user_notifications DEFAULT;

-- 7. Pembuatan Indeks Performa pada Tabel Utama Berpartisi
CREATE INDEX IF NOT EXISTS idx_user_notif_user_unread 
ON lms_sch.user_notifications(user_id, is_read) 
WHERE deleted_at IS NULL;
