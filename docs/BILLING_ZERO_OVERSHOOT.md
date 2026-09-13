# 后付费零超发护栏：运维手册

本文覆盖"余额模式后付费计费"的四层准入/结算护栏，说明**每个配置项的默认值、如何判断
护栏是否真的在跑、以及升级时的注意事项**。适用代码：`backend/internal/service/billing_cache_service.go`、
`backend/internal/service/gateway_request_spend_estimate.go`、`backend/internal/repository/billing_cache.go`、
`backend/internal/repository/usage_billing_repo.go`。

## 1. 问题背景

网关是**后付费**：请求先转发到上游（成本已经发生），结算时才从钱包扣钱。因此任何
"先放行、结算扣不动"的请求都直接变成坏账。历史上出现过三类表现：

1. 余额被扣到封底线后请求**全部继续成功返回**，`usage_log.actual_cost = 0`，日志只记
   `INSUFFICIENT_BALANCE` 但不阻断（"白嫖"）。
2. 余额贴近封底线时，**最后一笔付不满的请求**先被转发，结算只能扣到封底线，
   差额写成坏账（线上实证 `usage_log` 应收 0.19184 / 实收 0.099472）。
3. 并发突发时 N 个请求读到**同一份余额快照**，每个都判定"够付自己这一笔"而全部放行，
   结算时只有前几笔扣得动（坏账）——即使刚刚用 DB 真值复核过。

另外还有一个可用性问题：余额贴近封底时每笔请求都要 DB 复核，150–200 并发会把
PostgreSQL 连接槽打满，复核失败后 fail-closed 成大范围 **503**。

## 2. 四层护栏（按请求生命周期）

| 层 | 位置 | 作用 | 关键代码 |
| --- | --- | --- | --- |
| ① 准入阈值 + DB 真值复核带 | 转发前 | `balance <= reserve` 直接 403；`balance <= reserve + band` 时用 DB 真值复核；结算判定耗尽后写"钱包已耗尽"标记，预检命中即 403 | `checkBalanceEligibility` |
| ② 最坏费用闸门 | 转发前 | 要求 `balance >= reserve + 本次最坏费用`，付不满的请求**不转发**，不产生上游成本 | `estimateRequestSpendUpperBound` |
| ③ 在途预留（并发原子） | 转发前 | Redis 原子累加"已放行未结算"的最坏费用，要求 `balance - Σ预留 >= reserve`，使准入成为跨请求的原子操作 | `reserveSpendWithGuard` |
| ④ 结算封底 + 差额记账 | 结算 | `FOR UPDATE` 锁行后 `GREATEST(balance - amount, floor)`，差额记 write-off 并打耗尽标记；**永不为负** | `deductBalanceToFloorSQL` |

核心不变量：`balance − Σ在途预留 ≥ reserve`，且它在"结算扣钱"与"归还预留"以任意
先后顺序发生时都成立。

**订阅模式**（订阅计费）复用第 ③ 层：scope 为 `用户 × 分组`，护栏是
`usage + Σ预留 < daily/weekly/monthly limit`，堵住"并发请求共用同一份用量快照、
配额被超额消耗"的超卖。

## 3. 配置项清单

全部位于 `billing:` 段。

| 配置 | 默认 | 说明 |
| --- | --- | --- |
| `minimum_balance_reserve` | `0.1` | 钱包封底线（不可花）。预检 `balance <= 本值` 即 403。设 0 表示不保留。 |
| `balance_recheck_band` | `1.0` | DB 真值复核带宽（USD）。余额 `<= reserve + band`（且至少 `2*reserve`）时回源核对。**带宽越大越安全，但贴底用户的每笔请求都会多一次 DB 往返**。 |
| `request_spend_precheck_disabled` | `false` | 反向命名，零值安全。`true` 关闭第 ② 层（不建议）。 |
| `request_spend_default_max_output_tokens` | `8192` | 请求未声明输出上限时的缺省上界。 |
| `request_spend_safety_multiplier` | `1.0` | 预检安全系数。`>1` 更保守（多拦、零坏账），`<1` 更激进。 |
| `request_spend_min_output_tokens` | `8192` | **输出上界下限**。部分上游桥不执行请求声明的 `max_tokens`（实测声明 64/190 仍产出 999 token），信任小声明值会让预检低估。`>0` 时按 `max(声明值, 本值)` 预检。 |
| `database.max_open_conns` | `32` | 必须**显著低于** PG `max_connections`（默认 100）；多实例按实例数均摊。 |
| `database.max_idle_conns` | `8` | 建议为 `max_open_conns` 的 25%–50%。 |

