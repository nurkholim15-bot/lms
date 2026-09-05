-- =========================================================================
-- DML SQL: SEEDING PARAMETER FCM GOOGLE PUSH NOTIFICATION (LMS_SCH.GLOBAL_PARAMETERS)
-- =========================================================================

INSERT INTO lms_sch.global_parameters (key_name, key_value, description, created_user) VALUES
('FCM_SERVER_KEY', '', 'Server Key / Service Account Key Firebase Cloud Messaging (Google Push Notification Android)', 'SYSTEM')
ON CONFLICT (key_name) DO UPDATE SET 
    description = EXCLUDED.description,
    deleted_at = NULL;
