# Operator 更新记录

## 1.7.3(2026-01-08) {#cl-1.7.3}

- 新增支持 ClusterLoggingConfig CRD 的 `from_beginning_threshold_size` 配置项（#80）

## 1.7.2(2025-12-18) {#cl-1.7.2}

- 修复了注入 logfwd 可能存在的挂载错误问题（#78）

## 1.7.1(2025-12-17) {#cl-1.7.1}

- 修复了对旧配置的兼容问题（#77）
- 如果 `admission_inject_v2` 配置项 `namespace_selectors` 和 `label_selectors` 都为空，就不执行注入（#77）
- 注入 flameshot 新增配置项 `enable_prometheus_annotations` 允许添加 Prometheus.io Annotations（#74）
- 调整 flameshot 配置端口的方式，从环境变量 `FLAMESHOT_HTTP_LOCAL_ADDR` 改成 `FLAMESHOT_HTTP_LOCAL_PORT（#75）
- 修复注入 logfwd 时必须配置 `log_configs` 的错误问题，因为 logfwd 可以使用 CRD 配置源（#76）

## 1.7.0(2025-12-12) {#cl-1.7.0}

- 新增 `admission_inject_v2` 配置项，替代原有的 `admission_inject`，同时保持向后兼容（#73）
- 新增 flameshot 注入，替换原有的 Profiler 注入（#72）
- 选择器 `namespace_selectors` 和 `label_selectors` 同时配置时，两者的关系由 “或” 改为 “且”
- 移除对 Annotation `admission.datakit/logfwd.log_configs` 和 `admission.datakit/logfwd.volume_paths` 的支持（仅 v1.6.X 版本支持）
- 移除对 Annotation `admission.datakit/java-lib.version` 的支持（仍可通过 `admission.datakit/ddtrace.enabled:"false"` 禁用注入）

## 1.6.1(2025-12-02) {#cl-1.6.1}

- 优化对 Kubernetes ClusterLoggingConfig CRD 的错误检查，当资源不存在时会打印日志并停止该功能（#71）

## 1.6.0(2025-11-19) {#cl-1.6.0}

- 新增对 Kubernetes ClusterLoggingConfig CRD 的支持，新增 HTTP 接口返回日志采集配置（#70）
- 调整和优化了注入 logfwd 的方式，适配新的 CRD 日志采集配置（#70）

## 1.5.18(2025-07-15) {#cl-1.5.18}

- 支持在注入 ddtrace 时手动配置 Resources Requests 和 Resources Limits（#68）

## 1.5.17(2025-06-03) {#cl-1.5.17}

- 发布 image 支持 uos arm64 版本

## 1.5.16(2025-04-16) {#cl-1.5.16}

- 调整注入 logging 配置的匹配顺序（#64）

## 1.5.15(2025-04-15) {#cl-1.5.15}

- 优化在 Kubernetes 1.19 版本部署时不支持匹配 namespace 的问题（#63）

## 1.5.14(2025-04-01) {#cl-1.5.14}

- 优化 namespace 正则匹配的写法，现在要匹配 all 只需要写 `*` 而不是 `.*`（#59）

## 1.5.13(2025-03-31) {#cl-1.5.13}

- 支持正则表达式匹配 namespace 和 labels，以注入 DDtrace 和 Profiler（#58）
- 支持正则表达式匹配 namespace 和 labels 注入 logging 配置（#59）

## 1.5.12(2025-02-28) {#cl-1.5.12}

- 支持在注入 logfwd 和 profiler 时手动配置 Resources Requests 和 Resources Limits（#57）

## 1.5.11(2025-02-20) {#cl-1.5.11}

- 支持给目标 Pod 添加 datakit/logs 注解配置并自动挂载目录（#53）
- 优化注入 logfwd sidecar 的目录挂载逻辑，现在默认复用已有挂载（#55）

## 1.5.10(2024-12-10) {#cl-1.5.10}

- 调整环境变量 DD_TAGS 的注入顺序，始终可以引用前面环境变量的值（#51）

## 1.5.9(2024-11-25) {#cl-1.5.9}

- 修改注入镜像的 pullPolicy 为 Always（#50）

## 1.5.8(2024-10-10) {#cl-1.5.8}

- 修改 ddtrace Python 的注入逻辑，不再注入镜像和 lib，只添加基础环境变量（#35）

## 1.5.7(2024-09-18) {#cl-1.5.7}

- 修复 v1.5.5 和 v1.5.6 注入环境变量 DD_TAGS 的错误问题（#48）
- 支持在 Pod 上添加 Annotations（`admission.datakit/ddtrace.enabled="false"`、`admission.datakit/logfwd.enabled="false"` 和 `admission.datakit/profiler.enabled="false"`）更细致地关闭某一类注入（#49）

## 1.5.6(2024-09-13) {#cl-1.5.6}

- 优化注入 profiler volumeMount 逻辑，如果已经存在相同的 path 就不再注入（#46）

## 1.5.5(2024-09-02) {#cl-1.5.5}

- 优化注入 ddtrace 环境变量 `DD_TAGS` 逻辑，如果原 Pod 已经存在 `DD_TAGS`，现在会追加而不是忽略（#45）

## 1.5.4(2024-08-27) {#cl-1.5.4}

- 移除在注入 ddtrace 的 resource，降低 logfwd 和 profiler 的 resource（#44）

## 1.5.3(2024-04-09) {#cl-1.5.3}

- 修复在多容器情况下 ddtrace 挂载文件缺失的问题（#42）

## 1.5.2(2024-04-08) {#cl-1.5.2}

- 支持根据 labelSelector 批量注入 ddtrace（#41）

## 1.5.1(2024-04-07) {#cl-1.5.1}

- 在日志打印当前版本和编译信息（#41）

## 1.5.0(2024-03-18) {#cl-1.5.0}

- 支持在 Pod 上添加 Annotation `admission.datakit/enabled="false"` 关闭所有的注入，包括注入 ddtrace、logfwd 和 profiler（#39）
- 支持给指定的 namespace 批量注入 ddtrace（#38）
- 在注入 logfwd 时可以选择复用 volume，避免同路径多次挂载的报错问题（#34）

## 1.4.3(2023-12-21) {#cl-1.4.3}

- 修复在注入 logfwd 时，如果 logfiles 写通配路径会导致 mount 错误的问题（#31）
- 修复在注入 logfwd 时，如果该 Pod 有 2 个及以上的容器，会注入失败并影响原 Pod 启动的问题（#32）
- 优化一处日志打印

## 1.4.2(2023-09-18) {#cl-1.4.2}

- 修复注入的环境变量顺序不一致问题 (#28)

## 1.4.1(2023-09-15) {#cl-1.4.1}

- 支持给 logfwd 注入环境变量 (#27)

## 1.4.0(2023-09-13) {#cl-1.4.0}

- 支持以 Kubernetes DownloadAPI FieldRef 的方式配置环境变量 (#26)
- ddtrace 默认添加 DD_TAGS (#26)

## 1.3.1(2023-08-24) {#cl-1.3.1}

- 更新默认 Profiler image

## 1.3.0(2023-07-17) {#cl-1.3.0}

- 修改 admission 注入的最小单元为 Pod，需要将 yaml 更新到最新或 datakit-operator-v1.3.0.yaml (#22)
- 支持注入 profiler (#5)
- 注入的 sidecar 添加 Resource Limit (#20)

## 1.2.1(2023-06-28) {#cl-1.2.1}

- 修改默认的镜像仓库 (#19)
- 更新 Helm 结构

## 1.2.0(2023-06-13) {#cl-1.2.0}

- 支持以 JSON 的方式配置 Datakit Operator，现有的环境变量方式保持兼容 (#19)

## 1.0.5(2023-05-11) {#cl-1.0.5}

- 添加新的 ping API (#18)
- 添加 Datakit 选举专用 API，实现对 Datakit 采集器的任务分发 (#15)

## 1.0.4(2023-04-10) {#cl-1.0.4}

- 支持以 Kubernetes Admission 方式注入 logfwd 程序 (#12)
- 注入 ddtrace agent 时会默认添加 `DD_AGENT_HOST` 和 `DD_TRACE_AGENT_PORT` 两个环境变量 (#14)
- 修改 Admission 支持的 resources 列表，不再支持原生 Pod (#12)
- 修改代码的结构，补全 ddtrace 和 logfwd 的单元测试
- 优化 datakit.yaml 的结构
- 移除 docs 目录和文档，在 README 提供新的文档链接

## 1.0.3(2023-03-27) {#cl-1.0.3}

- 添加英文文档
- 支持以环境变量的方式配置 dd-agent 镜像地址 (#9)
- 优化几个环境变量的命名
- 修改 datakit.yaml，默认忽略 webhook 的报错 (#11)

## 1.0.2(2023-03-09) {#cl-1.0.2}

- 添加 CHANGELOG (#9)
- 修复因证书过期导致访问失败的问题，重新生成自签证书且延长过期时间 (#8)
- 修复发布 image 遇到的一个细节错误 (#10)
- 修改 yaml 安装方式，不在 yaml 中创建 namespace，在文档中补全说明 (#7)

## 1.0.1(2022-12-28) {#cl-1.0.1}

- 添加 Makefile、Dockerfile 和 CI 配置 (#2)
- 支持以 Kubernetes Admission 方式注入 ddtrace 文件和环境变量 (#1)

