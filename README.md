# Single Drive - 分布式文件存储系统

## 项目简介

Single Drive 是一个基于 Go 语言开发的分布式文件存储系统，支持文件的上传、下载、删除、目录管理等功能。系统使用 PostgreSQL 数据库存储文件元数据，采用闭包表（Closure Table）设计来维护文件树结构，实现高效的层级查询。

## 文件树结构

```
single_drive/
├── go.mod                      # Go 模块依赖管理
├── README.md                   # 项目说明文档
├── cmd/                        # 可执行程序入口
│   ├── server/
│   │   └── main.go            # 服务端启动程序
│   └── client/
│       └── main.go            # 客户端测试程序
├── server/
│   └── server.go              # 服务端核心逻辑（路由、数据库、处理器）
├── client/                    # 客户端实现（预留）
├── shared/
│   └── types.go              # 共享数据结构和工具函数
├── frontend/                  # 前端 Web 应用（React + TypeScript）
│   ├── src/
│   │   ├── components/       # React 组件
│   │   ├── pages/           # 页面组件
│   │   ├── services/        # API 服务
│   │   ├── types/           # TypeScript 类型定义
│   │   └── utils/           # 工具函数
│   ├── package.json         # 前端依赖配置
│   ├── vite.config.ts       # Vite 构建配置
│   └── start.ps1            # 前端启动脚本
├── uploads/                  # 服务端文件存储目录
├── download/                 # 下载目录（客户端）
└── test/                     # 测试数据目录
```

## 文件介绍

### 核心文件

- **`cmd/server/main.go`**: 服务端启动入口，初始化服务器并监听 8000 端口
- **`cmd/client/main.go`**: 客户端程序，提供文件上传、下载、列表查看等功能的测试接口
- **`server/server.go`**: 服务端核心代码，包含：
  - 数据库连接和初始化
  - RESTful API 路由定义
  - 文件上传/下载/删除处理逻辑
  - 目录管理（创建、删除）
  - 文件树构建算法
- **`shared/types.go`**: 共享类型定义，包含：
  - `MetaData`: 文件元数据结构
  - `FileObject`: 文件对象结构
  - `FileTree`: 文件树结构
  - 文件树读取和压缩/解压工具函数

### 数据库设计

系统使用两个主要表：

1. **`drivelist`**: 存储文件和目录的基本信息
   - `id`: 主键，自增
   - `name`: 文件/目录路径
   - `capacity`: 文件大小（目录为 0）
   - `created_at`: 创建时间

2. **`drivelist_closure`**: 闭包表，存储文件树的层级关系
   - `ancestor`: 祖先节点 ID
   - `descendant`: 后代节点 ID
   - `depth`: 层级深度（0 表示自己）
   - 外键级联删除，自动维护关系完整性

## 部署说明

### 前置要求

- Go 1.16 或更高版本
- PostgreSQL 数据库
- Git（可选）

### 本地部署（Windows/macOS）

#### 1. 安装 PostgreSQL

**Windows:**
- 下载并安装 PostgreSQL: https://www.postgresql.org/download/windows/
- 默认端口：5432

**macOS:**
```bash
brew install postgresql
brew services start postgresql
```

#### 2. 创建数据库

```bash
# 连接到 PostgreSQL
psql -U postgres

# 创建数据库
CREATE DATABASE tododb;

# 退出
\q
```

#### 3. 配置数据库连接

编辑 `server/server.go` 中的数据库连接字符串（第 28 行）：

```go
db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=你的密码 dbname=tododb sslmode=disable")
```

#### 4. 安装依赖

```bash
cd single_drive
go mod download
```

#### 5. 启动服务端

```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:8000` 启动

#### 6. 测试客户端（可选）

```bash
go run cmd/client/main.go
```

### Linux 服务器部署

#### 1. 安装 PostgreSQL

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

