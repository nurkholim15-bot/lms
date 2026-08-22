-- =========================================================================
-- DML SQL UNTUK MENAMBAHKAN MENU "Pengaturan Parameter" DAN "Produk Pinjaman"
-- SERTA OTORISASI ROLE_MENUS DI POSTGRESQL (PgAdmin)
-- =========================================================================

-- 1. Tambah/Pastikan Menu "Produk Pinjaman" (path: 'products') dan "Pengaturan Parameter" (path: 'parameters') ada di tabel lms_sch.menus
INSERT INTO lms_sch.menus (menu_id, parent_id, title, icon, path, order_seq) VALUES
(7, NULL, 'Produk Pinjaman', '📦', 'products', 7),
(8, NULL, 'Pengaturan Parameter', '⚙️', 'parameters', 8)
ON CONFLICT (menu_id) DO UPDATE SET 
    title = EXCLUDED.title, 
    icon = EXCLUDED.icon, 
    path = EXCLUDED.path,
    order_seq = EXCLUDED.order_seq;

-- 2. Berikan Hak Akses Otorisasi pada Tabel lms_sch.role_menus untuk Admin (Role ID: 1), Anggota (Role ID: 2), dan HRD (Role ID: 3)
-- Admin (Role 1): Produk Pinjaman & Pengaturan Parameter
INSERT INTO lms_sch.role_menus (role_id, menu_id) VALUES
(1, 7),
(1, 8)
ON CONFLICT DO NOTHING;

-- Anggota (Role 2): Produk Pinjaman
INSERT INTO lms_sch.role_menus (role_id, menu_id) VALUES
(2, 7)
ON CONFLICT DO NOTHING;

-- HRD (Role 3): Produk Pinjaman
INSERT INTO lms_sch.role_menus (role_id, menu_id) VALUES
(3, 7)
ON CONFLICT DO NOTHING;
