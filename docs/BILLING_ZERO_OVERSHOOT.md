# 后付费零超发护栏：运维手册

本文覆盖"余额模式后付费计费"的四层准入/结算护栏，说明**每个配置项的默认值、如何判断
护栏是否真的在跑、以及升级时的注意事项**。适用代码：`backend/internal/service/billing_cache_service.go`、
`backend/internal/service/gateway_request_spend_estimate.go`、`backend/internal/repository/billing_cache.go`、
`backend/internal/repository/usage_billing_repo.go`、`backend/internal/repository/billing_reservation_db.go`、
`backend/migrations/238_billing_balance_reservations.sql`。

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
| ③ 在途预留（并发原子） | 转发前 | 原子累加"已放行未结算"的最坏费用，要求 `balance - Σ预留 >= reserve`，使准入成为跨请求的原子操作。默认走 Redis（Lua 比较并累加）；Redis 不可用时**同一条不变量**改由 PostgreSQL 账本守住（见 2.1） | `reserveSpendWithGuard` |
| ④ 结算封底 + 差额记账 | 结算 | `FOR UPDATE` 锁行后 `GREATEST(balance - amount, floor)`，差额记 write-off 并打耗尽标记；**永不为负** | `deductBalanceToFloorSQL` |

核心不变量：`balance − Σ在途预留 ≥ reserve`，且它在"结算扣钱"与"归还预留"以任意
先后顺序发生时都成立。

**订阅模式**（订阅计费）复用第 ③ 层：scope 为 `用户 × 分组`，护栏是
`usage + Σ预留 < daily/weekly/monthly limit`，堵住"并发请求共用同一份用量快照、
配额被超额消耗"的超卖。

### 2.1 预留账本：Redis 热路径 + DB 兜底 + 余额缓存栅栏

在途预留有**两本账**，任一时刻只使用其中一本（参考 new-api 的 `TryReserveUserQuota`
"比较并扣减"与 `cacheInitToken` 栅栏语义实现）：

| 账本 | 何时使用 | 实现 | 关键不变量 |
| --- | --- | --- | --- |
| Redis（热路径） | 默认 | Lua 原子"读合计 → 比较 → 累加"（`TryReserveUserBalance`） | 与旧行为一致，单次往返 |
| PostgreSQL（兜底） | 共享兜底窗口生效期间 | 事务内 `pg_advisory_xact_lock(hashtextextended(scope, 0))` → 删过期行 → `SUM` 活跃金额 → 条件 `INSERT ... ON CONFLICT DO NOTHING`；拒绝即回滚 | 同一条 `balance − Σ预留 ≥ reserve`，跨实例仍是原子操作 |

切换协议（`billing_reservation_fallback_state` 单行表，迁移 `238`）：

1. 任一实例的 Redis 预留报错 → 抬起共享兜底窗口（`fallback_until = GREATEST(现值, now + 预留 TTL)`），并把本笔改走 DB 账本。
2. 所有实例的每笔预检先探测该窗口（进程内缓存 1s，避免每笔都查 DB）：窗口生效期间**统一**走 DB 账本，**绝不与 Redis 账本混用** —— 同一份额度被两本账各算一遍会变成重复计算（方向保守，但会误伤）。
3. 窗口不早于"切换时刻 + 预留 TTL"，即至少 10 分钟：切换前已存在的 Redis 预留最迟在窗口结束时全部过期，之后重新信任 Redis 才是安全的。
4. Redis 与 DB 兜底**同时**不可用时 fail-closed：返回 `ErrBillingServiceUnavailable`（HTTP 503），而不是退回"无预留放行"（审计 R2）。

余额缓存侧同样有一道栅栏（参考 new-api 的 `invalidateTokenCacheForMutation`，审计 R9）：

