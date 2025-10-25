# Single Drive - 分布式云盘文件管理系统

基于 Go + PostgreSQL + Gin 框架开发的轻量级云盘系统，支持文件上传、删除、列表查询，并将元数据持久化到数据库。采用面向对象的 Client 设计，提供完整的客户端/服务器架构。

## 核心特性

✅ **文件上传**: multipart/form-data 上传，支持任意文件类型  
✅ **自动去重**: 存在同名文件时自动更新容量而非插入重复记录  
✅ **元数据管理**: 文件名、大小、创建时间自动记录到 PostgreSQL  
✅ **文件列表**: 实时查询服务器上所有文件元数据  
✅ **文件删除**: 同时删除数据库记录和磁盘文件  
✅ **智能缓存**: 客户端上传/删除后自动刷新本地元数据缓存  
✅ **面向对象**: Client 类封装所有操作，代码清晰易维护  
✅ **跨平台**: 支持 Windows/Linux/macOS 编译和部署  

## 项目结构

```
single_drive/
├── cmd/                    # 可执行程序入口
│   ├── client/            # 客户端程序（文件上传）
│   │   └── main.go
│   └── server/            # 服务器程序（HTTP 服务）
│       └── main.go
├── server/                # 服务器业务逻辑包
│   └── server.go          # HTTP 路由、数据库操作、Server 结构体
├── shared/                # 共享数据结构和工具
│   └── types.go           # MetaData 和 FileObject 类型定义
├── test/                  # 测试文件
│   └── app.js             # 示例上传文件
├── uploads/               # 服务器文件存储目录（运行时自动创建）
├── go.mod                 # Go 模块定义
├── go.sum                 # 依赖版本锁定
└── README.md
```

## 技术栈

- **Web 框架**: Gin v1.11.0 (高性能 HTTP 路由)
- **数据库**: PostgreSQL 15+
- **数据库驱动**: github.com/lib/pq
- **开发语言**: Go 1.20+
- **架构模式**: RESTful API + Client-Server

## 快速开始

### 前置要求

- Go 1.20 或更高版本
- PostgreSQL 15+ (运行在 localhost:5432)
- Git (可选，用于克隆仓库)

### 1. 数据库初始化

在 PostgreSQL 中执行以下 SQL（使用 psql 或 pgAdmin）：

```sql
-- 创建数据库
CREATE DATABASE tododb;

-- 连接到数据库
\c tododb

-- 创建文件元数据表（服务器会自动创建，但手动创建可自定义字段）
CREATE TABLE IF NOT EXISTS drivelist (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  capacity BIGINT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- 创建索引（可选，提升查询性能）
CREATE INDEX idx_drivelist_name ON drivelist(name);
CREATE INDEX idx_drivelist_created ON drivelist(created_at DESC);
```

**注意**: 服务器启动时会自动执行 `CREATE TABLE IF NOT EXISTS`，所以如果你不需要自定义索引，可以跳过此步骤。

### 2. 安装项目依赖

```powershell
# 克隆仓库（如果还没有）
git clone <repository-url>
cd single_drive

# 下载 Go 依赖
go mod download

# 或者直接运行（会自动下载依赖）
go mod tidy
```

### 3. 配置数据库连接

当前服务器默认连接参数（`server/server.go` 第 25 行）：

```go
host=localhost 
port=5432 
user=postgres 
password=329426 
dbname=tododb 
sslmode=disable
```

**修改方法**：编辑 `server/server.go` 中的 `SetupDefaultSql()` 函数。

### 4. 启动服务器

```powershell
# 进入项目目录
cd d:\IHaveADream\single_drive

# 启动服务器（开发模式）
go run ./cmd/server
```

**成功输出**：
```
Starting server on :8000...
数据库连接成功
确保表 drivelist 存在
[GIN-debug] GET    /                         --> ...
[GIN-debug] POST   /upload                   --> ...
[GIN-debug] GET    /list                     --> ...
[GIN-debug] DELETE /delete                   --> ...
[GIN-debug] Listening and serving HTTP on :8000
```

