---
title     : 'SNMP'
summary   : '采集 SNMP 设备的指标和对象数据'
tags:
  - 'SNMP'
__int_icon      : 'icon/snmp'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

{{.AvailableArchs}}

---

本文主要介绍 [SNMP](https://en.wikipedia.org/wiki/Simple_Network_Management_Protocol){:target="_blank"} 数据采集。

## 术语  {#terminology}

- `SNMP` (Simple network management protocol): 用于收集有关裸机网络设备信息的网络协议。
- `OID` (Object identifier): 设备上的唯一 ID 或地址，轮询时返回该值的响应代码。例如，OID 是 CPU 或设备风扇速度。
- `sysOID` (System object identifier): 定义设备类型的特定地址。所有设备都有一个定义它的唯一 ID。例如，`Meraki` 基础 sysOID 是“1.3.6.1.4.1.29671”。
- `MIB` (Managed information base): 与 MIB 相关的所有可能的 OID 及其定义的数据库或列表。例如，“IF-MIB”（接口 MIB）包含有关设备接口的描述性信息的所有 OID。

## 关于 SNMP 协议 {#config-pre}

SNMP 协议分为 3 个版本：v1/v2c/v3，其中：

- v1 和 v2c 是兼容的，很多 SNMP 设备只提供 v2c 和 v3 两种版本的选择。v2c 版本，兼容性最好，很多旧设备只支持这个版本；
- 如果对安全性要求高，选用 v3。安全性也是 v3 版本与之前版本的主要区别；

DataKit 支持以上所有版本。

### 选择 v1/v2c 版本 {#config-v2}

如果选择 v1/v2c 版本，需要提供 `v2_community_string`（团体名/团体字符串/未加密的口令），与 SNMP 设备进行交互需要提供这个进行鉴权。另外，有的设备会进一步进行细分，分为*只读团体名*和*读写团体名*：顾名思义：

- 只读团体名：设备只会向该方提供内部指标数据，不能修改内部的一些配置（DataKit 用这个就够了）
- 读写团体名：提供方拥有设备内部指标数据查询与部分配置修改权限

### 选择 v3 版本 {#config-v3}

如果选择 v3 版本，需要提供 `v3_user/v3_auth_protocol/v3_auth_key/v3_priv_protocol/v3_priv_key` 等，各个设备要求不同，根据设备侧的配置进行填写。

## 配置 {#config}

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/samples` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

### 多种配置格式 {#configuration-formats}

#### 内置 Profile 格式 {#advanced-custom-oid}

内置 Profile 使用 YAML 描述设备的 `sysObjectID`、采集 OID、指标类型、标签和元数据。

该格式适合以下场景：

- 内置 Profile 未覆盖当前设备型号；
- 需要增加厂商私有 MIB 指标；
- 设备将数值以 String/OCTET STRING 形式返回；
- 需要补充设备元数据。

Profile YAML 与 Zabbix Template、Prometheus `snmp_exporter` 格式不同，字段不能混用。

##### 准备 OID 信息 {#profile-prepare-oid}

编写 Profile 前，应先获取设备 MIB/OID 手册，并使用 `snmpget` 或 `snmpwalk` 确认设备的实际返回结果。至少需要确认：

- `sysObjectID`，即 OID `1.3.6.1.2.1.1.2.0` 的值；
- 指标是标量还是表格列；
- 指标 OID、数据类型、含义和单位；
- String/OCTET STRING 的完整返回格式；
- 表格索引和能够区分每一行的标签列。

##### 新增 Profile {#profile-add}

DataKit 会将内置 Profile 释放到安装目录的 `conf.d/snmp/profiles/`，该目录可能在启动或升级时被覆盖。用户新增或覆盖内置 Profile 时，建议将 YAML 文件放入 `conf.d/snmp/extra_profiles/`，DataKit 会在默认 Profile 加载流程中合并这两个目录，且 `extra_profiles` 优先级更高。

示例：

```text
/usr/local/datakit/conf.d/snmp/extra_profiles/vendor-router.yaml
```

文件要求：

- 扩展名必须为 `.yaml`；
- 文件名不能以 `_` 开头，以 `_` 开头的文件只作为继承模板；
- 与内置 Profile 同名时，`extra_profiles` 中的文件优先；如仍需继承原内置内容，可在该文件中 `extends` 同名文件；
- 不同文件名不应配置相同的 `sysobjectid`，否则自动匹配时会按重复 Profile 处理。

新增后重启 DataKit。

##### 基本结构 {#profile-structure}

```yaml
extends:
  - generic-router.yaml

sysobjectid: 1.3.6.1.4.1.<enterprise_id>.<product_id>

device:
  vendor: vendor_name

static_tags:
  - "device_type:router"

metadata:
  device:
    fields:
      model:
        symbol:
          OID: 1.3.6.1.4.1.<enterprise_id>.1.1.0
          name: vendorModel

metric_tags:
  - symbol:
      OID: 1.3.6.1.2.1.1.5.0
      name: sysName
    tag: snmp_host

metrics:
  - MIB: VENDOR-MIB
    symbol:
      OID: 1.3.6.1.4.1.<enterprise_id>.2.1.0
      name: vendor.system.cpu_usage
      metric_type: gauge
```

根级字段说明：

| 字段 | 是否必需 | 说明 |
| --- | --- | --- |
| `extends` | 否 | 继承其他 Profile，常用于复用通用指标和元数据 |
| `sysobjectid` | 自动匹配时必需 | 一个字符串或字符串列表，支持 `*` 通配符 |
| `device.vendor` | 否 | 生成 `device_vendor` 标签，并作为设备厂商元数据的回退值 |
| `static_tags` | 否 | 应用于 Profile 数据的固定 `key:value` 标签 |
| `metadata` | 否 | 设备元数据定义 |
| `metric_tags` | 否 | 从标量 OID 生成 Profile 级动态标签 |
| `metrics` | 否 | 标量和表格指标定义 |

`MIB` 和 `table` 用于提高 YAML 可读性，不决定实际 SNMP 查询。DataKit 实际查询的是 `symbol`、`symbols`、`metric_tags` 和 `metadata` 中配置的 OID。

##### 继承 Profile {#profile-inheritance}

`extends` 用于继承已有 Profile。被继承 Profile 中的指标、标签和元数据会合并到当前 Profile，可以减少标准 MIB 的重复配置。

```yaml
extends:
  - _base.yaml
  - _generic-if.yaml
```

DataKit 已安装的内置 Profile 位于 `conf.d/snmp/profiles/`，用户补充 Profile 位于 `conf.d/snmp/extra_profiles/`。编写 YAML 前，建议先查看内置目录中以 `_` 开头的文件，并根据当前 DataKit 版本选择可复用的模板。常用模板包括：

| Profile | 用途 |
| --- | --- |
| `_base.yaml` | 通用设备标签和基础设备元数据 |
| `_generic-if.yaml` | IF-MIB 接口指标和接口元数据 |
| `_generic-ip.yaml` | IP-MIB 指标 |
| `_generic-tcp.yaml` | TCP-MIB 指标 |
| `_generic-udp.yaml` | UDP-MIB 指标 |
| `_generic-ospf.yaml` | OSPF-MIB 指标 |
| `_generic-bgp4.yaml` | BGP4-MIB 指标 |
| `_generic-lldp.yaml` | LLDP 相关指标和元数据 |
| `_generic-entity-sensor.yaml` | ENTITY-SENSOR-MIB 传感器指标 |
| `_generic-host-resources.yaml` | HOST-RESOURCES-MIB 主机资源指标 |
| `_generic-ups.yaml` | 通用 UPS 指标 |
| `_cisco-generic.yaml` | Cisco 通用接口、IP、TCP、UDP、OSPF、BGP、CPU、内存和元数据 |
| `_huawei.yaml` | Huawei 通用接口和厂商元数据 |
| `_juniper.yaml` | Juniper 通用基础配置和厂商元数据 |

以 `_` 开头的文件只作为继承模板，不独立参与 `sysObjectID` 自动匹配。也可以继承不以 `_` 开头的完整 Profile，例如 `generic-router.yaml`：

```yaml
extends:
  - generic-router.yaml
```

`generic-router.yaml` 已经继承 `_base.yaml`、`_generic-if.yaml`、`_generic-ip.yaml`、`_generic-tcp.yaml`、`_generic-udp.yaml` 和 `_generic-ospf.yaml`。因此继承它以后，不需要再次显式继承这些文件。

Cisco 设备可以继承通用 Cisco 模板，并追加设备专属指标：

```yaml
extends:
  - _base.yaml
  - _cisco-generic.yaml

sysobjectid: 1.3.6.1.4.1.9.1.<product_id>
```

注意：

- `extends` 中的文件名会优先从 `conf.d/snmp/extra_profiles/` 查找，再从 `conf.d/snmp/profiles/` 查找；
- `extra_profiles` 中的同名 Profile 可以通过 `extends` 自身文件名继承内置同名 Profile；
- 支持继承多个 Profile 和多层继承，但不能循环继承；
- 指标、动态标签和静态标签采用追加方式合并；
- 当前 Profile 已定义的同名元数据字段不会被继承内容覆盖；
- 不要重复继承已经被上层 Profile 包含的模板，否则可能产生重复指标或标签；
- 继承的通用模块越多，需要查询的 OID 越多，应按设备实际支持的 MIB 选择。

##### sysObjectID 匹配 {#profile-sysobjectid}

精确匹配：

```yaml
sysobjectid: 1.3.6.1.4.1.99999.1.2
```

通配匹配：

```yaml
sysobjectid: 1.3.6.1.4.1.99999.*
```

匹配多个型号：

```yaml
sysobjectid:
  - 1.3.6.1.4.1.99999.1.*
  - 1.3.6.1.4.1.99999.2.*
```

多个 Profile 同时匹配时，DataKit 选择更具体的规则。不同 Profile 不应配置完全相同的 `sysobjectid`。应尽量使用厂商或产品专属值，避免宽泛规则误匹配其他设备。

例如 `1.3.6.1.4.1.8072.3.2.10` 是常见的 Net-SNMP Linux `sysObjectID`，并非某个设备厂商独占。

##### 标量指标 {#profile-scalar-metrics}

标量只有一个值，使用 `symbol` 定义。标量 OID 通常以 `.0` 结尾：

```yaml
metrics:
  - MIB: VENDOR-SYSTEM-MIB
    symbol:
      OID: 1.3.6.1.4.1.99999.1.2.0
      name: vendor.system.cpu_usage
      metric_type: gauge
```

##### 表格指标 {#profile-table-metrics}

表格有多行数据，使用 `symbols` 定义需要采集的列，并使用 `metric_tags` 区分每一行：

```yaml
metrics:
  - MIB: IF-MIB
    table:
      OID: 1.3.6.1.2.1.2.2
      name: ifTable
    symbols:
      - OID: 1.3.6.1.2.1.2.2.1.10
        name: vendor.interface.in_octets
        metric_type: monotonic_count
      - OID: 1.3.6.1.2.1.2.2.1.16
        name: vendor.interface.out_octets
        metric_type: monotonic_count
    metric_tags:
      - symbol:
          OID: 1.3.6.1.2.1.31.1.1.1.1
          name: ifName
        tag: interface
```

表格必须至少配置一个能够区分行的标签，否则多行指标可能拥有相同标签，最终只能保留一行。

也可以从 OID 行索引中生成标签。`index` 从 `1` 开始：

```yaml
metric_tags:
  - index: 1
    tag: disk_index
```

如果完整索引为 `3.24`，以下配置会生成 `slot:3` 和 `port:24`：

```yaml
metric_tags:
  - index: 1
    tag: slot
  - index: 2
    tag: port
```

##### String 转数值 {#profile-string-to-number}

指标字段值需要是数值类型。纯数字 String/OCTET STRING（如 `"52.20"`）可直接转换；值中包含单位或其他字符时，使用 `extract_value` 提取第一个捕获组。

设备返回 `STRING: "5331MB"`：

```yaml
- OID: 1.3.6.1.4.1.99999.2.1.4
  name: vendor.disk.used
  extract_value: '^([0-9]+[.]?[0-9]*)MB$'
  metric_type: gauge
```

设备返回 `STRING: "52.20%"`：

```yaml
- OID: 1.3.6.1.4.1.99999.2.1.6
  name: vendor.disk.used_percent
  extract_value: '^([0-9]+[.]?[0-9]*)%?$'
  metric_type: gauge
```

注意：

- 正则必须至少包含一个捕获组，DataKit 只使用第一个捕获组；
- 建议使用单引号包裹正则，并使用 `^` 和 `$` 完整匹配；
- 使用 Go 正则语法，不支持 lookahead 和 lookbehind；
- 正则不匹配时，当前值不会上报；
- 普通文本 String 不应作为时序指标值，应配置为标签或元数据。

##### 指标类型和数值换算 {#profile-metric-type}

自定义 Profile 推荐使用以下 `metric_type`：

| 类型 | 适用场景 |
| --- | --- |
| `gauge` | 使用率、温度、容量、连接数等可增可减的当前值 |
| `monotonic_count` | 字节数、包数、请求数等单调递增累计值 |

可以在具体 `symbol` 中配置，也可以在表格指标根级配置并应用于该表格的所有 `symbols`。具体 `symbol` 中的配置优先级更高。不配置时，DataKit 根据 SNMP 返回类型推断，无法推断时按 `gauge` 处理。

`scale_factor` 用于数值换算。例如设备返回 `5234`，实际值为 `52.34%`：

```yaml
- OID: 1.3.6.1.4.1.99999.1.2.0
  name: vendor.system.cpu_usage
  scale_factor: 0.01
  metric_type: gauge
```

`scale_factor` 只作用于最终上报的指标数值。String 指标可先使用 `extract_value` 提取数值，再应用 `scale_factor`。

##### 固定上报数值 1 {#profile-constant-value-one}

表格表示一组实体，但没有合适的数值列时，可使用 `constant_value_one` 为每一行固定上报 `1`：

```yaml
metrics:
  - MIB: VENDOR-DISK-MIB
    table:
      OID: 1.3.6.1.4.1.99999.3.1
      name: vendorDiskTable
    symbols:
      - name: vendor.disk.present
        constant_value_one: true
    metric_tags:
      - symbol:
          OID: 1.3.6.1.4.1.99999.3.1.1
          name: diskName
        tag: disk_name
      - symbol:
          OID: 1.3.6.1.4.1.99999.3.1.2
          name: diskState
        tag: disk_state
        mapping:
          1: normal
          2: warning
          3: failed
```

`constant_value_one` 只能用于表格 `symbols`，不能用于标量。使用时不填写指标 OID，但必须配置 `name`，并至少配置一个不含 `index_transform`、带 OID 的 `metric_tags.symbol`，以便 DataKit 发现表格行。

##### 标签 {#profile-tags}

Profile 根级 `static_tags` 用于增加固定标签：

```yaml
static_tags:
  - "device_type:router"
  - "environment:production"
```

Profile 根级 `metric_tags` 从标量 OID 生成应用于设备指标的动态标签：

```yaml
metric_tags:
  - symbol:
      OID: 1.3.6.1.2.1.1.5.0
      name: sysName
    tag: snmp_host
```

表格内的 `metric_tags` 可以使用 `mapping` 将原始值转换为可读标签：

```yaml
metric_tags:
  - symbol:
      OID: 1.3.6.1.4.1.99999.3.1.2
      name: diskState
    tag: disk_state
    mapping:
      1: normal
      2: warning
      3: failed
```

使用 `match` 和 `tags` 可从一个表格列值生成多个标签：

```yaml
metric_tags:
  - symbol:
      OID: 1.3.6.1.4.1.99999.4.1.2
      name: interfaceLabel
    match: '^([A-Za-z]+)-([0-9]+)$'
    tags:
      interface_type: '$1'
      interface_number: '$2'
```

原始值为 `ethernet-12` 时，将生成 `interface_type:ethernet` 和 `interface_number:12`。配置 `match` 时必须同时配置非空的 `tags`，正则不匹配时不会生成标签。

当指标表和标签来源表的索引结构不一致时，可以使用 `index_transform` 截取并重组索引：

```yaml
metric_tags:
  - symbol:
      OID: 1.3.6.1.4.1.99999.5.1.2
      name: parentName
    index_transform:
      - start: 1
        end: 2
      - start: 6
        end: 7
    tag: parent_name
```

如果当前索引为 `1.2.3.4.5.6.7.8`，转换结果为 `2.3.7.8`。`start` 和 `end` 都从 `0` 开始，且 `end` 所在位置包含在结果中。只有在跨表关联时才需要使用该配置。

##### 元数据 {#profile-metadata}

自定义 Profile 可以通过 `metadata.device` 补充设备元数据，支持 `name`、`description`、`sys_object_id`、`location`、`serial_number`、`vendor`、`version`、`product_name`、`model`、`os_name`、`os_version`、`os_hostname` 和 `type` 字段。字段值可以来自标量 OID，也可以使用固定 `value`：

```yaml
metadata:
  device:
    fields:
      name:
        symbol:
          OID: 1.3.6.1.2.1.1.5.0
          name: sysName
      serial_number:
        symbol:
          OID: 1.3.6.1.4.1.99999.1.1.0
          name: vendorSerialNumber
      vendor:
        value: vendor_name
      type:
        value: router
```

一个字段可配置多个候选 `symbols`，DataKit 按顺序使用第一个能够取得值的 OID。`match_pattern` 和 `match_value` 可从文本中提取或替换内容：

```yaml
metadata:
  device:
    fields:
      model:
        symbols:
          - OID: 1.3.6.1.4.1.99999.1.2.0
            name: vendorModel
          - OID: 1.3.6.1.2.1.1.1.0
            name: sysDescr
            match_pattern: 'Model[=: ]+([A-Za-z0-9._-]+)'
            match_value: '$1'
```

省略 `match_value` 时默认使用第一个捕获组 `$1`；正则不匹配时，DataKit 尝试下一个候选 `symbol`。

继承 `_generic-if.yaml` 或 `generic-router.yaml` 后，通用接口元数据已经包含在继承内容中，通常不需要重复配置。

##### 指标命名 {#profile-metric-naming}

指标名按以下规则转换后上报：

1. 名称中包含下划线 `_` 时，将所有点号 `.` 替换为下划线 `_`；
1. 名称中不包含下划线 `_` 时，删除点号 `.`，并将点号后第一个字母转换为大写；
1. 名称中没有点号 `.` 时，保持不变。

示例：

| Profile 中的 `name` | 最终上报字段名 |
| --- | --- |
| `sangfor.disk.used` | `sangforDiskUsed` |
| `sangfor.disk.used_percent` | `sangfor_disk_used_percent` |
| `vendor.interface.in_octets` | `vendor_interface_in_octets` |
| `cpu_usage` | `cpu_usage` |

该转换也适用于标签键，但不会修改标签值。

同一个 Profile 应统一命名风格。可以使用不含下划线的点号分层名称，例如 `vendor.disk.used`；也可以直接使用不含点号的下划线名称，例如 `vendor_disk_used`。避免混用点号和下划线，因为点号最终会转换为下划线。现有内置 Profile 常使用 `huawei.hwEntityTemperature` 这类名称，最终上报为 `huaweiHwEntityTemperature`。

##### 完整示例 {#profile-example}

以下示例继承通用路由器指标，采集设备元数据、数值标量、String 类型磁盘指标和实体存在性指标：

```yaml
extends:
  - generic-router.yaml

sysobjectid: 1.3.6.1.4.1.99999.1.2

device:
  vendor: vendor_name

metadata:
  device:
    fields:
      model:
        symbol:
          OID: 1.3.6.1.4.1.99999.1.1.0
          name: vendorModel

metrics:
  - MIB: VENDOR-SYSTEM-MIB
    symbol:
      OID: 1.3.6.1.4.1.99999.1.2.0
      name: vendor.system.cpu_usage
      metric_type: gauge

  - MIB: VENDOR-DISK-MIB
    table:
      OID: 1.3.6.1.4.1.99999.2.1
      name: vendorDiskTable
    symbols:
      - OID: 1.3.6.1.4.1.99999.2.1.4
        name: vendor.disk.used
        extract_value: '^([0-9]+[.]?[0-9]*)MB$'
        metric_type: gauge
      - OID: 1.3.6.1.4.1.99999.2.1.6
        name: vendor.disk.used_percent
        extract_value: '^([0-9]+[.]?[0-9]*)%?$'
        metric_type: gauge
      - name: vendor.disk.present
        constant_value_one: true
    metric_tags:
      - index: 1
        tag: disk_index
      - symbol:
          OID: 1.3.6.1.4.1.99999.2.1.2
          name: diskName
        tag: disk_name
```

##### 验证和排错 {#profile-troubleshooting}

常见问题：

- Profile 未匹配：检查设备实际 `sysObjectID`、通配符范围以及是否存在更具体的匹配规则；
- YAML 未加载：检查文件扩展名、文件名、缩进和 DataKit 日志中的 Profile 校验错误；
- OID 无数据：使用与 DataKit 相同的 SNMP 版本和认证信息运行 `snmpget`/`snmpwalk`；
- String 指标未上报：检查 `extract_value` 是否包含捕获组并匹配实际完整值；
- 表格只上报一行：为表格增加能够区分每一行的 `metric_tags`；
- 指标被跳过：确认最终值能够转换为数值，并检查 `metric_type` 和 `scale_factor`。

#### Zabbix 格式 {#format-zabbix}

- 配置

    ```toml
      [[inputs.snmp.zabbix_profiles]]
        profile_name = "xxx.yaml"
        ip_list = ["ip1", "ip2"]
        class = "server"

      [[inputs.snmp.zabbix_profiles]]
        profile_name = "yyy.xml"
        ip_list = ["ip3", "ip4"]
        class = "firewall"

      # ...
    ```

    `profile_name` 可以是全路径或只包含文件名，只包含文件名的话，文件要放到 *./conf.d/snmp/userprofiles/* 子目录下。

    您可以去 Zabbix 官方下载对应的的配置，也可以去 [社区](https://github.com/zabbix/community-templates){:target="_blank"} 下载。

    如果您对下载到的 yaml 或 xml 文件不满意，也可以自行修改。

- 自动发现
    - 自动发现在引入的多个 yaml 配置里面匹配采集规则，进行采集。
    - 自动发现请尽量按 C 段配置，配置 B 段可能会慢一些。
    - 万一自动发现匹配不到 yaml ，是因为已有的 yaml 里面没有被采集设备的生产商特征码。
        - 可以在 yaml 的 items 里面人为加入一条 oid 信息，引导自动匹配过程。

          ```yaml
          zabbix_export:
            templates:
            - items:
              - snmp_oid: 1.3.6.1.4.1.2011.5.2.1.1.1.1.6.114.97.100.105.117.115.0.0.0.0
          ```

        - 拟加入的 oid 通过执行以下命令获得，后面加上 .0.0.0.0 是为了防止产生无用的指标。

        ```shell
        $ snmpwalk -v 2c -c public <ip> 1.3.6.1.2.1.1.2.0
        iso.3.6.1.2.1.1.2.0 = OID: iso.3.6.1.4.1.2011.2.240.12

        $ snmpgetnext -v 2c -c public <ip> 1.3.6.1.4.1.2011.2.240.12
        iso.3.6.1.4.1.2011.5.2.1.1.1.1.6.114.97.100.105.117.115 = STRING: "radius"
        ```

#### Prometheus 格式 {#format-Prometheus}

- 配置

    ```toml
      [[inputs.snmp.prom_profiles]]
        profile_name = "xxx.yml"
        ip_list = ["ip1", "ip2"]
        class = "server"

      [[inputs.snmp.prom_profiles]]
        profile_name = "yyy.yml"
        ip_list = ["ip3", "ip4"]
        class = "firewall"

      # ...
    ```

    profile 参考 Prometheus [snmp_exporter](https://github.com/prometheus/snmp_exporter){:target="_blank"} 的 snmp.yml 文件，
    建议把不同 class 的 [module](https://github.com/prometheus/snmp_exporter?tab=readme-ov-file#prometheus-configuration){:target="_blank"} 拆分成  不同 .yml 配置。

    Prometheus 的 profile 允许为 module 单独配置团体名 community，这个团体名优先于采集器配置的团体名。

    ```yml
    switch:
      walk:
      ...
      get:
      ...
      metrics:
      ...
      auth:
        community: xxxxxxxxxxxx
    ```

- 自动发现

    SNMP 采集器支持通过 Consul 服务发现来发现被采集对象，服务注入格式参考 [prom 官网](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#consul_sd_config){:target="_blank"}。

???+ tip

    上述配置完成后，可以使用 `datakit debug --input-conf` 命令来测试配置是否正确，示例如下：

    ```sh
    sudo datakit debug --input-conf /usr/local/datakit/conf.d/snmp/snmp.conf
    ```

    如果正确会输出行协议信息，否则看不到行协议信息。

???+ note

    1. 上面配置的 `inputs.snmp.tags` 中如果与原始 fields 中的 key 同名重复，则会被原始数据覆盖
    1. 设备的 IP 地址(指定设备模式)/网段(自动发现模式)、SNMP 协议的版本号及相对应的鉴权字段是必填字段
    1. 「指定设备」模式和「自动发现」模式，两种模式可以共存，但设备间的 SNMP 协议的版本号及相对应的鉴权字段必须保持一致
<!-- markdownlint-enable MD046 -->

### 配置被采集 SNMP 设备 {#config-snmp}

SNMP 设备在默认情况下，一般 SNMP 协议处于关闭状态，需要进入管理界面手动打开。同时，需要根据实际情况选择协议版本和填写相应信息。

<!-- markdownlint-disable MD046 -->
???+ tip

    有些设备为了安全需要额外配置放行 SNMP，具体因设备而异。比如华为系防火墙，需要在 "启用访问管理" 中勾选 SNMP 以放行。
    可以使用 `snmpwalk` 命令来测试采集侧与设备侧是否配置连通成功（在 DataKit 运行的主机上运行以下命令）：

    ```shell
    # 适用 v2c 版本
    snmpwalk -O bentU -v 2c -c [community string] [SNMP_DEVICE_IP] 1.3.6
    # 适用 v3 版本
    snmpwalk -v 3 -u user -l authPriv -a sha -A [认证密码] -x aes -X [加密密码] [SNMP_DEVICE_IP] 1.3.6
    ```

    如果配置没有问题的话，该命令会输出大量数据。`snmpwalk` 是运行在采集侧的一个测试工具，MacOS 下自带，Linux 安装方法：

    ```shell
    sudo yum install net-snmp net-snmp-utils # CentOS
    sudo apt–get install snmp                # Ubuntu
    ```
<!-- markdownlint-enable MD046 -->

#### SNMPv3 示例 {#snmpv3-example}

我们以一个 Linux 上的 `snmpd` 作为示例，来演示如何采集 SNMP v3 的采集。

- 在一台 Ubuntu 机器上，我们可以安装 snmpd 服务：

    ```shell
    sudo apt install snmp snmpd libsnmp-dev
    ```

- 准备如下一个简易的 *snmpd.conf* 配置：

    ```conf title="my-snmpd.conf"
    # 设定 UDP 161 端口
    agentaddress udp:161

    # 设定用户名以及几个认证相关的配置
    createUser snmpv3user1 SHA "authPassAgent1" AES "privPassAgent1"

    # 授予用户访问权限 (rouser: read-only, rwuser: read-write)
    rouser snmpv3user1 priv .1.3.6
    ```

- 先停掉 `snmpd` 服务，手动启动 snmpd 程序：

    ```shell
    sudo /usr/sbin/snmpd -f -Lo -C \
      -p x.pid -Ddump,usm,acl,header,context,pdu,snmpv3 \
      -c my-snmpd.conf
    ```

- 用 `snmpwalk` 命令检测一下，预期将输出很多 OID 设备信息：

    ```shell
    snmpwalk -v3 -l authPriv \
      -u snmpv3user1 \
      -a SHA \
      -A "authPassAgent1" \
      -x AES \
      -X "privPassAgent1" \
      udp:127.0.0.1:161 .1.3.6.1.2.1.1
    ```

- 如果 `snmpwalk` 命令成功，就可以在 DataKit 上开启本采集器，通过 SNMPv3 来采集指标。关键配置如下：

    ```toml title="conf.d/snmp/snmp.conf"
    specific_devices = ["127.0.0.1"] # 此处不要填 localhost
    snmp_version     = 3
    port             = 161

    v3_user                = "snmpv3user1"
    v3_auth_protocol       = "SHA" # MD5/SHA/SHA224/SHA256/SHA384/SHA512 or empty
    v3_auth_key            = "authPassAgent1"
    v3_priv_protocol       = "AES" # DES/AES/AES192/AES192C/AES256/AES256C or empty
    v3_priv_key            = "privPassAgent1"
    # v3_context_engine_id = "" # optional
    # v3_context_name      = "" # optional
    ```

## LLDP 网络拓扑采集 {#lldp-topology}

DataKit 支持通过 SNMP 协议采集网络设备的 LLDP（Link Layer Discovery Protocol，链路层发现协议）邻居信息，用于自动构建网络拓扑。

### 什么是 LLDP {#what-is-lldp}

LLDP 是一个标准化的链路层协议，允许网络设备（如交换机、路由器）向直连的邻居设备广播自己的身份和能力信息。通过采集 LLDP 数据，可以：

- 自动发现网络拓扑关系
- 了解设备间的物理连接情况
- 获取邻居设备的端口、主机名、系统描述等信息
- 构建可视化的网络拓扑图

### 启用 LLDP 采集 {#enable-lldp}

推荐将 LLDP 邻居作为链路元数据附加到 `snmp_object` 对象：

```toml
[[inputs.snmp]]
  ## 随对象采集 LLDP 邻居，并将拓扑链路写入 snmp_object 的 links 字段
  collect_topology = true
```

`collect_topology` 默认关闭。启用后，DataKit 在对象采集周期内额外查询 LLDP/CDP MIB，并将本地设备、接口以及对端设备、接口等链路元数据编码为 JSON 数组，写入 `snmp_object` 的 `links` 字段。存在 LLDP 链路时优先使用 LLDP；仅在没有 LLDP 链路时回退到 CDP。

该配置仅适用于上报 `snmp_object` 的内置 Profile 采集模式，不适用于用户 Profile 生成的 `snmp_<class>` 自定义对象。

#### `links` 字段 {#topology-links-field}

下面是 `links` 字段解码后的 LLDP 链路示例，`//` 表示字段说明注释：

```jsonc
[
  {
    // 本地设备上的拓扑邻居记录 ID。
    // LLDP 格式为 <device_namespace>:<管理 IP>:<lldpRemLocalPortNum>.<lldpRemIndex>；
    // CDP 格式为 <device_namespace>:<管理 IP>:<cdpCacheIfIndex>.<cdpCacheDeviceIndex>。
    "id": "default:192.0.2.10:7.1",
    // 邻居数据来源，取值为 lldp 或 cdp。
    "source_type": "lldp",
    // 采集集成名称，当前固定为 snmp。
    "integration": "snmp",
    // 被 DataKit 直接采集的本地链路端点。
    "local": {
      // 本地设备。
      "device": {
        // 已解析的本地设备 ID，格式为 <device_namespace>:<管理 IP>。
        "resolved_id": "default:192.0.2.10"
      },
      // 本地接口。
      "interface": {
        // 已解析的本地接口 ID，格式为 <local.device.resolved_id>:<ifIndex>。
        // LLDP 本地端口无法关联到 IF-MIB 接口时可能缺失。
        "resolved_id": "default:192.0.2.10:12",
        // LLDP 的 lldpLocPortId；CDP 链路当前为空字符串。
        "id": "82:a5:6e:a5:c9:01",
        // 由 lldpLocPortIdSubtype 转换的本地端口标识类型；CDP 链路中通常缺失。
        "id_type": "mac_address"
      }
    },
    // 邻居协议发现的对端链路端点，不代表该设备一定已被 DataKit 直接采集。
    "remote": {
      // 对端设备。
      "device": {
        // LLDP 的 lldpRemChassisId 或 CDP 的 cdpCacheDeviceId。
        "id": "01:00:00:00:01:02",
        // 由 lldpRemChassisIdSubtype 转换的 LLDP Chassis ID 类型；CDP 链路中通常缺失。
        "id_type": "mac_address",
        // LLDP 的 lldpRemSysName 或 CDP 的 cdpCacheSysName。
        "name": "switch-b",
        // LLDP 的 lldpRemSysDesc 或 CDP 的 cdpCacheVersion。
        "description": "remote switch",
        // 对端管理 IP。LLDP 从远端管理地址表索引解析；
        // CDP 依次尝试主、备用和缓存管理地址。设备未上报时缺失。
        "ip_address": "10.250.0.6"
      },
      // 对端接口。
      "interface": {
        // LLDP 的 lldpRemPortId 或 CDP 的 cdpCacheDevicePort。
        "id": "Ethernet1/7",
        // LLDP 远端端口标识类型；CDP 固定为 interface_name。
        "id_type": "interface_name",
        // LLDP 的 lldpRemPortDesc；CDP 链路中通常缺失。
        "description": "remote uplink"
      }
    }
  }
]
```

LLDP 的设备和端口标识类型可能包括 `mac_address`、`network_address`、`interface_name`、`interface_alias`、`port_component` 和 `local` 等。除结构必需字段外，名称、描述、管理 IP 以及未成功解析的对象 ID 都可能不出现在结果中。

现有的 `enable_lldp` 是独立的 LLDP 采集入口，按 `lldp_interval` 运行并上报 `snmp_lldp` 日志数据，保留用于兼容已有配置：

```toml
[[inputs.snmp]]
  enable_lldp = true
  lldp_interval = "10m"
```

两个开关相互独立且默认均为 `false`。通常只需启用一种输出；若同时启用，DataKit 会在各自周期内分别查询 LLDP 数据，并同时产生对象链路和日志两种输出。

### 被采集设备配置要求 {#lldp-device-config}

网络设备端需要：

- **启用 LLDP**

- **配置 SNMP 访问 LLDP MIB 权限**


### 验证配置 {#lldp-verify}

**在设备上验证 LLDP 邻居：**

```bash
# 华为设备
display lldp neighbor

# Cisco 设备
show lldp neighbor
```

**从 DataKit 主机验证 SNMP 能否查询 LLDP 数据：**

```bash
# SNMPv2c 验证
snmpwalk -v2c -c [COMMUNITY_STRING] [DEVICE_IP] 1.0.8802.1.1.2.1.4.1

# SNMPv3 验证
snmpwalk -v3 -u [USERNAME] -l authPriv \
  -a SHA -A [AUTH_PASSWORD] \
  -x AES -X [PRIV_PASSWORD] \
  [DEVICE_IP] 1.0.8802.1.1.2.1.4.1
```

## 指标 {#metric}

以下所有数据采集，默认会追加全局选举 tag，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
[inputs.{{.InputName}}.tags]
 # some_tag = "some_value"
 # more_tag = "some_other_value"
 # ...
```

<!-- markdownlint-disable MD046 -->
???+ note

    以下所有指标集以及其指标，只包含部分常见的字段，一些设备特定的字段，根据配置和设备型号不同，会额外多出一些字段。
<!-- markdownlint-enable MD046 -->

<!-- markdownlint-disable MD024 -->
{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}

## 对象 {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}

## 日志 {#logging}

<!-- markdownlint-disable MD024 -->
{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}
<!-- markdownlint-enable MD024 -->

## FAQ {#faq}

### DataKit 是如何发现设备的? {#faq-discover}

DataKit 支持 "指定设备" 和 "自动发现" 两种模式。两种模式可以同时开启。

指定设备模式下，DataKit 与指定 IP 的设备使用 SNMP 协议进行通信，可以获知其目前在线状态。

自动发现模式下，DataKit 向指定 IP 网段内的所有地址逐一发送 SNMP 协议数据包，如果其响应可以匹配到相应的 Profile，那么 DataKit 认为该 IP 上有一个 SNMP 设备。

### 设备不支持采集 {#faq-not-support}

DataKit 可以从所有 SNMP 设备中收集通用的基线指标。如果你发现被采集的设备上报的数据中没有你想要的指标，那么，你可能需要为该设备[自定义一份 Profile](snmp.md#advanced-custom-oid)。

为了完成上述工作，你很可能需要从设备厂商的官网下载该设备型号的 OID 手册。

### 开启 SNMP 设备采集但看不到指标 {#faq-no-metrics}

尝试为你的设备放开 ACLs/防火墙 规则。

可以在运行 DataKit 的主机上运行命令 `snmpwalk -O bentU -v 2c -c <COMMUNITY_STRING> <IP_ADDRESS>:<PORT> 1.3.6`。如果得到一个没有任何响应的超时，很可能是有什么东西阻止了 DataKit 从你的设备上收集指标。
