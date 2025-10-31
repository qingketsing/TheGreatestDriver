# Single Drive Frontend

这是 Single Drive 项目的前端应用，基于 React + TypeScript + Vite + Ant Design 构建。

## 功能特性

- 📁 **文件管理** - 支持文件和文件夹的浏览、上传、下载、删除、重命名
- 🌲 **树形结构** - 支持多层级文件夹导航
- 📤 **拖拽上传** - 支持拖拽文件上传，带进度显示
- 📊 **多种视图** - 列表视图和网格视图切换
- 🔍 **搜索过滤** - 支持文件搜索和类型过滤
- 📱 **响应式设计** - 适配各种屏幕尺寸
- 🎨 **现代化UI** - 基于 Ant Design 的美观界面

## 技术栈

- **React 18** - UI 框架
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **Ant Design 5** - UI 组件库
- **React Router** - 路由管理
- **Axios** - HTTP 请求

## 开始使用

### 安装依赖

```bash
npm install
```

### 开发模式

```bash
npm run dev
```

应用将在 http://localhost:12000 启动

### 构建生产版本

```bash
npm run build
```

### 预览生产版本

```bash
npm run preview
```

## 项目结构

```
src/
├── components/          # 组件
│   ├── Layout/         # 布局组件
│   ├── FileList/       # 文件列表相关组件
│   └── Upload/         # 上传相关组件
├── pages/              # 页面
├── services/           # API 服务
├── types/              # TypeScript 类型定义
├── utils/              # 工具函数
├── App.tsx             # 主应用组件
├── main.tsx            # 入口文件
└── index.css           # 全局样式
```

## API 配置

前端通过代理访问后端 API，配置在 `vite.config.ts` 中：

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

确保后端服务运行在 `http://localhost:8000`

## 主要功能

### 文件浏览
- 支持列表和网格两种视图模式
- 面包屑导航，快速切换目录
- 文件类型图标和标签

### 文件上传
- 拖拽上传支持
- 批量上传文件
- 实时显示上传进度

### 文件操作
- 下载文件/文件夹（文件夹自动打包为 zip）
- 删除文件/文件夹
- 重命名
- 创建新文件夹

### 批量操作
- 多选文件
- 批量下载
- 批量删除

## License

MIT
