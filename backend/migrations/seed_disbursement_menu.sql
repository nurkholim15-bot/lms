-- =========================================================================
-- DML SQL UNTUK MENAMBAHKAN MENU "Pencairan Dana" (path: 'disbursement')
-- SERTA OTORISASI ROLE_MENUS DI POSTGRESQL (PgAdmin / LMS)
-- =========================================================================

-- 1. Tambah / Pastikan Menu "Pencairan Dana" (path: 'disbursement') ada di tabel lms_sch.menus
INSERT INTO lms_sch.menus (title, icon, path, parent_id, order_seq)
SELECT 'Pencairan Dana', '💸', 'disbursement', NULL, 5
WHERE NOT EXISTS (
    SELECT 1 FROM lms_sch.menus WHERE path = 'disbursement'
);

-- Jika menu dengan path 'disbursement' sudah ada, update informasi title, icon, dan order_seq
UPDATE lms_sch.menus 
SET title = 'Pencairan Dana', 
    icon = '💸', 
    order_seq = 5 
WHERE path = 'disbursement';

-- 2. Berikan Hak Akses Otorisasi pada Tabel lms_sch.role_menus untuk Admin (Role ID: 1) dan HRD (Role ID: 3)
INSERT INTO lms_sch.role_menus (role_id, menu_id)
SELECT 1, menu_id FROM lms_sch.menus WHERE path = 'disbursement'
ON CONFLICT DO NOTHING;

INSERT INTO lms_sch.role_menus (role_id, menu_id)
SELECT 3, menu_id FROM lms_sch.menus WHERE path = 'disbursement'
ON CONFLICT DO NOTHING;

-- Verifikasi Data Menu & Hak Akses Role setelah Eksekusi
SELECT m.menu_id, m.title, m.icon, m.path, rm.role_id 
FROM lms_sch.menus m
LEFT JOIN lms_sch.role_menus rm ON m.menu_id = rm.menu_id
WHERE m.path = 'disbursement';