### 5. 运行客户端

在**另一个终端**窗口运行：

```powershell
cd d:\IHaveADream\single_drive
go run ./cmd/client
```

**客户端执行流程**：
1. 读取 `test/app.js` 文件
2. 保存副本到本地 `D:\drivetest\` 目录
3. 上传文件到服务器 `http://localhost:8000/upload`
4. 自动调用 `/list` 刷新服务器文件列表
5. 打印所有文件元数据

**成功输出**：
```
文件对象创建成功: &{Name:app.js Capacity:1816}
文件存储成功
文件 app.js 已上传到服务器
已刷新元数据列表，共 1 项
文件上传成功

服务器文件列表: [{Name:app.js Capacity:1816}]
```

## 编译与部署

### 开发环境运行

```powershell
# 启动服务器
go run ./cmd/server

# 启动客户端
go run ./cmd/client
```

### 编译可执行文件

**Windows 版本**:
```powershell
# 编译服务器
go build -o server.exe ./cmd/server

# 编译客户端
go build -o client.exe ./cmd/client

# 运行
.\server.exe
.\client.exe
```

**Linux 版本（在 Windows 上交叉编译）**:
```powershell
# 编译 Linux 服务器
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o server_linux ./cmd/server

# 编译 Linux 客户端
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o client_linux ./cmd/client

# 重置环境变量
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
```

### 生产环境部署

**方案一：直接运行**

```bash
# 上传文件到服务器
scp server_linux user@your-server:/opt/single_drive/

# SSH 登录
ssh user@your-server

# 运行服务器
cd /opt/single_drive
chmod +x server_linux
./server_linux
```

**方案二：使用 systemd（推荐）**

创建服务文件 `/etc/systemd/system/single-drive.service`:

```ini
[Unit]
Description=Single Drive File Server
After=network.target postgresql.service

[Service]
Type=simple
User=your-user
WorkingDirectory=/opt/single_drive
ExecStart=/opt/single_drive/server_linux
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable single-drive
sudo systemctl start single-drive
sudo systemctl status single-drive
```

**方案三：Docker 容器化**

创建 `Dockerfile`:

```dockerfile
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8000
CMD ["./server"]
```

构建并运行：
```bash
docker build -t single-drive:latest .
docker run -d -p 8000:8000 --name single-drive single-drive:latest
```

## API 接口文档

服务器默认运行在 `http://localhost:8000`，提供以下 RESTful API：

### 1. 健康检查

```http
GET /
```

**响应**: 
```
Hello! This is the Single Drive server.
```

**用途**: 验证服务器是否正常运行

---

### 2. 文件上传

```http
POST /upload
Content-Type: multipart/form-data
```

**请求体**:
- `file`: 文件二进制数据（必需）
- `meta`: JSON 字符串格式的元数据（必需）
  ```json
  {
    "name": "example.txt",
    "capacity": 2048
  }
  ```
- `path`: 自定义路径（可选，当前版本未使用）

**PowerShell 示例**:
```powershell
# 使用 Invoke-RestMethod 上传
$filePath = "test.txt"
$meta = @{name="test.txt"; capacity=100} | ConvertTo-Json
$form = @{
    file = Get-Item $filePath
    meta = $meta
}
Invoke-RestMethod -Uri http://localhost:8000/upload -Method Post -Form $form

# 使用 curl.exe
$meta = '{"name":"test.txt","capacity":100}'
curl.exe -X POST -F "file=@test.txt" -F "meta=$meta" http://localhost:8000/upload
```

**成功响应** (200 OK):
```json
{
  "message": "File uploaded successfully",
  "filename": "example.txt",
  "path": "./uploads/example.txt"
}
```

**错误响应** (400/500):
```json
{
  "error": "Invalid meta data: unexpected end of JSON input"
}
```

