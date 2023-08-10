
# IPMI

---

{{.AvailableArchs}}

---

IPMI metrics show the current, voltage, power consumption, occupancy rate, fan speed, temperature and equipment status of the monitored equipment.

### IPMI Introduction {#introduction}

IPMI is the abbreviation of Intelligent Platform Management Interface, which is an industry standard for managing peripheral devices used in enterprise systems based on Intel structure. This standard is formulated by Intel, Hewlett-Packard, NEC, Dell Computer and SuperMicro. Users can use IPMI to monitor the physical health characteristics of the server, such as temperature, voltage, fan working status, power status, etc.

IPMI enables the operation and maintenance system to obtain the operation health indicators of monitored servers and other devices **without intrusion**, thus ensuring information security.

## Preconditions {#precondition}

- Install the `ipmitool` Toolkit

DataKit collects IPMI data through the [ipmitool][1]  tool, so it needs to be installed on the machine. It can be installed by the following command:

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

    1. IP address refers to the IP address of the IPMI port of the server that you remotely manage
    1. Server `IPMI Settings -> Enable IPMI on LAN` needs to be checked
    1. Server `Channel Privilege Level Restrictions` operator level requirements and `<User Name>` keep level consistent
    1. `ipmitool` toolkit is installed on the machine running DataKit.

## Configuration  {#input-config}

=== "Host deployment"

    Go to the `conf.d/ipmi` directory under the DataKit installation directory, copy `ipmi.conf.sample` and name it `ipmi.conf`. Examples are as follows:
    
    ```toml
        
    [[inputs.ipmi]]
      ## If you have so many servers that 10 seconds can't finish the job.
      ## You can start multiple collectors.
    
      ## (Optional) collect interval: (defaults to "10s").
      interval = "10s"
    
      ## Set true to enable election
      election = true
    
      ## The binPath of ipmitool
      ## (Example) bin_path = "/usr/bin/ipmitool"
      bin_path = "/usr/bin/ipmitool"

      ## (Optional) The envs of LD_LIBRARY_PATH
      ## (Example) envs = [ "LD_LIBRARY_PATH=XXXX:$LD_LIBRARY_PATH" ]
    
      ## The ips of ipmi servers
      ## (Example) ipmi_servers = ["192.168.1.1"]
      ipmi_servers = ["192.168.1.1"]
    
      ## The interfaces of ipmi servers: (defaults to []string{"lan"}).
      ## If len(ipmi_users)<len(ipmi_ips), will use ipmi_users[0].
      ## (Example) ipmi_interfaces = ["lanplus"]
      ipmi_interfaces = ["lanplus"]
    
      ## The users name of ipmi servers: (defaults to []string{}).
      ## If len(ipmi_users)<len(ipmi_ips), will use ipmi_users[0].
      ## (Example) ipmi_users = ["root"]
      ## (Warning!) You'd better use hex_keys, it's more secure.
      ipmi_users = ["root"]
    
      ## The passwords of ipmi servers: (defaults to []string{}).
      ## If len(ipmi_passwords)<len(ipmi_ips), will use ipmi_passwords[0].
      ## (Example) ipmi_passwords = ["calvin"]
      ## (Warning!) You'd better use hex_keys, it's more secure.
      ipmi_passwords = ["calvin"]
    
      ## (Optional) provide the hex key for the IMPI connection: (defaults to []string{}).
      ## If len(hex_keys)<len(ipmi_ips), will use hex_keys[0].
      ## (Example) hex_keys = ["XXXX"]
      # hex_keys = []
    
      ## (Optional) Schema Version: (defaults to [1]).input.go
      ## If len(metric_versions)<len(ipmi_ips), will use metric_versions[0].
      ## (Example) metric_versions = [2]
      metric_versions = [2]
    
      ## (Optional) exec ipmitool timeout: (defaults to "5s").
      timeout = "5s"
    
      ## (Optional) ipmi server drop warning delay: (defaults to "300s").
      ## (Example) drop_warning_delay = "300s"
      drop_warning_delay = "300s"
    
      ## key words of current.
      ## (Example) regexp_current = ["current"]
      regexp_current = ["current"]
    
      ## key words of voltage.
      ## (Example) regexp_voltage = ["voltage"]
      regexp_voltage = ["voltage"]
    
      ## key words of power.
      ## (Example) regexp_power = ["pwr"]
      regexp_power = ["pwr"]
    
      ## key words of temp.
      ## (Example) regexp_temp = ["temp"]
      regexp_temp = ["temp"]
    
      ## key words of fan speed.
      ## (Example) regexp_fan_speed = ["fan"]
      regexp_fan_speed = ["fan"]
    
      ## key words of usage.
      ## (Example) regexp_usage = ["usage"]
      regexp_usage = ["usage"]
    
      ## key words of usage.
      ## (Example) regexp_count = []
      # regexp_count = []
    
      ## key words of status.
      ## (Example) regexp_status = ["fan","slot","drive"]
      regexp_status = ["fan","slot","drive"]
    
    [inputs.ipmi.tags]
      # some_tag = "some_value"
      # more_tag = "some_other_value"
    ```
    
    After configuration, restart DataKit.

=== "Kubernetes"

    Modification of configuration parameters as environment variables is supported in Kubernetes (effective only when the DataKit is running in K8s daemonset mode, which is not supported on host-deployed DataKits):
    
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
    - Refer to ipmitool -I The data returned by the command, then the keywords are reasonably configured



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
`conf.d/ipmi/ipmi.conf` file opens the `election` function:

```
  ## Set true to enable election
  election = true
```
-->

## IPMI Measurements {#measurements}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.ipmi.tags]`:

``` toml
 [inputs.ipmi.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```



-  Tag


| Tag Name | Description    |
|  ----  | --------|
|`host`|被监测主机名|
|`unit`|设备内单元名|

- Metrics List


| Metrics | Description| Data Type | Unit   |
| ---- |---- | :---:    | :----: |
|`count`|Count.|int|count|
|`current`|Current.|float|ampere|
|`fan_speed`|Fan speed.|int|RPM|
|`power_consumption`|Power consumption.|float|watt|
|`status`|Status of the unit.|int|-|
|`temp`|Temperature.|float|C|
|`usage`|Usage.|float|percent|
|`voltage`|Voltage.|float|volt|
|`warning`|Warning on/off.|int|-|



## Configuration of Alarm Notification for Service Withdrawal of Monitored Equipment {#warning-config}

```
 [Monitor]-> [Monitor]-> [New Monitor] Select [Threshold Detection]-> Enter [Rule Name]
 Select [indicator]-> [indicator set], select [ipmi]-> [specific indicator], select [warning]-> next column, select [Max]-> by [detection dimension], select [host]
 Input [999] in [Urgent] -> Input [1] in [Important] -> Input [888] in [Warning] -> Input [N] in [Normal]
```

[1]: https://github.com/ipmitool/ipmitool
