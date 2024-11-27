---
title     : 'Zabbix 数据接入'
summary   : 'Zabbix realTime data 数据接入'
tags:
  - '外部数据接入'
__int_icon      : 'icon/zabbix'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

[:octicons-tag-24: Version-1.37.0](../datakit/changelog.md#cl-1.37.0)

---

采集 Zabbix 服务的实时数据并发送到观测云中心。

Zabbix 从 5.0 到 7.0 版本都支持将实时数据写到文件中。实时数据中有三种数据格式：`events/history/trends` ，其中 `history` 和 `trends` 都是以指标形式展示。而 `events` 则可以通过 [Webhook](https://www.zabbix.com/documentation/5.4/en/manual/config/notifications/media/webhook?hl=Webhook%2Cwebhook){:target="_blank"} 方式发送到观测云。

## 配置 {#config}

### 前置条件 {#requirements}

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

### 采集器配置 {#input-config}

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
   {{ CodeBlock .InputSample 4 }}
```

配置注意：

1. 由于采集文件中的数据，所以一定是本机， zabbix server 就应该配置成 `localhost`
2. `measurement_config_dir` 是 yaml 配置
3. `objects`: zabbix 导出的数据共三种，现在仅 item 接入最完善
4. `crontab`: 必填字段，全量更新数据的时间表达式
5. `mysql`: 必填字段，需要从数据库中获取全量的 item 表数据

配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

## 数据缓存 {#cache}

1 从 mysql 中将 items 表全盘取出，存入到内存中。从 zabbix 导出数据后需要用 item id 从 items 中查询，所以 全表数据使用 `map[itemid]itemc` 进行存储

```text
mysql> select itemid,name,type,key_,hostid,units from items where itemid=29167;
+--------+--------------------+--------+-----------------------------+--------+-------+
| itemid | name               |   type | key_                        | hostid | units |
+--------+--------------------+--------+-----------------------------+--------+-------+
|  29167 | CPU interrupt time |   4    | system.cpu.util[,interrupt] |  10084 | %     |
+--------+--------------------+--------+-----------------------------+--------+-------+
```

type 有一个映射表。使用 item_type 作为 tag 的 key.

2 从 mysql 中取出 interface 表存入内存中。

```text
mysql> select * from zabbix.interface;
+-------------+--------+------+------+-------+---------------+-----+-------+
| interfaceid | hostid | main | type | useip | ip            | dns | port  |
+-------------+--------+------+------+-------+---------------+-----+-------+
|           1 |  10084 |    1 |    1 |     1 | 127.0.0.1     |     | 10050 |
|           2 |  10438 |    1 |    1 |     1 | 10.200.14.226 |     | 10050 |
+-------------+--------+------+------+-------+---------------+-----+-------+

字段含义
main 0: 默认网卡， 1: 非默认网卡
type 1: "Agent", 2: "SNMP", 3: "IPMI", 4: "JMX",
```

根据 host id 可以将 ip 取出，所以 interface 表的存储格式是 `map[hostid]interfaceC`

type 同样也有一个映射表。

3 measurement 缓存

去 `measurement_config_dir` 目录读取所有 yaml 结尾的文件，加载到内存中。

## 数据组装 {#Assembly}

***以下以 itemid = 29167 为例***

从 exporter file 取到的一行数据 如下：

```json
{"host":{"host":"Zabbix server","name":"Zabbix server"},"groups":["Zabbix servers"],"applications":["CPU"],"itemid":29167,"name":"CPU interrupt time","clock":1728611707,"ns":570308079,"value":0.000000,"type":0}
```

从 items 表中取出的数据：

```text
itemid , name , key_ , units
29167,CPU interrupt time,"system.cpu.util[,interrupt]",%
```

组装步骤：

通过 item id 去 items 表（已经缓存到内存中）中查询获取 name，key_,units，这时候： name="CPU interrupt time" key_ = "system.cpu.util[,interrupt]" units=%(百分比格式)

再从 measurement 表中取出该 name 的数据为：

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

根据 item 表中 key（去掉中括号）后的字符串，从 measurement 找到了这个数据。那么根据中括号中的对应关系 得知以下 tags： `cpu=""` , `type="interrupt"` , `mode="avg1"` , `logical_or_physical="logical"`

最后：该指标的最终形态为：

```text
指标集 zabbix-server
指标名：system_cpu_util 值为 0.000000 所拥有的标签有：

cpu=""
type="interrupt"
mode="avg1"
logical_or_physical="logical"
measurement="System"
applications="CPU"
groups="Zabbix servers"
host="Zabbix server"
item_type = "Zabbix agent"
interface_type = "Agent"
interface_main = "default"
ip = "127.0.0.1"
hostname="Zabbix server"
time=1728611707570308079
```

## Zabbix 服务 API {#api}

通过 zabbix 暴露的 api 接口请求，获取数据。

第一步，先通过登录接口获取 token 。token 是后续一切请求的认证。

第二步，通过 `itemid` 获取该 item 的 `name`,`key_`,`type`,`hostid`,`unit` 信息，返回的数据都是 string 格式。需要将他们转成对应的格式。

注意： 返回的格式并不是**固定**的，如果返回的是一个 item，就是 string 格式，如果请求多个则是数组格式。

## 参考文档 {#documents}

- 官方配置文档：[5.0 配置数据导出](https://www.zabbix.com/documentation/5.0/en/manual/appendix/install/real_time_export?hl=export){:target="_blank"}
- [6.0 数据导出](https://www.zabbix.com/documentation/6.0/en/manual/appendix/install/real_time_export){:target="_blank"}
- [7.0 数据导出](https://www.zabbix.com/documentation/current/en/manual/config/export/files){:target="_blank"}
