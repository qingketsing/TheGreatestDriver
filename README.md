# Single Drive - 分布式云盘系统

这是一个使用 Go 语言开发的分布式云盘系统，支持文件上传、下载和存储管理。

## 项目结构

```
single_drive/
├── cmd/                    # 可执行程序入口
│   ├── client/            # 客户端程序
│   │   └── main.go
│   └── server/            # 服务器程序
│       └── main.go
├── server/                # 服务器业务逻辑包
│   └── server.go          # HTTP 路由和处理器
├── shared/                # 共享数据结构和工具
│   └── types.go           # 共享类型定义
├── test/                  # 测试文件
│   └── app.js
├── uploads/               # 服务器上传文件存储目录（运行时生成）
├── go.mod                 # Go 模块定义
└── README.md
```

## 功能特性

- ✅ 文件上传（支持元数据传输）
- ✅ 本地文件存储
- ✅ HTTP multipart/form-data 传输
- ✅ JSON 元数据序列化
- ✅ RESTful API 设计

## 环境要求

- Go 1.20 或更高版本
- Windows/Linux/macOS

## 快速开始

### 1. 安装依赖

```powershell
cd d:\IHaveADream\single_drive
go mod tidy
```

### 2. 启动服务器

在服务器端运行：

```powershell
# Windows PowerShell
Set-Location -Path 'd:\IHaveADream\single_drive'
go run ./cmd/server
```

或者编译后运行：

```bash
# 在项目根目录（例如 /root/single_drive）下执行
go build -o server ./cmd/server
./server
```


服务器将在 `http://localhost:8000` 启动。

### 3. 运行客户端

在另一个终端窗口中运行：

```powershell
# 确保 test/app.js 文件存在
Set-Location -Path 'd:\IHaveADream\single_drive'
go run ./cmd/client
```

或者编译后运行：

```powershell
go build -o client.exe ./cmd/client
.\client.exe
```

## API 接口

### 上传文件

**端点**: `POST /upload`

**请求格式**: `multipart/form-data`

**表单字段**:
- `file`: 文件内容（二进制）
- `meta`: 文件元数据（JSON 字符串）
  ```json
  {
    "name": "文件名",
    "path": "原始路径",
    "capacity": 文件大小（字节）
  }
  ```
- `path`: 存储路径（可选）

**响应**:
```json
{
  "message": "File uploaded successfully",
  "filename": "app.js",
  "path": "./uploads/app.js"
}
```

### 健康检查

**端点**: `GET /`

**响应**: `Hello! This is the Single Drive server.`

## 部署到生产环境

### Linux 服务器部署

1. **编译 Linux 版本**（在开发机上）：

```powershell
# 在 Windows 上交叉编译 Linux 版本
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o server_linux ./cmd/server
```

2. **上传到服务器**：

```bash
scp server_linux user@your-server-ip:~/
```

3. **在服务器上运行**：

```bash
ssh user@your-server-ip
chmod +x ~/server_linux
nohup ~/server_linux > server.log 2>&1 &
```

### 修改客户端上传地址

编辑 `cmd/client/main.go`，将：

```go
uploadURL := "http://localhost:8000/upload"
```

修改为你的服务器地址：

```go
uploadURL := "http://your-server-ip:8000/upload"
```

然后重新编译客户端。

## 配置说明

### 客户端配置

- **本地存储目录**: `D:\drivetest`（可在 `cmd/client/main.go` 中修改）
- **上传服务器地址**: 默认 `localhost:8000`

### 服务器配置

- **监听端口**: `8000`（可在 `cmd/server/main.go` 中修改）
- **文件存储目录**: `./uploads`（可在 `server/server.go` 中修改）

## 开发说明

### 添加新功能

1. **服务器端添加新路由**: 编辑 `server/server.go`，在 `SetupDefaultRouter()` 函数中添加
2. **客户端添加新功能**: 编辑 `cmd/client/main.go`
3. **共享数据结构**: 在 `shared/types.go` 中定义

### 运行测试

```powershell
go test ./...
```

## 常见问题

### Q: 上传失败，提示连接被拒绝
A: 确保服务器已经启动并且监听在正确的端口上。检查防火墙设置。

### Q: 编译时出现 VCS 错误
A: 这是 Git 相关的警告，不影响编译。可以运行以下命令解决：
```powershell
git config --global --add safe.directory D:/
```

### Q: 如何修改上传文件大小限制
A: 在 `server/server.go` 中，Gin 默认的请求体大小限制是 32MB。可以通过中间件修改。

## 下一步计划

- [ ] 用户认证系统（JWT）
- [ ] 文件下载功能
- [ ] 文件列表查询
- [ ] 文件秒传（哈希去重）
- [ ] 断点续传
- [ ] 数据库支持（PostgreSQL/MySQL）
- [ ] 分布式存储
- [ ] Docker 容器化

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