- 扣费/加款落库后 `InvalidateUserBalance` 用一条 Lua **原子地"抬栅栏 + 删缓存"**；栅栏 TTL 10s，覆盖"DB 读 → 缓存写回"的整段间隙。
- 回源写缓存（`InitUserBalance`）只在**键缺失且无栅栏**时建立；键已存在时只刷新 TTL、不覆盖。
- "以 DB 真值覆写"（`SetUserBalanceFenced`）同样在栅栏存在时拒绝写入 —— 否则并发扣费的结果会被扣费前读到的旧快照覆盖（缓存"复活"，预检按虚高余额放行）。

## 3. 配置项清单

全部位于 `billing:` 段。

| 配置 | 默认 | 说明 |
| --- | --- | --- |
| `minimum_balance_reserve` | `0.1` | 钱包封底线（不可花）。预检 `balance <= 本值` 即 403。设 0 表示不保留。 |
| `balance_recheck_band` | `1.0` | DB 真值复核带宽（USD）。余额 `<= reserve + band`（且至少 `2*reserve`）时回源核对。**带宽越大越安全，但贴底用户的每笔请求都会多一次 DB 往返**。 |
| `request_spend_precheck_disabled` | `false` | 反向命名，零值安全。`true` 关闭第 ② 层（不建议）。 |
| `request_spend_default_max_output_tokens` | `8192` | 请求未声明输出上限时的缺省上界。 |
| `request_spend_safety_multiplier` | `1.0` | 预检安全系数。`>1` 更保守（多拦、零坏账），`<1` 更激进。 |
| `request_spend_min_output_tokens` | `0` | **输出上界下限（默认关）**。部分上游桥不执行请求声明的 `max_tokens`（实测声明 64/190 仍产出 999 token），信任小声明值会让预检低估。`>0` 时按 `max(声明值, 本值)` 预检。**只抬高偏小的声明值，对大声明值逐位不变**。 |
| `request_spend_cjk_tokens_per_rune` | `0`（= 内置 `1`） | CJK 每字 token 上界，**校验范围 0–8**（超出即启动失败；运行期再钳到 8，防止误配 1e6 让 CJK 请求全量 403、或极大值让乘法溢出把闸门变成 fail-open）。内置 `1` 是按生产上游（deepseek）实测 0.5–1 token/字 校准的**已校准上游**上界，对**字节回退型分词器不是上界**（cl100k/o200k 生僻汉字、Llama 系 CJK 词表可达 2–3 token/字）。路由到这类模型（GPT 系 / Llama 系）的部署应设为 `2`（或 `3`）。 |
| `inflight_reservation_budget_multiplier` | `1.0` | 在途预留聚合闸门的预算倍数：`1.0` = 严格（`balance − Σ预留 ≥ reserve`，并发下零坏账，但并发量被"可花余额 / 单笔最坏费用"卡住）；调高即用**有界坏账**换并发（`settlement_shortfall_count` 会随之上长）。**可被后台 `/admin/settings` 的同名项覆盖，后台值优先，保存后立即生效**。 |
| `database.max_open_conns` | `32` | 必须**显著低于** PG `max_connections`（默认 100）；多实例按实例数均摊。 |
| `database.max_idle_conns` | `8` | 建议为 `max_open_conns` 的 25%–50%。 |

### 3.1 DB 兜底相关的固定参数（不可配置）

| 参数 | 值 | 说明 |
| --- | --- | --- |
| 预留 TTL | `10m` | Redis / DB 两本账共用；长请求由心跳按 TTL/3 续期。 |
| 兜底窗口长度 | `≥ 10m` | 每次激活/续期都取 `max(现值, now + 10m)`，保证切换前的 Redis 预留全部过期。 |
| 兜底窗口探测 | 间隔 `1s` / 超时 `500ms` | 进程内缓存探测结果；探测失败不缓存、按未激活处理（保留 Redis 热路径可用性）。 |
| DB 兜底单次超时 | `3s` | 只作用于 Redis 故障后的降级路径，不进正常热路径。 |

### 关于 `request_spend_min_output_tokens`

默认 `0` = **信任请求声明的值**（与历史行为一致，升级零行为变化）。

