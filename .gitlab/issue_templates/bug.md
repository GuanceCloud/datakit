## Bug 问题描述（必填）

## DataKit/操作系统版本信息/部署信息（必填）

- DataKit 版本：？
- 操作系统：？
- K8s Daemonset 部署：？

<!--
建议提供命令 datakit version 的输出
-->

## 预期的行为是怎样的？（可选）
<!--
尽可能详细说明预期行为，如果是文档中就有的功能说明，可贴出文档链接
-->

## 尽可能提供复现步骤（必填）

复现步骤：

1. xxxxx
1. yyyyy
1. zzzzz

datakit.conf 配置：

```toml
这里贴上 datakit 配置
```

或者贴上 datakit.yaml：

```yaml
这里贴上 datakit k8s 中的 yaml 配置
```

文字、截屏、视屏等不限

贴图片的推荐方式：

<img src="/uploads/1d10e09cb7292de571axxxxxxxxxxxxx/image-1.png"  width="730">

<img src="/uploads/1d10e09cb7292de571axxxxxxxxxxxxx/image-2.png"  width="730">

这里建议用 730 作为图片宽度，连续的图片之间，用空行分割下。单个图片不要超过 250KB。

## 如果是采集器问题，请给出对应采集器的配置文件


如 MySQL 采集器问题，给出 .conf.d/db/mysql.conf 配置

```toml
这里贴上采集器配置
```

## 请给出 Datakit 运行信息截图，便于排查问题

命令行执行 datakit monitor -V，贴图方式参照上面的说明。

## 尽可能给出 DataKit 的 ERROR/WARN 日志

- Windows 位于 `C:\Program Files\datakit\log`

# 通过 Powershell 给出最近 10 个 ERROR, WARN 级别的日志

```powershell
Select-String -Path 'C:\Program Files\datakit\log' -Pattern "ERROR", "WARN"  | Select-Object Line -Last 10
```

Linux/Mac 位于 `/var/log/datakit/log`

```shell
# 给出最近 10 个 ERROR, WARN 级别的日志
cat /var/log/datakit/log | grep "WARN\|ERROR" | head 10
```