**CentOS/RHEL:**
```bash
sudo yum install postgresql-server postgresql-contrib
sudo postgresql-setup initdb
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

#### 2. 配置 PostgreSQL

```bash
# 切换到 postgres 用户
sudo -u postgres psql

# 创建数据库和用户
CREATE DATABASE tododb;
CREATE USER youruser WITH PASSWORD 'yourpassword';
GRANT ALL PRIVILEGES ON DATABASE tododb TO youruser;
\q
```

#### 3. 配置防火墙（如果需要）

```bash
# 允许 8000 端口
sudo firewall-cmd --permanent --add-port=8000/tcp
sudo firewall-cmd --reload
```

#### 4. 克隆项目并安装

```bash
# 克隆项目
git clone https://github.com/qingketsing/TheGreatestDriver.git
cd TheGreatestDriver/single_drive

# 安装 Go（如果未安装）
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# 安装依赖
go mod download
```

#### 5. 配置服务

编辑数据库连接信息（`server/server.go`），然后构建：

```bash
# 编译服务端
go build -o single-drive-server cmd/server/main.go

# 创建上传目录
mkdir -p uploads
```

#### 6. 使用 systemd 管理服务

创建服务文件 `/etc/systemd/system/single-drive.service`:

```ini
[Unit]
Description=Single Drive File Storage Service
After=network.target postgresql.service

[Service]
Type=simple
User=youruser
WorkingDirectory=/path/to/single_drive
ExecStart=/path/to/single_drive/single-drive-server
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl start single-drive
sudo systemctl enable single-drive
sudo systemctl status single-drive
```

#### 7. 配置 Nginx 反向代理（可选）

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        client_max_body_size 100M;
    }
}
```

## API 接口文档

基础 URL: `http://your-server:8000`

### 1. 首页检查

**接口**: `GET /`

**描述**: 检查服务是否正常运行

**响应**:
```
Hello! This is the Single Drive server.
```

---

### 2. 上传文件

**接口**: `POST /upload`

**描述**: 上传文件到服务器，支持指定目录路径

**请求参数** (multipart/form-data):
- `file`: 文件内容（必需）
- `meta`: JSON 格式的元数据（必需）
  ```json
  {
    "name": "文件路径",
    "capacity": 文件大小（字节）
  }
  ```
- `path`: 可选，指定上传到的目录路径（如 "test/data"）

**示例**:
```bash
curl -X POST http://localhost:8000/upload \
  -F "file=@/path/to/file.txt" \
  -F 'meta={"name":"file.txt","capacity":1024}' \
  -F "path=test/data"
```

**成功响应**:
```json
{
  "message": "File uploaded successfully",
  "filename": "file.txt",
  "path": "uploads/test/data/file.txt"
}
```

---

### 3. 列出文件

**接口**: `GET /list`

**描述**: 获取文件列表，支持树形结构和扁平列表两种格式

**请求参数**:
- `format`: 可选，`simple` 或 `flat` 返回扁平列表，默认返回树形结构

**树形结构响应**:
```json
{
  "total": 10,
  "roots": [
    {
      "id": 1,
      "name": "test",
      "capacity": 0,
      "is_dir": true,
      "path": "test",
      "children": [
        {
          "id": 2,
          "name": "file.txt",
          "capacity": 1024,
          "is_dir": false,
          "path": "test/file.txt",
          "children": []
        }
      ]
    }
  ]
}
```

**扁平列表响应** (`?format=simple`):
```json
[
  {
    "name": "test",
    "capacity": 0
  },
  {
    "name": "test/file.txt",
    "capacity": 1024
  }
]
```

---

### 4. 下载文件/目录

**接口**: `GET /download`

**描述**: 下载文件或目录（目录会自动打包为 ZIP）

**请求参数**:
- `name`: 文件或目录的路径（必需）

**示例**:
```bash
# 下载单个文件
curl -O http://localhost:8000/download?name=test/file.txt

# 下载目录（自动压缩）
curl -O http://localhost:8000/download?name=test
```