- **作用范围（先读这个）**：本下限只"抬高偏小的声明值"。声明 `max_tokens: 64` 且下限为
  8192 时按 8192 估；声明 `max_tokens: 64000` 时**逐位不变**、完全不受影响。因此它
  **不会**让"大 max_tokens 请求更早被 403" —— 那来自最坏费用闸门按声明上限计价，
  属于零超发的既有取舍，与本下限无关。
- **何时打开**：观察 ops 端点的 `settlement_shortfall_count`，斜率 > 0（仍在产生
  write-off，典型是 responses 桥不执行 max_tokens）时把它配成 `8192`（= 未声明时的
  缺省上界）；恒为 0 则保持 0 即可。
- 打开后影响范围有限：只有**余额贴近封底**的用户才可能因此被提前 403；余额充足的
  用户完全不受影响（`balance` 远大于最坏费用）。

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
| `reservation_fail_open` | Redis 预留失败且**装配中没有 DB 兜底能力** → 并发护栏静默失效（仅轻量/降级装配才会出现） | 检查 Redis 可用性；确认正式部署已挂 `userRepo` 的 DB 兜底 |
| `reservation_redis_failure` | Redis 预留失败，已切到 DB 兜底账本（护栏仍生效，但预留走 DB） | 检查 Redis 可用性；持续增长说明故障未恢复 |
| `reservation_fail_closed` | Redis 与 DB 兜底**同时**不可用 → 预检 503 | 检查 Redis 与 PG 连接、迁移 238 是否执行 |
| `reservation_fallback_activated` | 本实例抬起共享 DB 兜底窗口（首次降级信号） | 与 `reservation_redis_failure` 对照；持续增长说明 Redis 长期不可用 |
| `reservation_fallback_activate_error` / `reservation_fallback_probe_error` | 兜底窗口抬升/探测失败 → 跨实例账本一致性变弱 | 检查 DB 连接与迁移 238 |
| `balance_init_fenced` / `balance_write_fenced` | 余额缓存回源写被变更栅栏拦截（说明栅栏在起作用，不是故障） | 持续增长说明该用户余额写频繁，属正常并发/扣费节奏 |
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
2. **`request_spend_min_output_tokens` 保持默认 0**（不改变现有部署行为，见第 3 节）。
   它是按 write-off 指标决定是否打开的"修补开关"，不是本次升级的强制变更。
3. 新增 `billing.balance_recheck_band`（默认 1.0）。它使余额 ≤ 1.1 的用户每笔请求都
   回源查 DB；同一用户的并发复核已由 singleflight 合并为一次。若你的流量中"贴底小额
   高频"占比很高，可适当调小该带宽。
4. **`minimum_balance_reserve` 默认值从 `0.000001` 变为 `0.1`**：钱包保留金从"名义上
   保留"变成"真实不可花 $0.10"，余额 ≤ 0.1 的用户会在转发前 403。配置文件里显式写了
   旧默认值 `0.000001` 的部署会被自动迁移为 `0.1`（`config.go` 的 legacy 迁移，仅匹配
   该精确值并打 WARN 日志）；想保持可花完余额请显式配 `0`。
5. 配置里的 `billing.minimum_balance_reserve` / `balance_recheck_band` /
   `inflight_reservation_budget_multiplier` / `request_spend_safety_multiplier`
   现在会拒绝 `NaN` / `±Inf`（`NaN` 曾能绕过所有 `< 0` 判断并让封底、最坏费用闸门与
   DB 复核静默失效，`+Inf` 会让预留闸门永久放开）。
6. 配置变更后请用第 4.1 的端点确认 `runtime` 段与实际配置一致。
7. **必须执行迁移 `238_billing_balance_reservations.sql`**：DB 兜底账本依赖两张新表
   `billing_reservations`（在途凭据）与 `billing_reservation_fallback_state`（共享兜底窗口）。
   表缺失时 Redis 故障期的 DB 兜底会报错并 fail-closed（503），而不是静默放行；
   `degraded_signals` 会同步提示“检查迁移 238”。

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

