-- =========================================================================
-- DDL/SQL SCRIPT UNTUK MEMBUAT INDEX FIELD member_no PADA TABEL lms_sch.users
-- MANUAL EXECUTION VIA PGADMIN (TANPA AUTO-MIGRATION / SEEDING BACKEND)
-- =========================================================================

-- 1. Buat B-Tree Index pada kolom member_no di tabel lms_sch.users jika belum ada
CREATE INDEX IF NOT EXISTS idx_users_member_no 
ON lms_sch.users (member_no);

-- 2. Query Verifikasi Keberadaan Index setelah Eksekusi di PgAdmin
SELECT 
    schemaname, 
    tablename, 
    indexname, 
    indexdef 
FROM pg_indexes 
WHERE schemaname = 'lms_sch' 
  AND tablename = 'users' 
  AND indexname = 'idx_users_member_no';
