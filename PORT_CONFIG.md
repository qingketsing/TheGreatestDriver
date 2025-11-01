# 端口配置更新说明

## ✅ 配置已更新

前端和后端的端口配置已按您的要求修改完成！

## 🔧 新的端口配置

### 后端端口
- **端口**: 8000
- **访问地址**: http://localhost:8000
- **配置文件**: `cmd/server/main.go`

### 前端端口
- **端口**: 12000
- **访问地址**: http://localhost:12000
- **配置文件**: `frontend/vite.config.ts`

## 📝 修改的文件清单

### 核心配置文件
1. ✅ `frontend/vite.config.ts`
   - 前端端口: 3000 → **12000**
   - API 代理目标: http://localhost:8080 → **http://localhost:8000**

2. ✅ `frontend/start.ps1`
   - 更新启动提示信息

### 文档更新
3. ✅ `frontend/README.md`
   - 更新端口说明
   - 更新 API 配置示例

4. ✅ `frontend/SUMMARY.md`
   - 更新访问地址
   - 更新配置说明
   - 更新 API 代理示例

5. ✅ `README.md`
   - 更新完整启动流程
   - 更新访问地址
   - 更新 API 配置

6. ✅ `QUICKSTART.md`
   - 更新启动说明
   - 更新验证步骤
   - 更新 API 测试命令
   - 更新故障排查

## 🚀 如何启动

### 1. 启动后端（端口 8000）

```bash
cd cmd/server
go run main.go
```

看到以下输出表示成功：
```
[GIN-debug] Listening and serving HTTP on :8000
```

访问: http://localhost:8000

### 2. 启动前端（端口 12000）

```powershell
cd frontend
.\start.ps1
```

或者：
```powershell
npm run dev
```

看到以下输出表示成功：
```
➜  Local:   http://localhost:12000/
```

访问: http://localhost:12000

## 🔍 验证配置

### 测试后端
```bash
# 浏览器访问
http://localhost:8000

# 或使用 curl
curl http://localhost:8000
```

应该看到: `Hello! This is the Single Drive server.`

### 测试前端
浏览器访问: http://localhost:12000

应该看到 Single Drive 的用户界面

### 测试前后端连接
1. 打开浏览器访问 http://localhost:12000
2. 尝试创建文件夹或上传文件
3. 检查浏览器开发者工具的 Network 标签
4. 请求应该发送到 `/api/xxx`，实际代理到 `http://localhost:8000/xxx`

## ⚙️ 配置详情

### Vite 配置 (frontend/vite.config.ts)

```typescript
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 12000,  // ← 前端端口
    proxy: {
      '/api': {
        target: 'http://localhost:8000',  // ← 后端地址
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
})
```

### API 请求流程

```
前端发起请求
    ↓
http://localhost:12000/api/list
    ↓
Vite 代理拦截 /api 前缀
    ↓
转发到后端
    ↓
http://localhost:8000/list
    ↓
后端处理并返回
    ↓
前端接收响应
```

## 📌 注意事项

1. **修改端口后需要重启**
   - 修改配置后需要重启前端开发服务器
   - 后端修改端口也需要重启

2. **防火墙设置**
   - 确保防火墙允许 8000 和 12000 端口
   - Windows 防火墙可能会提示授权

3. **端口冲突**
   - 如果端口被占用，会提示错误
   - 可以在配置文件中修改为其他端口

4. **浏览器缓存**
   - 如果界面没有更新，清除浏览器缓存
   - 或使用隐私/无痕模式

## 🎯 快速命令

```bash
# 检查端口占用 (Windows)
netstat -ano | findstr :8000
netstat -ano | findstr :12000

# 检查端口占用 (Linux/Mac)
lsof -i :8000
lsof -i :12000

# 测试后端 API
curl http://localhost:8000/list

# 测试创建文件夹
curl -X POST "http://localhost:8000/createdir?path=test"
```

## ✅ 配置完成

所有文件已更新，新的端口配置为：
- 🔵 **后端**: http://localhost:8000
- 🟢 **前端**: http://localhost:12000

现在可以按照新的端口启动应用了！

---

**更新日期**: 2025年10月31日
