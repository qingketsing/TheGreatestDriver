# Single Drive - 云存储文件管理系统

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.16+-00ADD8?logo=go)](https://golang.org/)
[![React Version](https://img.shields.io/badge/React-18.2-61DAFB?logo=react)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.2-3178C6?logo=typescript)](https://www.typescriptlang.org/)

> 现代化的云存储文件管理系统，采用 Go + PostgreSQL + React 构建，支持分块上传、秒传、文件树管理等特性。

## ✨ 核心特性

### 后端
- 🚀 高性能 RESTful API（Gin 框架）
- 📦 PostgreSQL + 闭包表实现文件树
- 📤 分块上传 + 秒传（SHA256 哈希检测）
- 📁 完整文件操作（上传/下载/删除/重命名/移动）
- 🗜️ 文件夹 ZIP 打包下载
- 🔐 路径安全检查

### 前端
- 🎨 现代化 UI（Ant Design 5）
- 📊 列表/网格视图切换
- ⬆️ 拖拽批量上传
- 📈 实时进度显示
- 🗂️ 面包屑导航
- 📱 响应式设计

## 🚀 快速开始

### 前置要求

- Go 1.16+
- PostgreSQL 12+
- Node.js 18+（前端可选）

### 三步启动

```powershell
# 1. 初始化数据库
.\scripts\init_db.ps1

# 2. 启动后端
go run cmd/server/main.go

# 3. 访问服务
# http://localhost:8000
```

### 测试分块上传

```powershell
# 自动化测试（创建文件 → 上传 → 秒传 → 子目录上传）
.\test_chunk_upload.ps1
```

## 📖 文档

- **[开发指南](DEVELOPMENT.md)** - 详细的环境配置、API 说明、故障排查

## 🏗️ 技术栈

**后端**: Go + Gin + PostgreSQL  
**前端**: React + TypeScript + Vite + Ant Design  
**存储**: 闭包表（Closure Table）文件树结构

## 📁 项目结构

```
single_drive/
├── client/              # 客户端工具
│   ├── chunk_upload.go  # 分块上传客户端
│   └── create_test_file.go
├── cmd/                 # 主程序
│   ├── server/main.go   # 服务端启动
│   └── client/main.go
├── server/              # 服务器核心
│   └── server.go
├── shared/              # 共享类型
│   └── types.go
├── frontend/            # React 前端
├── scripts/             # 数据库脚本
│   ├── init_database.sql
│   └── init_db.ps1
├── uploads/             # 文件存储目录
└── test_chunk_upload.ps1  # 自动化测试
```

## 🔌 API 端点

### 文件操作
- `POST /upload` - 普通上传
- `POST /upload/quick` - 秒传检测
- `POST /upload/chunk` - 分块上传
- `GET /upload/progress/:id` - 上传进度
- `GET /download?name=` - 下载文件
- `DELETE /delete?name=` - 删除文件
- `PUT /rename` - 重命名
- `PUT /move` - 移动文件

### 目录操作
- `POST /createdir` - 创建目录
- `DELETE /deletedir?name=` - 删除目录
- `GET /downloaddir?name=` - 打包下载目录

### 查询
- `GET /list` - 文件列表（树形结构）
- `GET /info?name=` - 文件详情
- `GET /search?q=` - 搜索文件

### 调试
- `GET /debug/drivelist` - 查看数据库记录
- `GET /debug/closure` - 查看闭包表
- `GET /debug/subtree/:id` - 查看子树

## 💡 使用示例

### 上传文件

```powershell
# 使用测试客户端
cd client
go run chunk_upload.go ../myfile.pdf

# 上传到指定目录
go run chunk_upload.go ../myfile.pdf "documents/work"
```

### API 调用

```powershell
# 查看文件列表
curl http://localhost:8000/list

# 下载文件
curl "http://localhost:8000/download?name=myfile.pdf" -o myfile.pdf

# 创建目录
curl -X POST http://localhost:8000/createdir -H "Content-Type: application/json" -d '{"name":"newfolder"}'
```

## 📊 数据库设计

### 核心表结构

**drivelist** - 文件/目录元数据
```sql
id         SERIAL PRIMARY KEY
name       TEXT NOT NULL          -- 文件/目录路径
capacity   BIGINT NOT NULL        -- 大小（0=目录）
created_at TIMESTAMPTZ DEFAULT now()
```

**drivelist_closure** - 闭包表（文件树关系）
```sql
ancestor   INTEGER NOT NULL       -- 祖先节点ID
descendant INTEGER NOT NULL       -- 后代节点ID
depth      INT NOT NULL           -- 层级深度
created_at TIMESTAMPTZ DEFAULT now()
```

闭包表优势：
- ✅ 快速查询任意节点的所有子节点
- ✅ 快速查询任意节点的所有祖先
- ✅ 简化移动/删除操作
- ✅ 支持高效的子树查询

## 🔧 开发

### 启动开发环境

```powershell
# 后端（热重载需安装 air）
go run cmd/server/main.go

# 前端
cd frontend
npm install
npm run dev
```

### 运行测试

```powershell
# 分块上传测试
.\test_chunk_upload.ps1

# 手动测试
cd client
go run chunk_upload.go ../test.txt
```

### 清理数据

```powershell
# 清理上传文件
Remove-Item -Recurse uploads\*

# 重置数据库
psql -U postgres -d tododb -c "TRUNCATE TABLE drivelist_closure, drivelist RESTART IDENTITY CASCADE;"
```

## 🐛 故障排查

常见问题请参考 **[开发指南](DEVELOPMENT.md#故障排查)**

## 📄 许可证

MIT License

---

**遇到问题？** 请查看 [DEVELOPMENT.md](DEVELOPMENT.md) 获取详细的开发文档和故障排查指南。
