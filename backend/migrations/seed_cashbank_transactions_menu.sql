-- =========================================================================
-- DML SKRIP REGISTRASI MENU TRANSAKSI & MUTASI KAS BANK DAN LAPORAN AUDIT (WITHOUT SEED DATA)
-- =========================================================================

-- 1. Register Menu Transaksi & Mutasi Kas Bank dan Laporan Audit ke lms_sch.menus
INSERT INTO lms_sch.menus (menu_id, parent_id, title, icon, path, order_seq) VALUES
(915, 9, 'Transaksi & Mutasi Kas Bank', '💸', 'master-cashbank-transactions', 915),
(916, 9, 'Laporan Audit Kas & Bank', '📊', 'report-cashbank-transactions', 916)
ON CONFLICT (menu_id) DO UPDATE SET 
    parent_id = EXCLUDED.parent_id,
    title = EXCLUDED.title, 
    icon = EXCLUDED.icon, 
    path = EXCLUDED.path,
    order_seq = EXCLUDED.order_seq,
    deleted_at = NULL;

-- 2. Berikan Otorisasi Hak Akses ke Admin (role_id: 1) di lms_sch.role_menus
INSERT INTO lms_sch.role_menus (role_id, menu_id, created_user) VALUES
(1, 915, 'SYSTEM'),
(1, 916, 'SYSTEM')
ON CONFLICT (role_id, menu_id) DO UPDATE SET deleted_at = NULL;
