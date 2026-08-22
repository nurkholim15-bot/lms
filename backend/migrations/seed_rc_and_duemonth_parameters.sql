-- =========================================================================
-- DML SQL UNTUK PARAMETER LOAN_DUEMONTH, LOG_SUBMIT_LOAN_PATH, DAN RESPONSE CODE (RC)
-- DI POSTGRESQL (PgAdmin / LMS)
-- =========================================================================

INSERT INTO lms_sch.global_parameters (key_name, key_value, description) VALUES
('LOAN_DUEMONTH', '0', 'Offset bulan jatuh tempo angsuran pertama (0 = bulan ini, 1 = bulan depan)'),
('LOG_SUBMIT_LOAN_PATH', './logs/submit_loan.log', 'Lokasi path dan nama file log transaksi pengajuan pinjaman (submit loan)'),
('RC_SUBMIT_LOAN_SUCCESS', '00', 'Response Code transaksi pengajuan pinjaman berhasil (SUCCESS)'),
('RC_SUBMIT_LOAN_NON_KOPKARA', '10', 'Response Code ditolak bukan anggota Kopkara'),
('RC_SUBMIT_LOAN_NON_ADIRA', '11', 'Response Code ditolak bukan karyawan Adira'),
('RC_SUBMIT_LOAN_TENOR', '12', 'Response Code ditolak Tenor Melebihi Batas Maksimum'),
('RC_SUBMIT_LOAN_CREDIT_LIMIT', '13', 'Response Code ditolak Jumlah Pinjaman Melebihi Credit Limit'),
('RC_SUBMIT_LOAN_PERIOD',
    'RC_SUBMIT_LOAN_OTHERS', '14', 'Response Code ditolak Di Luar Periode Tanggal Pengajuan'),
('RC_SUBMIT_LOAN_OTHERS', '99', 'Response Code ditolak alasan lainnya (OTHERS)')
ON CONFLICT (key_name) DO UPDATE SET
    key_value = EXCLUDED.key_value,
    description = EXCLUDED.description;

-- Verifikasi Data Parameter setelah Eksekusi
SELECT id, key_name, key_value, description 
FROM lms_sch.global_parameters 
WHERE key_name IN (
    'LOAN_DUEMONTH', 
    'LOG_SUBMIT_LOAN_PATH', 
    'RC_SUBMIT_LOAN_SUCCESS', 
    'RC_SUBMIT_LOAN_NON_KOPKARA', 
    'RC_SUBMIT_LOAN_NON_ADIRA', 
    'RC_SUBMIT_LOAN_TENOR', 
    'RC_SUBMIT_LOAN_CREDIT_LIMIT', 
    'RC_SUBMIT_LOAN_PERIOD',
    'RC_SUBMIT_LOAN_OTHERS'
)
ORDER BY id ASC;
