-- =============================================================================
-- SQL DML CLEANUP SCRIPT UNTUK DUMMY DATA SEEDER LMS KOPKARA
-- =============================================================================
-- Jalankan query SQL ini di pgAdmin / DBeaver / psql untuk menghapus data dummy.

-- 1. Hapus Loan Schedules dummy
DELETE FROM lms_sch.loan_schedules 
WHERE loan_no BETWEEN 400001 AND 415000;

-- 2. Hapus Loans dummy
DELETE FROM lms_sch.loans 
WHERE loan_no BETWEEN 400001 AND 415000;

-- 3. Hapus Loan Trackings dummy
DELETE FROM lms_sch.loan_trackings 
WHERE application_no BETWEEN 300001 AND 315000;

-- 4. Hapus Loan Contracts dummy
DELETE FROM lms_sch.loan_contracts 
WHERE application_no BETWEEN 300001 AND 315000;

-- 5. Hapus Loan Applications dummy
DELETE FROM lms_sch.loan_applications 
WHERE application_no BETWEEN 300001 AND 315000;

-- 6. Hapus Members dummy
DELETE FROM lms_sch.members 
WHERE member_no BETWEEN 200001 AND 210000;

-- 7. Hapus Employees dummy
DELETE FROM lms_sch.employees 
WHERE employee_id BETWEEN 100001 AND 110000;

-- 8. Hapus Loan Products dummy
DELETE FROM lms_sch.loan_products 
WHERE id IN (1, 2, 3);

-- 9. Hapus Master Parent Tables dummy
DELETE FROM lms_sch.departments WHERE deptno IN ('DEPT01', 'DEPT02', 'DEPT03');
DELETE FROM lms_sch.roles WHERE role_id IN (1, 2);
DELETE FROM lms_sch.employee_categories WHERE category_code IN ('PERM', 'CONT');
DELETE FROM lms_sch.employee_statuses WHERE status_code IN ('ACTIVE', 'RESIGNED');
DELETE FROM lms_sch.kopkara_statuses WHERE status_code IN ('ACTIVE', 'INACTIVE');
