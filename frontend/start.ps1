# 检查 Node.js 是否安装
Write-Host "检查 Node.js..." -ForegroundColor Cyan
if (!(Get-Command node -ErrorAction SilentlyContinue)) {
    Write-Host "错误: 未安装 Node.js。请先安装 Node.js (https://nodejs.org/)" -ForegroundColor Red
    exit 1
}

$nodeVersion = node --version
Write-Host "Node.js 版本: $nodeVersion" -ForegroundColor Green

# 进入前端目录
Set-Location -Path $PSScriptRoot

# 检查是否已安装依赖
if (!(Test-Path "node_modules")) {
    Write-Host "`n正在安装依赖..." -ForegroundColor Cyan
    npm install
    if ($LASTEXITCODE -ne 0) {
        Write-Host "依赖安装失败！" -ForegroundColor Red
        exit 1
    }
    Write-Host "依赖安装成功！" -ForegroundColor Green
} else {
    Write-Host "`n依赖已安装" -ForegroundColor Green
}

# 启动开发服务器
Write-Host "`n启动开发服务器..." -ForegroundColor Cyan
Write-Host "前端将运行在: http://localhost:3000" -ForegroundColor Yellow
Write-Host "请确保后端服务运行在: http://localhost:8080`n" -ForegroundColor Yellow

npm run dev