**服务器处理逻辑**:
1. 解析 `meta` 字段（JSON）
2. 接收文件并保存到 `./uploads/<filename>`
3. 检查数据库是否已存在同名文件：
   - 存在：更新 `capacity` 字段
   - 不存在：插入新记录
4. 返回成功响应

---

### 3. 获取文件列表

```http
GET /list
```

**响应** (200 OK):
```json
[
  {
    "name": "app.js",
    "capacity": 1816
  },
  {
    "name": "test.txt",
    "capacity": 2048
  }
]
```

**PowerShell 示例**:
```powershell
# 查看所有文件
Invoke-RestMethod -Uri http://localhost:8000/list

# 格式化输出
(Invoke-RestMethod -Uri http://localhost:8000/list) | Format-Table
```

**用途**: 获取数据库中所有文件的元数据（不包括文件内容）

---

### 4. 删除文件

```http
DELETE /delete?name=<filename>
```

**查询参数**:
- `name`: 要删除的文件名（必需）

**PowerShell 示例**:
```powershell
# 删除指定文件
Invoke-RestMethod -Uri "http://localhost:8000/delete?name=test.txt" -Method Delete

# 使用 curl.exe
curl.exe -X DELETE "http://localhost:8000/delete?name=test.txt"
```

**成功响应** (200 OK):
```json
{
  "message": "File and record deleted successfully",
  "rows_affected": 1
}
```

**错误响应** (400):
```json
{
  "error": "Missing 'name' query parameter"
}
```

**服务器处理逻辑**:
1. 验证 `name` 参数存在
2. 从数据库删除记录：`DELETE FROM drivelist WHERE name=$1`
3. 从磁盘删除文件：`os.Remove("./uploads/<name>")`
4. 返回受影响的行数

## 客户端 SDK 使用

客户端采用面向对象设计，所有操作封装在 `Client` 类中。

### Client 类结构

```go
type Client struct {
    Files   []shared.FileObject  // 本地文件缓存
    Metas   []shared.MetaData    // 服务器元数据缓存
    BaseURL string               // 服务器地址
}
```

### 方法列表

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `NewClient(baseURL)` | 服务器地址 | `*Client` | 创建客户端实例 |
| `StoreFileObject(fo)` | 文件对象 | `error` | 保存文件到本地 `D:\drivetest` |
| `UploadFileObject(fo, meta)` | 文件对象, 元数据 | `error` | 上传文件并自动刷新缓存 |
| `RefreshMetaList()` | - | `error` | 从服务器获取最新文件列表 |
| `DeleteFile(filename)` | 文件名 | `error` | 删除服务器文件并刷新缓存 |

### 基础使用示例

```go
package main

import (
    "fmt"
    "path/filepath"
    "single_drive/shared"
)

func main() {
    // 1. 创建客户端（自动读取 UPLOAD_URL 环境变量，默认 localhost:8000）
    client := NewClient("")

    // 2. 读取文件
    filePath, _ := filepath.Abs("test/app.js")
    fileObj, meta, err := shared.NewFileObject(filePath)
    if err != nil {
        panic(err)
    }

    // 3. 保存到本地
    if err := client.StoreFileObject(fileObj); err != nil {
        panic(err)
    }
    fmt.Println("本地保存成功")

    // 4. 上传到服务器（会自动刷新 client.Metas）
    if err := client.UploadFileObject(fileObj, meta); err != nil {
        panic(err)
    }
    fmt.Println("上传成功，服务器文件:", client.Metas)

    // 5. 手动刷新列表
    if err := client.RefreshMetaList(); err != nil {
        panic(err)
    }
    fmt.Printf("当前共 %d 个文件\n", len(client.Metas))

    // 6. 删除文件（会自动刷新 client.Metas）
    if err := client.DeleteFile("app.js"); err != nil {
        panic(err)
    }
    fmt.Println("删除成功")
}
```

### 环境变量配置

