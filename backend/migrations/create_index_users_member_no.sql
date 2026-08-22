-- =========================================================================
-- DDL/SQL SCRIPT UNTUK MEMBUAT UNIQUE INDEX FIELD member_no PADA TABEL lms_sch.users
-- MANUAL EXECUTION VIA PGADMIN (TANPA AUTO-MIGRATION / SEEDING BACKEND)
-- =========================================================================

-- OPSE 1: UNIQUE INDEX (Direkomendasikan agar 1 Member_No hanya memiliki 1 Akun User)
-- Di PostgreSQL, UNIQUE index tetap mengizinkan multiple NULL (untuk akun Admin/Non-Anggota)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_member_no 
ON lms_sch.users (member_no);

-- OPSE 2: NON-UNIQUE INDEX (Gunakan skrip di bawah jika 1 Member_No diperbolehkan memiliki > 1 akun)
-- DROP INDEX IF EXISTS lms_sch.idx_users_member_no;
-- CREATE INDEX IF NOT EXISTS idx_users_member_no ON lms_sch.users (member_no);

-- Verifikasi Keberadaan Index setelah Eksekusi di PgAdmin
SELECT 
    schemaname, 
    tablename, 
    indexname, 
    indexdef 
FROM pg_indexes 
WHERE schemaname = 'lms_sch' 
  AND tablename = 'users' 
  AND indexname = 'idx_users_member_no';
