-- =========================================================================
-- DML ALTER & UPDATE: PENAMBAHAN KOLOM NOTIFICATION_TYPE PADA LMS_SCH.MENUS
-- =========================================================================
-- Skrip DML murni tanpa data seeding dummy.
-- Nilai notification_type:
-- 0 = Tidak Kirim Notifikasi (OFF)
-- 1 = Kirim Email Saja
-- 2 = Kirim WhatsApp Saja
-- 3 = Kirim Email & WhatsApp (Keduanya)
-- =========================================================================

-- 1. Tambahkan kolom notification_type ke tabel lms_sch.menus jika belum ada
ALTER TABLE lms_sch.menus 
ADD COLUMN IF NOT EXISTS notification_type INT DEFAULT 0;

COMMENT ON COLUMN lms_sch.menus.notification_type IS '0=None/OFF, 1=Email Only, 2=WhatsApp Only, 3=Email & WhatsApp';

-- 2. Update default notification_type untuk menu transaksi utama (Contoh: Pencairan = 3, Pelunasan = 3)
UPDATE lms_sch.menus SET notification_type = 3 WHERE path IN ('disbursement', 'manual-repayment', 'payroll-reconciliation') AND notification_type = 0;
