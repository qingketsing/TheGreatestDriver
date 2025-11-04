# Single Drive 数据库初始化脚本
# 用于在本地 PostgreSQL 数据库中自动创建所需的表

param(
    [string]$DbHost = "localhost",
    [string]$DbPort = "5432",
    [string]$DbUser = "postgres",
    [string]$DbPassword = "329426",
    [string]$DbName = "tododb"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Single Drive Database Initialization" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Set environment variable to avoid password prompt
$env:PGPASSWORD = $DbPassword

# 1. Check PostgreSQL installation
Write-Host "1. Checking PostgreSQL installation..." -ForegroundColor Yellow
try {
    $psqlVersion = & psql --version 2>&1
    Write-Host "  [OK] PostgreSQL installed: $psqlVersion" -ForegroundColor Green
} catch {
    Write-Host "  [ERROR] PostgreSQL not found in PATH" -ForegroundColor Red
    Write-Host "    Please download from https://www.postgresql.org/download/" -ForegroundColor White
    exit 1
}
Write-Host ""

# 2. Test database connection
Write-Host "2. Testing database connection..." -ForegroundColor Yellow
Write-Host "  Host: $DbHost" -ForegroundColor Gray
Write-Host "  Port: $DbPort" -ForegroundColor Gray
Write-Host "  User: $DbUser" -ForegroundColor Gray
Write-Host "  Database: $DbName" -ForegroundColor Gray

try {
    $testQuery = 'SELECT version();'
    $result = & psql -h $DbHost -p $DbPort -U $DbUser -d postgres -c $testQuery -t 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  [OK] Database connection successful" -ForegroundColor Green
    } else {
        Write-Host "  [ERROR] Database connection failed: $result" -ForegroundColor Red
        Write-Host "    Please check if PostgreSQL service is running" -ForegroundColor White
        exit 1
    }
} catch {
    Write-Host "  [ERROR] Cannot connect to database: $_" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 3. Check/Create database
Write-Host "3. Checking/Creating database..." -ForegroundColor Yellow
$dbExists = & psql -h $DbHost -p $DbPort -U $DbUser -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DbName'" 2>&1

if ($dbExists -eq "1") {
    Write-Host "  [OK] Database '$DbName' already exists" -ForegroundColor Green
} else {
    Write-Host "  [INFO] Database '$DbName' not found, creating..." -ForegroundColor Yellow
    & psql -h $DbHost -p $DbPort -U $DbUser -d postgres -c "CREATE DATABASE $DbName;" 2>&1 | Out-Null
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  [OK] Database '$DbName' created successfully" -ForegroundColor Green
    } else {
        Write-Host "  [ERROR] Failed to create database" -ForegroundColor Red
        exit 1
    }
}
Write-Host ""

# 4. Check existing tables
Write-Host "4. Checking existing tables..." -ForegroundColor Yellow
$tableCheck = & psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -tAc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name IN ('drivelist', 'drivelist_closure');" 2>&1

$tableCount = [int]$tableCheck

if ($tableCount -eq 2) {
    Write-Host "  [OK] Tables exist (drivelist, drivelist_closure)" -ForegroundColor Green
    Write-Host ""
    Write-Host "  WARNING: This will DROP and recreate all tables!" -ForegroundColor Yellow
    Write-Host "  All data will be lost!" -ForegroundColor Red
    $confirm = Read-Host "  Type 'YES' to continue, any other key to cancel"
    
    if ($confirm -ne "YES") {
        Write-Host "  Operation cancelled" -ForegroundColor Gray
        Write-Host ""
        Write-Host "========================================" -ForegroundColor Cyan
        Write-Host "Tables already exist. You can start the server:" -ForegroundColor White
        Write-Host "  cd cmd/server" -ForegroundColor Gray
        Write-Host "  go run main.go" -ForegroundColor Gray
        Write-Host "========================================" -ForegroundColor Cyan
        exit 0
    }
    
    Write-Host "  [INFO] Dropping existing tables..." -ForegroundColor Yellow
    & psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -c "DROP TABLE IF EXISTS drivelist_closure CASCADE; DROP TABLE IF EXISTS drivelist CASCADE;" 2>&1 | Out-Null
    Write-Host "  [OK] Existing tables dropped" -ForegroundColor Green
} elseif ($tableCount -eq 1) {
    Write-Host "  [INFO] Partial tables detected, dropping and recreating" -ForegroundColor Yellow
    & psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -c "DROP TABLE IF EXISTS drivelist_closure CASCADE; DROP TABLE IF EXISTS drivelist CASCADE;" 2>&1 | Out-Null
} else {
    Write-Host "  [OK] Database is empty, ready to initialize" -ForegroundColor Green
}
Write-Host ""

# 5. Execute initialization SQL
Write-Host "5. Executing database initialization script..." -ForegroundColor Yellow

$sqlFile = Join-Path $PSScriptRoot "init_database.sql"

if (-not (Test-Path $sqlFile)) {
    Write-Host "  [ERROR] SQL script not found: $sqlFile" -ForegroundColor Red
    Write-Host "    Please ensure init_database.sql is in the scripts directory" -ForegroundColor White
    exit 1
}

Write-Host "  Executing: $sqlFile" -ForegroundColor Gray

$output = & psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -f $sqlFile 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "  [OK] SQL script executed successfully" -ForegroundColor Green
    
    # Show NOTICE messages from script
    $output | Where-Object { $_ -match "NOTICE" } | ForEach-Object {
        $msg = $_ -replace "^NOTICE:\s*", ""
        Write-Host "    $msg" -ForegroundColor Cyan
    }
} else {
    Write-Host "  [ERROR] SQL script execution failed" -ForegroundColor Red
    Write-Host "    Error: $output" -ForegroundColor White
    exit 1
}
Write-Host ""

# 6. Verify table creation
Write-Host "6. Verifying table structure..." -ForegroundColor Yellow

$verifyQuery = @'
SELECT 
    table_name,
    (SELECT COUNT(*) FROM information_schema.columns WHERE table_name = t.table_name) as column_count
FROM information_schema.tables t
WHERE table_schema='public' AND table_name IN ('drivelist', 'drivelist_closure')
ORDER BY table_name;
'@

$tables = & psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -c $verifyQuery 2>&1

Write-Host $tables -ForegroundColor Gray
Write-Host ""

# 7. Show indexes
Write-Host "7. Verifying indexes..." -ForegroundColor Yellow

$indexQuery = @'
SELECT 
    tablename,
    indexname
FROM pg_indexes
WHERE schemaname='public' AND tablename IN ('drivelist', 'drivelist_closure')
ORDER BY tablename, indexname;
'@

$indexes = & psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -c $indexQuery 2>&1
Write-Host $indexes -ForegroundColor Gray
Write-Host ""

# Clean up environment variable
$env:PGPASSWORD = ""

# Done
Write-Host "========================================" -ForegroundColor Green
Write-Host "  [SUCCESS] Database initialization complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. Start the server:" -ForegroundColor White
Write-Host "   cd cmd/server" -ForegroundColor Gray
Write-Host "   go run main.go" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Test upload:" -ForegroundColor White
Write-Host "   cd client" -ForegroundColor Gray
Write-Host "   go run chunk_upload_test.go ../test_file.txt" -ForegroundColor Gray
Write-Host ""
Write-Host "3. View data:" -ForegroundColor White
Write-Host "   Browser: http://localhost:8000/debug/drivelist" -ForegroundColor Gray
Write-Host ""