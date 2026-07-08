# Flameshot Python 采集文档

Python 采集为后续支持项，当前 Flameshot 还没有实现 Python 语言的采集执行流程。

## 当前状态

- `FLAMESHOT_PROCESSES[].language` 当前不支持 `python`。
- 当前版本不会为 Python 进程启动 Python profiler。
- HTTP 手动触发、阈值触发和定时触发都依赖语言采集器实现，因此暂不能用于 Python Profile 上传。

## 后续文档结构

Python 支持落地后，本文件应补充以下内容：

- Python profiler 选型和运行方式。
- 业务进程需要开启的参数或依赖。
- Flameshot 进程规则字段。
- Profile 文件格式和 DataKit 上传格式。
- Kubernetes Sidecar 示例。
- 本机测试方式。
- 常见问题和排障命令。

## 通用配置

Flameshot 的进程匹配、阈值、定时采集、HTTP 手动触发等公共能力见 [总文档](./readme.md)。Python 语言采集实现后，应复用这些公共触发能力。
