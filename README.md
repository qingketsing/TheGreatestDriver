# Single Drive - 云存储文件管理系统

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.16+-00ADD8?logo=go)](https://golang.org/)
[![React Version](https://img.shields.io/badge/React-18.2-61DAFB?logo=react)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.2-3178C6?logo=typescript)](https://www.typescriptlang.org/)

## 📋 项目简介

Single Drive 是一个现代化的云存储文件管理系统，采用前后端分离架构。后端使用 Go + Gin + PostgreSQL 构建高性能 RESTful API，前端使用 React + TypeScript + Ant Design 打造类似 OneDrive 的用户界面。系统采用闭包表（Closure Table）设计维护文件树结构，支持高效的层级查询和文件管理操作。

### ✨ 主要特性

**后端特性**：
- 🚀 高性能 RESTful API（基于 Gin 框架）
- 📦 PostgreSQL 数据库存储元数据
- 🌲 闭包表实现高效文件树管理
- 📁 完整的文件操作（上传、下载、删除、重命名、移动）
- 📂 目录管理（创建、删除、遍历）
- 🔐 路径安全检查，防止路径穿越攻击
- 💾 事务处理保证数据一致性
- 🗜️ 文件夹自动打包为 ZIP 下载

**前端特性**：
- 🎨 现代化 UI 设计（基于 Ant Design 5）
- 📊 列表视图和网格视图自由切换
- ⬆️ 拖拽上传，支持批量上传
- 📈 实时上传进度显示
- 🗂️ 面包屑导航，快速目录切换
- 🖱️ 右键菜单快捷操作
- 📱 响应式设计，适配各种屏幕
- ⚡ Vite 构建，快速热更新

## 🏗️ 技术栈

### 后端
- **语言**: Go 1.16+
- **Web 框架**: Gin
- **数据库**: PostgreSQL
- **ORM**: database/sql (原生)
- **文件处理**: archive/zip

### 前端
- **框架**: React 18.2
- **语言**: TypeScript 5.2
- **构建工具**: Vite 5.0
- **UI 库**: Ant Design 5.12
- **路由**: React Router 6
- **HTTP 客户端**: Axios 1.6
- **日期处理**: Day.js

## 📁 项目结构

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

## 🗄️ 数据库设计

### 数据表结构

#### 1. drivelist（文件元数据表）
```sql
CREATE TABLE drivelist (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,              -- 文件/目录完整路径
    capacity BIGINT NOT NULL,         -- 文件大小（目录为0）
    created_at TIMESTAMPTZ DEFAULT now()
);
```

#### 2. drivelist_closure（闭包表）
```sql
CREATE TABLE drivelist_closure (
    ancestor INTEGER NOT NULL,        -- 祖先节点ID
    descendant INTEGER NOT NULL,      -- 后代节点ID
    depth INT NOT NULL,               -- 层级深度（0表示自己）
    created_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (ancestor, descendant),
    FOREIGN KEY (ancestor) REFERENCES drivelist(id) ON DELETE CASCADE,
    FOREIGN KEY (descendant) REFERENCES drivelist(id) ON DELETE CASCADE
);

-- 性能优化索引
CREATE INDEX idx_closure_ancestor ON drivelist_closure(ancestor);
CREATE INDEX idx_closure_descendant ON drivelist_closure(descendant);
CREATE INDEX idx_closure_depth ON drivelist_closure(depth);
```

### 闭包表优势

- ✅ **高效查询**: O(1) 时间复杂度查询所有祖先/后代
- ✅ **简化操作**: 移动/删除节点时自动维护关系
- ✅ **级联删除**: 利用外键约束自动清理
- ✅ **支持任意深度**: 不受树深度限制

### 核心文件说明

**后端核心**：
- `cmd/server/main.go` - 服务端启动入口（监听 8000 端口）
- `server/server.go` - 核心业务逻辑（约 1000 行代码）
  - 数据库连接和初始化
  - RESTful API 路由定义
  - 文件上传/下载/删除处理
  - 目录管理和文件树构建
  - 闭包表维护逻辑
- `shared/types.go` - 共享数据结构和工具函数

**前端核心**：
- `frontend/src/App.tsx` - 主应用组件
- `frontend/src/components/Layout/MainLayout.tsx` - 布局框架
- `frontend/src/pages/HomePage.tsx` - 文件管理主页面
- `frontend/src/components/FileList/` - 列表和网格视图
- `frontend/src/components/Upload/` - 上传组件
- `frontend/src/services/api.ts` - API 服务封装（15+ 接口）
- `frontend/src/utils/helpers.ts` - 工具函数库

## 🚀 快速开始

### 前置要求

**后端**：
- Go 1.16 或更高版本
- PostgreSQL 数据库

**前端**：
- Node.js 18 或更高版本
- npm 或 yarn

### 安装步骤

#### 1. 克隆项目

```bash
git clone https://github.com/qingketsing/TheGreatestDriver.git
cd single_drive
```

#### 2. 配置数据库

**安装 PostgreSQL**：

Windows: https://www.postgresql.org/download/windows/

macOS:
```bash
brew install postgresql
brew services start postgresql
```

**创建数据库**：

```bash
# 连接到 PostgreSQL
psql -U postgres

# 在 psql 命令行中执行
CREATE DATABASE tododb;
\q
```

**配置连接信息**：

编辑 `server/server.go`（约第 33 行）：
```go
db, err := sql.Open("postgres", 
    "host=localhost port=5432 user=postgres password=你的密码 dbname=tododb sslmode=disable")
```

#### 3. 启动后端服务

```bash
cd cmd/server
go run main.go
```

看到以下输出表示成功：
```
数据库连接成功
确保表 drivelist 存在
确保表 drivelist_closure 存在
确保 drivelist_closure 索引存在
[GIN-debug] Listening and serving HTTP on :8000
```

后端 API: `http://localhost:8000`

#### 4. 启动前端服务

打开新的终端窗口：

```powershell
cd frontend

# 首次运行需要安装依赖
npm install

# 启动开发服务器
npm run dev
# 或使用启动脚本
.\start.ps1
```

看到以下输出表示成功：
```
➜  Local:   http://localhost:12000/
```

前端应用: `http://localhost:12000`

#### 5. 访问应用

打开浏览器访问: **http://localhost:12000**

### 生产环境部署

#### 后端编译

```bash
cd cmd/server
go build -o single_drive_server
./single_drive_server
```

#### 前端构建

```bash
cd frontend
npm run build
# 构建产物在 dist/ 目录
```

#### 配置外网访问

**修改前端配置** (`frontend/vite.config.ts`)：
```typescript
server: {
  host: '0.0.0.0',  // 允许外网访问
  port: 12000,
}
```

**开放防火墙端口**：
```bash
# Linux (firewalld)
sudo firewall-cmd --zone=public --add-port=8000/tcp --permanent
sudo firewall-cmd --zone=public --add-port=12000/tcp --permanent
sudo firewall-cmd --reload

# 或 ufw
sudo ufw allow 8000/tcp
sudo ufw allow 12000/tcp
```

**云服务器安全组**：在云服务商控制台开放 8000 和 12000 端口

## 📡 API 文档

### 基础信息
- **Base URL**: `http://localhost:8000`
- **Content-Type**: `application/json` (除文件上传外)

### 文件操作

#### 上传文件
```http
POST /upload
Content-Type: multipart/form-data

Parameters:
- file: 文件对象 (required)
- meta: JSON字符串 {"name": "path/to/file.txt", "capacity": 1024}
- path: 上传目录路径 (optional)

Response:
{
  "message": "File uploaded successfully",
  "filename": "file.txt",
  "path": "/uploads/path/to/file.txt"
}
```

#### 下载文件/文件夹
```http
GET /download?name=path/to/file.txt

Response: 文件流（文件夹自动打包为 ZIP）
```

#### 删除文件
```http
DELETE /delete?name=path/to/file.txt

Response:
{
  "message": "File and record deleted successfully",
  "rows_affected": 1
}
```

#### 重命名文件
```http
PUT /rename?oldName=old.txt&newName=new.txt

Response:
{
  "message": "File renamed successfully",
  "old_name": "old.txt",
  "new_name": "new.txt"
}
```

#### 移动文件
```http
PUT /move?oldpath=folder1/file.txt&newparent=folder2

Response:
{
  "message": "File/folder moved successfully",
  "old_path": "folder1/file.txt",
  "new_path": "folder2/file.txt"
}
```

#### 获取文件信息
```http
GET /info?name=path/to/file.txt

Response:
{
  "name": "file.txt",
  "size": 1024,
  "mode": "-rw-r--r--",
  "mod_time": "2025-11-01T12:00:00Z",
  "is_directory": false
}
```

### 目录操作

#### 创建目录
```http
POST /createdir?path=new/folder/path

Response:
{
  "message": "Directory created successfully",
  "id": 123,
  "path": "new/folder/path"
}
```

#### 删除目录
```http
DELETE /deletedir?dirname=folder/path

Response:
{
  "message": "Directory and its contents deleted successfully",
  "path": "folder/path"
}
```

#### 下载目录（ZIP）
```http
GET /downloaddir?dirname=folder/path

Response: ZIP 文件流
```

### 文件列表

#### 获取文件树
```http
GET /list

Response:
{
  "total": 10,
  "roots": [
    {
      "id": 1,
      "name": "folder1",
      "capacity": 0,
      "is_dir": true,
      "path": "folder1",
      "children": [...]
    }
  ]
}
```

#### 获取简单列表
```http
GET /list?format=simple

Response:
[
  {"name": "file1.txt", "capacity": 1024},
  {"name": "folder1", "capacity": 0}
]
```

### 调试接口

#### 查看数据库记录
```http
GET /debug/drivelist
GET /debug/closure
GET /debug/subtree/:id
```

### 待实现接口

以下接口已定义路由但功能待实现（标记为 TODO）：

- `DELETE /batch-delete` - 批量删除
- `POST /batch-download` - 批量下载
- `GET /search` - 文件搜索
- `GET /filter/type` - 按类型过滤
- `GET /filter/date` - 按日期过滤
- `GET /filter/size` - 按大小过滤

## 💡 核心功能实现

### 1. 文件上传

支持指定目录上传，自动创建不存在的父目录，维护闭包表关系：

```go
// 1. 保存文件到 uploads/path/file.txt
// 2. 插入数据库记录
// 3. 维护闭包表关系（自己到自己 + 与父节点的关系）
```

### 2. 文件移动

使用事务保证文件系统和数据库一致性：

```go
// 1. 开始事务
// 2. 移动文件系统文件
// 3. 更新数据库路径
// 4. 删除旧的闭包关系
// 5. 建立新的闭包关系
// 6. 更新所有子节点路径
// 7. 提交事务（失败则回滚）
```

### 3. 目录删除

利用闭包表级联删除所有后代节点：

```go
// 1. 查询目录 ID
// 2. 删除闭包表中所有后代（CASCADE 自动处理）
// 3. 删除文件系统目录
```

### 4. 文件树构建

从闭包表高效构建完整文件树：

```go
// 1. 查询所有节点
// 2. 查询 depth=1 的父子关系
// 3. 构建树形结构
// 4. 返回根节点列表
```

## 🎨 前端界面

### 主要页面

1. **文件浏览页面** (`/`)
   - 面包屑导航
   - 工具栏（新建、上传、刷新、视图切换）
   - 文件列表（表格或卡片）
   - 右键菜单

2. **上传模态框**
   - 拖拽上传区域
   - 文件队列管理
   - 实时进度显示

3. **操作确认对话框**
   - 删除确认
   - 重命名输入
   - 新建文件夹

### 组件架构

```
App
├── MainLayout (布局)
│   ├── Header (顶部导航)
│   ├── Sider (侧边栏)
│   └── Content (内容区)
│       └── HomePage (文件管理页面)
│           ├── FileList (列表视图)
│           ├── FileGrid (网格视图)
│           └── UploadModal (上传组件)
```

### API 集成

前端通过 Vite 代理访问后端：

```
前端请求: http://localhost:12000/api/list
      ↓ (Vite 代理)
后端接收: http://localhost:8000/list
```

配置在 `frontend/vite.config.ts`：
```typescript
server: {
  host: '0.0.0.0',
  port: 12000,
  proxy: {
    '/api': {
      target: 'http://localhost:8000',
      changeOrigin: true,
      rewrite: (path) => path.replace(/^\/api/, ''),
    },
  },
}
```

## 🔒 安全注意事项

### 当前状态（开发版本）

⚠️ **本项目为学习和开发用途，不建议直接用于生产环境**

当前缺少的安全功能：
- ❌ 用户认证和授权系统
- ❌ 文件访问权限控制
- ❌ 路径遍历攻击防护（需添加路径验证）
- ❌ 文件类型和大小限制
- ❌ CSRF 防护
- ❌ HTTPS 加密传输

### 生产环境建议

如需在生产环境使用，请务必添加：

1. **身份验证**
   - JWT Token 或 Session 管理
   - 用户注册和登录系统

2. **权限控制**
   - 文件所有权验证
   - 读写权限分离
   - 共享链接功能

3. **输入验证**
   ```go
   // 防止路径遍历攻击
   if strings.Contains(path, "..") {
       return errors.New("invalid path")
   }
   ```

4. **文件限制**
   - 上传大小限制
   - 文件类型白名单
   - 病毒扫描集成

5. **网络安全**
   - HTTPS 证书配置
   - CORS 策略设置
   - Rate Limiting

6. **数据备份**
   - 数据库定期备份
   - 文件存储冗余

## 🛠️ 技术亮点

### 1. 闭包表设计

使用闭包表（Closure Table）管理文件树，相比传统方法的优势：

| 特性 | 闭包表 | 邻接表 | 路径枚举 |
|------|--------|--------|----------|
| 查询所有后代 | O(1) | O(N*logN) | O(N) |
| 插入节点 | O(depth) | O(1) | O(1) |
| 移动子树 | O(nodes) | O(N) | O(N) |
| 删除子树 | O(1)（CASCADE） | O(N) | O(N) |
| 存储开销 | O(N²) | O(N) | O(N) |

### 2. 事务一致性

文件移动操作使用数据库事务确保文件系统和数据库的强一致性：

```go
tx, _ := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
        // 回滚文件系统操作
        os.Rename(newFullPath, oldFullPath)
    }
}()
// ... 数据库操作
tx.Commit()
```

### 3. 前端性能优化

- **虚拟滚动**（未来计划）：处理大量文件列表
- **图标缓存**：复用文件类型图标
- **懒加载**：按需加载目录内容
- **批量操作**：减少网络请求

## 🐛 故障排查

### 后端问题

**问题：数据库连接失败**
```
panic: pq: password authentication failed for user "postgres"
```

解决：
1. 检查 `server/server.go` 中的数据库连接字符串
2. 确认 PostgreSQL 服务已启动
3. 验证用户名和密码

**问题：端口被占用**
```
bind: address already in use
```

解决：
```bash
# Windows PowerShell
Get-Process -Id (Get-NetTCPConnection -LocalPort 8000).OwningProcess | Stop-Process

# Linux/macOS
lsof -ti:8000 | xargs kill
```

### 前端问题

**问题：API 请求 404**
```
GET http://localhost:12000/api/list 404 (Not Found)
```

解决：
1. 确认后端服务已启动（localhost:8000）
2. 检查 `vite.config.ts` 代理配置
3. 查看浏览器开发者工具 Network 标签

**问题：依赖安装失败**
```
npm ERR! code ERESOLVE
```

解决：
```bash
# 清理缓存
npm cache clean --force
rm -rf node_modules package-lock.json

# 重新安装
npm install --legacy-peer-deps
```

### 外网访问问题

**问题：外网无法访问前端**

检查清单：
- [ ] `vite.config.ts` 设置 `host: '0.0.0.0'`
- [ ] 防火墙已开放 12000 端口
- [ ] 云服务器安全组规则已配置
- [ ] 使用正确的公网 IP 访问

详见 [PORT_CONFIG.md](PORT_CONFIG.md)

## 📚 相关文档

- [快速开始指南](QUICKSTART.md) - 详细安装步骤
- [端口配置说明](PORT_CONFIG.md) - 外网访问配置
- [前端开发文档](frontend/README.md) - 前端技术细节
- [功能总结](frontend/SUMMARY.md) - 功能清单
- [Sprint 任务](TODO_sprint_1.md) - 开发任务列表

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

### 代码规范

**Go 代码**：
- 遵循 `gofmt` 格式化
- 使用有意义的变量名
- 添加必要的注释

**TypeScript/React**：
- 使用 ESLint 和 Prettier
- 组件使用 PascalCase
- 函数使用 camelCase

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 👥 作者

- **qingketsing** - *初始工作* - [GitHub](https://github.com/qingketsing)

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) - Go Web 框架
- [React](https://react.dev/) - 前端框架
- [Ant Design](https://ant.design/) - UI 组件库
- [PostgreSQL](https://www.postgresql.org/) - 数据库系统

## 📮 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 [Issue](https://github.com/qingketsing/TheGreatestDriver/issues)
- 发送邮件到项目维护者

---

⭐ 如果这个项目对你有帮助，请给个 Star！

**项目状态**: 🚧 活跃开发中


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

---

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 👥 作者

- **qingketsing** - *初始工作* - [GitHub](https://github.com/qingketsing)

## 🙏 致谢

- [Gin](https://github.com/gin-gonic/gin) - Go Web 框架
- [React](https://react.dev/) - 前端框架
- [Ant Design](https://ant.design/) - UI 组件库
- [PostgreSQL](https://www.postgresql.org/) - 数据库系统

## 📮 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 [Issue](https://github.com/qingketsing/TheGreatestDriver/issues)
- 发送邮件到项目维护者

---

⭐ 如果这个项目对你有帮助，请给个 Star！

**项目状态**: 🚧 活跃开发中

