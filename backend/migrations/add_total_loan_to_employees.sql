-- Migration: Add total_loan column to lms_sch.employees table
-- Description: Multi-Loan Credit Limit Validation (total_loan = Pokok + Bunga)

ALTER TABLE lms_sch.employees 
ADD COLUMN IF NOT EXISTS total_loan NUMERIC(15, 2) DEFAULT 0.00;
