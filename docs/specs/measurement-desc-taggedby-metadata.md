# 设计：补全 Measurement 描述和 FieldInfo Taggedby

由 /office-hours 于 2026-05-13 生成
分支：units.desc
仓库：datakit
状态：完成
模式：Builder

## 问题说明

一些指标字段在上报时带有采集器主动构建的维度 tag，但对应的 `inputs.FieldInfo` 文档不一定声明了 `Taggedby`。文档应该告诉用户某个字段由哪些 tag 限定范围，并且以真实 point 构建代码作为事实来源。

同时，部分 `MeasurementInfo` 只有 `Name`、`Cat`、`Fields`、`Tags`，缺少 measurement 级别的 `Desc` / `DescZh`。这些字段应该根据指标集整体语义补齐，让文档读者先理解这个指标集覆盖什么对象、从哪里采集、按什么维度上报。

这次目标只限元数据文档更新：补 `FieldInfo.Taggedby`，并补 `MeasurementInfo.Desc` / `MeasurementInfo.DescZh`。不要为这轮文档更新新增 measurement-refine 测试。

## 现有模式

Kafka、Redis、JVM、PostgreSQL、MySQL、Oracle 和 SQL Server 已经有比较合适的结构：

- 为一个语义分组构建独立的 field map。
- 为同一个分组构建匹配的 tag map。
- 调用 `addTaggedbyToFields(fields, TagGroupX)`。
- helper 会把 tag key 复制到该分组每个字段的 `FieldInfo.Taggedby` 中。

示例：

- `internal/plugins/inputs/redis/metric_measurement.go`：`TagGroupInfo`、`TagGroupCommand`、`TagGroupReplica`、`TagGroupDatabase`。
- `internal/plugins/inputs/kafka/measurement.go`：`TagGroupPurgatory`、`TagGroupRequest`、`TagGroupTopic`、`TagGroupPartition`。
- `internal/plugins/inputs/jvm/measurement.go`：`TagGroupGC`、`TagGroupPool`。
- DB 类采集器在各自的 `metric_measurement.go` 中已经有较完整的 helper 覆盖。

## 前提

1. `Taggedby` 应该来自实际上报 point 的 tag，而不是从字段名推断。
2. `MeasurementInfo.Desc` / `DescZh` 应该描述指标集整体语义，不重复字段列表，也不写实现细节噪声。
3. `Taggedby` 跟 SQL `GROUP BY` 没有直接绑定关系，核心是构建 point 时实际使用了哪些 tag key。
4. 如果 tag key 是采集器主动加到 point 上的字段维度，就应该进入对应字段的 `Taggedby`；global host tag、global election tag 和用户额外配置的全局 tag 不进入 `Taggedby`。
5. 追踪 point tag 时不能只看 `AddTag`。还要检查 `point.KVs`、`point.NewTags(...)`、`append(point.NewTags(tags), ...)`、helper 返回的 tag map，以及 `vendor/github.com/GuanceCloud/cliutils/point/` 支持的其他构建路径。
6. 这项工作应该按小批量采集器推进，因为仓库里 query 类采集器很多，只按名字扫描会产生误报。

## 备选方案

### 方案 A：按采集器做 source-traced 小批量

摘要：选择 point 构建路径清晰的采集器，追踪实际上报时由采集器主动加入的 tag key，把 `Taggedby` 加到对应的 `FieldInfo` 分组，并按指标集语义补齐 `MeasurementInfo.Desc` / `DescZh`。

工作量：M
风险：低
优点：
- 符合之前 metadata cleanup 的推进方式。
- 以真实上报 point 结构为准。
- `Desc` / `DescZh` 和 `Taggedby` 在同一批 source trace 中一起补，语义一致。
- 产出的提交容易 review。
缺点：
- 比机械扫描慢。
- 每个采集器都需要手动 source trace。
复用：
- 既有 `addTaggedbyToFields` helper 和 tag-group 模式。

### 方案 B：机械 AST/SQL 扫描器

摘要：写一个扫描器提取候选 tag 构建路径，例如 `AddTag`、`point.NewTags(...)`、tag map 合并和 helper 封装，然后机械 patch `Taggedby`，再对缺少 `Desc` / `DescZh` 的 measurement 做模板化补齐。

