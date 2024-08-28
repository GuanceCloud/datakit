---
title     : 'Zabbix_export'
summary   : 'Zabbix real-time data exporter'
__int_icon      : 'icon/zabbix'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
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

Collect real-time data from the Zabbix service and send it to the GuanCe cloud center. Currently, Zabbix supports writing real-time data to files from version 5.0 to 7.0.
ExportType allows to specify which entity types (events, history, trends) will be exported.

## Zabbix Config {zabbix_config}

Zabbix config file: */etc/zabbix/zabbix_server.conf* :

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

Modify the configuration items:

```toml
ExportDir=/data/zbx/datakit
ExportFileSize=32M
```

Permission to modify files:

```shell
mkdir -p /data/zbx/datakit
chown zabbix:zabbix -R /data/zbx/datakit
chmod u+rw -R /data/zbx/datakit/
```

Attention: When the size of the configuration file is small, it is measured based on the host configuration. Files that are too large can easily cause insufficient disk space. And the `.old` files should be deleted regularly.

Restart server:

```shell
systemctl restart zabbix-server
```

## Config {#config}

Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

```toml
   {{ CodeBlock .InputSample 4 }}
```

Restart DataKit server.

## Docs {#documents}

- Zabbix Documentations: [5.0 real_time_export](https://www.zabbix.com/documentation/5.0/en/manual/appendix/install/real_time_export?hl=export){:target="_blank"}
- [6.0 real_time_export](https://www.zabbix.com/documentation/6.0/en/manual/appendix/install/real_time_export){:target="_blank"}
- [7.0 real_time_export](https://www.zabbix.com/documentation/current/en/manual/config/export/files){:target="_blank"}
