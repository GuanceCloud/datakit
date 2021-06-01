# DataKit/操作系统版本信息（建议提供 datakit --version 的输出）

# 你预期的行为是怎样的？

# 尽可能提供复现步骤（文字、截屏、视屏等不限）

# 如果是对应采集器（如 MySQL）问题，请给出对应采集器的配置文件

# 如果可以，请给出 http://datakit-ip:9529/monitor 截图，便于排查问题

# 尽可能给出 DataKit 的 ERROR/WARN 日志

- Windows 位于 `C:\Program Files\datakit\log`

```powershell
# 通过 Powershell 给出最近 10 个 ERROR, WARN 级别的日志
Select-String -Path 'C:\Program Files\datakit\log' -Pattern "ERROR", "WARN"  | Select-Object Line -Last 10
```
- Linux/Mac 位于 `/var/log/datakit/log`

```shell
# 给出最近 10 个 ERROR, WARN 级别的日志
cat /var/log/datakit/log | grep "WARN\|ERROR" | head 10
```