客户端支持通过环境变量指定服务器地址：

```powershell
# Windows PowerShell
$env:UPLOAD_URL = "http://192.168.1.100:8000"
go run ./cmd/client

# Linux/macOS
export UPLOAD_URL="http://192.168.1.100:8000"
go run ./cmd/client
```

### 自动缓存刷新

- **上传后**: `UploadFileObject` 成功后自动调用 `RefreshMetaList()`
- **删除后**: `DeleteFile` 成功后自动调用 `RefreshMetaList()`
- **手动刷新**: 随时调用 `client.RefreshMetaList()` 获取最新数据

**优势**: 客户端的 `Metas` 字段始终与服务器保持同步，无需手动管理。

## 数据库管理

### 表结构

**drivelist 表**:

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | SERIAL | PRIMARY KEY | 自增主键 |
| name | TEXT | NOT NULL | 文件名 |
| capacity | BIGINT | NOT NULL | 文件大小（字节）|
| created_at | TIMESTAMPTZ | DEFAULT now() | 创建时间（带时区）|

### 常用查询

```sql
-- 查看最近上传的 10 个文件
SELECT id, name, capacity, created_at 
FROM drivelist 
ORDER BY created_at DESC 
LIMIT 10;

-- 按文件名搜索
SELECT * FROM drivelist 
WHERE name LIKE '%app%';

-- 统计总文件数和总大小
SELECT 
    COUNT(*) as total_files, 
    SUM(capacity) as total_bytes,
    pg_size_pretty(SUM(capacity)) as total_size
FROM drivelist;

-- 查找重复文件名
SELECT name, COUNT(*) as count 
FROM drivelist 
GROUP BY name 
HAVING COUNT(*) > 1;

-- 查看表占用空间
SELECT pg_size_pretty(pg_total_relation_size('drivelist'));
```

### 数据维护

```sql
-- 删除 30 天前的旧文件记录
DELETE FROM drivelist 
WHERE created_at < NOW() - INTERVAL '30 days';

-- 清空表（保留结构）
TRUNCATE TABLE drivelist RESTART IDENTITY;

-- 重建索引（性能优化）
REINDEX TABLE drivelist;

-- 分析表（更新统计信息）
ANALYZE drivelist;
```

### 备份与恢复

```bash
# 备份数据库
pg_dump -U postgres -d tododb > backup_$(date +%Y%m%d).sql

# 仅备份 drivelist 表
pg_dump -U postgres -d tododb -t drivelist > drivelist_backup.sql

# 恢复数据库
psql -U postgres -d tododb < backup_20251025.sql

# 导出为 CSV
psql -U postgres -d tododb -c "COPY drivelist TO '/tmp/drivelist.csv' CSV HEADER;"
```

## 高级用法

### 配置服务器地址

**服务器端** (`server/server.go`):

| 配置项 | 位置 | 默认值 |
|--------|------|--------|
| 数据库 DSN | `SetupDefaultSql()` | `tododb@localhost:5432` |
| 监听端口 | `cmd/server/main.go` | `:8000` |
| 上传目录 | `SetupDefaultRouter()` | `./uploads` |

**客户端** (`cmd/client/main.go`):

```go
// 方式 1: 修改代码中的默认值
func NewClient(base string) *Client {
    if base == "" {
        base = "http://192.168.1.100:8000"  // 修改这里
    }
    // ...
}

// 方式 2: 通过环境变量（推荐）
$env:UPLOAD_URL = "http://192.168.1.100:8000"
go run ./cmd/client

// 方式 3: 代码中传参
client := NewClient("http://192.168.1.100:8000")
```

---

## 故障排除

### 问题 1: 数据库连接失败

**错误信息**:
```
pq: password authentication failed for user "postgres"
```

**解决方案**:
1. 检查 PostgreSQL 是否运行：
   ```powershell
   # Windows
   Get-Service postgresql*
   
   # Linux
   systemctl status postgresql
   ```

