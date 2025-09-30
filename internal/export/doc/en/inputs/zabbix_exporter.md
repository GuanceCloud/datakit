---
title     : 'Zabbix Export'
summary   : 'Zabbix real-time data exporter'
tags:
  - 'THIRD PARTY'
__int_icon      : 'icon/zabbix'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

[:octicons-tag-24: Version-1.37.0](../datakit/changelog.md#cl-1.37.0)

---

Collect real-time data from the Zabbix service and send it to the center. Currently, Zabbix supports writing real-time data to files from version 4.0 to 7.0.
ExportType allows to specify which entity types (events, history, trends) will be exported.

## Config {#config}

### Requirements Config {#requirements}

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

### DataKit Config {#config}

Go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

```toml
   {{ CodeBlock .InputSample 4 }}
```

Precautions when configuring:

1. Since the data in the collection file is collected, it must be the local machine, and the Zabbix server should be configured as `localhost`.
2. `measurement_config_dir` is a YAML configuration.
3. `objects`: There are three types of data exported by Zabbix, and currently only the item access is the most complete.
4. `crontab`: A required field, the time expression for full - scale data update.
5. `mysql`: A required field, full-scale item table data needs to be obtained from the database.
6. `module_version`: A required field. For versions from 5.0 to 7.0+ it is `v5`, and when the Zabbix version is lower than 5.0, it is `v4`.

Restart DataKit server.

## Data Caching {#cache}

1 Retrieve the entire `items` table from MySQL and store it in memory. After exporting data from Zabbix, you need to query the `items` table using the item ID. Therefore, the entire table is stored using `map[itemid]itemc`.

```text
mysql> select itemid, name, type, key_, hostid, units from items where itemid=29167;
+--------+--------------------+--------+-----------------------------+--------+-------+
| itemid | name               | type   | key_                        | hostid | units |
+--------+--------------------+--------+-----------------------------+--------+-------+
|  29167 | CPU interrupt time | 4      | system.cpu.util[,interrupt] |  10084 | %     |
+--------+--------------------+--------+-----------------------------+--------+-------+
```

There is a mapping table for `type`, using `item_type` as the key for the tag.

2 Retrieve the `interface` table from MySQL and store it in memory.

```text
mysql> select * from zabbix.interface;
+-------------+--------+------+------+-------+---------------+-----+-------+
| interfaceid | hostid | main | type | useip | ip            | dns | port  |
+-------------+--------+------+------+-------+---------------+-----+-------+
|           1 |  10084 |    1 |    1 |     1 | 127.0.0.1     |     | 10050 |
|           2 |  10438 |    1 |    1 |     1 | 10.200.14.226 |     | 10050 |
+-------------+--------+------+------+-------+---------------+-----+-------+

Field Meanings:
main 0: Default network interface, 1: Non-default network interface
type 1: "Agent", 2: "SNMP", 3: "IPMI", 4: "JMX",
```

The IP can be retrieved based on the host ID, so the storage format for the `interface` table is `map[hostid]interfaceC`.

There is also a mapping table for `type`.

3 Measurement Caching

Read all files ending in `.yaml` from the `measurement_config_dir` directory and load them into memory.

## Data Assembly {#Assembly}

***Using itemid = 29167 as an example***

A line of data retrieved from the exporter file is as follows:

```json
{"host":{"host":"Zabbix server","name":"Zabbix server"},"groups":["Zabbix servers"],"applications":["CPU"],"itemid":29167,"name":"CPU interrupt time","clock":1728611707,"ns":570308079,"value":0.000000,"type":0}
```

Data retrieved from the `items` table:

```text
itemid, name, key_, units
29167, CPU interrupt time, "system.cpu.util[,interrupt]", %
```

Assembly Steps:

Use the item ID to query the `items` table (already cached in memory) to obtain `name`, `key_`, and `units`. At this point: `name="CPU interrupt time"`, `key_ = "system.cpu.util[,interrupt]"`, `units=%` (percentage format).

Then, retrieve the data for this `name` from the measurement table:

```yaml
- measurement: System
  metric: system_cpu_util
  key: system.cpu.util
  params:
    - cpu
    - type
    - mode
    - logical_or_physical
  values:
    - ''
    - user
    - avg1
    - logical
```

Based on the string after the key in the `items` table (brackets removed), the data is found in the measurement. The corresponding relationship in the brackets indicates the following tags: `cpu=""`, `type="interrupt"`, `mode="avg1"`, `logical_or_physical="logical"`.

Finally, the final form of the metric is:

```text
Metric Set: zabbix-server
Metric Name: system_cpu_util
Value: 0.000000
Tags:

cpu=""
type="interrupt"
mode="avg1"
logical_or_physical="logical"
measurement="System"
applications="CPU"
groups="Zabbix servers"
host="Zabbix server"
item_type="Zabbix agent"
interface_type="Agent"
interface_main="default"
ip="127.0.0.1"
hostname="Zabbix server"
time=1728611707570308079
```

## Zabbix Service API {#zabbix_api}

Obtain data through the API interfaces exposed by Zabbix.

First, obtain a token through the login interface. The token is used for authentication in subsequent requests.

Second, use the `itemid` to retrieve the `name`, `key_`, `type`, `hostid`, and `unit` information for that item. The returned data is in string format and needs to be converted to the corresponding format.

Note: The returned format is not **fixed**. If a single item is returned, it is in string format; if multiple items are requested, it is in array format.


## Docs {#documents}

- Zabbix Documentations: [5.0 real_time_export](https://www.zabbix.com/documentation/5.0/en/manual/appendix/install/real_time_export?hl=export){:target="_blank"}
- [6.0 real_time_export](https://www.zabbix.com/documentation/6.0/en/manual/appendix/install/real_time_export){:target="_blank"}
- [7.0 real_time_export](https://www.zabbix.com/documentation/current/en/manual/config/export/files){:target="_blank"}
