<div align="center">

<img src="assets/logo.svg" alt="Sub2API UI Logo" width="128" />

# Sub2API UI1

![UI1 v1](https://img.shields.io/badge/UI1-v1-2563EB.svg)
![Vue](https://img.shields.io/badge/Vue-3.4%2B-42B883.svg)
![Vite](https://img.shields.io/badge/Vite-5%2B-646CFF.svg)
![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-3%2B-06B6D4.svg)

**Sub2API 的 UI 专属版本，采用 Scheme 3 Console 界面。**

[UI1 v1 说明](UI1_V1.md) | [上游 Sub2API](https://github.com/ziyue67/sub2api)

</div>

## 项目定位

本仓库用于维护 Sub2API 的 UI1 v1 界面。它基于上游 Sub2API `0.1.176`，重点是前端视觉、布局、交互和路由兼容性，不是一个独立的后端分支。

UI 代码位于 [`frontend/`](frontend/)。后端、数据库和部署文件保留在仓库中，是为了保证上游构建链和版本兼容；本分支的维护范围以 UI 适配为主。

## UI 特性

- **Scheme 3 Console**：面向日常运营的控制台布局和导航结构。
- **响应式界面**：适配桌面和移动视口，减少窄屏下的横向溢出。
- **明暗主题**：覆盖用户端、管理端、弹窗和 Teleport 内容。
- **用户端页面**：仪表盘、用量、密钥、排行榜、模型广场、支付和个人设置等页面的统一视觉风格。
- **管理端页面**：账号、渠道、分组、用量、系统设置和运维页面的统一组件样式。
- **交互细节**：表格、筛选器、提示框、公告 Markdown、工具提示和加载/错误状态保持一致。

## 版本信息

| 项目 | 内容 |
| --- | --- |
| UI 版本 | UI1 v1 |
| UI 方案 | Scheme 3 Console |
| 上游基线 | Sub2API `0.1.176` |
| 发布分支 | `ui1-v1` |
| 发布标签 | `ui1-v1` |
| 前端目录 | `frontend/` |

## 本地开发

要求 Node.js 运行环境和 `pnpm`。

```bash
cd frontend
pnpm install
pnpm dev
```

开发服务器启动后，按终端显示的地址打开 UI。要连接真实 API，请将前端请求地址指向对应的 Sub2API 后端实例。

## 验证命令

在 `frontend/` 目录执行：

```bash
pnpm typecheck
pnpm lint:check
pnpm build
```

UI1 v1 的专项检查：

```bash
pnpm test:run -- \
  src/components/common/__tests__/Select.spec.ts \
  src/components/common/__tests__/DataTable.spec.ts \
  src/features/channel-monitor-v2/__tests__/designSystem.structure.spec.ts \
  src/features/channel-monitor-v2/__tests__/Scheme3V2Toggle.spec.ts
```

专项审计覆盖已登录的用户端和管理端路由、明暗主题、桌面和移动视口、Teleport 弹窗、端点工具提示、公告 Markdown、页面错误状态、旧版壳节点以及横向溢出。

## 相关文件

- [`UI1_V1.md`](UI1_V1.md)：UI1 v1 的版本范围和审计记录。
- [`frontend/`](frontend/)：前端源代码、组件、路由和样式。
- [`assets/`](assets/)：UI 使用的品牌和静态资源。
- [上游 Sub2API](https://github.com/ziyue67/sub2api)：后端平台和上游项目。

## 说明

本仓库的 UI 改动应优先保持与上游 API 契约兼容。涉及后端接口、数据库迁移或生产部署的变更，请回到上游项目或对应的后端维护分支处理。
