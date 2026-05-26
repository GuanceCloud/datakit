# Relation 通配匹配设计

## 目标

`source -> script` 的 `relation` 映射支持：

- 精准匹配优先
- 精准未命中时再走通配
- 通配符：
  - `*`：匹配 0 个或多个字符
  - `?`：匹配 1 个字符
- 多条通配命中时，保持稳定优先级：
  - 字面量字符更多优先
  - `*` 更少优先
  - 模式更长优先
  - 字典序兜底

## 当前实现

相关代码位于：

- `manager/relation.go`
- `manager/relation_test.go`
- `manager/relation_anchor_test.go`
- `manager/relation_bench_test.go`

### 索引结构

当前生产实现使用 hybrid 索引：

- `prefixTrie`
  - 第一个通配符之前存在稳定字面前缀
  - 例如：`nginx-*`、`svc-??`
- `suffixTrie`
  - 没有稳定前缀，但最后一个通配符之后存在稳定字面后缀
  - 例如：`*-error`
- `literal-anchor index`
  - 没有稳定前后缀，但能抽出较强字面量片段
  - 例如：`*prod*api*`
- `generic`
  - 连 anchor 都抽不出来的规则兜底
  - 例如：`*`、`?*?`

另外，prefix/suffix trie 节点内部会继续按长度分桶：

- 无 `*` 的规则按 `patternLen` 分桶
- 有 `*` 的规则按 `minLen` 分桶

这样在高重叠前缀场景下，不需要把节点上的所有规则都逐条尝试。

### 查询流程

`Query(cat, source)` 的流程：

1. 先查精确 map
2. 如果 miss，查当前快照上的 wildcard result cache
3. 如果 cache miss，再从 wildcard 索引里收集候选规则：
   - generic fallback
   - literal-anchor index
   - prefix trie
   - suffix trie
4. 对候选规则做最终 wildcard match
5. 多条 wildcard 同时命中时，按统一优先级选最优规则
6. 将 wildcard 结果写入当前快照缓存

这里要特别注意两点：

- 候选规则的收集顺序不代表最终命中顺序
- 最终返回结果由统一的“更具体优先”规则决定，因此不是随机的，也不是按配置顺序决定的

### 多条通配同时命中时的优先级

当一个 `source` 同时命中多条 wildcard pattern 时，最终选择遵循固定优先级：

1. 字面量字符更多优先
2. `*` 更少优先
3. 模式更长优先
4. pattern 字典序兜底

这套规则的目标是优先选择“约束更强、更具体”的 pattern，而不是谁先被扫描到就返回谁。

例如，有下面几条 relation：

```text
svc-*        -> p1
svc-??       -> p2
svc-prod-*   -> p3
*prod*       -> p4
```

查询 `source = "svc-prod-ab"` 时：

- `svc-*` 命中
- `svc-prod-*` 命中
- `*prod*` 命中
- `svc-??` 不命中

最终返回 `p3`，因为 `svc-prod-*` 的字面量部分更多，模式更具体。

再例如：

```text
ab*cd   -> p1
ab?cd   -> p2
```

查询 `source = "abxcd"` 时，两条都命中，但最终返回 `p2`。

原因是：

- 两条规则的字面量字符数相同
- `ab?cd` 的 `*` 更少
- `?` 比 `*` 更严格，因此优先级更高

## 高并发设计

当前实现专门考虑了高并发查询：

- `UpdateRelation()` 先构建完整不可变快照，再用原子方式整体切换
- `Query()` 优先读快照，不在正常路径上长期持有 `RWMutex`
- anchor 查询内部的临时标记数组通过 `sync.Pool` 复用
- wildcard 查询结果缓存按快照绑定，更新时跟快照一起切换，不做跨版本失效协调

这意味着：

- 更新仍然是整体替换语义
- 查询路径适合读多写少场景
- 更新与查询并发时不会读到半更新状态
- 当 `source` 重复率高时，重复 wildcard 查询可以快速命中缓存

## 复杂度

记：

- `L`：`source` 长度
- `W`：某个 `category` 下的总通配规则数
- `C`：经过索引剪枝后的候选规则数

则当前生产实现的主路径近似为：

- `O(L + C * L)`

在存在稳定前后缀、字面量 anchor 或高重复查询的场景下，实际成本会明显低于线性扫描。

## 基准结论

基于当前 benchmark，生产路径大致表现为：

- 串行 wildcard 查询：约 `22~23 ns/op`
- `RunParallel` 并发查询：约 `3.6~3.7 ns/op`
- 持续更新 + 并发查询：约 `4.2~4.5 ns/op`

按规则分布看：

- `prefix-heavy`：表现稳定
- `generic-heavy`：受益于 literal-anchor index
- `overlap-heavy`：受益于 trie 节点长度分桶

## 测试覆盖

当前保留的测试重点包括：

- 精准匹配优先
- `*` / `?` 通配语义
- generic anchor 命中
- 多 goroutine 并发查询
- 更新与查询并发进行
- `-race` 下的数据竞争检测
