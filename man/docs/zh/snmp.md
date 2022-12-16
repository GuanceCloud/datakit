{{.CSS}}
# SNMP
---

{{.AvailableArchs}}

---

本文主要介绍 [SNMP](https://en.wikipedia.org/wiki/Simple_Network_Management_Protocol/){:target="_blank"} 数据采集。

## 关于 SNMP 协议 {#config-pre}

SNMP 协议分为 3 个版本: v1/v2c/v3，其中：

    - v1 和 v2c 是兼容的。很多 SNMP 设备只提供 v2c 和 v3 两种版本的选择。v2c 版本，兼容性最好，很多旧设备只支持这个版本
    - 如果对安全性要求高，选用 v3。安全性也是 v3 版本与之前版本的主要区别

如果选择 v2c 版本，需要提供 `community string`，中文翻译为 `团体名/团体字符串`。`未加密的口令` 即密码，与 SNMP 设备进行交互需要提供这个进行鉴权。另外，有的设备会进行区别，分为 `只读团体名` 和 `读写团体名`。顾名思义:

- `只读团体名`: 设备只会向该方提供内部指标数据，不能修改内部的一些配置(DataKit 用这个就够了)
- `读写团体名`: 提供方拥有设备内部指标数据查询与部分配置修改权限

如果选择 v3 版本，需要提供 `用户名`、`认证算法/密码`、`加密算法/密码`、`上下文` 等，各个设备不同，根据要求进行填写。

## 配置采集器 {#config-input}

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下: 
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。

---
    
???+ attention

    上面配置的 `inputs.snmp.tags` 中如果与原始 fields 中的 key 同名重复，则会被原始数据覆盖。


### 配置 SNMP {#config-snmp}

- 在设备侧, 配置 SNMP 协议

SNMP 设备在默认情况下, 一般 SNMP 协议处于关闭状态, 需要进入管理界面手动打开。同时, 需要根据实际情况选择协议版本和填写相应信息。

???+ tip

    有些设备为了安全需要额外配置放行 SNMP，具体因设备而异。比如华为系防火墙，需要在 "启用访问管理" 中勾选 SNMP 以放行。可以使用 `snmpwalk` 命令来测试采集侧与设备侧是否配置连通成功: 

    ```shell
    # 适用 v2c 版本
    snmpwalk -O bentU -v 2c -c [community string] [IP] 1.3.6` 
    # 适用 v3 版本
    snmpwalk -v 3 -u user -l authPriv -a sha -A [认证密码] -x aes -X [加密密码] [IP] 1.3.6 
    ```

    如果配置没有问题的话，该命令会输出大量数据。`snmpwalk` 是运行在采集侧的一个测试工具，MacOS 下自带，Linux 安装方法: 

    ```shell
    sudo yum install net–snmp–utils # CentOS
    sudo apt–get install snmp       # Ubuntu
    ```

- 在 DataKit 侧, 配置采集。

## 自定义设备的 OID 配置 {#custom-oid}

如果你发现被采集的设备上报的数据中没有你想要的指标，那么，你可以需要为该设备额外定义一份 Profile。

设备的所有 OID 一般都可以在其官网上下载。Datakit 定义了一些通用的 OID，以及 Cisco/Dell/HP 等部分设备。根据 SNMP 协议，各设备生产商可以自定义 [OID](https://www.dpstele.com/snmp/what-does-oid-network-elements.php)，用于标识其内部特殊对象。如果想要标识这些，你需要自定义设备的配置(我们这里称这种配置为 Profile，即 "自定义 Profile")，方法如下。

在 Datakit 的安装目录的路径 `conf.d/snmp/profiles` 下，如下所示创建 yml 文件 `cisco-3850.yaml` (这里以 Cisco 3850 为例):

``` yaml
# Backward compatibility shim. Prefer the Cisco Catalyst profile directly
# Profile for Cisco 3850 devices

extends:
  - _base.yaml
  - _cisco-generic.yaml
  - _cisco-catalyst.yaml

sysobjectid: 1.3.6.1.4.1.9.1.1745 # cat38xxstack

device:
  vendor: "cisco"

# Example sysDescr:
#   Cisco IOS Software, IOS-XE Software, Catalyst L3 Switch Software (CAT3K_CAA-UNIVERSALK9-M), Version 03.06.06E RELEASE SOFTWARE (fc1) Technical Support: http://www.cisco.com/techsupport Copyright (c) 1986-2016 by Cisco Systems, Inc. Compiled Sat 17-Dec-

metadata:
  device:
    fields:
      serial_number:
        symbol:
          MIB: OLD-CISCO-CHASSIS-MIB
          OID: 1.3.6.1.4.1.9.3.6.3.0
          name: chassisId
```

如上所示，定义了一个 `sysobjectid` 为 `1.3.6.1.4.1.9.1.1745` 的设备，下次 Datakit 如果采集到 `sysobjectid` 相同的设备时，便会应用该文件，在此情况下，采集到 OID 为 `1.3.6.1.4.1.9.3.6.3.0` 的数据便会上报为名称是 `chassisId` 的指标。

> 注意: `conf.d/snmp/profiles` 这个文件夹需要 SNMP 采集器运行一次后才会出现。

## 指标集 {#measurements}

以下所有数据采集，默认会追加名为 `host`(值为 SNMP 设备的名称)，也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

???+ attention

    以下所有指标集以及其指标，只包含部分常见的字段，一些设备特定的字段，根据配置和设备型号不同，会额外多出一些字段。

### 指标 {#metrics}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

### 对象 {#objects}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}
