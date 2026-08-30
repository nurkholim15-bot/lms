-- =========================================================================
-- MIGRATION & DDL/DML MODUL SUB-LEDGER KAS & BANK, REKENING OPERASIONAL,
-- MASTER TIPE TRANSAKSI, SERTA PARTISI BULANAN 2026 & 2027
-- =========================================================================

-- 1. Buat Tabel Master Rekening Operasional Kas & Bank Koperasi
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_accounts (
    account_id SERIAL NOT NULL,
    account_number VARCHAR(30) NOT NULL,
    account_name VARCHAR(100) NOT NULL,
    bank_code VARCHAR(15) NOT NULL,
    currency VARCHAR(3) DEFAULT 'IDR',
    initial_balance NUMERIC(15,2) DEFAULT 0.00,
    current_balance NUMERIC(15,2) DEFAULT 0.00,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_user VARCHAR(100) DEFAULT 'SYSTEM',
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_user VARCHAR(100),
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_user VARCHAR(100),
    CONSTRAINT pk_cashbank_accounts PRIMARY KEY (account_id),
    CONSTRAINT fk_cashbank_accounts_bank FOREIGN KEY (bank_code) 
        REFERENCES lms_sch.banks (bank_code) ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_cashbank_account_number_active ON lms_sch.cashbank_accounts(account_number) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cb_acc_bank_code ON lms_sch.cashbank_accounts(bank_code);
CREATE INDEX IF NOT EXISTS idx_cb_acc_active ON lms_sch.cashbank_accounts(is_active);

-- 2. Buat Tabel Master Tipe Transaksi Kas & Bank
CREATE TABLE IF NOT EXISTS lms_sch.transaction_types (
    type_code CHAR(5) NOT NULL,
    type_name VARCHAR(60) NOT NULL,
    direction CHAR(3) NOT NULL CHECK (direction IN ('IN', 'OUT')),
    description TEXT,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_user VARCHAR(100) DEFAULT 'SYSTEM',
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_user VARCHAR(100),
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_user VARCHAR(100),
    CONSTRAINT pk_transaction_types PRIMARY KEY (type_code)
);

-- Seed Initial Data Tipe Transaksi
INSERT INTO lms_sch.transaction_types (type_code, type_name, direction, description) VALUES
('DISB', 'Pencairan Pinjaman (Disbursement)', 'OUT', 'Pencairan dana pinjaman ke rekening anggota'),
('RPMN', 'Pelunasan Manual (Manual Repayment)', 'IN', 'Pembayaran angsuran/pelunasan manual via transfer/kas'),
('RPIP', 'Pelunasan Import Payroll', 'IN', 'Pembayaran potong gaji massal via import file excel/csv'),
('ADJI', 'Penyesuaian Masuk (Adjustment In)', 'IN', 'Koreksi penambahan saldo kas/bank (bunga bank, pengembalian, dll)'),
('ADJO', 'Penyesuaian Keluar (Adjustment Out)', 'OUT', 'Koreksi pengurangan saldo kas/bank (biaya admin, denda, dll)'),
('FEEB', 'Biaya Admin Transfer Bank', 'OUT', 'Biaya administrasi transfer antar bank saat pencairan')
ON CONFLICT (type_code) DO UPDATE SET 
    type_name = EXCLUDED.type_name,
    direction = EXCLUDED.direction,
    description = EXCLUDED.description;

-- 3. Buat Tabel Induk Mutasi Sub-Ledger Kas & Bank (Partitioned Table)
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions (
    transaction_id BIGSERIAL,
    transaction_no VARCHAR(30) NOT NULL,
    transaction_date TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    bank_code VARCHAR(15) NOT NULL,
    bank_account_no VARCHAR(30) NOT NULL,
    type_code CHAR(5) NOT NULL,
    direction CHAR(3) NOT NULL CHECK (direction IN ('IN', 'OUT')),
    amount NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    balance_before NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    balance_after NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    reference_type VARCHAR(30) NOT NULL,
    reference_no VARCHAR(50) NOT NULL,
    member_no BIGINT,
    employee_id BIGINT,
    description TEXT,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_user VARCHAR(100) DEFAULT 'SYSTEM',
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_user VARCHAR(100),
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_user VARCHAR(100),
    CONSTRAINT pk_cashbank_transactions PRIMARY KEY (transaction_id, transaction_date),
    CONSTRAINT fk_cb_bank FOREIGN KEY (bank_code) REFERENCES lms_sch.banks(bank_code) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_cb_type FOREIGN KEY (type_code) REFERENCES lms_sch.transaction_types(type_code) ON UPDATE CASCADE ON DELETE RESTRICT
) PARTITION BY RANGE (transaction_date);

-- 4. Buat Tabel Partisi Bulanan Tahun 2026 (202601 s/d 202612)
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202601 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2026-01-01 00:00:00') TO ('2026-02-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202602 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2026-02-01 00:00:00') TO ('2026-03-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202603 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2026-03-01 00:00:00') TO ('2026-04-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202604 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2026-04-01 00:00:00') TO ('2026-05-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202605 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2026-05-01 00:00:00') TO ('2026-06-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202606 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2026-06-01 00:00:00') TO ('2026-07-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202607 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2026-07-01 00:00:00') TO ('2026-08-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202608 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2026-08-01 00:00:00') TO ('2026-09-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202609 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2026-09-01 00:00:00') TO ('2026-10-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202610 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2026-10-01 00:00:00') TO ('2026-11-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202611 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2026-11-01 00:00:00') TO ('2026-12-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202612 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2026-12-01 00:00:00') TO ('2027-01-01 00:00:00');

-- 5. Buat Tabel Partisi Bulanan Tahun 2027 (202701 s/d 202712)
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202701 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2027-01-01 00:00:00') TO ('2027-02-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202702 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2027-02-01 00:00:00') TO ('2027-03-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202703 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2027-03-01 00:00:00') TO ('2027-04-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202704 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2027-04-01 00:00:00') TO ('2027-05-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202705 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2027-05-01 00:00:00') TO ('2027-06-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202706 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2027-06-01 00:00:00') TO ('2027-07-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202707 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2027-07-01 00:00:00') TO ('2027-08-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202708 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2027-08-01 00:00:00') TO ('2027-09-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202709 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2027-09-01 00:00:00') TO ('2027-10-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202710 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2027-10-01 00:00:00') TO ('2027-11-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202711 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2027-11-01 00:00:00') TO ('2027-12-01 00:00:00');
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_202712 PARTITION OF lms_sch.cashbank_transactions FOR VALUES FROM ('2027-12-01 00:00:00') TO ('2028-01-01 00:00:00');

-- 6. Partisi Default
CREATE TABLE IF NOT EXISTS lms_sch.cashbank_transactions_default PARTITION OF lms_sch.cashbank_transactions DEFAULT;

-- 7. Register Menu ke lms_sch.menus dan lms_sch.role_menus
INSERT INTO lms_sch.menus (menu_id, parent_id, title, icon, path, order_seq) VALUES
(913, 9, 'Master Rekening Bank', '🏦', 'master-cashbank-accounts', 913),
(914, 9, 'Master Tipe Transaksi Kas/Bank', '🏷️', 'master-transaction-types', 914),
(915, 9, 'Transaksi & Mutasi Kas Bank', '💸', 'master-cashbank-transactions', 915)
ON CONFLICT (menu_id) DO UPDATE SET 
    parent_id = EXCLUDED.parent_id,
    title = EXCLUDED.title, 
    icon = EXCLUDED.icon, 
    path = EXCLUDED.path,
    order_seq = EXCLUDED.order_seq,
    deleted_at = NULL;

INSERT INTO lms_sch.role_menus (role_id, menu_id, created_user) VALUES
(1, 913, 'SYSTEM'),
(1, 914, 'SYSTEM'),
(1, 915, 'SYSTEM')
ON CONFLICT (role_id, menu_id) DO UPDATE SET deleted_at = NULL;