**响应**: 文件流或 ZIP 压缩包

---

### 5. 下载目录（ZIP）

**接口**: `GET /downloaddir`

**描述**: 下载指定目录的 ZIP 压缩包

**请求参数**:
- `dirname`: 目录路径（必需）

**示例**:
```bash
curl -O http://localhost:8000/downloaddir?dirname=test/data
```

**响应**: ZIP 文件流

---

### 6. 删除文件

**接口**: `DELETE /delete`

**描述**: 删除指定文件（同时删除数据库记录和文件系统文件）

**请求参数**:
- `name`: 文件路径（必需）

**示例**:
```bash
curl -X DELETE "http://localhost:8000/delete?name=test/file.txt"
```

**成功响应**:
```json
{
  "message": "File and record deleted successfully",
  "rows_affected": 1
}
```

---

### 7. 删除目录

**接口**: `DELETE /deletedir`

**描述**: 删除指定目录及其所有内容（包括子目录和文件）

**请求参数**:
- `dirname`: 目录路径（必需）

**示例**:
```bash
curl -X DELETE "http://localhost:8000/deletedir?dirname=test/data"
```

**成功响应**:
```json
{
  "message": "Directory and its contents deleted successfully",
  "path": "test/data"
}
```

**注意**: 此操作会递归删除所有子节点，使用闭包表确保数据一致性

---

### 8. 创建目录

**接口**: `POST /createdir`

**描述**: 创建新目录，自动创建父目录并维护闭包表关系

**请求参数**:
- `path`: 目录路径（必需）

**示例**:
```bash
curl -X POST "http://localhost:8000/createdir?path=test/newdir"
```

**成功响应**:
```json
{
  "message": "Directory created successfully",
  "id": 15,
  "path": "test/newdir"
}
```

---

### 调试接口

#### 9. 查看所有数据库记录

**接口**: `GET /debug/drivelist`

**描述**: 查看 drivelist 表的所有记录

**响应**:
```json
{
  "count": 5,
  "items": [
    {
      "id": 1,
      "name": "test",
      "capacity": 0,
      "created_at": "2025-10-29T10:00:00Z"
    }
  ]
}
```

---

#### 10. 查看闭包表

**接口**: `GET /debug/closure`

**描述**: 查看 drivelist_closure 表的所有关系

**响应**:
```json
{
  "count": 10,
  "items": [
    {
      "ancestor": 1,
      "descendant": 2,
      "depth": 1,
      "ancestor_name": "test",
      "descendant_name": "test/file.txt",
      "descendant_capacity": 1024
    }
  ]
}
```

---

#### 11. 查看子树

**接口**: `GET /debug/subtree/:id`

**描述**: 查看指定节点的所有后代节点

**示例**:
```bash
curl http://localhost:8000/debug/subtree/1
```

**响应**:
```json
{
  "root_id": "1",
  "count": 3,
  "items": [
    {
      "id": 1,
      "name": "test",
      "capacity": 0,
      "depth": 0
    },
    {
      "id": 2,
      "name": "test/file.txt",
      "capacity": 1024,
      "depth": 1
    }
  ]
}
```

---

## 错误处理

所有接口在发生错误时返回统一的 JSON 格式：

```json
{
  "error": "错误描述信息"
}
```

常见的 HTTP 状态码：
- `200 OK`: 请求成功
- `400 Bad Request`: 请求参数错误
- `404 Not Found`: 资源不存在
- `500 Internal Server Error`: 服务器内部错误

---

## 安全说明

1. **路径穿越防护**: 系统会检查并拒绝包含 `..`、绝对路径等不安全的路径
2. **数据库事务**: 删除操作使用事务确保数据一致性
3. **级联删除**: 使用外键约束自动维护闭包表完整性

**生产环境建议**:
- 添加身份认证和授权机制
- 配置 HTTPS
- 限制文件上传大小
- 添加访问日志和监控
- 定期备份数据库