### 关于 `request_spend_min_output_tokens` 的默认值

该值默认取 `8192`（与"未声明时的缺省上界"一致），语义是**不再信任任何小于 8192 的
声明值**。这是修掉"responses 桥不执行 max_tokens"这个已知坏账洞所需的取值。

- 影响范围有限：只有**余额贴近封底**的用户才可能因此被提前 403；余额充足的用户
  完全不受影响（`balance` 远大于最坏费用）。
- 要回到"信任声明值"的旧行为，显式配置 `request_spend_min_output_tokens: 0`，
  并接受 write-off 指标持续非零。

## 4. 如何验证护栏在跑

### 4.1 护栏健康端点

```
GET /api/v1/admin/ops/billing-guard
```

只读进程内原子计数，不查库、不依赖 ops 监控开关，因此故障期可直接 curl。返回：

```json
{
  "stats": { "...": "见下表" },
  "runtime": { "reservation_ttl_seconds": 600, "reject_reasons": ["..."] }
}
```

### 4.2 指标判读（按优先级）

**第一优先：护栏失效类。这些只要非零，就说明护栏在某些条件下"根本没拦"，而不是"拦得多"。**

| 字段 | 非零意味着 | 处置 |
| --- | --- | --- |
| `settlement_shortfall_count` | 仍在产生 write-off，预检口径有漏网 | 检查 `request_spend_min_output_tokens` / `default_max_output_tokens` 是否偏小 |
| `reservation_fail_open` | Redis 预留失败 → 并发护栏**静默失效** | 检查 Redis 可用性与延迟 |
| `reservation_release_error` | 预留归还失败 → 额度滞留到 TTL | 检查 Redis 写入与网络 |
| `reservation_renew_error` | 续期失败 → 超长请求可能失去保护 | 同上 |
| `reservation_abandoned` | 结算任务被 drop 语义丢弃 → 该笔未扣费 | 检查 usage_record worker 池容量与 drop 配置 |
| `recheck_skipped_no_user_repo` | 无法读 DB 真值（降级装配）→ 坏账窗口回到修复前 | 检查依赖注入装配 |
| `recheck_fail_closed` | 复核失败 fail-closed → 用户看到 503 | 检查 PG 连接池与 `max_open_conns` |
| `precheck_disabled` | 第 ② 层被配置关闭 | 确认是否为有意为之 |
| `precheck_unavailable` | 部分请求无法给出最坏费用上界（无定价/依赖缺失） | 检查模型定价配置 |

`stats.healthy == false` 且 `degraded_signals` 非空即为上述任一情形的汇总。

**第二优先：误伤程度。** 这些是**正常业务拒绝**，斜率上升不一定有问题，但要区分原因：

| 字段 | 含义 |
| --- | --- |
| `reject_below_reserve` / `reject_marker_active` / `reject_db_truth_below_reserve` | 用户**真的没钱**，正常拒绝 |
| `reject_worst_case_db_truth` / `reject_worst_case_cache_only` | 用户还有余额，但付不满"这一笔"的最坏费用 |
| `reject_inflight_reservation` | 已被同一用户的在途请求占满额度 |
| `worst_case_reject_ratio` | 上述"护栏类"拒绝占全部预检拒绝的比例。持续偏高 = 预检过保守（可能误伤），可下调 `request_spend_safety_multiplier` 或调低 reserve |

**第三优先：在途预留堆积。** `reservation_reserved` 与 `reservation_released` 的差值
趋势反映当前在途量；`reservation_expired_release` 增长说明有请求在预留 TTL 内未完成
结算（结算任务丢失，或心跳未生效）。

