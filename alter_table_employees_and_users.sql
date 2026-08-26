-- ==============================================================================
-- DDL SCRIPT: PENAMBAHAN KOLOM NO KTP, NO HP & SECURITY PARAMETERS DI LMS
-- SKEMA: lms_sch
-- TANGGAL: 24 AGUSTUS 2026
-- ==============================================================================

-- 1. Penambahan Kolom no_ktp, phone_number, & email pada tabel lms_sch.employees
ALTER TABLE "lms_sch"."employees" 
  ADD COLUMN IF NOT EXISTS "no_ktp" VARCHAR(16),
  ADD COLUMN IF NOT EXISTS "phone_number" VARCHAR(20),
  ADD COLUMN IF NOT EXISTS "email" VARCHAR(100);

COMMENT ON COLUMN "lms_sch"."employees"."no_ktp" IS 'Nomor KTP / NIK 16 digit karyawan';
COMMENT ON COLUMN "lms_sch"."employees"."phone_number" IS 'Nomor HP aktif karyawan untuk OTP & Login';
COMMENT ON COLUMN "lms_sch"."employees"."email" IS 'Email korporat/pribadi karyawan';

-- 2. Penambahan Kolom PIN 6-Digit & Lockout State pada tabel lms_sch.users
ALTER TABLE "lms_sch"."users"
  ADD COLUMN IF NOT EXISTS "no_ktp" VARCHAR(16),
  ADD COLUMN IF NOT EXISTS "phone_number" VARCHAR(20),
  ADD COLUMN IF NOT EXISTS "pin" VARCHAR(255),
  ADD COLUMN IF NOT EXISTS "failed_pin_attempts" INT DEFAULT 0,
  ADD COLUMN IF NOT EXISTS "pin_locked_until" TIMESTAMP;

COMMENT ON COLUMN "lms_sch"."users"."pin" IS 'Hash Bcrypt dari PIN 6-digit Mobile App';
COMMENT ON COLUMN "lms_sch"."users"."failed_pin_attempts" IS 'Jumlah kegagalan input PIN berturut-turut';
COMMENT ON COLUMN "lms_sch"."users"."pin_locked_until" IS 'Waktu kuncian akun jika salah PIN 3 kali';

-- 3. Inserksi Parameter Keamanan Dinamis pada tabel lms_sch.global_parameters
INSERT INTO "lms_sch"."global_parameters" ("key_name", "key_value", "description", "created_at", "created_user", "updated_at", "updated_user")
VALUES
  ('PIN_MAX_FAILED_ATTEMPTS', '3', 'Maksimal kegagalan PIN berturut-turut sebelum terkunci', NOW(), 'SYSTEM_AUTO', NOW(), 'SYSTEM_AUTO'),
  ('PIN_LOCKOUT_DURATION_MINUTES', '15', 'Durasi penguncian aplikasi (menit) jika salah PIN 3x', NOW(), 'SYSTEM_AUTO', NOW(), 'SYSTEM_AUTO'),
  ('PIN_IDLE_TIMEOUT_MINUTES', '3', 'Batas waktu idle background (menit) sebelum auto-lock', NOW(), 'SYSTEM_AUTO', NOW(), 'SYSTEM_AUTO'),
  ('PASSWORD_MAX_FAILED_ATTEMPTS', '3', 'Maksimal kesalahan password berturut-turut di Web Portal', NOW(), 'SYSTEM_AUTO', NOW(), 'SYSTEM_AUTO'),
  ('PASSWORD_ROTATION_DAYS', '90', 'Masa berlaku password (hari) sebelum wajib diganti', NOW(), 'SYSTEM_AUTO', NOW(), 'SYSTEM_AUTO'),
  ('PASSWORD_MIN_LENGTH', '9', 'Panjang minimal karakter password', NOW(), 'SYSTEM_AUTO', NOW(), 'SYSTEM_AUTO')
ON CONFLICT (key_name) DO UPDATE 
SET "key_value" = EXCLUDED."key_value", "updated_at" = NOW(), "updated_user" = 'SYSTEM_AUTO';
