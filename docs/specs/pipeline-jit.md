# Pipeline JIT 配置与性能说明

适用交付分支：`feat-jit-aot80`。本页性能对应 Rust `fdd63c39`，DK 基准代码 `c8946f4cc8`；测试报告提交 `f5dbaf8940`。这是本地验证版本，不代表发布产物已构建或部署。

## 启用

JIT 默认关闭。需要支持 `pipeline_jit` 的 DK 构建及匹配架构的 Rust 动态库；当前支持 Linux amd64/arm64 glibc。构建与打包见 [发布构建说明](pipeline-jit-build.md)。普通静态 DK 包不能仅靠配置获得 JIT 能力。

在 `datakit.conf` 中添加以下配置；已有 `[pipeline.jit]` 时修改原表，不要重复声明：

```toml
[pipeline.jit]
  enabled = true
  mode = "auto"
  on_init_error = "go"
  max_cached_programs = 256
```

动态库默认从 `<DK 安装目录>/lib/libplatypus_jit.so` 加载。上述配置使用该默认路径，不需要维护脚本 allowlist。修改配置后重启 DK 生效；本页不依赖配置热更新。

自定义动态库位置时，在同一表中设置 `runtime_path`。生产发布可以同时设置 `runtime_sha256`，值必须来自对应架构的可信发布产物；不要复制性能测试库的哈希作为其他产物的哈希。

```toml
# 添加到上面的 [pipeline.jit] 表内，按实际产物填写：
# runtime_path = "/opt/datakit/lib/libplatypus_jit.so"
# runtime_sha256 = "<可信发布产物的64位SHA-256>"
```

## 配置字段

| 字段 | 默认值 | 行为 |
|---|---|---|
| `enabled` | `false` | JIT 总开关 |
| `mode` | `"auto"` | 自动检查脚本并选择可执行路径；不支持的脚本在执行前选择 Go |
| `on_init_error` | `"go"`（空值同此行为） | `"go"` 允许首次初始化的 runtime 准备失败时选 Go；`"error"` 返回初始化错误 |
| `runtime_path` | 安装目录下 `lib/libplatypus_jit.so` | Rust 动态库路径 |
| `runtime_sha256` | 空 | 可显式固定可信产物 SHA-256；allowlist 模式必填 |
| `max_cached_programs` | `256` | 配置为 0 也使用默认值；允许 0–4096 |
| `force_go` | 空 | 按 category、namespace、script 精确指定走 Go 的脚本 |
| `allow` | 空 | allowlist 模式使用的脚本及 bundle 哈希规则；auto 模式无需维护 |

`auto` 不是“任何脚本都强制 native”，也不保证所有脚本更快。有些函数可能使用 Go host callback；需要结合路由指标确认实际执行路径。

## 指定脚本走 Go

例如让默认命名空间的 Redis 内置脚本保持 Go：

```toml
[[pipeline.jit.force_go]]
  category = "logging"
  namespace = "default"
  script = "redis.p"
```

匹配完整身份，不支持通配符。namespace 可为 `default`、`gitrepo`、`confd`、`remote`；使用脚本实际加载的命名空间和名称。

## 关闭与故障处理

全局回退：把 `enabled` 改为 `false`，重启 DK；后续执行走 Go。保留原有 Pipeline 脚本和采集器配置即可。

`on_init_error = "go"` 只覆盖首次初始化、尚未发布 Pipeline 时的 runtime 准备失败，不兜底非法配置，也不是运行期任意错误后重试的开关。已经提交给 JIT 的当前批次失败时，不会自动重新执行 Go，以免重复外部副作用。运行期隔离针对后续路由，不能追回当前失败批次，也不提供进程崩溃隔离。

灰度时先在少量实例开启 auto；对比输出、错误率、CPU、内存和处理延迟。性能落后的脚本可以用 `force_go` 保留 Go。

## 确认生效

启动日志应出现 `pipeline JIT enabled with runtime`，随后记录动态库 SHA-256 和 policy。首次降级会记录 `pipeline JIT initial load failed; selecting Go before execution`。

可观察以下指标：

- `datakit_pipeline_jit_control_state`：`state="enabled"` 和 `state="init_degraded"`。
- `datakit_pipeline_jit_route_records_total`：检查 native 提交量；只有开关打开不足以证明实际执行 JIT。
- `datakit_pipeline_jit_fallback_records_total`：执行前路由到 Go 的记录。
- `datakit_pipeline_jit_quarantines_total`：脚本或 runtime 隔离情况。

## 性能与验证

[内置脚本 E2E 与 runtime 报告](pipeline-jit-performance.md) 保留最终测量结果、适用范围及历史原始数据和复现脚本的归档入口。

8 个未修改的内置脚本、10 份 DK 原始 LogExamples，包括多行 MySQL 慢日志。完整 E2E 包含 Point 构造、路由、执行、回写和释放，不含采集、上传和首次编译。四轮交替运行取中位数。

| 场景 | batch=10 Go/JIT | batch=1 Go/JIT |
|---|---:|---:|
| Nginx access | 1.23× | 1.19× |
| Nginx error | 1.77× | 1.26× |
| Apache | 1.21× | 0.97× |
| MySQL 普通日志 | 1.15× | 0.82× |
| MySQL 多行慢日志 | 1.97× | 1.73× |
| Redis | 0.94× | 0.74× |
| Elasticsearch 慢日志 | 3.08× | 2.50× |
| MongoDB | 1.52× | 1.19× |
| Consul | 4.26× | 2.96× |
| TDengine | 2.24× | 1.71× |

倍率大于 1 表示 JIT 更快。本轮 batch=10 为 9/10 领先，batch=1 为 7/10 领先；独立 runtime 测量为 10/10 领先。Redis 的 runtime 优势未抵消完整调用路径成本。插桩 runtime 不能直接从独立 E2E 数据中相减得到对接耗时。

所有基准预热的完整 Point 输出一致性检查通过（仅排除 `_pl_cost`），确认 native 提交；上线前控制测试通过。这是 WSL 本地样例回放，仍有宿主机波动，不能当作客户生产吞吐承诺。本轮未重新测量 AOT。
