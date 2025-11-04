# 开发指南

## 📋 目录

1. [环境要求](#环境要求)
2. [快速开始](#快速开始)
3. [数据库初始化](#数据库初始化)
4. [启动服务](#启动服务)
5. [功能测试](#功能测试)
6. [故障排查](#故障排查)

---

## 环境要求

### 必需软件

- ✅ **Go** 1.16+ - [下载](https://golang.org/dl/)
- ✅ **PostgreSQL** 12+ - [下载](https://www.postgresql.org/download/)
- ✅ **Node.js** 18+ (前端) - [下载](https://nodejs.org/)

### 检查安装

```powershell
# 检查 Go
go version

# 检查 PostgreSQL
psql --version

# 检查 Node.js
node --version
npm --version
```

---

## 快速开始

### 一键启动（推荐）

```powershell
# 1. 初始化数据库
.\scripts\init_db.ps1

# 2. 启动后端服务器
go run cmd/server/main.go

# 3. (可选) 启动前端
cd frontend
npm install
npm run dev
```

---

## 数据库初始化

### 自动初始化

```powershell
# 在项目根目录执行
.\scripts\init_db.ps1
```

### 执行流程

脚本会自动完成：

1. ✓ 检查 PostgreSQL 是否安装
2. ✓ 测试数据库连接
3. ✓ 检查/创建数据库 (tododb)
4. ✓ 检查现有表
5. ✓ 执行初始化 SQL
6. ✓ 验证表结构和索引

### 手动初始化

如果脚本执行失败，可以手动初始化：

```powershell
# 1. 连接到 PostgreSQL
psql -U postgres

# 2. 创建数据库
CREATE DATABASE tododb;
\c tododb

# 3. 执行初始化脚本
\i scripts/init_database.sql
```

### 数据库配置

默认配置（在 `server/server.go` 中）：

```go
host=localhost 
port=5432 
user=postgres 
password=329426 
dbname=tododb 
sslmode=disable
```

修改密码或其他配置请编辑 `server/server.go` 中的 `SetupDefaultSql()` 函数。

---

## 启动服务

### 后端服务器

```powershell
# 方法1: 直接运行
go run cmd/server/main.go

# 方法2: 编译后运行
go build -o server.exe cmd/server/main.go
.\server.exe
```

服务器将在 `http://localhost:8000` 启动。

**验证服务器运行**:
```powershell
# 访问主页
curl http://localhost:8000/

# 查看文件列表
curl http://localhost:8000/list

# 查看数据库记录
curl http://localhost:8000/debug/drivelist
```

### 前端服务

```powershell
cd frontend
npm install
npm run dev
```

前端将在 `http://localhost:5173` 启动。

---

## 功能测试

### 分块上传和秒传功能

#### 自动化测试（推荐）

```powershell
# 在项目根目录运行
.\test_chunk_upload.ps1
```

测试内容：
- ✅ 创建 5MB 测试文件
- ✅ 第一次上传（分块上传）
- ✅ 第二次上传（秒传）
- ✅ 上传到子目录
- ✅ 自动清理

#### 手动测试

```powershell
# 1. 启动服务器
go run cmd/server/main.go

# 2. 创建测试文件（可选）
cd client
go run create_test_file.go

# 3. 测试上传
go run chunk_upload.go ../test_file.bin

# 4. 测试秒传（再次上传同一文件）
go run chunk_upload.go ../test_file.bin

# 5. 上传到指定目录
go run chunk_upload.go ../test_file.bin "folder/subfolder"
```

### API 功能说明

#### 1. 秒传（Quick Upload）

**接口**: `POST /upload/quick`

**功能**: 检查文件哈希，如果文件已存在则直接返回成功

**参数**:
- `fileHash`: 文件的 SHA256 哈希值
- `fileName`: 文件名
- `path`: 目标路径（可选）
- `totalSize`: 文件大小

**返回**:
```json
{
  "needUpload": false,
  "existing_id": 123,
  "message": "quick upload success"
}
```

#### 2. 分块上传（Chunk Upload）

**接口**: `POST /upload/chunk`

**功能**: 支持大文件分块上传，默认每块 1MB

**参数**:
- `uploadId`: 上传会话ID
- `fileName`: 文件名
- `fileHash`: 文件哈希
- `totalChunks`: 总分片数
- `chunkIndex`: 当前分片索引（从1开始）
- `totalSize`: 文件总大小
- `path`: 目标路径（可选）
- `chunk`: 分片文件数据

#### 3. 上传进度查询

**接口**: `GET /upload/progress/:uploadId`

**返回**:
```json
{
  "uploadId": "xxx",
  "status": "uploading",
  "receivedChunks": 3,
  "totalChunks": 5,
  "percent": 60.0,
  "receivedBytes": 3145728,
  "totalBytes": 5242880
}
```

### 其他 API 测试

```powershell
# 文件列表
Invoke-WebRequest http://localhost:8000/list

# 上传文件
$file = Get-Item test.txt
$form = @{
    file = $file
    meta = '{"name":"test.txt","capacity":'+$file.Length+'}'
}
Invoke-WebRequest -Uri http://localhost:8000/upload -Method Post -Form $form

# 下载文件
Invoke-WebRequest -Uri "http://localhost:8000/download?name=test.txt" -OutFile downloaded.txt

# 删除文件
Invoke-WebRequest -Uri "http://localhost:8000/delete?name=test.txt" -Method Delete

# 创建目录
$body = @{name="newfolder"} | ConvertTo-Json
Invoke-WebRequest -Uri http://localhost:8000/createdir -Method Post -Body $body -ContentType "application/json"
```

---

## 故障排查

### 数据库连接问题

**问题**: `connection refused` 或 `database does not exist`

**解决方案**:
```powershell
# 1. 检查 PostgreSQL 服务是否运行
Get-Service -Name postgresql*

# 2. 启动服务（如果未运行）
Start-Service postgresql-x64-14  # 版本号根据实际情况调整

# 3. 检查数据库是否存在
psql -U postgres -c "\l"

# 4. 如果不存在，运行初始化脚本
.\scripts\init_db.ps1
```

### 端口占用问题

**问题**: `bind: Only one usage of each socket address is normally permitted`

**解决方案**:
```powershell
# 查找占用端口 8000 的进程
netstat -ano | findstr :8000

# 终止进程（替换 PID）
Stop-Process -Id <PID> -Force
```

### Go 依赖问题

**问题**: `cannot find package` 或 `missing module`

**解决方案**:
```powershell
# 下载依赖
go mod download

# 整理依赖
go mod tidy

# 如果还有问题，清理缓存
go clean -modcache
go mod download
```

### 前端问题

**问题**: `Cannot find module` 或 `ENOENT`

**解决方案**:
```powershell
cd frontend

# 删除 node_modules 和重新安装
Remove-Item -Recurse -Force node_modules
npm install

# 清理缓存
npm cache clean --force
npm install
```

### 文件上传失败

**问题**: 上传大文件失败或超时

**检查**:
1. 服务器日志输出
2. 临时目录权限 (`uploads/_tmp/`)
3. 磁盘空间

**解决方案**:
```powershell
# 确保 uploads 目录存在且有写权限
New-Item -ItemType Directory -Force -Path uploads
```

### 查看详细日志

**服务器端**:
- 控制台会输出所有请求和错误信息
- 观察 `[GIN]` 开头的日志

**客户端**:
- chunk_upload.go 会输出详细的上传进度
- 检查返回的错误信息

---

## 开发技巧

### 调试 API

```powershell
# 查看所有数据库记录
curl http://localhost:8000/debug/drivelist

# 查看闭包表关系
curl http://localhost:8000/debug/closure

# 查看子树结构
curl http://localhost:8000/debug/subtree/1
```

### 清理测试数据

```powershell
# 清理上传文件
Remove-Item -Recurse -Force uploads\*

# 重置数据库
psql -U postgres -d tododb -c "TRUNCATE TABLE drivelist_closure, drivelist RESTART IDENTITY CASCADE;"
```

### 性能测试

```powershell
# 测试大文件上传（创建 100MB 文件）
cd client
go run create_test_file.go 100

# 上传并测试
go run chunk_upload.go ../test_file.bin
```

---

## 项目结构

```
single_drive/
├── client/              # 客户端工具
│   ├── chunk_upload.go  # 分块上传客户端
│   └── create_test_file.go
├── cmd/                 # 主程序
│   ├── client/main.go
│   └── server/main.go
├── frontend/            # React 前端
│   ├── src/
│   └── package.json
├── scripts/             # 数据库脚本
│   ├── init_database.sql
│   └── init_db.ps1
├── server/              # 服务器核心
│   └── server.go
├── shared/              # 共享类型
│   └── types.go
├── uploads/             # 上传目录
├── test_chunk_upload.ps1  # 测试脚本
├── go.mod
└── README.md
```

---

## 参考资料

- [Gin 框架文档](https://gin-gonic.com/docs/)
- [PostgreSQL 文档](https://www.postgresql.org/docs/)
- [React 文档](https://react.dev/)
- [Ant Design 文档](https://ant.design/)

---

**遇到问题？** 请检查：
1. 所有服务是否正常运行
2. 数据库连接配置是否正确
3. 端口是否被占用
4. 日志输出中的错误信息
