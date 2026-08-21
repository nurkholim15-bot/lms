-- DDL & DML Migration: Move Role Control to lms_sch.users (No Default Role)

-- 1. Tambah kolom role_id pada lms_sch.users tanpa default value
ALTER TABLE lms_sch.users ADD COLUMN IF NOT EXISTS role_id BIGINT;

-- 2. DML Update role_id pada lms_sch.users HANYA berdasarkan data eksisting
UPDATE lms_sch.users SET role_id = 1 WHERE LOWER(role) = 'admin';
UPDATE lms_sch.users SET role_id = 3 WHERE LOWER(role) = 'hrd';
UPDATE lms_sch.users SET role_id = 2 WHERE LOWER(role) = 'anggota';

-- 3. Drop kolom string 'role' dari lms_sch.users
ALTER TABLE lms_sch.users DROP COLUMN IF EXISTS role;

-- 4. Drop kolom 'role_id' dan 'role' dari lms_sch.employees
ALTER TABLE lms_sch.employees DROP COLUMN IF EXISTS role_id;
ALTER TABLE lms_sch.employees DROP COLUMN IF EXISTS role;
