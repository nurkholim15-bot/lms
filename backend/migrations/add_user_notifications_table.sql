-- =========================================================================
-- DML SQL: PEMBUATAN TABEL LENDING IN-APP NOTIFICATIONS (LMS_SCH.USER_NOTIFICATIONS)
-- =========================================================================

CREATE TABLE IF NOT EXISTS lms_sch.user_notifications (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL,                  -- Member No / Employee ID Penerima
    title           VARCHAR(255) NOT NULL,            -- Judul Singkat Notifikasi
    message         TEXT NOT NULL,                    -- Isi Pesan Ringkas
    category        VARCHAR(50) NOT NULL,             -- DISBURSEMENT, REPAYMENT, APPROVAL, SYSTEM
    reference_type  VARCHAR(50),                      -- LOAN_APPLICATION, PAYROLL_IMPORT, MANUAL_REPAYMENT
    reference_no    VARCHAR(100),                     -- No. Pengajuan / No. Transaksi / Loan No
    action_url      VARCHAR(255),                     -- Router Path UI (misal: 'loans?app_no=123')
    is_read         BOOLEAN DEFAULT FALSE,            -- Status Dibaca (FALSE = Unread, TRUE = Read)
    read_at         TIMESTAMP WITH TIME ZONE,         -- Waktu Dibaca
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_user    VARCHAR(100) DEFAULT 'SYSTEM',
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_user    VARCHAR(100) DEFAULT 'SYSTEM',
    deleted_at      TIMESTAMP WITH TIME ZONE,
    deleted_user    VARCHAR(100)
);

-- Index untuk Performa Kueri Cepat (< 5ms)
CREATE INDEX IF NOT EXISTS idx_user_notif_user_unread 
ON lms_sch.user_notifications(user_id, is_read) 
WHERE deleted_at IS NULL;