---

## 常见问题

### 1. 清空数据库表

```sql
TRUNCATE TABLE drivelist CASCADE;
```

### 2. 查看服务日志

```bash
# 如果使用 systemd
sudo journalctl -u single-drive -f

# 或直接运行时查看终端输出
```

### 3. 修改监听端口

编辑 `cmd/server/main.go`，将 `:8000` 改为其他端口

---

---

## 前端 Web 应用

### 技术栈

- **React 18** - 现代化 UI 框架
- **TypeScript** - 类型安全的 JavaScript
- **Vite** - 快速的前端构建工具
- **Ant Design 5** - 企业级 UI 组件库
- **React Router** - 单页应用路由
- **Axios** - HTTP 请求库

### 功能特性

✨ **文件管理**
- 📁 文件和文件夹的浏览、上传、下载、删除、重命名
- 🌲 支持多层级文件夹导航
- 📊 列表视图和网格视图自由切换
- 🔍 文件搜索和类型过滤

✨ **上传功能**
- 📤 拖拽上传支持
- 📦 批量上传文件
- 📈 实时显示上传进度
- 📂 支持指定目录上传

✨ **用户体验**
- 🎨 基于 Ant Design 的现代化界面
- 📱 响应式设计，适配各种屏幕
- ⚡ 快速的加载和交互体验
- 🔄 实时刷新文件列表

### 前端部署

#### 1. 安装 Node.js

**Windows:**
- 下载并安装 Node.js LTS 版本: https://nodejs.org/
- 推荐版本：Node.js 18 或更高

**macOS:**
```bash
brew install node
```

#### 2. 安装依赖

```powershell
# 进入前端目录
cd frontend

# 安装依赖
npm install
```

#### 3. 启动前端开发服务器

**方式一：使用启动脚本（推荐）**
```powershell
# Windows PowerShell
.\start.ps1
```

**方式二：手动启动**
```powershell
npm run dev
```

前端将在 `http://localhost:12000` 启动

#### 4. 构建生产版本

```powershell
npm run build
```

构建产物将输出到 `dist/` 目录

### 完整启动流程

1. **启动后端服务**
```bash
cd cmd/server
go run main.go
```
后端服务运行在 `http://localhost:8000`

2. **启动前端服务**
```powershell
cd frontend
.\start.ps1
```
前端应用运行在 `http://localhost:12000`

3. **访问应用**
打开浏览器访问 `http://localhost:12000`

### 前端 API 配置

前端通过 Vite 代理访问后端 API，配置在 `frontend/vite.config.ts`：

```typescript
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8000',
      changeOrigin: true,
      rewrite: (path) => path.replace(/^\/api/, ''),
    },
  },
}
```

如果修改了后端端口，需要同步修改此配置。

### 前端页面说明

1. **主页面** (`/`)
   - 文件浏览和管理
   - 面包屑导航
   - 工具栏（新建、上传、刷新、视图切换）

2. **文件列表视图**
   - 表格展示，支持排序
   - 显示文件名、类型、大小、修改时间
   - 右键菜单：下载、重命名、删除

3. **文件网格视图**
   - 卡片式展示
   - 大图标显示文件类型
   - 适合图片和媒体文件浏览

4. **上传模态框**
   - 拖拽上传区域
   - 批量上传队列
   - 实时进度显示

### 开发建议

- **开发模式**: 使用 `npm run dev` 启动热重载开发服务器
- **代码检查**: 使用 `npm run lint` 进行代码规范检查
- **类型检查**: TypeScript 提供完整的类型安全
- **组件复用**: 所有组件都在 `src/components` 目录下

---

## 许可证

本项目采用 MIT 许可证

## 联系方式

- GitHub: https://github.com/qingketsing/TheGreatestDriver
- Issues: https://github.com/qingketsing/TheGreatestDriver/issues
