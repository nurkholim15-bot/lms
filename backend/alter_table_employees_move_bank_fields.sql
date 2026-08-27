-- =====================================================================
-- SCRIPT DDL: PEMINDAHAN FIELD REKENING BANK & DROP DARI TABEL MEMBERS
-- Dijalankan di pgAdmin Query Tool
-- Target Schema: lms_sch
-- =====================================================================

-- 1. Tambahkan 3 kolom informasi bank ke tabel lms_sch.employees
ALTER TABLE lms_sch.employees ADD COLUMN IF NOT EXISTS bank_name VARCHAR(100);
ALTER TABLE lms_sch.employees ADD COLUMN IF NOT EXISTS bank_account_no VARCHAR(50);
ALTER TABLE lms_sch.employees ADD COLUMN IF NOT EXISTS bank_account_name VARCHAR(100);

-- 2. Migrasikan data eksisting dari lms_sch.members ke lms_sch.employees (jika ada)
UPDATE lms_sch.employees e
SET bank_name = COALESCE(e.bank_name, m.bank_name),
    bank_account_no = COALESCE(e.bank_account_no, m.bank_account_no),
    bank_account_name = COALESCE(e.bank_account_name, m.bank_account_name)
FROM lms_sch.members m
WHERE e.employee_id = m.employee_id
  AND (e.bank_name IS NULL OR e.bank_account_no IS NULL OR e.bank_account_name IS NULL);

-- 3. Hapus (DROP) 3 kolom bank lama dari tabel lms_sch.members
ALTER TABLE lms_sch.members DROP COLUMN IF EXISTS bank_name;
ALTER TABLE lms_sch.members DROP COLUMN IF EXISTS bank_account_no;
ALTER TABLE lms_sch.members DROP COLUMN IF EXISTS bank_account_name;

-- 4. Verifikasi hasil pemindahan di tabel lms_sch.employees
SELECT employee_id, name, bank_name, bank_account_no, bank_account_name 
FROM lms_sch.employees 
ORDER BY employee_id ASC;
