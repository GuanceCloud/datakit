
# SNMP
---

{{.AvailableArchs}}

---

This article focuses on [SNMP](https://en.wikipedia.org/wiki/Simple_Network_Management_Protocol/){:target="_blank"} data collection.

## About SNMP Protocol {#config-pre}

The SNMP protocol is divided into three versions: v1/v2c/v3, of which:

    - V1 and v2c are compatible. Many SNMP devices only offer v2c and v3 versions. v2c version, the best compatibility, many older devices only support this version.
    - If the safety requirements are high, choose v3. Security is also the main difference between v3 version and previous versions.

If you choose v2c version, you need to provide `community string`, which translates into `community name/community string`in Chinese. An `unencrypted password` is a password, which is required for authentication when interacting with an SNMP device. In addition, some devices will be distinguished into `read-only community name` and `read-write community name`. As the name implies:

- `Read-only community name`: The device will only provide internal metrics data to that party, and cannot modify some internal configurations (this is enough for DataKit).
- `Read and write community name`: The provider has the permission to query the internal metrics data of the equipment and modify some configurations.

If you choose v3 version, you need to provide `username`, `authentication algorithm/password`, `encryption algorithm/password`, `context`, etc. Each device is different and should be filled in as required.

## Configure Collector {#config-input}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](datakit-service-how-to.md#manage-service) is sufficient.

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](datakit-daemonset-deploy.md#configmap-setting).

---

???+ attention

    If the `inputs.snmp.tags` configured above duplicates the key in the original fields with the same name, it will be overwritten by the original data.


### Configure SNMP {#config-snmp}

- On the device side, configure the SNMP protocol

When SNMP devices are in the default, the general SNMP protocol is closed, you need to enter the management interface to open manually. At the same time, it is necessary to select the protocol version and fill in the corresponding information according to the actual situation.

???+ tip

    Some devices require additional configuration to release SNMP for security, which varies from device to device. For example, Huawei is a firewall, so it is necessary to check SNMP in "Enable Access Management" to release it. You can use the `snmpwalk` command to test whether the acquisition side and the device side are configured to connect successfully:
    
    ```shell
    # Applicable v2c version
    snmpwalk -O bentU -v 2c -c [community string] [IP] 1.3.6` 
    # Applicable v3 version
    snmpwalk -v 3 -u user -l authPriv -a sha -A [认证密码] -x aes -X [加密密码] [IP] 1.3.6 
    ```
    
    If there is no problem with the configuration, the command will output a large amount of data. `snmpwalk` is a test tool running on the collection side, which comes with MacOS. Linux installation method:
    
    ```shell
    sudo yum install net–snmp–utils # CentOS
    sudo apt–get install snmp       # Ubuntu
    ```

- On the DataKit side, configure collection.

## Custom Device OID c=Configuration {#custom-oid}

If you find that the data reported by the collected device does not contain the indicators you want, then you may need to define an additional Profile for the device.

All OIDs of devices can generally be downloaded from their official website. Datakit defines some common OIDs, as well as some devices such as Cisco/Dell/HP. According to snmp protocol, each device manufacturer can customize [OID](https://www.dpstele.com/snmp/what-does-oid-network-elements.php) to identify its internal special objects. If you want to identify these, you need to customize the configuration of the device (we call this configuration Profile here, that is, "Custom Profile"), as follows.

Create the yml file `cisco-3850.yaml` under the path `conf.d/snmp/profiles` of the Datakit installation directory (in this case, Cisco 3850) as follows:

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

As shown above, a device with `sysobjectid` of `1.3.6.1.4.1.9.1.1745` is defined, and the file will be applied the next time Datakit collects a device with the same `sysobjectid`, in which case the collected data with an OID of `1.3.6.1.4.1.9.3.6.3.0` will be reported as an indicator with the name `chassisId`.

> Note: The folder `conf.d/snmp/profiles` requires the SNMP collector to run once before it appears.

## Measurements {#measurements}

All of the following data collections are appended by default with the name `host` (the value is the name of the SNMP device), or other labels can be specified in the configuration by `[inputs.snmp.tags]`:

``` toml
 [inputs.snmp.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

???+ attention

    All the following measurements and their metrics contain only some common fields, some device-specific fields, and some additional fields will be added according to different configurations and device models.

### Metrics {#metrics}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- field list

{{$m.FieldsMarkdownTable}} {{end}}

{{ end }}


### Object {#objects}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

#### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- field list

{{$m.FieldsMarkdownTable}} {{end}}

{{ end }}
