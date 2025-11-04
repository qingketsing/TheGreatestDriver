# Chunk Upload Test Script

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Chunk Upload Functionality Test" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. Create test file (5MB)
Write-Host "1. Creating test file..." -ForegroundColor Yellow
$testFile = "test_large_file.bin"
$fileSize = 5MB

if (Test-Path $testFile) {
    Remove-Item $testFile
}

# Create a 5MB random data file
$bytes = New-Object byte[] $fileSize
(New-Object Random).NextBytes($bytes)
[System.IO.File]::WriteAllBytes($testFile, $bytes)

Write-Host "  [OK] Created $testFile ($fileSize bytes)" -ForegroundColor Green
Write-Host ""

# 2. Check server status
Write-Host "2. Checking server status..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8000/" -TimeoutSec 2 -ErrorAction Stop
    Write-Host "  [OK] Server is running" -ForegroundColor Green
} catch {
    Write-Host "  [ERROR] Server not running!" -ForegroundColor Red
    Write-Host "    Please start the server first:" -ForegroundColor White
    Write-Host "      cd cmd/server" -ForegroundColor Gray
    Write-Host "      go run main.go" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  Cleaning up test file..." -ForegroundColor Yellow
    if (Test-Path $testFile) {
        Remove-Item $testFile
    }
    exit 1
}
Write-Host ""

# 3. First upload (chunk upload)
Write-Host "3. First upload (chunk upload)..." -ForegroundColor Yellow
Set-Location client
go run chunk_upload.go ../$testFile
$exitCode1 = $LASTEXITCODE
Set-Location ..

if ($exitCode1 -ne 0) {
    Write-Host "  [ERROR] First upload failed" -ForegroundColor Red
    Write-Host ""
    Write-Host "  Cleaning up test file..." -ForegroundColor Yellow
    if (Test-Path $testFile) {
        Remove-Item $testFile
    }
    exit 1
}
Write-Host ""

# 4. Second upload of same file (test quick upload)
Write-Host "4. Second upload of same file (test quick upload)..." -ForegroundColor Yellow
Set-Location client
go run chunk_upload.go ../$testFile
$exitCode2 = $LASTEXITCODE
Set-Location ..

if ($exitCode2 -ne 0) {
    Write-Host "  [WARNING] Second upload failed (expected to succeed via quick upload)" -ForegroundColor Yellow
}
Write-Host ""

# 5. Test upload to subdirectory
Write-Host "5. Test upload to subdirectory..." -ForegroundColor Yellow
Set-Location client
go run chunk_upload.go ../$testFile "test/chunks"
$exitCode3 = $LASTEXITCODE
Set-Location ..

if ($exitCode3 -ne 0) {
    Write-Host "  [WARNING] Subdirectory upload failed" -ForegroundColor Yellow
}
Write-Host ""

# 6. Clean up test file
Write-Host "6. Cleaning up test file..." -ForegroundColor Yellow
if (Test-Path $testFile) {
    Remove-Item $testFile
    Write-Host "  [OK] Test file deleted" -ForegroundColor Green
}
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Test Completed" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Tips:" -ForegroundColor Yellow
Write-Host "  - View uploaded files: uploads/ directory" -ForegroundColor White
Write-Host "  - View database records: http://localhost:8000/debug/drivelist" -ForegroundColor White
Write-Host "  - Temporary chunks: uploads/_tmp/" -ForegroundColor White
Write-Host ""