- **Gemini 文本生成**（`/v1beta` → `gemini_v1beta_handler.go`，走 `RecordUsage` 按余额计费）
- **OpenAI embeddings**（`/v1/embeddings` → `openai_embeddings.go`，同样按余额计费）
- `/images/generations/async`、`/images/edits/async`、`/images/tasks/:id`
  （异步图片任务，`AsyncImage` handler）
- Grok 平台图片：`/v1/images/generations` 在 Grok 分组下走 `GrokImages`（`grok_media.go`）。
  该入口的主链路只做基础阈值检查，**没有最坏费用闸门、也不建立在途预留**；只有"选号失败后
  切换分组路由"这条支路在 PR#16 后补上了按新分组重估的最坏费用闸门
- `/v1/live`、`/realtime/calls`（实时语音）、`/v1/realtime`、OpenAI Responses WebSocket
  长连接、Grok realtime audio WebSocket
- 视频生成（按秒计费，`CalculateVideoCost`：接入点与图片同构，需按"时长上限 × 数量"另做上界）

> 已从本清单移出：**Anthropic 提示过长后的兜底分组重试**（`gateway_handler.go` 的 fallback
> group 分支）。PR#16 起该支路会按兜底分组重新估算最坏费用并**替换**旧分组的在途预留
> （`recheckSelectedGroupRouteEligibility`），不再只做基础阈值检查。

其他已知边界：

- **Redis 不可用 ⇒ 切 DB 兜底，而不是 fail-open**：`TryReserveUserBalance` 报错时先抬起共享
  兜底窗口（`reservation_redis_failure` / `reservation_fallback_activated`），再用 PostgreSQL 账本
  继续守同一条不变量；只有 Redis 与 DB 兜底**同时**不可用时才 fail-closed（`reservation_fail_closed`，
  HTTP 503）。仍然 fail-open 的只剩“装配中没有 DB 兜底能力”的降级部署（轻量/测试装配），
  见 `reservation_fail_open` 指标。
- **账本切换的残留边界（已知）**：切换到 DB 账本时，Redis 上**切换前已存在**的预留不会被搬进
  DB 账本（DB 只统计自己的行），因此切换前那批请求靠“共享窗口 ≥ 预留 TTL”兜住：窗口结束前
  不重新信任 Redis，等它们全部自然过期。影响方向是保守的（期间可能多拦几笔），不会放行超额。
- **DB 兜底账本不是扣费**：`billing_reservations` 只是准入用的在途账本，结算仍只写钱包
  （`users.balance`）；凭据行按 `(scope, request_id)` 唯一，过期行由 TTL + 事务内清理回收。
- **预留后端返回无法解释的状态**：`tryReserveBalanceScript` 只认 `'1'`/`'0'` 两种答复，
  其它值一律按**拒绝**处理（交给调用方的 guard 再判一次），不走 fail-open —— 避免"状态不一致"
  被误当作"后端故障"从而在完全没有预留的情况下放行。
- **预留 TTL 自愈**：预留默认 TTL 10 分钟，长请求由心跳（TTL/3）续期；若结算任务
  丢失且心跳已停止，额度最多滞留 10 分钟。
- **凭据过期后的残留份额**：若某请求的预留凭据被 TTL 回收而聚合总额的自愈 TTL 尚未到，
  聚合键会短暂偏高（保守方向，不会放行超额，最多让该用户早一点被 403）。
