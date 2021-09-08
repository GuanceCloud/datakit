## Bug 问题描述

## DataKit/操作系统版本信息

<!--
建议提供 datakit --version 的输出
-->

## 预期的行为是怎样的？
<!--
尽可能详细说明预期行为，如果是文档中就有的功能说明，可贴出文档链接
-->

## 尽可能提供复现步骤

<!--
文字、截屏、视屏等不限
-->

## 如果是采集器问题，请给出对应采集器的配置文件

<!--
如 MySQL 采集器问题，给出 .conf.d/db/mysql.conf 配置
-->

## 请给出 Datakit 运行信息截图，便于排查问题

<!--
如果可以，请给出 http://datakit-ip:9529/monitor （或者命令行 datakit -M --vvv）截图，便于排查问题
-->

## 尽可能给出 DataKit 的 ERROR/WARN 日志

<!--
- Windows 位于 `C:\Program Files\datakit\log`

# 通过 Powershell 给出最近 10 个 ERROR, WARN 级别的日志
Select-String -Path 'C:\Program Files\datakit\log' -Pattern "ERROR", "WARN"  | Select-Object Line -Last 10

Linux/Mac 位于 `/var/log/datakit/log`

# 给出最近 10 个 ERROR, WARN 级别的日志
cat /var/log/datakit/log | grep "WARN\|ERROR" | head 10
-->
