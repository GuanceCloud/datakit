{{.CSS}}
# IPMI

- 操作系统支持：{{.AvailableArchs}} | *Author：张连山*

IPMI 指标展示被监测设备的电流、电压、功耗、占用率、风扇转速、温度以及设备状态等信息。

### IPMI 介绍 {#introduction}

IPMI 是智能型平台管理接口（Intelligent Platform Management Interface）的缩写，是管理基于 Intel 结构的企业系统中所使用的外围设备采用的一种工业标准，该标准由英特尔、惠普、NEC、美国戴尔电脑和 SuperMicro 等公司制定。用户可以利用 IPMI 监视服务器的物理健康特征，如温度、电压、风扇工作状态、电源状态等。

IPMI 可以让运维系统**无侵入**获得被监控服务器等设备的运行健康指标，保障信息安全。

## 前置条件 {#precondition}

- 安装 `ipmitool` 工具包

DataKit 是通过 [ipmitool][1] 这个工具来采集 IPMI 数据的，故需要机器上安装这个工具。可通过如下命令安装：

```shell
# CentOS
yum -y install ipmitool
# Ubuntu
sudo apt-get update && sudo apt -y install ipmitool
# macOS
brew install ipmitool # macOS
```

- 加载模块

```shell
modprobe ipmi_msghandler
modprobe ipmi_devintf
```

安装成功后，运行如下命令，即可以看到 ipmi 服务器输出的信息：

```shell
ipmitool -I lanplus -H <IP地址> -U <用户名> -P <密码> sdr elist

SEL              | 72h | ns  |  7.1 | No Reading
Intrusion        | 73h | ok  |  7.1 | 
Fan1A RPM        | 30h | ok  |  7.1 | 2160 RPM
Fan2A RPM        | 32h | ok  |  7.1 | 2280 RPM
Fan3A RPM        | 34h | ok  |  7.1 | 2280 RPM
Fan4A RPM        | 36h | ok  |  7.1 | 2400 RPM
Fan5A RPM        | 38h | ok  |  7.1 | 2280 RPM
Fan6A RPM        | 3Ah | ok  |  7.1 | 2160 RPM
Inlet Temp       | 04h | ok  |  7.1 | 23 degrees C
Exhaust Temp     | 01h | ok  |  7.1 | 37 degrees C
Temp             | 0Fh | ok  |  3.2 | 45 degrees C
... more
```

???+ attention

    1. IP地址指的是被您远程管理服务器的 IPMI 口 IP 地址
    1. 服务器的 `IPMI设置 -> 启用 LAN 上的 IPMI` 需要勾选
    1. 服务器 `信道权限级别限制` 操作员级别需要和 `<用户名>` 保持级别一致
    1. `ipmitool` 工具包是安装到运行 DataKit 的机器里。

## 配置  {#input-config}

=== "主机部署"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，重启 DataKit 即可。

=== "Kubernetes"

    Kubernetes 中支持以环境变量的方式修改配置参数（只在 DataKit 以 K8s daemonset 方式运行时生效，主机部署的 DataKit 不支持此功能）：

    | 环境变量名                          | 对应的配置参数项     | 参数示例                                                     |
    | :------------------------           | ---                  | ---                                                          |
    | `ENV_INPUT_IPMI_TAGS`               | `tags`               | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它 |
    | `ENV_INPUT_IPMI_INTERVAL`           | `interval`           | `10s`                                                        |
    | `ENV_INPUT_IPMI_TIMEOUT`            | `timeout`            | `5s`                                                         |
    | `ENV_INPUT_IPMI_DEOP_WARNING_DELAY` | `drop_warning_delay` | `300s`                                                       |
    | `ENV_INPUT_IPMI_BIN_PATH`           | `bin_path`           | `"/usr/bin/ipmitool"`                                        |
    | `ENV_INPUT_IPMI_SERVERS`            | `ipmi_servers`       | `["192.168.1.1"]`                                            |
    | `ENV_INPUT_IPMI_INTERFACES`         | `ipmi_interfaces`    | `["lanplus"]`                                                |
    | `ENV_INPUT_IPMI_USERS`              | `ipmi_users`         | `["root"]`                                                   |
    | `ENV_INPUT_IPMI_PASSWORDS`          | `ipmi_passwords`     | `["calvin"]`                                                 |
    | `ENV_INPUT_IPMI_HEX_KEYS`           | `hex_keys`           | `["50415353574F5244"]`                                       |
    | `ENV_INPUT_IPMI_METRIC_VERSIONS`    | `metric_versions`    | `[2]`                                                        |
    | `ENV_INPUT_IPMI_REGEXP_CURRENT`     | `regexp_current`     | `["current"]`                                                |
    | `ENV_INPUT_IPMI_REGEXP_VOLTAGE`     | `regexp_voltage`     | `["voltage"]`                                                |
    | `ENV_INPUT_IPMI_REGEXP_POWER`       | `regexp_power`       | `["pwr"]`                                                    |
    | `ENV_INPUT_IPMI_REGEXP_TEMP`        | `regexp_temp`        | `["temp"]`                                                   |
    | `ENV_INPUT_IPMI_REGEXP_FAN_SPEED`   | `regexp_fan_speed`   | `["fan"]`                                                    |
    | `ENV_INPUT_IPMI_REGEXP_USAGE`       | `regexp_usage`       | `["usage"]`                                                  |
    | `ENV_INPUT_IPMI_REGEXP_COUNT`       | `regexp_count`       | `[]`                                                         |
    | `ENV_INPUT_IPMI_REGEXP_STATUS`      | `regexp_status`      | `["fan","slot","drive"]`                                     |
    

<!--
## 选举配置 {#election-config}

IPMI 采集器支持选举功能，当多台机器运行 DataKit 时，通过选举，防止大家重复采集数据。

`/conf.d/datakit.conf`文件打开`选举`功能：
```
[election]
  # 开启选举
  enable = true

  # 设置选举的命名空间(默认 default)
  namespace = "default"

  # 允许在数据上追加选举空间的 tag
  enable_namespace_tag = false
```
`conf.d/{{.Catalog}}/{{.InputName}}.conf`文件打开`选举`功能：
```
  ## Set true to enable election
  election = true
```
-->

## IPMI 指标集 {#measurements}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 被监测设备退服告警通知配置 {#warning-config}

```
 [监控] -> [监控器] -> [新建监控器] 选 [阈值检测] -> 输入[规则名称]
 [指标] 选 [指标] -> [指标集] 选 [ipmi] -> [具体指标]选 [warning] -> 下一栏选 [Max] -> by[检测维度] 选 [host]
 [紧急] 填写 [999] -> [重要] 填写 [1] -> [警告] 填写 [888] -> [正常] 填写 [N]
```

[1]: https://github.com/ipmitool/ipmitool