- **保守方向的误伤**：输入 token 按**四段口径**估计 —— 普通文本 `字节数/2`；
  **CJK 主导文本**（非结构字节中 CJK 占比 ≥ 90%）按 `每字 1 token`
  （`request_spend_cjk_tokens_per_rune` 可调高，上限 8；该费率按生产上游校准，对
  字节回退型分词器不是上界）；**多模态块内**的 base64（`url`/`image_url`/`data`/
  `file_data` 字段承载，含 OpenAI Responses 的字符串形态 `image_url` 与**带 mime
  参数的 data URI**（`data:image/svg+xml;charset=utf-8;base64,…`），即真实图片/音频
  负载）按固定 allowance（1600/块，与分辨率无关）。其中 `data` 是极常见的通用键名
  （任何自定义 JSON 都可能用它表示数据），因此还要由**父键**确认是否为多模态：
  只有 `source`（Anthropic）/ `inline_data`·`inlineData`（Gemini）/ `input_audio`
  （OpenAI）才算多模态块；父键明确但不是上述之一（如 `{"payload":{"data":"…"}}`）
  时回落稠密/文本口径，父键无法确证（顶层 `data`、数组元素）时保持按多模态折算，
  以免把真实图片降级为稠密口径而误 403。
  `url` / `file_data` 同属通用键名（`{"image":{"url":…}}` 是媒体，`{"payload":{"url":…}}`
  就可能只是长文本），只有父键命中媒体白名单（`image_url`/`input_image`/`input_audio`/
  `image`/`audio`、`input_file`/`file`/`document`）或值前带 `;base64,` 标记才算多模态块；
  **落在灰区时口径取 `max(稠密, 1 块固定额度)`**：按字节折算（1 token/字节）高于块额度时
  用稠密，否则用块额度。两种口径在原始约 1.6KB 处相交，只取其中之一必然在某一侧低估 ——
  这正是审计 R1 实测到的交叉点（固定额度与按字节折算在 512B–1.6KB 一侧相差最多 2.3 倍）。
  **其它位置的 base64**（含没有
  任何 `;base64,` / `"data":"` 标记的裸长串，判定依据是长串前最近的 JSON 键名）按
  **1 token/字节** 稠密计（实测 20KB base64 被上游分词为 18907 token ≈ 0.92 token/字节，
  若按图片折算会低估 4 倍、生产实测产生过一笔 $0.0034 的 write-off）。
  分段写法（MIME/PEM 每 76 字符换行、url-safe 的 `-`/`_`、转义 `\/`）按**同一条串**
  合并识别；无标记的长串再按"字符种类 ≥ 24"判定是否真是 base64（英文文本的 base64
  实测 37 种、十六进制 16 种、`abab…` 2 种），避免把低熵长串按稠密费率高估。
  对超长单行文本等异常内容仍会高估（方向安全）。
- **键名回溯有字节预算**（1024）：`jsonKeyBefore` 只回扫有限字节，避免"A×512;"这类
  构造体触发 Θ(n²) 反向扫描（实测 1.64MB 请求体曾耗 6.19s CPU）；超预算按"未知键"
  处理（归稠密，方向保守）。
- **`userRepo` 未装配的降级部署**：无法做 DB 真值复核，护栏退化为"只按缓存判断"，
  会通过 `recheck_skipped_no_user_repo` 计数并出现在 `degraded_signals` 中。
- **计费幂等的安全前提（重要）**：`usage_billing_dedup` 的主键是
  `(request_id, api_key_id)`，其中 `request_id` **只取服务端可信来源**
  （强持久 money-event id（`web_search:` 等）> 上游响应 id > 服务端生成
  `generated:`），**绝不读取客户端可控的 `X-Client-Request-ID` / `X-Request-ID`**；
  这两个 header 也**不在上游透传白名单内**（否则"回显客户端 id 为响应
  `x-request-id`"的上游会让客户端重新取得幂等键的控制权 —— 恒定一个 id 就能让多笔
  真实调用折叠成一笔、静默去重不扣费，使整条护栏失效）。
  由此推出两条运维含义：
  1. **上游必须返回可信、互不相同的 request id**（或返回空 → 由服务端生成）。若某上游
     对多笔调用复用同一个 id，第二笔起会走"指纹不同 → 换服务端新 id 重新落账"（不丢单，
     会在日志打 `ALERT: usage billing request id conflict`）；该笔 `usage_log.request_id`
     会同步为实际计费键（形如 `<原id>:conflict:<服务端id>`），便于按 request_id 对账。
  2. 幂等以"同一笔调用只提交一次"为前提（worker 池的 `TrySubmit` 成功即仅入队一次，
     失败按 overflow policy 单次降级）。若将来引入记账重试/人工重放，必须复用同一个
     上游 id，否则会被视为两笔独立扣费。