### 4.3 服务器实测口径（复测建议）

1. 把某测试用户余额打到贴近封底（如 0.2），用小额请求消耗到 0.1 附近。
2. 小额请求（最坏费用 < 可花余额）→ 应放行，`usage_log.actual_cost` 精确。
3. 大 `max_tokens` 请求（最坏费用 > 可花余额）→ 应**转发前** 403，`reject_worst_case_*` +1。
4. 150 并发打贴底用户 → 预期 `403 insufficient balance`，**0 个 503**；同时
   `settlement_shortfall_count` 不应增长。
5. 充值 +0.1 后 → 应立即恢复放行（耗尽标记在加款路径被清除）。

## 5. 升级注意事项

1. **`database.max_open_conns` 默认值从 256 降到 32**。若你的部署高吞吐、且未显式配置
   该值，升级后吞吐可能下降，需要显式调大并**同步调大 PG `max_connections`**；
   多实例部署按实例数均摊（例：3 实例建议 ≤ 30）。
2. **`request_spend_min_output_tokens` 默认从 0 改为 8192**（见第 3 节）。升级后贴底
   用户的大 `max_tokens` 请求会更早被 403 —— 这是零超发的预期代价。回退方式见上。
3. 新增 `billing.balance_recheck_band`（默认 1.0）。它使余额 ≤ 1.1 的用户每笔请求都
   回源查 DB；同一用户的并发复核已由 singleflight 合并为一次。若你的流量中"贴底小额
   高频"占比很高，可适当调小该带宽。
4. 配置变更后请用第 4.1 的端点确认 `runtime` 段与实际配置一致。

## 6. 已知边界与未覆盖入口

护栏当前覆盖以下入口：

**余额模式的 6 条文本链路**：`/v1/messages`（Anthropic 链 + OpenAI 链）、
`/v1/responses`（两条链）、`/v1/chat/completions`（两条链）。

**按次计费的 OpenAI 图片链路**：`/v1/images/generations`、`/v1/images/edits`
（OpenAI 平台 → `OpenAIGatewayHandler.Images`）。图片走**独立的最坏费用上界**
（`EstimateImageRequestSpendUpperBound`：同源于结算用的 `calculateOpenAIImageCost`，
按尺寸档 × 张数计；分组配置了图片单价时取**最贵档**作为上界），并与文本链路共用同一套
在途预留槽位。

`/images/batches` 走独立的 batch image hold 机制，已单独覆盖。

以下入口**尚未接入**上述护栏，仍可能出现"先服务后写坏账"：

- `/images/generations/async`、`/images/edits/async`、`/images/tasks/:id`
  （异步图片任务，`AsyncImage` handler）
- Grok 平台图片：`/v1/images/generations` 在 Grok 分组下走 `GrokImages`（`grok_media.go`）
- `/v1/live`、`/realtime/calls`（实时语音）、`/v1/realtime`、OpenAI Responses WebSocket
  长连接、Grok realtime audio WebSocket
- 视频生成（按秒计费，`CalculateVideoCost`：接入点与图片同构，需按"时长上限 × 数量"另做上界）

其他已知边界：

- **预留 TTL 自愈**：预留默认 TTL 10 分钟，长请求由心跳（TTL/3）续期；若结算任务
  丢失且心跳已停止，额度最多滞留 10 分钟。
- **凭据过期后的残留份额**：若某请求的预留凭据被 TTL 回收而聚合总额的自愈 TTL 尚未到，
  聚合键会短暂偏高（保守方向，不会放行超额，最多让该用户早一点被 403）。
- **保守方向的误伤**：输入 token 按 `非二进制字节数/2 + 图片固定折算` 估计，
  对超长单行文本（无空格、无换行）以及异常编码内容仍会高估，贴底用户可能被提前拦截。
- **`userRepo` 未装配的降级部署**：无法做 DB 真值复核，护栏退化为"只按缓存判断"，
  会通过 `recheck_skipped_no_user_repo` 计数并出现在 `degraded_signals` 中。
