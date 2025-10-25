# Single Drive - 云盘文件上传系统

基于 Go + PostgreSQL + Gin 框架开发的文件上传与元数据管理系统，支持文件上传到本地存储并将元数据持久化到数据库。

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

- **后端框架**: Gin (HTTP 路由和中间件)
- **数据库**: PostgreSQL 15+
- **数据库驱动**: github.com/lib/pq
- **语言**: Go 1.20+

## 环境要求

- Go 1.20 或更高版本
- PostgreSQL 15+ (需运行在 localhost:5432)
- Windows/Linux/macOS

## 数据库准备

### 1. 创建数据库和表

在 PostgreSQL 中执行以下命令（使用 psql 或 pgAdmin）：

```bash
# 连接到 PostgreSQL
psql -U postgres -h localhost -p 5432
```

```sql
-- 创建数据库
CREATE DATABASE drivertest;

-- 连接到新数据库
\c drivertest

-- 创建文件元数据表
CREATE TABLE IF NOT EXISTS drivelist (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,              -- 文件名
  capacity BIGINT NOT NULL,        -- 文件大小（字节）
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- 创建索引（可选，提升查询性能）
CREATE INDEX IF NOT EXISTS idx_drivelist_name ON drivelist(name);
```

### 2. 配置数据库连接

当前服务器默认使用以下连接参数（在 `server/server.go` 中）：

```go
host=localhost 
port=5432 
user=postgres 
password=329426 
dbname=drivertest 
sslmode=disable
```

**如需修改**：编辑 `server/server.go` 的 `SetupDefaultSql()` 函数中的 DSN 字符串。

## 快速开始

### 1. 安装 Go 依赖

```powershell
# 进入项目根目录
Set-Location -Path 'd:\IHaveADream\single_drive'

# 安装依赖
go get github.com/gin-gonic/gin@v1.11.0
go get github.com/lib/pq@latest

# 整理模块
go mod tidy
```

### 2. 启动服务器

确保 PostgreSQL 已启动并且 `drivertest` 数据库已创建。

```powershell
# Windows PowerShell
Set-Location -Path 'd:\IHaveADream\single_drive'
go run ./cmd/server
```

服务器启动后会：
1. 连接到 PostgreSQL 数据库
2. 从 `drivelist` 表加载现有元数据到内存
3. 在 `http://localhost:8000` 启动 HTTP 服务

**成功输出示例**：
```
Starting server on :8000...
数据库连接成功
[GIN-debug] Listening and serving HTTP on :8000
```

### 3. 运行客户端上传文件

在另一个终端窗口：

```powershell
# 确保 test/app.js 文件存在
Set-Location -Path 'd:\IHaveADream\single_drive'
go run ./cmd/client
```

**客户端执行流程**：
1. 读取 `test/app.js` 文件
2. 将文件保存到本地 `D:\drivetest` 目录
3. 上传文件和元数据到服务器 `http://localhost:8000/upload`

**成功输出示例**：
```
文件对象创建成功: &{Name:app.js Capacity:1816}
文件存储成功
文件 app.js 已上传到服务器
```

### 4. 验证上传结果

**查看服务器文件**：
- 文件保存在 `single_drive/uploads/app.js`

**查看数据库记录**：
```sql
-- 在 psql 中查询
SELECT * FROM drivelist ORDER BY id DESC LIMIT 10;
```

## 编译生产版本

### Windows 可执行文件

```powershell
# 编译服务器
go build -o server.exe ./cmd/server

# 编译客户端
go build -o client.exe ./cmd/client

# 运行
.\server.exe
.\client.exe
```

### Linux 服务器部署

```powershell
# 在 Windows 上交叉编译 Linux 版本
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o server_linux ./cmd/server
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o client_linux ./cmd/client
```

**上传并运行**：
```bash
# 上传到服务器
scp server_linux user@your-server-ip:~/

# SSH 登录服务器
ssh user@your-server-ip

# 给予执行权限
chmod +x ~/server_linux

# 后台运行
nohup ~/server_linux > server.log 2>&1 &
```

## API 接口文档

### 1. 健康检查

**端点**: `GET /`

**响应**: 
```
Hello! This is the Single Drive server.
```

### 2. 文件上传

**端点**: `POST /upload`

**请求格式**: `multipart/form-data`

**表单字段**:
- `file`: 文件内容（二进制数据）
- `meta`: 文件元数据（JSON 字符串）
  ```json
  {
    "name": "app.js",
    "capacity": 1816
  }
  ```

**请求示例（PowerShell）**:
```powershell
# 使用 curl 上传文件
$meta = '{"name":"test.txt","capacity":100}'
curl.exe -X POST -F "file=@test.txt" -F "meta=$meta" http://localhost:8000/upload
```

**成功响应** (200 OK):
```json
{
  "message": "File uploaded successfully",
  "filename": "app.js",
  "path": "./uploads/app.js"
}
```

**服务器端处理流程**:
1. 解析表单中的 `meta` JSON 字段
2. 接收并保存文件到 `./uploads/` 目录
3. 将元数据（name, capacity）插入到 `drivelist` 表
4. 返回成功响应

## 配置说明

### 服务器配置 (`server/server.go`)

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| 监听端口 | `:8000` | HTTP 服务端口 |
| 数据库地址 | `localhost:5432` | PostgreSQL 地址 |
| 数据库名 | `drivertest` | 数据库名称 |
| 数据库用户 | `postgres` | 登录用户名 |
| 数据库密码 | `329426` | 登录密码 |
| 文件存储目录 | `./uploads` | 上传文件保存路径 |

