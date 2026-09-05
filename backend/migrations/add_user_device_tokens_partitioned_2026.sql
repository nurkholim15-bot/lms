-- =========================================================================
-- DDL POSTGRESQL: TABEL USER_DEVICE_TOKENS BERPARTISI BULANAN (2026 & 2027)
-- =========================================================================
-- Mengubah / Mengkonversi lms_sch.user_device_tokens menjadi Tabel Berpartisi 
-- berbasis Range kolom created_at (Format Nama Partisi: user_device_tokens_YYYYMM).
-- =========================================================================

-- 1. Drop tabel unpartitioned lama jika sudah pernah dibuat tanpa partisi
DROP TABLE IF EXISTS lms_sch.user_device_tokens CASCADE;

-- 2. Sequence untuk Auto-Increment Primary Key ID
CREATE SEQUENCE IF NOT EXISTS lms_sch.user_device_tokens_id_seq;

-- 3. Pembuatan Tabel Utama Berpartisi Range (created_at)
CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens (
    id            BIGINT NOT NULL DEFAULT nextval('lms_sch.user_device_tokens_id_seq'),
    user_id       BIGINT NOT NULL,                  -- Member No / Employee ID
    fcm_token     TEXT NOT NULL,                    -- Unique FCM Device Token dari Google
    device_model  VARCHAR(100),                     -- Contoh: 'Samsung Galaxy A54'
    device_os     VARCHAR(50) DEFAULT 'ANDROID',     -- ANDROID / IOS
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_user  VARCHAR(100) DEFAULT 'SYSTEM',
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_user  VARCHAR(100) DEFAULT 'SYSTEM',
    deleted_at    TIMESTAMP WITH TIME ZONE,
    deleted_user  VARCHAR(100),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- 4. Pembuatan Partisi Bulanan Tahun 2026 (Format: user_device_tokens_YYYYMM)
CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202601 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2026-01-01 00:00:00+00') TO ('2026-02-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202602 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2026-02-01 00:00:00+00') TO ('2026-03-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202603 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2026-03-01 00:00:00+00') TO ('2026-04-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202604 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2026-04-01 00:00:00+00') TO ('2026-05-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202605 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2026-05-01 00:00:00+00') TO ('2026-06-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202606 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2026-06-01 00:00:00+00') TO ('2026-07-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202607 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2026-07-01 00:00:00+00') TO ('2026-08-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202608 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2026-08-01 00:00:00+00') TO ('2026-09-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202609 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2026-09-01 00:00:00+00') TO ('2026-10-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202610 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2026-10-01 00:00:00+00') TO ('2026-11-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202611 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2026-11-01 00:00:00+00') TO ('2026-12-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202612 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2026-12-01 00:00:00+00') TO ('2027-01-01 00:00:00+00');

-- 5. Pembuatan Partisi Bulanan Tahun 2027 (Format: user_device_tokens_YYYYMM)
CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202701 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2027-01-01 00:00:00+00') TO ('2027-02-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202702 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2027-02-01 00:00:00+00') TO ('2027-03-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202703 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2027-03-01 00:00:00+00') TO ('2027-04-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202704 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2027-04-01 00:00:00+00') TO ('2027-05-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202705 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2027-05-01 00:00:00+00') TO ('2027-06-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202706 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2027-06-01 00:00:00+00') TO ('2027-07-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202707 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2027-07-01 00:00:00+00') TO ('2027-08-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202708 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2027-08-01 00:00:00+00') TO ('2027-09-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202709 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2027-09-01 00:00:00+00') TO ('2027-10-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202710 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2027-10-01 00:00:00+00') TO ('2027-11-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202711 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2027-11-01 00:00:00+00') TO ('2027-12-01 00:00:00+00');

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_202712 PARTITION OF lms_sch.user_device_tokens
    FOR VALUES FROM ('2027-12-01 00:00:00+00') TO ('2028-01-01 00:00:00+00');

-- 6. Partisi Catch-All Default
CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens_default PARTITION OF lms_sch.user_device_tokens DEFAULT;

-- 7. Indeks Performa pada Tabel Utama Berpartisi
CREATE INDEX IF NOT EXISTS idx_user_fcm_token ON lms_sch.user_device_tokens(user_id, fcm_token) WHERE deleted_at IS NULL;
