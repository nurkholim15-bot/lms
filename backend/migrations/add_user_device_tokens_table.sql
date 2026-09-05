-- =========================================================================
-- DDL SQL: PEMBUATAN TABEL FCM DEVICE TOKENS (LMS_SCH.USER_DEVICE_TOKENS)
-- =========================================================================

CREATE TABLE IF NOT EXISTS lms_sch.user_device_tokens (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL,                  -- Member No / Employee ID
    fcm_token     TEXT NOT NULL,                    -- Unique FCM Device Token dari Google
    device_model  VARCHAR(100),                     -- Contoh: 'Samsung Galaxy A54'
    device_os     VARCHAR(50) DEFAULT 'ANDROID',     -- ANDROID / IOS
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_user  VARCHAR(100) DEFAULT 'SYSTEM',
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_user  VARCHAR(100) DEFAULT 'SYSTEM',
    deleted_at    TIMESTAMP WITH TIME ZONE,
    deleted_user  VARCHAR(100)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_fcm_token ON lms_sch.user_device_tokens(user_id, fcm_token) WHERE deleted_at IS NULL;
