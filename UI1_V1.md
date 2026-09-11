# UI1 v1

这是 UI1 的 v1 版本：前端保留 Scheme 3 console 界面，后端与构建链路已完全同步上游仓库。

- UI: Scheme 3 console interface
- Synced upstream: Sub2API `0.2.4`
- Upstream commit: `2d46dfaa5352d1b67d206082c75cf6129734b84e` (ziyue67/sub2api main)
- Style baseline: `bd1f46d35f6cde0557f52ea919edb22e7d6336db` (0.2.1, UI1 样式起点)
- Release branch: `ui1-v1`
- Docker image: `ghcr.io/ziyue67/sub2api-ui` (仅 GHCR，不使用 Docker Hub)
- Scope: frontend UI styling + GHCR-only release chain; the upstream backend is synced verbatim.

## 同步说明 (2026-09-11)

- 后端、部署脚本、CI 全部取自上游 `ziyue67/sub2api` main（`2d46dfaa5`，v0.2.4）。
- 前端为 UI1 v1 样式在 `bd1f46d35` 基点上的三方重放，保留 Scheme 3 视觉、布局与交互。
- Docker 发布仅走 GitHub Container Registry：

  | 工作流 | 触发 | 产物 |
  | --- | --- | --- |
  | `.github/workflows/ghcr.yml` | push `main` / 手动 | `ghcr.io/ziyue67/sub2api-ui:latest`、`:sha-*`、`:0.2.4`、`:0.2.4-amd64` |
  | `.github/workflows/release.yml` | tag `v*` / 手动 | GoReleaser 打包 + GHCR 镜像 |
  | `.github/workflows/sync-fork-release.yml` | GHCR 发布完成后 | 自动创建 `vX.Y.Z` Release |

- 仓库内已不含任何 Docker Hub 工作流（`dockerhub.yml`、`cleanup-dockerhub.yml`、`probe-dockerhub.yml`、`restore-ghcr-tags.yml` 等均已移除）。

## Docker 测试

```bash
docker pull ghcr.io/ziyue67/sub2api-ui:latest
docker run --rm -p 8080:8080 ghcr.io/ziyue67/sub2api-ui:latest
```

Compose 方式（`deploy/docker-compose.yml` 已指向 `ghcr.io/ziyue67/sub2api-ui:latest`）：

```bash
cd deploy && docker compose up -d
```

## Verification

From `frontend/`:

```text
pnpm typecheck
pnpm lint:check
pnpm build
```

The UI1 v1 acceptance covers authenticated user and administrator routes, feature-flagged navigation, model plaza and embedded model plaza, model-square pricing and filters, monitor-v2, dialogs, and desktop/mobile layout checks.
The UI1 v1 audit covers authenticated user and administrator routes in light and dark themes at desktop and mobile viewports. It also checks teleported dialogs, endpoint tooltips, announcement Markdown, page errors, legacy shell nodes, and horizontal overflow.