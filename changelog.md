## v1.1.0

#### 发布说明

本版本主要涉及部分采集器的 bug 修复以及 datakit 主配置的调整。

* Breaking Changes * ：采用新的版本号机制，原来形如 `v1.0.0-2002-g1fe9f870` 这样的版本号将不再使用，改用 `v1.2.3` 这样的版本号
* Breaking Changes * ：原 DataKit 顶层目录的 `datakit.conf` 配置移入 `conf.d` 目录
* Breaking Changes * ：原 `network/net.conf` 移入 `host/net.conf`
* Breaking Changes * ：原 `pattern` 目录转移到 `pipeline` 目录下
* Breaking Changes * ：原 grok 中内置的 pattern，如 `%{space}` 等，都改成大写形式 `%{SPACE}`。**之前写好的 grok 需全量替换**
* Breaking Changes * ：移除 `datakit.conf` 中 `uuid` 字段，单独用 `.id` 文件存放，便于统一 DataKit 所有配置文件
* Breaking Changes * ：移除 ansible 采集器事件数据上报

#### Bug 修复

- 修复 `prom`、`oraclemonitor` 采集不到数据的问题
- `self` 采集器将主机名字段 hostname 改名成 host，并置于 tag 上
- 修复 `mysqlMonitor` 同时采集 MySQL 和 MariaDB 类型冲突问题

#### 特性

- 新增采集器/主机黑白名单功能（暂不支持正则）
- 重构主机、进程、容器等对象采集器采集器
- 新增 pipeline/grok 调试工具
- `-version` 参数除了能看当前版本，还将提示线上新版本信息以及更新命令
- 支持 DDTrace 数据接入
- `tailf` 采集器新日志匹配改成正向匹配
- 其它一些细节问题修复