工作量：L
风险：高
优点：
- 盘点速度快。
- 适合找候选项。
缺点：
- point tag 构建路径多样，机械扫描容易漏掉 helper、map 合并或动态 tag。
- SQL alias、字段重命名和 point 转换容易导致文档错误。
- 模板化 `Desc` / `DescZh` 容易变成泛泛描述，不能体现指标集语义。
- 动态/custom query 采集器很难安全 patch。
- 很容易给对用户没有帮助的字段加上噪声 `Taggedby`。
复用：
- 之前 metadata audit 中的临时 AST 扫描模式。

### 方案 C：共享 helper 重构

摘要：引入一个所有采集器共用的标准 helper，让字段分组在一个地方声明 tag scope。

工作量：XL
风险：中
优点：
- 长期一致性更好。
- 减少重复 helper 代码。
缺点：
- 对一次 metadata pass 来说范围太大。
- 为文档目标触碰太多采集器。
- review 负担明显上升。
复用：
- 既有 helper 形态，但需要大范围重构。

## 推荐方案

选择方案 A。

从 point 构建和文档已经比较接近的采集器开始：

1. `tdengine`：point 构建通过 `selectSQL.tags` 把 `database_name`、`dnode_ep`、`endpoint/status_code`、`client_ip` 等 tag 加入上报点，适合先补 `Taggedby`。
2. `kingbase`：`collect.go` 中逐个构建 point，显式加入 `queryid`、`query`、`spcname`、`lock_type`、`state/wait_event`、`schemaname/funcname` 等维度 tag。
3. `dameng`：point 构建会添加 `tablespace_name`、`pool_name`、`sess_id`、`database`、`host` 等 tag；只在字段确实被这些 tag 限定范围时添加 `Taggedby`，同时补对应指标集的中英文描述。
4. 然后针对剩余没有 `Taggedby`、但上报 point 含有非通用 tag 的 `FieldInfo` 做定向扫描。

不要从全仓库机械 patch 开始。那样会生成看起来正确、实际细节错误的文档；这种细微错误的文档比缺文档更糟。

## 成功标准

- 按分组或资源上报的字段有 `Taggedby`，并列出限定该字段范围的 tag。
- 缺少 measurement 级别说明的指标集补齐 `Desc` 和 `DescZh`。
- `Desc` / `DescZh` 能说明指标集语义，例如对象类型、采集来源、主要指标范围或上报维度。
- 没有无关 metadata churn。
- 不新增 measurement-refine 测试。
- 每批都通过 `gofmt`、`git diff --check` 和相关的 `go test ./internal/plugins/inputs/<collector> -run '^$' -count=1`。
- 如果这成为另一轮可跟踪 audit pass，则更新 matrix 或进度文档。

## 执行结论

- 本 spec 已完成 2026-06-08 续改扫描。对应的执行记录见 `measurement-desc-taggedby-progress-matrix.md`，当前非测试 Go 文件中已无缺少 `MeasurementInfo.Desc` / `DescZh` 的 `inputs.MeasurementInfo` 字面量。
- 本轮补齐了剩余静态采集器和动态 measurement wrapper 的 measurement 级别 `Desc` / `DescZh`。动态/custom/profile 驱动的采集器仍不伪造字段级 `Taggedby`，只在能够从 point 构造路径确认固定字段维度时更新 `Taggedby`。
- 当前仍保留 `Source traced` 状态，用于标记完成事实追踪但没有字段级 metadata 变更的采集器；这类条目均记录了来源与标签语义，不会在无事实依据时伪造 `Taggedby`。

## 下一步

1. 后续只在新增采集器或新增 measurement 时继续执行相同口径：先 source trace，再补 `Desc` / `DescZh`，字段级 `Taggedby` 只按实际 point tag 归属添加。
2. 如果未来需要继续 type/unit/field desc 清理，应另开 audit pass，不与本轮 measurement 级描述补齐混在一起。

## 观察

- 这个请求的边界是对的：以既有上报 point 作为事实来源，而不是以文档反推。
- 这次新增的 `Desc` / `DescZh` 是 measurement 级别语义描述，不是再做一轮广义字段 type/unit/desc 清理。这样可以补齐文档入口信息，同时避免任务重新膨胀成全仓库级别的泥潭。
