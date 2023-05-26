
# SNMP
---

{{.AvailableArchs}}

---

This article focuses on [SNMP](https://en.wikipedia.org/wiki/Simple_Network_Management_Protocol/){:target="_blank"} data collection.

## Terminology  {#terminology}

- `SNMP` (Simple network management protocol): A network protocol that is used to collect information about bare metal networking gear.
- `OID` (Object identifier): A unique ID or address on a device that when polled returns the response code of that value. For example, OIDs are CPU or device fan speed.
- `sysOID` (System object identifier): A specific address that defines the device type. All devices have a unique ID that defines it. For example, the Meraki base sysOID is `1.3.6.1.4.1.29671`.
- `MIB` (Managed information base): A database or list of all the possible OIDs and their definitions that are related to the MIB. For example, the `IF-MIB` (interface MIB) contains all the OIDs for descriptive information about a device’s interface.

## About SNMP Protocol {#config-pre}

The SNMP protocol is divided into three versions: v1/v2c/v3, of which:

- V1 and v2c are compatible. Many SNMP devices only offer v2c and v3 versions. v2c version, the best compatibility, many older devices only support this version.
- If the safety requirements are high, choose v3. Security is also the main difference between v3 version and previous versions.

Datakit supports all of the above versions.

### Choosing v1/v2c version {#config-v2}

If you choose v1/v2c version, you need to provide `community string`, AKA `community name/community string/unencrypted password`, which is required for authentication when interacting with an SNMP device. In addition, some devices will be distinguished into `read-only community name` and `read-write community name`. As the name implies:

- `Read-only community name`: The device will only provide internal metrics data to that party, and cannot modify some internal configurations (this is enough for DataKit).
- `Read-write community name`: The provider has the permission to query the internal metrics data of the equipment and modify some configurations.

### Choosing v3 version {#config-v3}

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

???+ tip

    Once the above configuration is complete, you can use the `datakit debug --input-conf` command to test if the configuration is correct, as shown in the following example:

    ```sh
    sudo datakit debug --input-conf /usr/local/datakit/conf.d/snmp/snmp.conf
    ```

    If correct the line protocol information would print out in output, otherwise no line protocol information is seen.

???+ attention

    1. If the `inputs.snmp.tags` configured above duplicates the key in the original fields with the same name, it will be overwritten by the original data.
    2. The IP address (required in specified device mode)/segment (required in auto-discovery mode) of the device, the version number of the SNMP protocol and the corresponding authentication fields are required.
    3. "Specified device mode" and "auto-discovery mode", the two modes can coexist, but the SNMP protocol version number and the corresponding authentification fields must be the same among devices.


### Configure SNMP {#config-snmp}

- On the device side, configure the SNMP protocol

When SNMP devices are in the default, the general SNMP protocol is closed, you need to enter the management interface to open manually. At the same time, it is necessary to select the protocol version and fill in the corresponding information according to the actual situation.

???+ tip

    Some devices require additional configuration to release SNMP for security, which varies from device to device. For example, Huawei is a firewall, so it is necessary to check SNMP in "Enable Access Management" to release it. You can use the `snmpwalk` command to test whether the acquisition side and the device side are configured to connect successfully(These commands runs on the host which Datakit running on):
    
    ```shell
    # Applicable v2c version
    snmpwalk -O bentU -v 2c -c [community string] [SNMP_DEVICE_IP] 1.3.6
    # Applicable v3 version
    snmpwalk -v 3 -u user -l authPriv -a sha -A [AUTH_PASSWORD] -x aes -X [ENCRYPT_PASSWORD] [SNMP_DEVICE_IP] 1.3.6
    ```
    
    If there is no problem with the configuration, the command will output a large amount of data. `snmpwalk` is a test tool running on the collection side, which comes with MacOS. Linux installation method:
    
    ```shell
    sudo yum install net-snmp net-snmp-utils # CentOS
    sudo apt–get install snmp                # Ubuntu
    ```

- On the DataKit side, configure collection.

## Advanced features {#advanced-features}

### Custom Device OID configuration {#advanced-custom-oid}

If you find that the data reported by the collected device does not contain the indicators you want, then you may need to define an additional Profile for the device.

All OIDs of devices can generally be downloaded from their official website. Datakit defines some common OIDs, as well as some devices such as Cisco/Dell/HP. According to snmp protocol, each device manufacturer can customize [OID](https://www.dpstele.com/snmp/what-does-oid-network-elements.php){:target="_blank"} to identify its internal special objects. If you want to identify these, you need to customize the configuration of the device (we call this configuration Profile here, that is, "Custom Profile"), as follows.

To add metrics or a custom configuration, list the MIB name, table name, table OID, symbol, and symbol OID, for example:

```yaml
- MIB: EXAMPLE-MIB
    table:
      # Identification of the table which metrics come from.
      OID: 1.3.6.1.4.1.10
      name: exampleTable
    symbols:
      # List of symbols ('columns') to retrieve.
      # Same format as for a single OID.
      # Each row in the table emits these metrics.
      - OID: 1.3.6.1.4.1.10.1.1
        name: exampleColumn1
```

Here is an example of operation.

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

???+ attention

    The folder `conf.d/snmp/profiles` requires the SNMP collector to run once before it appears.

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

## FAQ {#faq}

### How dows Datakit find devices? {#faq-discover}

Datakit supports "Specified device mode" and "auto-discovery mode" two modes. The two modes can enabled at the same time.

In "specified device mode", Datakit communicates with the specificed IP device using the SNMP protocol to know its current online status.

In "auto-discovery mode", Datakit sends SNMP packets to all address in the specified IP segment one by one, and if the response matches the corresponding profile, Datakit assumes that there is a SNMP device on that IP.

### I can't find metrics I'm looking for in [Guance](https://console.guance.com/){:target="_blank"}, what should I do?  {#faq-not-support}

Datakit collects generic base-line metrics from all devices. If you can't find the metric you want, you can [write a custom profile](snmp.md#advanced-custom-oid).

To archiving this, you probably needs to download the device's OID manual from its official website.

### Why I can't see any metrics in [Guance](https://console.guance.com/){:target="_blank"} after I completed configruation? {#faq-no-metrics}

Try loosening ACLs/firewall rules for your devices.

Run `snmpwalk -O bentU -v 2c -c <COMMUNITY_STRING> <IP_ADDRESS>:<PORT> 1.3.6` from the host Datakit is running on. If you get a timeout without any response, there is likely something blocking Datakit from collecting metrics from your device.
