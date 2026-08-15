# UI1 v1

这是 UI1 的 v1 版本。

- UI: Scheme 3 console interface
- Base: upstream Sub2API `0.1.176`
- Release branch: `ui1-v1`
- Release tag: `ui1-v1`
- Scope: frontend UI adaptation and compatibility styling; the repository keeps the upstream backend and build history intact.

## Verification

From `frontend/`:

```text
pnpm typecheck
pnpm lint:check
pnpm build
pnpm test:run -- src/components/common/__tests__/Select.spec.ts src/components/common/__tests__/DataTable.spec.ts src/features/channel-monitor-v2/__tests__/designSystem.structure.spec.ts src/features/channel-monitor-v2/__tests__/Scheme3V2Toggle.spec.ts
```

The UI1 v1 audit covers authenticated user and administrator routes in light and dark themes at desktop and mobile viewports. It also checks teleported dialogs, endpoint tooltips, announcement Markdown, page errors, legacy shell nodes, and horizontal overflow.
