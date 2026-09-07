# UI1 v1

这是 UI1 的 v1 版本，基于 Sub2API 上游魔改版本适配。

- UI: Scheme 3 console interface
- Compatible upstream: Sub2API `0.2.1`
- Upstream commit: `bd1f46d35f6cde0557f52ea919edb22e7d6336db`
- Release branch: `ui1-v1`
- Scope: frontend UI adaptation and compatibility styling; the upstream backend and repository build history remain unchanged.

## Verification

From `frontend/`:

```text
pnpm typecheck
pnpm lint:check
pnpm build
```

The UI1 v1 acceptance covers authenticated user and administrator routes, feature-flagged navigation, model plaza and embedded model plaza, model-square pricing and filters, monitor-v2, dialogs, and desktop/mobile layout checks.
The UI1 v1 audit covers authenticated user and administrator routes in light and dark themes at desktop and mobile viewports. It also checks teleported dialogs, endpoint tooltips, announcement Markdown, page errors, legacy shell nodes, and horizontal overflow.
