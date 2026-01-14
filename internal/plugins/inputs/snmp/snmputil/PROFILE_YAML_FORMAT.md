# SNMP Profile YAML 格式说明

本文档详细说明 SNMP Profile YAML 文件的格式和配置选项。

## 目录

- [Profile 结构概览](#profile-结构概览)
- [字段详细说明](#字段详细说明)
  - [extends](#extends)
  - [sysobjectid](#sysobjectid)
  - [device](#device)
  - [static_tags](#static_tags)
  - [metadata](#metadata)
  - [metrics](#metrics)
  - [metric_tags](#metric_tags)
- [SymbolConfig 配置选项](#symbolconfig-配置选项)
- [示例](#示例)
- [继承机制](#继承机制)

## Profile 结构概览

一个完整的 SNMP Profile YAML 文件包含以下主要部分：

```yaml
extends:                    # 继承的 base profile（可选）
  - _base.yaml
  - _generic-if.yaml

sysobjectid:                # 设备 sysObjectID 匹配规则（必需）
  - 1.3.6.1.4.1.9.*
  - 1.3.6.1.4.1.9.1.*

device:                     # 设备元数据（可选）
  vendor: "cisco"

static_tags:                # 静态标签（可选）
  - "tag1:value1"
  - "tag2:value2"

metadata:                   # 元数据配置（可选）
  device:
    fields:
      name:
        symbol:
          OID: 1.3.6.1.2.1.1.5.0
          name: sysName
  interface:
    fields:
      name:
        symbol:
          OID: 1.3.6.1.2.1.31.1.1.1.1
          name: ifName
    id_tags:
      - symbol:
          OID: 1.3.6.1.2.1.31.1.1.1.1
          name: ifName
        tag: interface

metric_tags:                # Profile 级别的 metric tags（可选）
  - OID: 1.3.6.1.2.1.1.5.0
    symbol: sysName
    tag: snmp_host

metrics:                    # 指标配置（可选）
  - MIB: IF-MIB
    symbol:
      OID: 1.3.6.1.2.1.2.1.0
      name: ifNumber
  - MIB: IF-MIB
    table:
      OID: 1.3.6.1.2.1.2.2
      name: ifTable
    symbols:
      - OID: 1.3.6.1.2.1.2.2.1.10
        name: ifInOctets
    metric_tags:
      - symbol:
          OID: 1.3.6.1.2.1.31.1.1.1.1
          name: ifName
        tag: interface
```

## 字段详细说明

### extends

**类型**: `[]string`  
**必需**: 否  
**说明**: 指定要继承的 base profile 文件列表。Base profile 通常以 `_` 开头（如 `_base.yaml`），用于定义通用的配置。

**示例**:
```yaml
extends:
  - _base.yaml
  - _generic-if.yaml
```

**注意**:
- Base profile 会被先加载，然后当前 profile 的配置会覆盖或合并到 base profile 的配置上
- 支持多层继承（base profile 也可以继承其他 base profile）
- 继承顺序：从左到右，后面的会覆盖前面的
- 不能有循环依赖

### sysobjectid

**类型**: `string` 或 `[]string`  
**必需**: 是  
**说明**: 设备的 sysObjectID 匹配规则，用于自动识别设备类型并应用对应的 profile。

**格式**:
- 支持通配符 `*`，匹配任意数字
- 支持多个匹配规则（数组形式）
- 匹配优先级：更具体的 OID > 通配符 OID

**示例**:
```yaml
# 单个匹配规则
sysobjectid: 1.3.6.1.4.1.9.*

# 多个匹配规则
sysobjectid:
  - 1.3.6.1.4.1.9.1.*
  - 1.3.6.1.4.1.9.2.*
```

**匹配规则**:
- `1.3.6.1.4.1.9.*` - 匹配所有 Cisco 设备
- `1.3.6.1.4.1.9.1.*` - 匹配 Cisco 路由器
- `1.3.6.1.4.1.3375.2.1.3.4.43` - 精确匹配特定设备型号

### device

**类型**: `object`  
**必需**: 否  
**说明**: 设备相关的静态元数据。

**字段**:
- `vendor` (string): 设备厂商名称

**示例**:
```yaml
device:
  vendor: "cisco"
```

### static_tags

**类型**: `[]string`  
**必需**: 否  
**说明**: 静态标签列表，会添加到所有采集的数据点上。

**格式**: `"key:value"`

**示例**:
```yaml
static_tags:
  - "device_type:router"
  - "vendor:cisco"
  - "location:datacenter"
```

### metadata

**类型**: `object`  
**必需**: 否  
**说明**: 定义设备或接口等资源的元数据字段。用于上报 object 类型的数据。

**支持的资源类型**:
- `device`: 设备级别的元数据（使用标量 OID）
- `interface`: 接口级别的元数据（使用列 OID）

**结构**:
```yaml
metadata:
  <resource_name>:
    fields:
      <field_name>:
        symbol:          # 单个 OID（标量或列）
          OID: "..."
          name: "..."
        symbols:         # 多个 OID（列）
          - OID: "..."
            name: "..."
        value: "..."     # 固定值（不查询 OID）
    id_tags:            # 用于标识资源的标签（仅列资源）
      - symbol:
          OID: "..."
          name: "..."
        tag: "..."
```

**device 资源示例**:
```yaml
metadata:
  device:
    fields:
      name:
        symbol:
          OID: 1.3.6.1.2.1.1.5.0
          name: sysName
      description:
        symbol:
          OID: 1.3.6.1.2.1.1.1.0
          name: sysDescr
      vendor:
        value: "cisco"  # 固定值，不查询 OID
      type:
        value: "router"
```

**interface 资源示例**:
```yaml
metadata:
  interface:
    fields:
      name:
        symbol:
          OID: 1.3.6.1.2.1.31.1.1.1.1
          name: ifName
      admin_status:
        symbol:
          OID: 1.3.6.1.2.1.2.2.1.7
          name: ifAdminStatus
    id_tags:  # 用于标识每个接口的标签
      - symbol:
          OID: 1.3.6.1.2.1.31.1.1.1.1
          name: ifName
        tag: interface
```

### metric_tags

**类型**: `[]object`  
**必需**: 否  
**说明**: Profile 级别的 metric tags，会应用到所有 metrics 上。

**结构**:
```yaml
metric_tags:
  - OID: "..."           # 标量 OID（旧语法，已废弃）
    symbol: "..."        # 符号名称（旧语法，已废弃）
    tag: "..."           # 标签名称
  - symbol:              # 新语法（推荐）
      OID: "..."
      name: "..."
    tag: "..."
    mapping:             # 值映射（可选）
      "1": "up"
      "2": "down"
```

**示例**:
```yaml
metric_tags:
  - OID: 1.3.6.1.2.1.1.5.0
    symbol: sysName
    tag: snmp_host
  - symbol:
      OID: 1.3.6.1.2.1.1.5.0
      name: sysName
    tag: device_hostname
```

### metrics

**类型**: `[]object`  
**必需**: 否  
**说明**: 定义要采集的 SNMP 指标。

**两种类型**:
1. **标量指标** (Scalar Metric): 单个值，使用 `symbol` 字段
2. **表格指标** (Table Metric): 多个值（表格），使用 `symbols` 字段

**标量指标结构**:
```yaml
metrics:
  - MIB: IF-MIB                    # MIB 名称（描述性，可选）
    symbol:                         # 标量 OID
      OID: 1.3.6.1.2.1.2.1.0
      name: ifNumber
    metric_type: gauge              # 指标类型（可选）
    static_tags:                    # 该指标的静态标签（可选）
      - "tag1:value1"
```

**表格指标结构**:
```yaml
metrics:
  - MIB: IF-MIB
    table:                          # 表格信息（描述性，可选）
      OID: 1.3.6.1.2.1.2.2
      name: ifTable
    symbols:                        # 表格列 OID 列表
      - OID: 1.3.6.1.2.1.2.2.1.10
        name: ifInOctets
        metric_type: monotonic_count
      - OID: 1.3.6.1.2.1.2.2.1.16
        name: ifOutOctets
        metric_type: monotonic_count
    metric_type: monotonic_count_and_rate  # 应用到所有 symbols（可选）
    metric_tags:                    # 该指标的标签（可选）
      - symbol:
          OID: 1.3.6.1.2.1.31.1.1.1.1
          name: ifName
        tag: interface
        mapping:                    # 值映射（可选）
          "1": "up"
          "2": "down"
```

**完整示例**:
```yaml
metrics:
  # 标量指标
  - MIB: IF-MIB
    symbol:
      OID: 1.3.6.1.2.1.2.1.0
      name: ifNumber
    metric_type: gauge

  # 表格指标
  - MIB: IF-MIB
    table:
      OID: 1.3.6.1.2.1.2.2
      name: ifTable
    symbols:
      - OID: 1.3.6.1.2.1.2.2.1.10
        name: ifInOctets
        metric_type: monotonic_count
      - OID: 1.3.6.1.2.1.2.2.1.16
        name: ifOutOctets
        metric_type: monotonic_count
    metric_tags:
      - symbol:
          OID: 1.3.6.1.2.1.31.1.1.1.1
          name: ifName
        tag: interface
```

## SymbolConfig 配置选项

`SymbolConfig` 用于定义单个 OID 的配置，支持以下选项：

### 基本字段

```yaml
symbol:
  OID: "1.3.6.1.2.1.1.5.0"    # OID（必需）
  name: "sysName"              # 符号名称（必需）
```

### 高级选项

```yaml
symbol:
  OID: "1.3.6.1.2.1.1.5.0"
  name: "sysName"
  
  # 值提取（使用正则表达式）
  extract_value: "([0-9]+)"
  
  # 值匹配
  match_pattern: ".*"
  match_value: "test"
  
  # 缩放因子
  scale_factor: 0.001
  
  # 格式化
  format: "mac_address"  # 支持: mac_address
  
  # 常量值（仅用于表格指标）
  constant_value_one: true  # 如果为 true，每个表格行都报告值为 1
  
  # 指标类型（覆盖默认类型）
  metric_type: gauge  # gauge, rate, monotonic_count, monotonic_count_and_rate
```

### extract_value

**类型**: `string`  
**说明**: 使用正则表达式从 SNMP 返回值中提取部分内容。

**示例**:
```yaml
symbol:
  OID: 1.3.6.1.2.1.1.1.0
  name: sysDescr
  extract_value: "Version ([0-9.]+)"  # 从描述中提取版本号
```

### match_pattern 和 match_value

**类型**: `string`  
**说明**: 用于匹配和过滤值。

**示例**:
```yaml
symbol:
  OID: 1.3.6.1.2.1.2.2.1.7
  name: ifAdminStatus
  match_pattern: "1|3"  # 只匹配值为 1 或 3 的接口
```

### scale_factor

**类型**: `float64`  
**说明**: 对采集到的值进行缩放。

**示例**:
```yaml
symbol:
  OID: 1.3.6.1.2.1.2.2.1.5
  name: ifSpeed
  scale_factor: 0.000001  # 将 bps 转换为 Mbps
```

### format

**类型**: `string`  
**说明**: 格式化输出值。

**支持的值**:
- `mac_address`: 将字节数组格式化为 MAC 地址（如 `00:11:22:33:44:55`）

**示例**:
```yaml
symbol:
  OID: 1.3.6.1.2.1.2.2.1.6
  name: ifPhysAddress
  format: mac_address
```

### constant_value_one

**类型**: `bool`  
**说明**: 仅用于表格指标。如果为 `true`，每个表格行都会报告值为 1，而不查询 OID 的实际值。常用于计数或存在性检查。

**限制**: 只能用于 `symbols`（表格列），不能用于 `symbol`（标量）。

**示例**:
```yaml
metrics:
  - MIB: CPS-MIB
    table:
      OID: 1.3.6.1.4.1.3808.1.1.3.2.4.1
      name: ePDULoadBankConfigTable
    symbols:
      - name: cyberpower.ePDULoadBankConfig
        constant_value_one: true  # 每个 bank 都报告 1
    metric_tags:
      - symbol:
          OID: 1.3.6.1.4.1.3808.1.1.3.2.4.1.1.1
          name: ePDULoadBankConfigIndex
        tag: e_pdu_load_bank_config_index
```

### metric_type

**类型**: `string`  
**说明**: 指定指标类型，覆盖从 SNMP 数据类型推断的默认类型。

**支持的值**:
- `gauge`: 仪表盘类型（当前值）
- `rate`: 速率类型（计算变化率）
- `monotonic_count`: 单调计数器（累计值）
- `monotonic_count_and_rate`: 单调计数器 + 速率（同时上报累计值和速率）

**使用位置**:
1. 在 `SymbolConfig` 中（`symbol.metric_type` 或 `symbols[].metric_type`）
2. 在 `MetricsConfig` 中（`metric_type`，应用到所有 symbols）

**优先级**: `SymbolConfig.metric_type` > `MetricsConfig.metric_type` > `MetricsConfig.forced_type` > 默认类型

**示例**:
```yaml
# 在 symbol 级别
symbol:
  OID: 1.3.6.1.2.1.2.2.1.10
  name: ifInOctets
  metric_type: monotonic_count

# 在 metrics 级别（应用到所有 symbols）
metrics:
  - MIB: IF-MIB
    table:
      OID: 1.3.6.1.2.1.2.2
      name: ifTable
    metric_type: monotonic_count_and_rate
    symbols:
      - OID: 1.3.6.1.2.1.2.2.1.10
        name: ifInOctets
      - OID: 1.3.6.1.2.1.2.2.1.16
        name: ifOutOctets
```

## MetricTagConfig 配置选项

`MetricTagConfig` 用于定义指标的标签，支持以下选项：

```yaml
metric_tags:
  - symbol:                    # 新语法（推荐）
      OID: 1.3.6.1.2.1.31.1.1.1.1
      name: ifName
    tag: interface             # 标签名称（必需）
    
    # 值映射（可选）
    mapping:
      "1": "up"
      "2": "down"
      "3": "testing"
    
    # 索引转换（可选，仅用于表格指标）
    index_transform:
      - start: 0
        end: 2
      - start: 4
        end: 6
```

### mapping

**类型**: `map[string]string`  
**说明**: 将原始 SNMP 值映射为更易读的字符串。

**示例**:
```yaml
metric_tags:
  - symbol:
      OID: 1.3.6.1.2.1.2.2.1.7
      name: ifAdminStatus
    tag: if_admin_status
    mapping:
      "1": "up"
      "2": "down"
      "3": "testing"
```

### index_transform

**类型**: `[]object`  
**说明**: 仅用于表格指标。从表格索引中提取部分内容作为标签值。

**示例**:
```yaml
metric_tags:
  - symbol:
      OID: 1.3.6.1.2.1.2.2.1.1
      name: ifIndex
    tag: interface_index
    index_transform:
      - start: 0    # 从索引的第 0 位开始
        end: 2      # 到第 2 位结束（不包含）
```

## 示例

### 简单示例

```yaml
sysobjectid: 1.3.6.1.4.1.9.*

device:
  vendor: "cisco"

metadata:
  device:
    fields:
      name:
        symbol:
          OID: 1.3.6.1.2.1.1.5.0
          name: sysName

metrics:
  - MIB: IF-MIB
    symbol:
      OID: 1.3.6.1.2.1.2.1.0
      name: ifNumber
    metric_type: gauge
```

### 完整示例（带继承）

```yaml
extends:
  - _base.yaml
  - _generic-if.yaml

sysobjectid:
  - 1.3.6.1.4.1.9.1.*
  - 1.3.6.1.4.1.9.2.*

device:
  vendor: "cisco"

static_tags:
  - "device_type:router"
  - "vendor:cisco"

metadata:
  device:
    fields:
      vendor:
        value: "cisco"
      type:
        value: "router"

metrics:
  - MIB: IF-MIB
    symbol:
      OID: 1.3.6.1.2.1.2.1.0
      name: ifNumber
    metric_type: gauge

  - MIB: IF-MIB
    table:
      OID: 1.3.6.1.2.1.2.2
      name: ifTable
    metric_type: monotonic_count_and_rate
    symbols:
      - OID: 1.3.6.1.2.1.2.2.1.10
        name: ifInOctets
      - OID: 1.3.6.1.2.1.2.2.1.16
        name: ifOutOctets
    metric_tags:
      - symbol:
          OID: 1.3.6.1.2.1.31.1.1.1.1
          name: ifName
        tag: interface
      - symbol:
          OID: 1.3.6.1.2.1.2.2.1.7
          name: ifAdminStatus
        tag: if_admin_status
        mapping:
          "1": "up"
          "2": "down"
```

### 使用 constant_value_one 的示例

```yaml
extends:
  - _base.yaml

sysobjectid: 1.3.6.1.4.1.3808.1.1.*

metadata:
  device:
    fields:
      vendor:
        value: "cyberpower"
      type:
        value: "PDU"

metrics:
  - MIB: CPS-MIB
    table:
      OID: 1.3.6.1.4.1.3808.1.1.3.2.4.1
      name: ePDULoadBankConfigTable
    symbols:
      - name: cyberpower.ePDULoadBankConfig
        constant_value_one: true  # 每个 bank 都报告 1
    metric_tags:
      - symbol:
          OID: 1.3.6.1.4.1.3808.1.1.3.2.4.1.1.1
          name: ePDULoadBankConfigIndex
        tag: e_pdu_load_bank_config_index
```

## 继承机制

### 继承规则

1. **继承顺序**: 从左到右，后面的配置会覆盖前面的
2. **字段合并规则**:
   - `metrics`: 追加（append）
   - `metric_tags`: 追加（append）
   - `static_tags`: 追加（append）
   - `metadata.fields`: 如果字段不存在则添加，如果存在则不覆盖（target 优先）
   - `device`: 覆盖
   - `sysobjectid`: 覆盖

3. **循环依赖检测**: 系统会检测并拒绝循环依赖

### 示例

**base.yaml**:
```yaml
metric_tags:
  - OID: 1.3.6.1.2.1.1.5.0
    symbol: sysName
    tag: snmp_host

metadata:
  device:
    fields:
      name:
        symbol:
          OID: 1.3.6.1.2.1.1.5.0
          name: sysName
```

**custom-profile.yaml**:
```yaml
extends:
  - _base.yaml

metadata:
  device:
    fields:
      vendor:
        value: "cisco"  # 新增字段
      name:              # 已存在，不会被 base 覆盖
        symbol:
          OID: 1.3.6.1.2.1.1.5.0
          name: customSysName
```

**结果**: 
- `metric_tags` 包含 base 和 custom 的所有 tags
- `metadata.device.fields.name` 使用 custom 的定义（target 优先）
- `metadata.device.fields.vendor` 使用 custom 的定义（新增）

## 注意事项

1. **MIB 和 table 字段**: `MIB` 和 `table` 字段是描述性的，不会影响实际的 SNMP 查询。实际的查询基于 `symbol`/`symbols` 中的 OID。

2. **OID 格式**: OID 可以是点分十进制格式（如 `1.3.6.1.2.1.1.5.0`）或数字格式。

3. **name 字段**: `name` 字段用于标识符号，会作为指标名称的一部分。建议使用有意义的名称。

4. **metric_type 优先级**: 
   - `SymbolConfig.metric_type`（最高优先级）
   - `MetricsConfig.metric_type`
   - `MetricsConfig.forced_type`（已废弃）
   - 从 SNMP 数据类型推断（默认）

5. **constant_value_one 限制**: 只能用于表格指标（`symbols`），不能用于标量指标（`symbol`）。

6. **metadata 资源类型**:
   - `device`: 使用标量 OID（单个值）
   - `interface`: 使用列 OID（表格，每个接口一个值）

7. **Base Profile**: 以 `_` 开头的 profile 文件是 base profile，不会被直接加载为可用 profile，只能被其他 profile 继承。

