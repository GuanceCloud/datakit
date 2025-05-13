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

- **v1 和 v2c 是兼容的**。很多 SNMP 设备只提供 v2c 和 v3 两种版本的选择。v2c 版本，兼容性最好，很多旧设备只支持这个版本；
- 如果对安全性要求高，选用 v3。安全性也是 v3 版本与之前版本的主要区别；

DataKit 支持以上所有版本。

### 选择 v1/v2c 版本 {#config-v2}

如果选择 v1/v2c 版本，需要提供 `community string`，中文翻译为「团体名/团体字符串/未加密的口令」，即密码，与 SNMP 设备进行交互需要提供这个进行鉴权。另外，有的设备会进一步进行细分，分为「只读团体名」和「读写团体名」。顾名思义：

- 只读团体名：设备只会向该方提供内部指标数据，不能修改内部的一些配置（DataKit 用这个就够了）
- 读写团体名：提供方拥有设备内部指标数据查询与部分配置修改权限

### 选择 v3 版本 {#config-v3}

如果选择 v3 版本，需要提供 「用户名」、「认证算法/密码」、「加密算法/密码」、「上下文」 等，各个设备要求不同，根据设备侧的配置进行填写。

## 配置 {#config}

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

### 多种配置格式 {#configuration-formats}

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
<!-- markdownlint-enable -->

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
<!-- markdownlint-enable -->

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
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD024 -->
{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## 对象 {#object}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}
<!-- markdownlint-enable -->

## FAQ {#faq}

### DataKit 是如何发现设备的? {#faq-discover}

DataKit 支持 "指定设备" 和 "自动发现" 两种模式。两种模式可以同时开启。

指定设备模式下，DataKit 与指定 IP 的设备使用 SNMP 协议进行通信，可以获知其目前在线状态。

自动发现模式下，DataKit 向指定 IP 网段内的所有地址逐一发送 SNMP 协议数据包，如果其响应可以匹配到相应的 Profile，那么 DataKit 认为该 IP 上有一个 SNMP 设备。

<!-- markdownlint-disable MD013 -->
### 在<<<custom_key.brand_name>>>上看不到我想要的指标怎么办? {#faq-not-support}
<!-- markdownlint-enable -->

DataKit 可以从所有 SNMP 设备中收集通用的基线指标。如果你发现被采集的设备上报的数据中没有你想要的指标，那么，你可以需要为该设备[自定义一份 Profile](snmp.md#advanced-custom-oid)。

为了完成上述工作，你很可能需要从设备厂商的官网下载该设备型号的 OID 手册。

<!-- markdownlint-disable MD013 -->
### 为什么开启 SNMP 设备采集但看不到指标? {#faq-no-metrics}
<!-- markdownlint-enable -->

尝试为你的设备放开 ACLs/防火墙 规则。

可以在运行 DataKit 的主机上运行命令 `snmpwalk -O bentU -v 2c -c <COMMUNITY_STRING> <IP_ADDRESS>:<PORT> 1.3.6`。如果得到一个没有任何响应的超时，很可能是有什么东西阻止了 DataKit 从你的设备上收集指标。
