-- =========================================================================
-- DML SKRIP KONFIGURASI PARAMETER NOTIFIKASI EMAIL (GMAIL/SMTP) & WHATSAPP
-- =========================================================================

-- 1. Parameter Notifikasi Email (SMTP Gmail & Generic SMTP)
INSERT INTO lms_sch.global_parameters (key_name, key_value, description, created_user) VALUES
('SMTP_HOST', 'smtp.gmail.com', 'Host server SMTP untuk pengiriman Email (Contoh: smtp.gmail.com, smtp.office365.com, email-smtp.us-east-1.amazonaws.com)', 'SYSTEM'),
('SMTP_PORT', '587', 'Port server SMTP (587 untuk TLS/STARTTLS, 465 untuk SSL, 25 untuk Standard)', 'SYSTEM'),
('SMTP_USERNAME', 'your-email@gmail.com', 'Username/Email akun pengirim SMTP Gmail atau Server Email Koperasi', 'SYSTEM'),
('SMTP_PASSWORD', 'xxxx xxxx xxxx xxxx', 'Password Aplikasi Gmail (16 Karakter App Password) atau Password SMTP', 'SYSTEM'),
('SMTP_FROM_NAME', 'Kopkara LMS EWA System', 'Nama Pengirim yang muncul di Email Anggota', 'SYSTEM'),
('SMTP_FROM_EMAIL', 'your-email@gmail.com', 'Alamat Email Pengirim (Reply-To)', 'SYSTEM'),

-- 2. Parameter Channel Notifikasi WhatsApp (Fonnte / Meta Cloud API)
('FONNTE_TOKEN', '', 'Token API Gateway WhatsApp Fonnte (Opsional untuk kirim WA otomatis)', 'SYSTEM'),
('META_WA_PHONE_NUMBER_ID', '', 'ID Nomor Telepon Meta WhatsApp Business Cloud API', 'SYSTEM'),
('META_WA_ACCESS_TOKEN', '', 'Access Token Meta WhatsApp Business Cloud API', 'SYSTEM'),

-- 3. Parameter Konfigurasi Event Notifikasi
('NOTIFICATION_ENABLED_CHANNELS', 'EMAIL,WA', 'Channel Notifikasi Aktif Global. Pilihan: EMAIL,WA / EMAIL / WA / OFF', 'SYSTEM'),
('NOTIF_EVENT_DISBURSEMENT', 'EMAIL,WA', 'Channel Notifikasi Pencairan Pinjaman. Pilihan: EMAIL,WA / EMAIL / WA / OFF', 'SYSTEM'),
('NOTIF_EVENT_REPAYMENT', 'EMAIL,WA', 'Channel Notifikasi Pelunasan Angsuran. Pilihan: EMAIL,WA / EMAIL / WA / OFF', 'SYSTEM'),
('NOTIF_EVENT_APPROVAL', 'EMAIL,WA', 'Channel Notifikasi Persetujuan / Penolakan Pinjaman. Pilihan: EMAIL,WA / EMAIL / WA / OFF', 'SYSTEM')

ON CONFLICT (key_name) DO UPDATE SET 
    description = EXCLUDED.description,
    deleted_at = NULL;