2. 验证密码：
   ```bash
   psql -U postgres -h localhost -p 5432
   ```

3. 修改 `server/server.go` 第 25 行的 DSN 连接字符串

4. 检查 `pg_hba.conf` 文件允许本地连接：
   ```
   host    all    all    127.0.0.1/32    md5
   ```

---

### 问题 2: 表不存在

**错误信息**:
```
pq: relation "drivelist" does not exist
```

**解决方案**:
1. 确认已连接到正确的数据库（`tododb` 而非 `postgres`）
2. 服务器会在启动时自动创建表，检查启动日志是否有错误
3. 手动创建表（见"数据库初始化"章节）

---

### 问题 3: 客户端上传 404 错误

**错误信息**:
```
delete failed: server returned 404
```

**原因**: 
- 旧版本服务器使用 `GET /delete` 而非 `DELETE /delete`
- 路由方法不匹配

**解决方案**:
1. 确保服务器代码使用 `r.DELETE("/delete", ...)` 而非 `r.GET`
2. 重新编译并启动服务器：
   ```powershell
   go build ./cmd/server
   go run ./cmd/server
   ```

---

### 问题 4: 端口被占用

**错误信息**:
```
listen tcp :8000: bind: Only one usage of each socket address is permitted
```

**解决方案**:
```powershell
# Windows - 查找占用端口的进程
netstat -ano | findstr :8000

# 终止进程（PID 替换为实际值）
taskkill /PID <PID> /F

# Linux
lsof -i :8000
kill -9 <PID>
```

---

### 问题 5: 客户端连接被拒绝

**错误信息**:
```
dial tcp 127.0.0.1:8000: connectex: No connection could be made
```

**解决方案**:
1. 确保服务器已启动（`go run ./cmd/server`）
2. 检查防火墙设置允许 8000 端口
3. 验证服务器监听地址（`:8000` 表示所有网卡）
4. 测试连接：
   ```powershell
   Test-NetConnection -ComputerName localhost -Port 8000
   ```

---

### 问题 6: Git VCS 警告

**错误信息**:
```
error obtaining VCS status: exit status 128
fatal: detected dubious ownership in repository
```

**解决方案**:
```powershell
# 添加安全目录
git config --global --add safe.directory D:/IHaveADream/single_drive
```

---

### 问题 7: 上传后数据库无记录

**可能原因**:
1. 数据库事务未提交
2. 数据库连接中断
3. SQL 插入失败但未返回错误

**调试步骤**:
```go
// 在 server.go 中添加日志
log.Printf("Inserting: name=%s, capacity=%d", meta.Name, meta.Capacity)
result, err := s.DB.Exec("INSERT INTO drivelist ...")
if err != nil {
    log.Printf("Insert error: %v", err)
}
rowsAffected, _ := result.RowsAffected()
log.Printf("Rows affected: %d", rowsAffected)
```

## 性能优化

### 数据库优化

```go
// 设置连接池参数
func (s *Server) SetupDefaultSql() {
    db, err := sql.Open("postgres", dsn)
    // ...
    
    db.SetMaxOpenConns(25)                 // 最大打开连接数
    db.SetMaxIdleConns(5)                  // 最大空闲连接数
    db.SetConnMaxLifetime(5 * time.Minute) // 连接最大生命周期
    
    s.DB = db
}
```

### Gin 中间件优化

```go
func (s *Server) SetupDefaultRouter() {
    r := gin.Default()
    
    // 限制上传大小（32MB）
    r.MaxMultipartMemory = 32 << 20
    
    // 添加 CORS 支持
    r.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Next()
    })
    
    // 请求日志
    r.Use(gin.Logger())
    
    // 崩溃恢复
    r.Use(gin.Recovery())
    
    // ... 路由定义
}
```
## 开发规范

### 项目结构说明

