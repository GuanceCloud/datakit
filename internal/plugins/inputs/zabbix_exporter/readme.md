# Zabbix exporter 采集器

采集 Zabbix 导出的数据到中心。

## 配置

```toml
[[inputs.zabbix_exporter]]
  ## zabbix server web.
  localhostAddr = "http://localhost/zabbix/api_jsonrpc.php"
  user_name = "Admin"
  user_pw = "zabbix"
  
  ## measurement yaml Dir
  measurement_config_dir = "/data/zbx/yaml"

  ## exporting object.default is item. all is <trigger,item,trends>. 
  objects = "item"

  ## update items and interface data.
  ## like this: All data is updated at 2 o'clock every day.
  crontab = "0 2 * * *"

  # [inputs.zabbix_exporter.tags]
    # key1 = "value1"
    # key2 = "value2"
    # ...

  [inputs.zabbix_exporter.mysql]
    db_host = "192.168.10.12"
    db_port = "3306"
    user = "root"
    pw = "123456"

  # Zabbix server version 5.x.
  [inputs.zabbix_exporter.export_v5]
    # zabbix realTime exportDir path
    export_dir = "/data/zbx/datakit/"

```

配置注意：

1. 由于采集文件中的数据，所以一定是本机， zabbix server 就应该配置成 `localhost`
2. measurement_config_dir 是 yaml 配置
3. objects: zabbix 导出的数据共三种，现在仅 item 接入最完善
4. crontab: 全量更新数据的时间表达式
5. mysql: 必填字段，需要从数据库中获取全量的item表数据


## 数据缓存

1 从 mysql 中将 items 表全盘取出，存入到内存中。从 zabbix 导出数据后需要用 item id 从 items 中查询，所以 全表数据使用 `map[itemid]itemc` 进行存储

```text
mysql> select itemid,name,type,key_,hostid,units from items where itemid=29167;
+--------+--------------------+--------+-----------------------------+--------+-------+
| itemid | name               |   type | key_                        | hostid | units |
+--------+--------------------+--------+-----------------------------+--------+-------+
|  29167 | CPU interrupt time |   4    | system.cpu.util[,interrupt] |  10084 | %     |
+--------+--------------------+--------+-----------------------------+--------+-------+
```

type 有一个映射表。使用 item_type 作为tag的key

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
main 0: 默认网卡, 1: 非默认网卡
type 1: "Agent", 2: "SNMP", 3: "IPMI", 4: "JMX",
```

根据 host id 可以将 ip 取出，所以 interface 表的存储格式是 `map[hostid]interfaceC`

type 同样也有一个映射表。

3 measurement 缓存

去 `measurement_config_dir` 目录读取所有 yaml 结尾的文件，加载到内存中。

## 数据组装

***以下以 itemid = 29167 为例***

从exporter file取到的一行数据 如下：

```json
{"host":{"host":"Zabbix server","name":"Zabbix server"},"groups":["Zabbix servers"],"applications":["CPU"],"itemid":29167,"name":"CPU interrupt time","clock":1728611707,"ns":570308079,"value":0.000000,"type":0}
```

从items表中取出的数据：

```text
itemid , name , key_ , units
29167,CPU interrupt time,"system.cpu.util[,interrupt]",%
```

组装步骤：

通过 itemid 去items表（已经缓存到内存中）中查询获取 name，key_,units，这时候： name="CPU interrupt time" key_ = "system.cpu.util[,interrupt]" units=%(百分比格式)

再从 measurement 表中取出该name的数据为：
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

根据 item 表中 key（去掉中括号）后的字符串，从measurement找到了这个数据。那么根据中括号中的对应关系 得知以下 tags： `cpu=""` , `type="interrupt"` , `mode="avg1"` , `logical_or_physical="logical"`

最后：该指标的最终形态为：

```text
指标集 zabbix_server
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

## Zabbix 服务 API

通过 zabbix 暴露的 api 接口请求，获取数据。

第一步，先通过登录接口获取 token 。token 是后续一切请求的认证。

第二步，通过 itemid 获取该 item 的 name,key_,type,hostid,unit 信息，返回的数据都是string格式。需要将他们转成对应的格式。

注意： 返回的格式并不是**固定**的，如果返回的是一个item，就是string格式，如果请求多个则是数组格式。


## 采集指标

添加采集指标，采集数量 文件数量 api次数：
```text
# HELP datakit_input_zabbix_exporter_collect_file_total The files number of exporter file
# TYPE datakit_input_zabbix_exporter_collect_file_total counter
# 5个文件，此处 Object 有： item，trends，为采集的类型
datakit_input_zabbix_exporter_collect_file_total{object="item"} 5
# HELP datakit_input_zabbix_exporter_collect_metric_total exporter metric count number from start
# TYPE datakit_input_zabbix_exporter_collect_metric_total counter
# 从启动开始的指标数数量，按object 区分。
datakit_input_zabbix_exporter_collect_metric_total{object="item"} 3237

# api次数 status:success 和 failed
datakit_input_zabbix_exporter_request_api_total{status="success"} 1
```

## 问题汇总

- 观测云中心不允许指标数据的值为string，也就是说 如果有些指标的value的type类型为string，直接pass不做处理。
- 以一个`system_cpu_util`为例，items 表中的 id 就有 125 个之多。这个指标在观测云中心如果按照 tag 来分，能分出125条线（理论上可能更多）
- item 中 key_ 名字有很多重复，但包含的 tag 又不同，比如 `zabbix[xx,xx] zabbix[xx]`  就一个 zabbix 指标就有很多种。
- misskey 的时候，会查询一次api，如果misskey的比例很高，且查不到，就会浪费api资源。应该做速率限制。