### 客户端配置 (`cmd/client/main.go`)

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| 上传服务器地址 | `http://localhost:8000/upload` | 服务器上传接口 |
| 本地存储目录 | `D:\drivetest` | 客户端本地备份路径 |
| 测试文件路径 | `test/app.js` | 要上传的文件 |

**修改上传地址**（用于生产部署）：

编辑 `cmd/client/main.go`，将：
```go
uploadURL := "http://localhost:8000/upload"
```
改为：
```go
uploadURL := "http://your-server-ip:8000/upload"
```

## 数据库表结构

### drivelist 表

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | SERIAL | PRIMARY KEY | 自增主键 |
| name | TEXT | NOT NULL | 文件名 |
| capacity | BIGINT | NOT NULL | 文件大小（字节）|
| created_at | TIMESTAMP WITH TIME ZONE | DEFAULT now() | 创建时间 |

**查询示例**：
```sql
-- 查看最近上传的 10 个文件
SELECT id, name, capacity, created_at 
FROM drivelist 
ORDER BY created_at DESC 
LIMIT 10;

-- 按文件名搜索
SELECT * FROM drivelist WHERE name LIKE '%app%';

-- 统计总文件数和总大小
SELECT COUNT(*) as total_files, 
       SUM(capacity) as total_size_bytes,
       pg_size_pretty(SUM(capacity)) as total_size_human
FROM drivelist;
```

## 开发说明

### 项目架构

**Server 结构体** (`server/server.go`):
```go
type Server struct {
    DB       *sql.DB              // PostgreSQL 连接
    Metalist []shared.MetaData    // 内存中的元数据缓存
    Ge       *gin.Engine          // Gin 路由引擎
}
```

**初始化流程**:
1. `InitServer()` 创建 Server 实例
2. `SetupDefaultSql()` 连接数据库并加载现有数据到 `Metalist`
3. `SetupDefaultRouter()` 配置 HTTP 路由
4. `Ge.Run(":8000")` 启动服务

### 添加新功能

**1. 添加新的 HTTP 路由**:

编辑 `server/server.go` 的 `SetupDefaultRouter()` 方法：

```go
func (s *Server) SetupDefaultRouter() {
    r := gin.Default()
    
    // 现有路由...
    
    // 添加新路由：获取文件列表
    r.GET("/files", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "files": s.Metalist,
            "count": len(s.Metalist),
        })
    })
    
    s.Ge = r
}
```

**2. 扩展共享数据结构**:

编辑 `shared/types.go`：

```go
type MetaData struct {
    Name     string `json:"name"`
    Capacity int64  `json:"capacity"`
    // 新增字段
    SHA256   string `json:"sha256"`    // 文件哈希
    MimeType string `json:"mimeType"`  // MIME 类型
}
```

**3. 数据库迁移**:

在 `drivelist` 表添加新列：

```sql
ALTER TABLE drivelist ADD COLUMN sha256 TEXT;
ALTER TABLE drivelist ADD COLUMN mime_type TEXT;
CREATE INDEX idx_drivelist_sha256 ON drivelist(sha256);
```

### 运行测试

```powershell
# 运行所有测试
go test ./...

# 运行指定包的测试
go test ./server

# 带覆盖率
go test -cover ./...
```

### 代码规范

- 使用 `gofmt` 格式化代码
- 变量和函数使用驼峰命名
- 导出的类型、函数首字母大写
- 私有的首字母小写

## 常见问题

### Q: 数据库连接失败

**错误**: `pq: password authentication failed for user "postgres"`

**解决**:
1. 确认 PostgreSQL 正在运行：
   ```powershell
   # Windows 查看服务状态
   Get-Service postgresql*
   ```
2. 检查密码是否正确（默认 `329426`）
3. 修改 `server/server.go` 中的 DSN

### Q: 表不存在错误

**错误**: `pq: relation "drivelist" does not exist`

**解决**:
1. 确认已连接到 `drivertest` 数据库（不是 `postgres` 数据库）
2. 执行建表 SQL（见"数据库准备"章节）
3. 在 psql 中验证：
   ```sql
   \c drivertest
   \dt
   ```

### Q: 上传失败，提示连接被拒绝

**错误**: `dial tcp 127.0.0.1:8000: connectex: No connection could be made...`

**解决**:
1. 确保服务器已启动（`go run ./cmd/server`）
2. 检查端口 8000 是否被占用：
   ```powershell
   netstat -ano | findstr :8000
   ```
3. 检查防火墙设置

### Q: 编译时出现 VCS 错误

**错误**: `error obtaining VCS status: exit status 128`

**解决**:
```powershell
git config --global --add safe.directory D:/
```

### Q: 客户端本地存储目录创建失败

**错误**: Windows 上路径权限问题

**解决**:
修改 `cmd/client/main.go` 中的存储路径为有权限的目录，或以管理员身份运行。

## 性能优化建议

1. **数据库连接池**: 调整 `sql.DB` 的 `SetMaxOpenConns()` 和 `SetMaxIdleConns()`
2. **文件上传限制**: 在 Gin 中配置 `MaxMultipartMemory`
3. **索引优化**: 在频繁查询的字段（如 `name`、`sha256`）上创建索引
4. **分页查询**: 对大数据集使用 `LIMIT` 和 `OFFSET`
5. **缓存**: 使用 Redis 缓存热门文件的元数据

## 下一步开发计划

- [ ] 文件列表分页查询
- [ ] 文件删除接口 (`GET /delete?name=`)
- [ ] 文件下载接口 (`GET /download?name=`)
- [ ] 用户认证系统（JWT）
- [ ] 文件访问权限控制
- [ ] 断点续传支持
- [ ] 对象存储集成（MinIO/S3）
- [ ] Docker 容器化部署
- [ ] 日志系统（zap/logrus）
- [ ] 监控和指标采集（Prometheus）

## 许可证

MIT License