```
cmd/          - 应用程序入口（main 包）
server/       - 业务逻辑层（数据库、路由、Server 结构体）
shared/       - 共享类型和工具函数
test/         - 测试文件和测试数据
uploads/      - 运行时文件存储（gitignore）
```

### 代码风格

- **格式化**: 使用 `gofmt` 或 `goimports`
- **命名规范**:
  - 导出（公开）: `UploadFile`, `MetaData`
  - 私有: `uploadFile`, `metaData`
- **错误处理**: 永远检查 `err != nil`
- **日志**: 使用 `log.Printf` 记录关键操作

### Git 提交规范

```bash
feat: 添加文件下载接口
fix: 修复删除接口 404 错误
docs: 更新 README
refactor: 重构客户端为 OOP 设计
test: 添加上传接口单元测试
```

### 单元测试示例

创建 `server/server_test.go`:

```go
package server

import (
    "testing"
    "net/http/httptest"
)

func TestHealthCheck(t *testing.T) {
    s := InitServer()
    
    req := httptest.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()
    s.Ge.ServeHTTP(w, req)
    
    if w.Code != 200 {
        t.Errorf("Expected 200, got %d", w.Code)
    }
}
```

运行测试：
```powershell
go test ./server -v
```

---

## 项目架构

### 核心组件

**1. Server 结构体** (`server/server.go`):
```go
type Server struct {
    DB       *sql.DB              // PostgreSQL 连接
    Metalist []shared.MetaData    // 内存元数据缓存
    Ge       *gin.Engine          // HTTP 路由引擎
}
```

**2. Client 结构体** (`cmd/client/main.go`):
```go
type Client struct {
    Files   []shared.FileObject  // 本地文件缓存
    Metas   []shared.MetaData    // 服务器元数据缓存
    BaseURL string               // 服务器地址
}
```

**3. 共享数据结构** (`shared/types.go`):
```go
type MetaData struct {
    Name     string `json:"name"`
    Capacity int64  `json:"capacity"`
}

type FileObject struct {
    Name     string
    Capacity int64
    Content  []byte
}
```

### 数据流

```
[用户] → [Client.UploadFileObject()]
    ↓
[HTTP POST /upload] → [Server 路由处理]
    ↓
[解析 multipart] → [保存文件到 ./uploads/]
    ↓
[检查数据库] → [INSERT 或 UPDATE drivelist]
    ↓
[返回成功] → [Client.RefreshMetaList()]
    ↓
[GET /list] → [更新 Client.Metas 缓存]
```

### 扩展建议

- **认证系统**: JWT + 用户表
- **对象存储**: 集成 MinIO/AWS S3
- **消息队列**: 异步处理大文件上传（RabbitMQ/Redis）
- **监控**: Prometheus + Grafana
- **日志**: 结构化日志（zap/logrus）
- **限流**: gin-limiter 中间件
- **文件预览**: 集成 Office 在线预览

---

## 贡献指南

欢迎提交 Issue 和 Pull Request！

### 开发流程

1. Fork 本仓库
2. 创建功能分支：`git checkout -b feature/my-feature`
3. 提交更改：`git commit -am 'Add some feature'`
4. 推送分支：`git push origin feature/my-feature`
5. 提交 Pull Request

### 需要帮助的领域

- [ ] 添加单元测试和集成测试
- [ ] 完善错误处理和日志系统
- [ ] 实现文件下载接口
- [ ] 添加用户认证和权限管理
- [ ] Docker 和 K8s 部署文档
- [ ] 性能基准测试
- [ ] 前端 Web UI（Vue/React）

---

## 许可证

MIT License

Copyright (c) 2025 Single Drive Project

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software.

---

## 致谢

- [Gin](https://github.com/gin-gonic/gin) - 高性能 Web 框架
- [PostgreSQL](https://www.postgresql.org/) - 强大的关系型数据库
- [lib/pq](https://github.com/lib/pq) - Go PostgreSQL 驱动

---

**⭐ 如果这个项目对你有帮助，请给个 Star！**


