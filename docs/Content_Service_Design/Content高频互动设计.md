# 1 设计背景

在短视频场景中，一个爆款视频可能在极短时间内承受大量点赞、取消点赞、发表评论、删除评论请求。

如果每次互动都直接同步更新 MySQL：

1. 点赞关系表会产生大量高频写入。
2. 视频主表中的 `favorite_count`、`comment_count` 会频繁更新，形成热点行竞争。
3. 评论列表查询会在热点视频下频繁扫描数据库。

因此，Content 服务的高频互动模块需要一套兼顾性能、正确性与可恢复性的方案。

# 2 设计目标

本方案的核心目标如下：

1. **Redis 扛热写**：让热点互动请求优先落在内存层，避免直接冲击 MySQL。
2. **MySQL 保真相**：核心关系和评论正文最终都要持久化到 MySQL，不接受永久丢失。
3. **最终一致**：允许短时间统计延迟，但最终计数与关系必须收敛一致。
4. **幂等安全**：重复点赞、重复取消、消息重放不能把计数刷乱。
5. **平滑演进**：尽量复用项目已有的 Redis、Kafka、Worker 与 DDD 目录结构。

# 3 一致性取舍

本方案的取舍原则如下：

## 3.1 可以接受的情况

1. 点赞数、评论数在短时间内存在轻微延迟。
2. Redis 缓存丢失后，热点数据需要重新回源和重建。
3. Worker 短时故障时，聚合计数刷盘延迟。

## 3.2 不可接受的情况

1. 用户点赞关系永久丢失。
2. 评论正文永久丢失。
3. 统计长期不一致且无法自动修复。

因此，Redis 在本方案中不是最终真相库，而是高并发缓冲层。

# 4 点赞设计

点赞分为两类数据：

1. **关系数据**：用户是否点赞了某个视频。
2. **统计数据**：视频当前有多少点赞。

## 4.1 MySQL 设计

新增表 `video_favorites`：

- `id`
- `video_id`
- `user_id`
- `status`
- `created_at`
- `updated_at`

其中：

- `status = 1` 表示当前已点赞
- `status = 0` 表示当前已取消点赞

并通过 `(video_id, user_id)` 唯一索引保证单用户对单视频只有一条关系记录。

这种设计的优势是：

1. 点赞和取消点赞都能通过 `upsert` 幂等落库。
2. 后续可以方便做审计、追踪、纠偏。
3. 即使 Redis 失效，也能从 MySQL 恢复最终关系。

## 4.2 Redis 设计

点赞在实现层采用 **单用户-单视频状态缓存**，而不是维护一个“全量点赞用户集合”。

原因：

1. 如果只维护局部集合缓存，在缓存重建不完整时会产生“假阴性”。
2. 对 `is_favorite` 场景来说，最常见的问题是：当前用户是否点赞了当前视频。
3. 使用 pair key 可以在缓存失效后按需懒加载，语义更精确。

点赞相关 Redis Key：

- `like:status:{videoId}:{userId}`
  - 类型：String
  - 含义：当前用户对当前视频的点赞状态
  - 值：`1` 或 `0`

- `cnt:video:like:{videoId}`
  - 类型：String
  - 含义：视频实时点赞数缓存

- `dirty:video:stats`
  - 类型：Set
  - 含义：记录哪些视频的聚合统计需要异步回刷

## 4.3 Kafka 设计

新增 Topic：

- `video_favorite_events`

Kafka 中保存点赞关系变更事件，而不是直接保存聚合计数。

事件字段：

- `event_id`
- `video_id`
- `user_id`
- `status`

含义：

- `status = 1` 表示最终状态为已点赞
- `status = 0` 表示最终状态为已取消

这样比“只记增量 +1/-1”更稳健，因为消息重放时可以直接做幂等 upsert。

## 4.4 点赞写链路

### 点赞

1. 读取 `like:status:{videoId}:{userId}`，若缓存未命中则回源 MySQL。
2. 若当前未点赞，则：
   1. 将状态缓存置为 `1`
   2. `cnt:video:like:{videoId}` 增加 1
   3. 将 `videoId` 放入 `dirty:video:stats`
3. 发送 Kafka 事件，异步落点赞关系表。
4. 若 Kafka 发送失败，则同步降级写 MySQL，保证关系不丢。

### 取消点赞

1. 读取点赞状态缓存，未命中则回源 MySQL。
2. 若当前已点赞，则：
   1. 将状态缓存置为 `0`
   2. `cnt:video:like:{videoId}` 减 1
   3. 将 `videoId` 放入 `dirty:video:stats`
3. 发送 Kafka 事件，异步落点赞关系表。
4. 若 Kafka 发送失败，则同步降级写 MySQL。

## 4.5 点赞读链路

当 Feed 或 PublishList 返回视频时：

1. 优先读 `cnt:video:like:{videoId}` 作为实时点赞数。
2. 若用户已登录，则读 `like:status:{videoId}:{userId}` 作为 `is_favorite`。
3. 若缓存未命中，则回源 MySQL，并把结果回填到 Redis。

这样可保证：

1. 热点查询优先走缓存。
2. 缓存失效后仍能保证正确结果。
3. 点赞读链路不会依赖热点集合重建。

# 5 评论设计

评论和点赞不同，评论正文是核心业务数据，不能只放在 Redis 中。

因此本方案采用：

- **评论正文同步写 MySQL**
- **评论列表和评论计数走 Redis 加速**
- **视频主表评论总数异步回刷**

## 5.1 MySQL 设计

