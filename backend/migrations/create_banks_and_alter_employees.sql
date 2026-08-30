-- 1. Create lms_sch.banks table
CREATE TABLE IF NOT EXISTS lms_sch.banks (
    bank_code VARCHAR(15) NOT NULL,
    bank_name VARCHAR(60) NOT NULL,
    bank_code_provider VARCHAR(15),
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_user VARCHAR(100) DEFAULT 'SYSTEM',
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_user VARCHAR(100),
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_user VARCHAR(100),
    CONSTRAINT pk_banks PRIMARY KEY (bank_code)
);

-- 2. Alter lms_sch.employees table
ALTER TABLE lms_sch.employees 
    ADD COLUMN IF NOT EXISTS bank_code VARCHAR(15),
    DROP COLUMN IF EXISTS bank_name,
    DROP COLUMN IF EXISTS bank_account_name;

-- 3. Add Foreign Key constraint (if not exists)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_employees_bank'
    ) THEN
        ALTER TABLE lms_sch.employees
            ADD CONSTRAINT fk_employees_bank 
            FOREIGN KEY (bank_code) 
            REFERENCES lms_sch.banks (bank_code) 
            ON UPDATE CASCADE 
            ON DELETE RESTRICT;
    END IF;
END $$;
