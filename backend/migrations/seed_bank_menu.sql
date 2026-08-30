-- =========================================================================
-- DML SQL UNTUK MENAMBAHKAN MENU "Master Bank" (path: 'master-banks')
-- DAN OTORISASI ROLE_MENUS UNTUK ADMIN (ROLE_ID: 1)
-- =========================================================================

-- 1. Tambah / Pastikan Menu "Master Bank" (menu_id: 912) ada di tabel lms_sch.menus
INSERT INTO lms_sch.menus (menu_id, parent_id, title, icon, path, order_seq) 
VALUES (912, 9, 'Master Bank', '🏦', 'master-banks', 912)
ON CONFLICT (menu_id) DO UPDATE SET 
    parent_id = EXCLUDED.parent_id,
    title = EXCLUDED.title, 
    icon = EXCLUDED.icon, 
    path = EXCLUDED.path,
    order_seq = EXCLUDED.order_seq,
    deleted_at = NULL;

-- 2. Berikan Hak Akses Otorisasi pada Tabel lms_sch.role_menus untuk Admin (Role ID: 1)
INSERT INTO lms_sch.role_menus (role_id, menu_id, created_user) 
VALUES (1, 912, 'SYSTEM')
ON CONFLICT (role_id, menu_id) DO UPDATE SET deleted_at = NULL;
