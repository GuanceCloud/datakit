---
title     : 'Zabbix_export'
summary   : 'Zabbix realTime data 数据接入'
__int_icon      : 'icon/zabbix'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Zabbix RealTime Exporter
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

[:octicons-tag-24: Version-1.37.0](../changelog.md#cl-1.37.0)

---

采集 Zabbix 服务的实时数据并发送到观测云中心。

Zabbix 从 5.0 到 7.0 版本都支持将实时数据写到文件中。实时数据中有三种数据格式：`events/history/trends` ，其中 `history` 和 `trends` 都是以指标形式展示。而 `events` 则可以通过 [Webhook](https://www.zabbix.com/documentation/5.4/en/manual/config/notifications/media/webhook?hl=Webhook%2Cwebhook){:target="_blank"} 方式发送到观测云。

## Zabbix 配置 {zabbix_config}

修改配置文件，一般位于 */etc/zabbix/zabbix_server.conf* :

```toml
### Option: ExportDir
#       Directory for real time export of events, history and trends in newline delimited JSON format.
#       If set, enables real time export.
#
# Mandatory: no
ExportDir=/data/zbx/datakit

### Option: ExportFileSize
#       Maximum size per export file in bytes.
#       Only used for rotation if ExportDir is set.
#
# Mandatory: no
# Range: 1M-1G
ExportFileSize=32M

### Option: ExportType
#       List of comma delimited types of real time export - allows to control export entities by their
#       type (events, history, trends) individually.
#       Valid only if ExportDir is set.
#
# Mandatory: no
# Default:
# ExportType=events,history,trends
```

修改其中的配置项：

```toml
ExportDir=/data/zbx/datakit
ExportFileSize=32M
```


修改文件的权限：

```shell
mkdir -p /data/zbx/datakit
chown zabbix:zabbix -R /data/zbx/datakit
chmod u+rw -R /data/zbx/datakit/
```

注意：配置文件大小时根据主机配置衡量，太大的文件容易造成磁盘空间不足。并且应当定期删除 `.old` 文件。这里设置成为 32M 就是考虑到文件系统的负载太大。

配置好之后，重启服务：

```shell
systemctl restart zabbix-server
```

## 采集器配置 {#config}

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
   {{ CodeBlock .InputSample 4 }}
```

配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

## 参考文档 {#documents}

- 官方配置文档：[5.0 配置数据导出](https://www.zabbix.com/documentation/5.0/en/manual/appendix/install/real_time_export?hl=export){:target="_blank"}
- [6.0 数据导出](https://www.zabbix.com/documentation/6.0/en/manual/appendix/install/real_time_export){:target="_blank"}
- [7.0 数据导出](https://www.zabbix.com/documentation/current/en/manual/config/export/files){:target="_blank"}
