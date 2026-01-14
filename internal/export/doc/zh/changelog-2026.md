# 更新日志

## 1.88.0(2026/01/14) {#cl-1.88.0}

本次发布属于迭代发布，主要有如下更新：

### 新加功能 {#cl-1.88.0-new}

- 新增数据采集[可用性指标采集](../integrations/ingestion_canary.md)（#2900）
- DCA 新增 DataKit 存活检测（#2910）

### 问题修复 {#cl-1.88.0-fix}

- 修复 Pod 内存采集数值虚高问题（#2933）
- 修复 Pod 重启后 KubernetesPrometheus 未能重新采集的问题（#2936）
- 修复无法采集 DDTrace 中 NodeJS profile 的问题（#2937）[^2937]
- 修复多步拨测重试问题（#2915）
- 修复 AWS Lambda 扩展采集异常问题（#2918）

[^2937]: 要完整支持 DDTrace NodeJS profile 采集，底座仍需升级到最新版本。

### 功能优化 {#cl-1.88.0-opt}

- DataKit 日志输出中，给 `ERROR` 级别的日志单独一个文件（默认为 *error.log*），避免其被其它日志覆盖掉，同时 bug report 中也会带上这个错误日志（#2940）
- 优化磁盘缓存模块（WAL），新增更多指标和日志暴露，同时优化 *.pos* 文件对磁盘 io 的影响（#2935）
- SNMP 采集新增更多 yaml 配置，修复一些历史遗留问题（#2923）
- 容器日志采集和 logfwd 新增 `from_beginning_threshold_size` 配置项（#2934）
- 多个采集器采集的数据上增加了 `collector_source_ip` 字段，表示其数据来源（#2819）[^2819]
- 其它优化（#2928/#2932/#2930）

[^2819]: 这些采集器包括 `zipkin/logstreaming/beats_output` 等。

### 兼容调整 {#cl-1.88.0-brk}

- SNMP 采集的数据中移除了对象数据中的 `all` 冗余字段（#2923）
