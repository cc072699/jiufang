# CLAUDE.md

> ERP 对话式查询助手前端 — 项目指南

## 项目概述

基于自然语言的企业 ERP 数据查询系统前端，使用 React + Ant Design 构建，后端由 Go + Gin 开发（前后端分离）。

- **仓库**: https://github.com/cc072699/jiufang
- **设计文档**: `/Users/macmima1234/Desktop/codex/久方设计文档/` (prd.md, detailed-design.md, architecture-overview.md, 前端开发方案与ClaudeCode提示词.md)
- **参考实现 (Vue 版)**: `/Users/macmima1234/Desktop/claude code/久方前端test1/erp-query-frontend/`

## 技术栈

- Vite 8 + React 19 + TypeScript 6
- Ant Design 6 + @ant-design/icons
- React Router 7 (createBrowserRouter)
- Axios (API Client, JWT Token 注入, 统一错误处理)
- TanStack Query 5 (服务端状态)
- Zustand 5 (本地状态: 登录态)
- ECharts 6 + echarts-for-react (图表)
- MSW 2 (Mock Service Worker)
- Vitest + Playwright (测试, 待配置)

## 运行命令

```bash
# 安装依赖
npm install

# 启动开发服务器 (含 MSW Mock, 端口 5173)
npm run dev

# TypeScript 类型检查
npx tsc --noEmit

# 生产构建
npm run build

# ESLint 检查
npm run lint
```

## 测试账号

| 账号 | 密码 | 角色 | 权限 |
|------|------|------|------|
| admin | admin123 | 管理员 | 全部功能 |
| user | user123 | 高管 | 工作台/查询/历史/收藏/个人中心 |
| disabled_user | pass123 | 已停用 | 无法登录 (BR-069 测试) |

## 项目结构

```
src/
  app/
    router.tsx          # 路由配置 (15 个页面路由)
    providers.tsx       # QueryClient + AntD + Auth 错误回调
    layout/
      AppLayout.tsx     # 侧边导航 + 顶部栏 + 内容区 (按角色裁剪菜单)
  shared/
    api/
      client.ts         # Axios 封装 (Token 注入, 401/403 处理, 错误码映射)
      types.ts          # 全部 API 类型定义 (28 个接口) + 错误码
      index.ts          # API 函数封装 (28 个)
    auth/
      token.ts          # localStorage Token/User 存取
      store.ts          # Zustand 认证状态
      guards.tsx        # AuthGuard / GuestGuard / AdminGuard
    ui/
      DataTable.tsx     # 查询结果表格 (空值展示规则)
      ChartRenderer.tsx # ECharts 图表 + 表格切换
      EmptyState.tsx    # 空状态
      ConfirmAction.tsx # 二次确认弹窗
    utils/
      encrypt.ts        # AES 密码加密适配层 (待后端提供 key/iv)
      nullDisplay.ts    # 空值展示规则
  pages/                # 15 个页面
  mocks/
    browser.ts          # MSW Service Worker 配置
    handlers.ts         # MSW Mock 处理器 (22 个 API)
  main.tsx              # 应用入口 (MSW 初始化 + React 挂载)
```

## 已实现功能 (四个阶段全部完成)

### 第一阶段: 项目骨架与联调底座
- [x] 项目初始化 (Vite + React + TS + AntD)
- [x] 路由系统 (15 个路由, 含 push-records)
- [x] 路由守卫 (AuthGuard / GuestGuard / AdminGuard + 403 提示)
- [x] API Client (Axios 拦截器, JWT 注入, {code,message,data} 统一处理, 401 跳登录, 403 提示)
- [x] BR-069 登录错误码细分 (40103 账号不存在 / 40104 已停用 / 40105 密码错误)
- [x] 登录页 (表单, 错误提示, AES 加密预留)
- [x] AppLayout (侧边导航, 角色菜单裁剪, 用户下拉菜单)

### 第二阶段: 核心查询闭环
- [x] 对话查询页 (左侧会话栏, 多轮追问, session_id, 输入校验 1-500 字, 失败重试, 联想问题)
- [x] 历史记录页 (分页, 关键词/状态筛选, 详情弹窗, 续接追问, 删除确认)
- [x] 收藏管理页 (分页列表, 一键查询, 删除确认)

### 第三阶段: 管理员后台
- [x] 人员管理 (列表, 分页, 搜索, 角色筛选, 新增/编辑/删除)
- [x] 权限管理 (用户组列表, 表级/字段级权限配置, 二次确认)
- [x] 操作日志 (分页列表, 操作类型筛选, JSON 详情格式化)

### 第四阶段: 报告、预警与体验补齐
- [x] 定时报告 (列表, 新建/编辑表单, 启停, 删除确认)
- [x] 预警规则 (列表, 新建/编辑表单, 启停, 删除确认)
- [x] 推送记录 (分页列表, 类型筛选)
- [x] 个人中心 (资料展示, 退出登录, 密码修改 UI 预留)

## MSW Mock 覆盖 (22/28 API)

已 mock: login, logout, query, history(CRUD), favorites(CRUD), users(CRUD), groups(CRUD), permissions, logs, reports(CRUD), alerts(CRUD), push-records

## 接口约定

- 所有业务 API 路径: `/api/v1/`
- 响应格式: `{ code: number, message: string, data?: T }`
- 登录: `POST /api/v1/auth/login` (无需 Token)
- 登出: `POST /api/v1/auth/logout`
- 除登录外请求头: `Authorization: Bearer {token}`
- 分页参数: `page` + `page_size`, 默认 20, 支持 20/50/100

## 切换真实后端

修改环境变量即可切换:

```bash
# .env.development
VITE_API_BASE_URL=http://your-backend-host:8080/api/v1
```

或关闭 MSW (在 main.tsx 中移除 DEV 环境的 MSW 初始化)。

## 待后端确认接口清单 (10 项)

| PRD 能力 | 缺口 |
|----------|------|
| 数据导出 Excel/PDF | 缺导出接口、异步进度、下载地址 |
| 反馈机制 | 缺满意/不满意反馈提交接口 |
| 首次登录强制改密 | 登录响应缺 must_change_password 字段, 缺改密接口 |
| 修改密码 | 缺修改密码接口 |
| 头像上传 | 缺头像上传接口、头像 URL 字段 |
| 人员管理 | 缺企业微信推送地址、重置密码、启用/停用专用接口 |
| 报告/预警编辑 | 缺详情接口, 列表字段不足以回填编辑表单 |
| 批量删除历史 | PRD 要求最多 50 条批量删除, API 只定义单条删除 |
| 用户组成员管理 | 缺成员搜索、添加、移除接口 |
| AES 密码传输 | PRD 要求 AES 加密, 未提供 key/iv/mode/padding |
