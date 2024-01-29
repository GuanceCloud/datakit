---
title     : 'IPMI'
summary   : 'Collect IPMI metrics'
__int_icon      : 'icon/ipmi'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# IPMI
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

IPMI metrics show the current, voltage, power consumption, occupancy rate, fan speed, temperature and equipment status of the monitored equipment.

IPMI is the abbreviation of Intelligent Platform Management Interface, which is an industry standard for managing peripheral devices used in enterprise systems based on Intel structure. This standard is formulated by Intel, Hewlett-Packard, NEC, Dell Computer and SuperMicro. Users can use IPMI to monitor the physical health characteristics of the server, such as temperature, voltage, fan working status, power status, etc.

IPMI enables the operation and maintenance system to obtain the operation health indicators of monitored servers and other devices **without intrusion**, thus ensuring information security.

## Configuration {#config}

### Preconditions {#requirements}

- Install the `ipmitool` Toolkit

DataKit collects IPMI data through the [`ipmitool`][1]  tool, so it needs to be installed on the machine. It can be installed by the following command:

```shell
# CentOS
yum -y install ipmitool

# Ubuntu
sudo apt-get update && sudo apt -y install ipmitool

# macOS
brew install ipmitool # macOS
```

- Loading Module

```shell
modprobe ipmi_msghandler
modprobe ipmi_devintf
```

After successful installation, you can see the information output by ipmi server by running the following command:

```shell
ipmitool -I lanplus -H <IP 地址> -U <用户名> -P <密码> sdr elist

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

<!-- markdownlint-disable MD046 -->
???+ attention

    1. IP address refers to the IP address of the IPMI port of the server that you remotely manage
    1. Server `IPMI Settings -> Enable IPMI on LAN` needs to be checked
    1. Server `Channel Privilege Level Restrictions` operator level requirements and `<User Name>` keep level consistent
    1. `ipmitool` toolkit is installed on the machine running DataKit.

### Collector Configuration {#input-config}

=== "Host deployment"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, restart DataKit.

=== "Kubernetes"

    Modification of configuration parameters as environment variables is supported in Kubernetes (effective only when the DataKit is running in K8s DaemonSet mode, which is not supported on host-deployed DataKits):
    
    | Environment Variable Name                          | Corresponding Configuration Parameter Item     | Parameter Example                                                     |
    | :------------------------           | ---                  | ---                                                          |
    | `ENV_INPUT_IPMI_TAGS`               | `tags`               | `tag1=value1,tag2=value2`; If there is a tag with the same name in the configuration file, it will be overwritten |
    | `ENV_INPUT_IPMI_INTERVAL`           | `interval`           | `10s`                                                        |
    | `ENV_INPUT_IPMI_TIMEOUT`            | `timeout`            | `5s`                                                         |
    | `ENV_INPUT_IPMI_DEOP_WARNING_DELAY` | `drop_warning_delay` | `300s`                                                       |
    | `ENV_INPUT_IPMI_BIN_PATH`           | `bin_path`           | `"/usr/bin/ipmitool"`                                        |
    | `ENV_INPUT_IPMI_ENVS`               | `envs`               | `["LD_LIBRARY_PATH=XXXX:$LD_LIBRARY_PATH"]`                  |
    | `ENV_INPUT_IPMI_SERVERS`            | `ipmi_servers`       | `["192.168.1.1"]`                                            |
    | `ENV_INPUT_IPMI_INTERFACES`         | `ipmi_interfaces`    | `["lanplus"]`                                                |
    | `ENV_INPUT_IPMI_USERS`              | `ipmi_users`         | `["root"]`                                                   |
    | `ENV_INPUT_IPMI_PASSWORDS`          | `ipmi_passwords`     | `["calvin"]`                                                 |
    | `ENV_INPUT_IPMI_HEX_KEYS`           | `hex_keys`           | `["50415353574F5244"]`                                       |
    | `ENV_INPUT_IPMI_METRIC_VERSIONS`    | `metric_versions`    | `[2]`                                                        |
    | `ENV_INPUT_IPMI_REGEXP_CURRENT`     | `regexp_current`     | `["current"]`                                                |
    | `ENV_INPUT_IPMI_REGEXP_VOLTAGE`     | `regexp_voltage`     | `["voltage"]`                                                |
    | `ENV_INPUT_IPMI_REGEXP_POWER`       | `regexp_power`       | `["pwr","power"]`                                            |
    | `ENV_INPUT_IPMI_REGEXP_TEMP`        | `regexp_temp`        | `["temp"]`                                                   |
    | `ENV_INPUT_IPMI_REGEXP_FAN_SPEED`   | `regexp_fan_speed`   | `["fan"]`                                                    |
    | `ENV_INPUT_IPMI_REGEXP_USAGE`       | `regexp_usage`       | `["usage"]`                                                  |
    | `ENV_INPUT_IPMI_REGEXP_COUNT`       | `regexp_count`       | `[]`                                                         |
    | `ENV_INPUT_IPMI_REGEXP_STATUS`      | `regexp_status`      | `["fan"]`                                                    |

???+ tip "Configuration"

    - The keywords for each parameter classification are all in lowercase
    - Refer to `ipmitool -I ...` The data returned by the command, then the keywords are reasonably configured
<!-- markdownlint-enable -->

<!--
## Election Configuration {#election-config}

IPMI collector supports election function. When multiple machines run DataKit, it prevents everyone from collecting data repeatedly through election.

`/conf.d/datakit.conf `file opens the `election `function:

```
[election]
  # Start election
  enable = true

  # Set the namespace of the election (default)
  namespace = "default"

  # Tag that allows election space to be appended to data
  enable_namespace_tag = false
```
`conf.d/{{.Catalog}}/{{.InputName}}.conf` file opens the `election` function:
```
  ## Set true to enable election
  election = true
```
-->

## Metric {#metric}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

- Tag

{{$m.TagsMarkdownTable}}

- Metrics List

{{$m.FieldsMarkdownTable}}

{{ end }}

[1]: https://github.com/ipmitool/ipmitool
