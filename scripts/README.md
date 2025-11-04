# 数据库初始化 - 使用说明

## 概述

本目录包含用于初始化 Single Drive 数据库的脚本和工具。

## 文件说明

- `init_database.sql` - SQL 初始化脚本（包含表结构和索引）
- `init_db.ps1` - PowerShell 自动化初始化脚本

## 快速开始

### 方法1：使用 PowerShell 脚本（推荐）

```powershell
# 在项目根目录执行
.\scripts\init_db.ps1
```

脚本会自动：
1. ✅ 检查 PostgreSQL 安装
2. ✅ 测试数据库连接
3. ✅ 创建数据库（如果不存在）
4. ✅ 检查并创建表
5. ✅ 创建索引
6. ✅ 验证结果

### 方法2：手动执行 SQL

```bash
# 1. 连接到 PostgreSQL
psql -U postgres

# 2. 创建数据库（如果不存在）
CREATE DATABASE tododb;

# 3. 连接到数据库
\c tododb

# 4. 执行初始化脚本
\i scripts/init_database.sql

# 5. 查看表
\dt

# 6. 退出
\q
```

### 方法3：使用 psql 命令行

```bash
# Windows
set PGPASSWORD=你的密码
psql -h localhost -U postgres -d tododb -f scripts/init_database.sql

# Linux/macOS
PGPASSWORD=你的密码 psql -h localhost -U postgres -d tododb -f scripts/init_database.sql
```

## 配置参数

如果需要修改数据库连接参数：

```powershell
# 使用自定义参数
.\scripts\init_db.ps1 -DbHost "localhost" -DbPort "5432" -DbUser "postgres" -DbPassword "你的密码" -DbName "tododb"
```

## 表结构说明

### 1. drivelist 表

存储文件和目录的元数据：

```sql
CREATE TABLE drivelist (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,              -- 文件/目录路径
    capacity BIGINT NOT NULL,        -- 文件大小
    created_at TIMESTAMPTZ DEFAULT now()
);
```

### 2. drivelist_closure 表

闭包表，用于存储文件树的层级关系：

```sql
CREATE TABLE drivelist_closure (
    ancestor INTEGER NOT NULL,       -- 祖先节点ID
    descendant INTEGER NOT NULL,     -- 后代节点ID
    depth INT NOT NULL,              -- 层级深度
    created_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (ancestor, descendant),
    FOREIGN KEY (ancestor) REFERENCES drivelist(id) ON DELETE CASCADE,
    FOREIGN KEY (descendant) REFERENCES drivelist(id) ON DELETE CASCADE
);
```

## 索引

脚本会自动创建以下索引以优化性能：

**drivelist 表**：
- `idx_drivelist_name` - 文件名索引
- `idx_drivelist_created_at` - 创建时间索引

**drivelist_closure 表**：
- `idx_closure_ancestor` - 祖先节点索引
- `idx_closure_descendant` - 后代节点索引
- `idx_closure_depth` - 深度索引
- `idx_closure_ancestor_depth` - 组合索引

## 验证安装

### 检查表是否创建

```sql
-- 查看所有表
SELECT table_name FROM information_schema.tables 
WHERE table_schema='public';

-- 查看表结构
\d drivelist
\d drivelist_closure
```

### 检查索引

```sql
SELECT tablename, indexname 
FROM pg_indexes 
WHERE schemaname='public';
```

### 测试插入

```sql
-- 插入测试数据
INSERT INTO drivelist (name, capacity) VALUES ('test', 0) RETURNING id;
-- 假设返回 id = 1

-- 插入闭包关系
INSERT INTO drivelist_closure (ancestor, descendant, depth) 
VALUES (1, 1, 0);

-- 查询
SELECT * FROM drivelist;
SELECT * FROM drivelist_closure;
```

## 故障排查

### 问题1：psql 命令未找到

**解决方法**：
1. 确认已安装 PostgreSQL
2. 将 PostgreSQL bin 目录添加到 PATH
   - Windows: `C:\Program Files\PostgreSQL\15\bin`
   - Linux/macOS: 通常已在 PATH 中

### 问题2：连接被拒绝

**检查清单**：
- [ ] PostgreSQL 服务是否运行
  - Windows: 检查服务 `postgresql-x64-15`
  - Linux: `sudo systemctl status postgresql`
- [ ] 端口 5432 是否正确
- [ ] 用户名密码是否正确

### 问题3：密码认证失败

**解决方法**：
1. 修改脚本中的密码参数
2. 或设置环境变量：
   ```powershell
   $env:PGPASSWORD = "你的密码"
   ```

### 问题4：权限不足

**解决方法**：
- 使用 postgres 超级用户
- 或确保当前用户有创建数据库和表的权限

### 问题5：表已存在

脚本会检测现有表并提示是否删除重建。如果需要保留数据，请取消操作。

## 重置数据库

如果需要完全重置：

```sql
-- 删除所有表
DROP TABLE IF EXISTS drivelist_closure CASCADE;
DROP TABLE IF EXISTS drivelist CASCADE;

-- 然后重新运行初始化脚本
```

或使用 PowerShell 脚本，会提示是否删除现有表。

## 下一步

数据库初始化完成后：

1. **启动服务器**
   ```bash
   cd cmd/server
   go run main.go
   ```

2. **测试上传**
   ```bash
   cd client
   go run chunk_upload_test.go ../test_file.txt
   ```

3. **查看数据**
   - 浏览器访问: http://localhost:8000/debug/drivelist
   - 或使用 psql 查询

## 备份和恢复

### 备份数据库

```bash
pg_dump -U postgres -d tododb > backup.sql
```

### 恢复数据库

```bash
psql -U postgres -d tododb < backup.sql
```

## 参考资料

- PostgreSQL 官方文档: https://www.postgresql.org/docs/
- Single Drive 项目文档: ../README.md
- 闭包表设计: https://www.slideshare.net/billkarwin/models-for-hierarchical-data