新增表 `video_comments`：

- `id`
- `video_id`
- `user_id`
- `content`
- `status`
- `created_at`
- `updated_at`

状态设计：

- `1`：有效评论
- `2`：已删除评论

这使得评论删除无需物理删行，后续也便于做审核、恢复与追踪。

## 5.2 Redis 设计

评论相关 Redis Key：

- `comment:list:{videoId}`
  - 类型：ZSet
  - 含义：评论 ID 有序列表
  - score：评论创建时间

- `comment:detail:{commentId}`
  - 类型：String(JSON)
  - 含义：评论详情缓存

- `cnt:video:comment:{videoId}`
  - 类型：String
  - 含义：视频实时评论数缓存

- `dirty:video:stats`
  - 类型：Set
  - 含义：标记该视频评论统计需要回刷

## 5.3 评论写链路

### 发表评论

1. 同步写入 MySQL 评论表，确保正文持久化。
2. 将评论详情写入 `comment:detail:{commentId}`。
3. 将评论 ID 写入 `comment:list:{videoId}`。
4. `cnt:video:comment:{videoId}` 增加 1。
5. 将 `videoId` 放入 `dirty:video:stats`。

### 删除评论

1. 先校验评论存在且归属当前用户。
2. 将 MySQL 中评论状态更新为已删除。
3. 从 Redis 评论详情和评论列表中删除对应缓存。
4. `cnt:video:comment:{videoId}` 减 1。
5. 将 `videoId` 放入 `dirty:video:stats`。

## 5.4 评论读链路

1. 优先从 `comment:list:{videoId}` 读取评论 ID 列表。
2. 再从 `comment:detail:{commentId}` 读取评论详情。
3. 若评论详情缓存缺失，则回源 MySQL 批量加载并回填缓存。
4. 用户信息通过 User 服务批量查询并装配。

这样可以让热点视频评论列表主要在 Redis 中完成读取，降低数据库压力。

# 6 Worker 设计

Worker 分为两类职责：

## 6.1 Favorite Worker

职责：

1. 消费 `video_favorite_events`
2. 将点赞最终状态幂等写入 `video_favorites`
3. 写入成功后，将视频标记为 `dirty:video:stats`

这样做保证了：

1. 点赞关系最终进入 MySQL
2. Kafka 消息重放不会破坏关系正确性
3. 点赞关系和聚合计数解耦

## 6.2 Interaction Stats Flusher

职责：

1. 定时扫描 `dirty:video:stats`
2. 分别统计：
   - MySQL 中有效点赞数
   - MySQL 中有效评论数
3. 回写 `videos.favorite_count` 和 `videos.comment_count`
4. 同步刷新 Redis 中的 `cnt:video:like:{videoId}` 与 `cnt:video:comment:{videoId}`
5. 清除 dirty 标记

这里采用的是**按脏视频精确重算**，而不是直接将增量盲目刷盘。

优势：

1. 避免因消息乱序导致计数错误
2. 避免重复消费时发生重复加减
3. 让计数最终收敛到 MySQL 中的真实结果

# 7 故障处理策略

## 7.1 Redis 挂掉

影响：

1. `is_favorite` 和计数缓存失效
2. 评论列表缓存失效

恢复方式：

1. 点赞状态从 MySQL 点赞关系表回源恢复
2. 评论列表从 MySQL 评论表回源恢复
3. 计数通过 Stats Flusher 重新回填

结论：

- Redis 丢的是缓存和短期热状态
- 不会导致核心关系或评论正文永久丢失

## 7.2 Kafka 挂掉

影响：

1. 点赞关系异步链路暂时不可用

处理方式：

1. 点赞接口检测到 Kafka 写失败后，同步降级写 MySQL
2. 保证点赞关系不丢
3. 只是性能退化，不破坏正确性

## 7.3 Worker 挂掉

影响：

1. 计数刷盘延迟
2. 点赞关系落库延迟

恢复方式：

1. Worker 重启后继续消费 Kafka
2. 脏视频会继续被 Stats Flusher 刷新

因此 Worker 故障只会造成**延迟一致**，不会造成永久错误。

# 8 为什么这是优雅方案

本方案优雅的原因主要有四点：

1. **正确性与性能兼顾**
   - Redis 负责高并发
   - MySQL 负责最终真相

2. **点赞与评论分而治之**
   - 点赞偏关系与计数，适合异步削峰
   - 评论偏正文内容，适合同步持久化

3. **幂等设计明确**
   - 点赞关系落库基于最终状态
   - 计数刷盘采用精确重算

4. **故障恢复路径清晰**
   - Redis 丢失可回源
   - Kafka 不可用可降级
   - Worker 故障可自动收敛

# 9 当前实现落点

本次实现已经在代码中完成以下落地：

1. 新增点赞关系模型与评论模型
2. 新增点赞仓储、评论仓储
3. 新增点赞接口、评论接口、评论列表接口
4. 在 Feed 与 PublishList 中补充互动计数与 `is_favorite`
5. 新增 Favorite Worker
6. 新增 Interaction Stats Flusher
7. 将高频互动验证流程补充到手工验证文档中

# 10 后续可演进方向

后续还可以继续演进：

1. 为评论接入审核状态机
2. 引入通知系统，给视频作者推送“新点赞/新评论”
3. 为点赞关系增加用户点赞列表接口
4. 对评论列表加入分页游标
5. 对热点视频互动缓存增加限长与淘汰策略
6. 对 Stats Flusher 做批量 SQL 优化

