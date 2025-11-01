# Single Drive 快速启动指南

## 📋 目录
1. [系统要求](#系统要求)
2. [后端启动](#后端启动)
3. [前端启动](#前端启动)
4. [验证部署](#验证部署)
5. [常见问题](#常见问题)

## 系统要求

### 后端
- ✅ Go 1.16 或更高版本
- ✅ PostgreSQL 数据库
- ✅ Git (可选)

### 前端
- ✅ Node.js 18 或更高版本
- ✅ npm 或 yarn

## 后端启动

### 第一步：安装 PostgreSQL

**Windows:**
1. 访问 https://www.postgresql.org/download/windows/
2. 下载并安装 PostgreSQL
3. 记住设置的密码（默认用户名是 postgres）

**macOS:**
```bash
brew install postgresql
brew services start postgresql
```

### 第二步：创建数据库

打开命令行/终端，执行：

```bash
# 连接到 PostgreSQL
psql -U postgres

# 在 psql 命令行中执行
CREATE DATABASE tododb;

# 退出 psql
\q
```

### 第三步：配置数据库连接

编辑 `server/server.go` 文件（约第 33 行），修改数据库连接信息：

```go
db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=你的密码 dbname=tododb sslmode=disable")
```

### 第四步：启动后端服务

```bash
# 进入项目目录
cd single_drive

# 进入服务器目录
cd cmd/server

# 运行服务器
go run main.go
```

看到以下输出表示启动成功：
```
2025/10/31 xx:xx:xx 数据库连接成功
2025/10/31 xx:xx:xx 确保表 drivelist 存在
2025/10/31 xx:xx:xx 确保表 drivelist_closure 存在
2025/10/31 xx:xx:xx 确保 drivelist_closure 索引存在
[GIN-debug] Listening and serving HTTP on :8000
```

后端服务现在运行在 `http://localhost:8000`

## 前端启动

### 第一步：安装 Node.js

**Windows:**
1. 访问 https://nodejs.org/
2. 下载并安装 LTS 版本
3. 安装完成后重启终端

**macOS:**
```bash
brew install node
```

验证安装：
```bash
node --version
npm --version
```

### 第二步：安装依赖

打开新的终端窗口：

```powershell
# 进入前端目录
cd single_drive/frontend

# 安装依赖（首次运行需要几分钟）
npm install
```

### 第三步：启动前端开发服务器

**Windows PowerShell（推荐）:**
```powershell
.\start.ps1
```

**或者手动启动:**
```powershell
npm run dev
```

看到以下输出表示启动成功：
```
  VITE v5.x.x  ready in xxx ms

  ➜  Local:   http://localhost:12000/
  ➜  Network: use --host to expose
  ➜  press h to show help
```

前端应用现在运行在 `http://localhost:12000`

## 验证部署

### 检查清单

1. ✅ **后端服务检查**
   - 打开浏览器访问 `http://localhost:8000`
   - 应该看到: `Hello! This is the Single Drive server.`

2. ✅ **前端服务检查**
   - 打开浏览器访问 `http://localhost:12000`
   - 应该看到 Single Drive 的用户界面

3. ✅ **功能测试**
   - 尝试创建文件夹
   - 尝试上传文件
   - 尝试下载文件
   - 尝试删除文件

### API 测试

使用浏览器或 curl 测试后端 API：

```bash
# 获取文件列表
curl http://localhost:8000/list

# 创建文件夹
curl -X POST "http://localhost:8000/createdir?path=test"
```

## 常见问题

### Q1: 后端启动失败 - "connection refused"

**原因**: PostgreSQL 未启动或连接信息错误

**解决**:
1. 检查 PostgreSQL 是否运行
   ```bash
   # Windows
   services.msc  # 查找 postgresql 服务
   
   # macOS
   brew services list
   ```

2. 检查 `server/server.go` 中的数据库连接信息
3. 确保数据库 `tododb` 已创建

### Q2: 前端启动失败 - "Cannot find module"

**原因**: 依赖未正确安装

**解决**:
```powershell
# 删除 node_modules
Remove-Item -Recurse -Force node_modules

# 删除 package-lock.json
Remove-Item package-lock.json

# 重新安装
npm install
```

### Q3: 前端无法访问后端 API

**原因**: 后端未启动或端口不匹配

**解决**:
1. 确保后端运行在 `http://localhost:8000`
2. 检查 `frontend/vite.config.ts` 中的代理配置
3. 查看浏览器控制台的网络请求

### Q4: 上传文件失败

**原因**: uploads 目录不存在或无权限

**解决**:
```bash
# 在项目根目录创建 uploads 文件夹
mkdir uploads

# 或让服务器自动创建（重启后端）
```

### Q5: 端口被占用

**后端端口 8000 被占用**:
- 修改 `cmd/server/main.go` 中的端口号
- 同时修改 `frontend/vite.config.ts` 中的代理目标端口

**前端端口 12000 被占用**:
- 修改 `frontend/vite.config.ts` 中的 `server.port`

## 开发模式 vs 生产模式

### 开发模式（当前）
- 前端：热重载，实时更新
- 后端：需要手动重启
- 数据：使用本地 PostgreSQL

### 生产模式（未来）
```bash
# 前端构建
cd frontend
npm run build

# 后端编译
cd cmd/server
go build -o single_drive_server

# 部署
# 将 dist/ 目录和编译后的二进制文件部署到服务器
```

## 下一步

现在您可以：

1. 🎯 **熟悉界面** - 浏览所有功能
2. 📁 **测试文件操作** - 上传、下载、删除文件
3. 🔍 **探索源码** - 了解系统实现
4. 🛠️ **自定义开发** - 添加新功能

## 获取帮助

遇到问题？

- 📖 查看完整 [README.md](../README.md)
- 🐛 提交 [Issue](https://github.com/qingketsing/TheGreatestDriver/issues)
- 💬 查看项目文档

---

**祝您使用愉快！** 🎉
