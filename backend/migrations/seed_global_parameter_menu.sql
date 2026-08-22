-- =========================================================================
-- DML SQL UNTUK MENAMBAHKAN MENU "Pengaturan Parameter" (GLOBAL PARAMETER)
-- DAN OTORISASI ROLE_MENUS DI POSTGRESQL (PgAdmin)
-- =========================================================================

-- 1. Tambahkan Menu "Pengaturan Parameter" ke tabel lms_sch.menus (ID: 8, path: 'parameters')
INSERT INTO lms_sch.menus (menu_id, parent_id, title, icon, path, order_seq) VALUES
(8, NULL, 'Pengaturan Parameter', '⚙️', 'parameters', 8)
ON CONFLICT (menu_id) DO UPDATE SET 
    title = EXCLUDED.title, 
    icon = EXCLUDED.icon, 
    path = EXCLUDED.path,
    order_seq = EXCLUDED.order_seq;

-- 2. Berikan Hak Akses Otorisasi Menu Parameter kepada Role Admin (role_id: 1) di tabel lms_sch.role_menus
INSERT INTO lms_sch.role_menus (role_id, menu_id) VALUES
(1, 8)
ON CONFLICT DO NOTHING;

-- 3. (Opsional) Inisialisasi Parameter Default di tabel lms_sch.global_parameters jika belum ada
INSERT INTO lms_sch.global_parameters (key_name, key_value, description) VALUES
('APP_TOKEN', 'ewa_token', 'Nama Cookie Token Parameter'),
('APP_USER', 'ewa_user', 'Nama LocalStorage User Parameter'),
('LOGIN_MAX_ATTEMPTS', '5', 'Maksimum Percobaan Login'),
('LOGIN_LOCKOUT_MINUTES', '15', 'Durasi Lockout Akun (Menit)'),
('SESSION_EXPIRY_MINUTES', '1440', 'Durasi Masa Berlaku Sesi (Menit)'),
('SINGLE_SESSION_MODE', 'FALSE', 'Mode Sesi Tunggal Per User')
ON CONFLICT (key_name) DO NOTHING;
