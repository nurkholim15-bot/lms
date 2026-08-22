# =========================================================================
# SKRIP TESTING AUTOMATED VALIDASI RESPONSE CODE (RC) & SUBMIT LOAN LOG
# LMS (Loan Management System) - Backend API Testing
# =========================================================================

$baseUrl = "https://localhost:8086/api/applications"
[System.Net.ServicePointManager]::ServerCertificateValidationCallback = {$true}
$headers = @{ "Content-Type" = "application/json"; "Authorization" = "Bearer mock-token-100001" }

Write-Host "=========================================================================" -ForegroundColor Cyan
Write-Host " MEMULAI TESTING VALIDASI BACKEND API & PENCATATAN LOG SUBMIT LOAN" -ForegroundColor Cyan
Write-Host "=========================================================================" -ForegroundColor Cyan

# -------------------------------------------------------------------------
# Test 1: RC 11 - Ditolak Bukan Karyawan Adira
# -------------------------------------------------------------------------
Write-Host "`n[TEST 1] Testing RC 11: Non-Adira Employee (Member/Employee ID: 999999)" -ForegroundColor Yellow
$body1 = '{"member_no": 999999, "product_id": 1, "requested_amount": 1000000, "tenor": 1}'
try {
    $res1 = Invoke-RestMethod -Uri $baseUrl -Method POST -Headers $headers -Body $body1
    Write-Host "Respon: " ($res1 | ConvertTo-Json)
} catch {
    Write-Host "Status Response: "$_.Exception.Response.StatusCode.Value__ -ForegroundColor Red
    Write-Host "Detail Eror: "$_.ErrorDetails.Message -ForegroundColor Red
}

# -------------------------------------------------------------------------
# Test 2: RC 12 - Ditolak Tenor Melebihi Batas Maksimum (Tenor: 24 bulan)
# -------------------------------------------------------------------------
Write-Host "`n[TEST 2] Testing RC 12: Exceeding Max Tenor (Tenor: 24 bulan)" -ForegroundColor Yellow
$body2 = '{"member_no": 100001, "product_id": 1, "requested_amount": 1000000, "tenor": 24}'
try {
    $res2 = Invoke-RestMethod -Uri $baseUrl -Method POST -Headers $headers -Body $body2
    Write-Host "Respon: " ($res2 | ConvertTo-Json)
} catch {
    Write-Host "Status Response: "$_.Exception.Response.StatusCode.Value__ -ForegroundColor Red
    Write-Host "Detail Eror: "$_.ErrorDetails.Message -ForegroundColor Red
}

# -------------------------------------------------------------------------
# Test 3: RC 13 - Ditolak Jumlah Pinjaman Melebihi Credit Limit (Rp 100.000.000)
# -------------------------------------------------------------------------
Write-Host "`n[TEST 3] Testing RC 13: Exceeding Credit Limit (Nominal: Rp 100.000.000)" -ForegroundColor Yellow
$body3 = '{"member_no": 100001, "product_id": 1, "requested_amount": 100000000, "tenor": 12}'
try {
    $res3 = Invoke-RestMethod -Uri $baseUrl -Method POST -Headers $headers -Body $body3
    Write-Host "Respon: " ($res3 | ConvertTo-Json)
} catch {
    Write-Host "Status Response: "$_.Exception.Response.StatusCode.Value__ -ForegroundColor Red
    Write-Host "Detail Eror: "$_.ErrorDetails.Message -ForegroundColor Red
}

# -------------------------------------------------------------------------
# Test 4: RC 00 - Pengajuan Pinjaman Berhasil (Nominal Wajar: Rp 1.000.000)
# -------------------------------------------------------------------------
Write-Host "`n[TEST 4] Testing RC 00: Valid Submission (Nominal: Rp 1.000.000, Tenor: 1)" -ForegroundColor Yellow
$body4 = '{"member_no": 100001, "product_id": 1, "requested_amount": 1000000, "tenor": 1}'
try {
    $res4 = Invoke-RestMethod -Uri $baseUrl -Method POST -Headers $headers -Body $body4
    Write-Host "Respon Sukses: " ($res4 | ConvertTo-Json) -ForegroundColor Green
} catch {
    Write-Host "Detail Eror: "$_.ErrorDetails.Message -ForegroundColor Red
}

# -------------------------------------------------------------------------
# Verifikasi File Log
# -------------------------------------------------------------------------
Write-Host "`n=========================================================================" -ForegroundColor Cyan
Write-Host " HASIL ISI FILE LOG (logs/submit_loan.log):" -ForegroundColor Cyan
Write-Host "=========================================================================" -ForegroundColor Cyan
if (Test-Path "./logs/submit_loan.log") {
    Get-Content "./logs/submit_loan.log" -Tail 10 | Write-Host -ForegroundColor White
} else {
    Write-Host "File log belum terbentuk di ./logs/submit_loan.log" -ForegroundColor Red
}
