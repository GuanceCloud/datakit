## v1.1.0

#### 发布说明

本版本主要涉及部分采集器的 bug 修复以及 datakit 主配置的调整。

#### Bug 修复

- 修复 `prom`、`oraclemonitor` 采集不到数据的问题
- `self` 采集器将主机名字段 hostname 改名成 host，并置于 tag 上
- 修复 `mysqlMonitor` 同时采集 MySQL 和 MariaDB 类型冲突问题

#### 变更

- 采集器主配置文件 `datakit.conf` 移入 `conf.d` 目录
- 移除 `datakit.conf` 中 `uuid` 字段，单独用 .id 文件存放
- 原 `pattern` 目录转移到 `pipeline` 目录下
- 原 grok 内置的默认 pattern 名称改成全大写形式

#### 特性

- 新增采集器/主机黑白名单功能
- 重构主机、进程、容器等对象采集器采集器
- 新增 pipeline/grok 调试工具
- `-version` 参数除了能看当前版本，还将提示线上新版本信息以及更新命令
- 支持 DDTrace 数据接入
- `tailf` 采集器新日志匹配改成正向匹配

#### 其它

- 移除 ansible 采集器事件数据上报
