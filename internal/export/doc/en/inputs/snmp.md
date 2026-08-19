---
title     : 'SNMP'
summary   : 'Collect metrics and object data from SNMP devices'
tags:
  - 'SNMP'
__int_icon      : 'icon/snmp'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
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

DataKit supports all of the above versions.

### Choosing v1/v2c version {#config-v2}

When selecting the v1/v2c version, a `v2_community_string` must be provided for authentication when interacting with SNMP devices. Additionally, some devices further subdivide this into *read-only community strings* and *read-write community strings*. As the names suggest:

- `Read-only community name`: The device will only provide internal metrics data to that party, and cannot modify some internal configurations (this is enough for DataKit).
- `Read-write community name`: The provider has the permission to query the internal metrics data of the equipment and modify some configurations.

### Choosing v3 version {#config-v3}

If you choose v3 version, you need to provide `v3_user/v3_auth_protocol/v3_auth_key/v3_priv_protocol/v3_priv_key`, etc. Each device is different and should be configured as same as configuration in SNMP device.

## Configuration {#config}

### Input Configuration {#config-input}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service) is sufficient.

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

---

???+ tip

    Once the above configuration is complete, you can use the `datakit debug --input-conf` command to test if the configuration is correct, as shown in the following example:

    ```sh
    sudo datakit debug --input-conf /usr/local/datakit/conf.d/snmp/snmp.conf
    ```

    If correct the line protocol information would print out in output, otherwise no line protocol information is seen.

???+ note

    1. If the `inputs.snmp.tags` configured above duplicates the key in the original fields with the same name, it will be overwritten by the original data.
    1. The IP address (required in specified device mode)/segment (required in auto-discovery mode) of the device, the version number of the SNMP protocol and the corresponding authentication fields are required.
    1. "Specified device mode" and "auto-discovery mode", the two modes can coexist, but the SNMP protocol version number and the corresponding authentication fields must be the same among devices.

<!-- markdownlint-enable MD046 -->

### Multiple configuration formats {#configuration-formats}

#### Built-in Profile Format {#advanced-custom-oid}

A built-in Profile uses YAML to describe the device `sysObjectID`, OIDs to collect, metric types, tags, and metadata.

This format is suitable when:

- The built-in Profiles do not cover the device model;
- Vendor-specific MIB metrics need to be added;
- The device returns numeric values as String/OCTET STRING;
- Device metadata needs to be added.

Profile YAML is different from Zabbix Template and Prometheus `snmp_exporter` formats. Their fields cannot be mixed.

##### Prepare OID Information {#profile-prepare-oid}

Before writing a Profile, obtain the device MIB/OID manual and use `snmpget` or `snmpwalk` to verify the actual values returned by the device. At minimum, confirm:

- The `sysObjectID`, which is the value of OID `1.3.6.1.2.1.1.2.0`;
- Whether each metric is a scalar or a table column;
- The metric OID, data type, meaning, and unit;
- The complete String/OCTET STRING value format;
- The table index and tag columns that distinguish each row.

##### Add a Profile {#profile-add}

DataKit releases built-in Profiles to `conf.d/snmp/profiles/` under the installation directory, and this directory may be overwritten during startup or upgrade. To add a site-specific Profile or override a built-in Profile, place the YAML file in `conf.d/snmp/extra_profiles/`. DataKit merges both directories during the default Profile loading flow, with `extra_profiles` taking precedence.

Example:

```text
/usr/local/datakit/conf.d/snmp/extra_profiles/vendor-router.yaml
```

File requirements:

- The extension must be `.yaml`;
- The file name must not start with `_`; files starting with `_` are inheritance templates only;
- If a file has the same name as a built-in Profile, the file in `extra_profiles` takes precedence; to keep the built-in content, extend the same file name in the supplemental Profile;
- Different file names should not use the same `sysobjectid`, otherwise automatic matching treats them as duplicate Profiles.

Restart DataKit after adding the file.

##### Basic Structure {#profile-structure}

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

Root-level fields:

| Field | Required | Description |
| --- | --- | --- |
| `extends` | No | Inherits other Profiles to reuse common metrics and metadata |
| `sysobjectid` | Required for automatic matching | A string or list of strings; supports the `*` wildcard |
| `device.vendor` | No | Generates the `device_vendor` tag and serves as the fallback device vendor metadata |
| `static_tags` | No | Fixed `key:value` tags applied to Profile data |
| `metadata` | No | Device metadata definitions |
| `metric_tags` | No | Profile-level dynamic tags generated from scalar OIDs |
| `metrics` | No | Scalar and table metric definitions |

`MIB` and `table` improve YAML readability but do not determine which SNMP queries are executed. DataKit queries the OIDs configured in `symbol`, `symbols`, `metric_tags`, and `metadata`.

##### Inherit Profiles {#profile-inheritance}

Use `extends` to inherit existing Profiles. Metrics, tags, and metadata from inherited Profiles are merged into the current Profile, reducing duplicate standard MIB configuration.

```yaml
extends:
  - _base.yaml
  - _generic-if.yaml
```

Built-in Profiles are located in `conf.d/snmp/profiles/`, and site-specific supplemental Profiles are located in `conf.d/snmp/extra_profiles/`. Before writing YAML, check the files starting with `_` in the built-in directory and select reusable templates provided by the current DataKit version. Common templates include:

| Profile | Purpose |
| --- | --- |
| `_base.yaml` | Common device tags and basic device metadata |
| `_generic-if.yaml` | IF-MIB interface metrics and interface metadata |
| `_generic-ip.yaml` | IP-MIB metrics |
| `_generic-tcp.yaml` | TCP-MIB metrics |
| `_generic-udp.yaml` | UDP-MIB metrics |
| `_generic-ospf.yaml` | OSPF-MIB metrics |
| `_generic-bgp4.yaml` | BGP4-MIB metrics |
| `_generic-lldp.yaml` | LLDP metrics and metadata |
| `_generic-entity-sensor.yaml` | ENTITY-SENSOR-MIB sensor metrics |
| `_generic-host-resources.yaml` | HOST-RESOURCES-MIB host resource metrics |
| `_generic-ups.yaml` | Common UPS metrics |
| `_cisco-generic.yaml` | Common Cisco interface, IP, TCP, UDP, OSPF, BGP, CPU, memory, and metadata definitions |
| `_huawei.yaml` | Common Huawei interface and vendor metadata definitions |
| `_juniper.yaml` | Common Juniper base configuration and vendor metadata |

Files starting with `_` are inheritance templates and do not independently participate in automatic `sysObjectID` matching. A complete Profile without the `_` prefix can also be inherited. For example:

```yaml
extends:
  - generic-router.yaml
```

`generic-router.yaml` already inherits `_base.yaml`, `_generic-if.yaml`, `_generic-ip.yaml`, `_generic-tcp.yaml`, `_generic-udp.yaml`, and `_generic-ospf.yaml`. Do not inherit these files again after inheriting `generic-router.yaml`.

For a Cisco device, inherit the common Cisco template and add device-specific metrics:

```yaml
extends:
  - _base.yaml
  - _cisco-generic.yaml

sysobjectid: 1.3.6.1.4.1.9.1.<product_id>
```

Notes:

- File names in `extends` are resolved from `conf.d/snmp/extra_profiles/` first, then from `conf.d/snmp/profiles/`;
- A same-name Profile in `extra_profiles` can extend its own file name to inherit the built-in Profile with the same name;
- Multiple and nested inheritance are supported, but circular inheritance is not allowed;
- Metrics, dynamic tags, and static tags are appended during merging;
- A metadata field defined by the current Profile is not overwritten by inherited content;
- Do not inherit a template that an upper-level Profile already includes, or duplicate metrics and tags may be generated;
- More inherited modules result in more OID queries. Select modules according to the MIBs supported by the device.

##### sysObjectID Matching {#profile-sysobjectid}

Exact match:

```yaml
sysobjectid: 1.3.6.1.4.1.99999.1.2
```

Wildcard match:

```yaml
sysobjectid: 1.3.6.1.4.1.99999.*
```

Match multiple models:

```yaml
sysobjectid:
  - 1.3.6.1.4.1.99999.1.*
  - 1.3.6.1.4.1.99999.2.*
```

When multiple Profiles match, DataKit selects the more specific rule. Different Profiles must not define the same `sysobjectid`. Use vendor- or product-specific values whenever possible to prevent a broad rule from matching unrelated devices.

For example, `1.3.6.1.4.1.8072.3.2.10` is a common Net-SNMP Linux `sysObjectID`, not a value exclusive to a specific device vendor.

##### Scalar Metrics {#profile-scalar-metrics}

A scalar has one value and is defined with `symbol`. Scalar OIDs usually end with `.0`:

```yaml
metrics:
  - MIB: VENDOR-SYSTEM-MIB
    symbol:
      OID: 1.3.6.1.4.1.99999.1.2.0
      name: vendor.system.cpu_usage
      metric_type: gauge
```

##### Table Metrics {#profile-table-metrics}

A table contains multiple rows. Use `symbols` to define the columns to collect and `metric_tags` to distinguish each row:

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

A table must have at least one tag that distinguishes its rows. Otherwise, multiple rows may have identical tags and only one row will be retained.

Tags can also be generated from the OID row index. `index` starts at `1`:

```yaml
metric_tags:
  - index: 1
    tag: disk_index
```

For a composite row index of `3.24`, the following configuration generates `slot:3` and `port:24`:

```yaml
metric_tags:
  - index: 1
    tag: slot
  - index: 2
    tag: port
```

##### Convert String Values to Numbers {#profile-string-to-number}

Metric field values must be numeric. A numeric String/OCTET STRING such as `"52.20"` can be converted directly. If the value contains a unit or other characters, use `extract_value` to extract the first capture group.

For a device value of `STRING: "5331MB"`:

```yaml
- OID: 1.3.6.1.4.1.99999.2.1.4
  name: vendor.disk.used
  extract_value: '^([0-9]+[.]?[0-9]*)MB$'
  metric_type: gauge
```

For a device value of `STRING: "52.20%"`:

```yaml
- OID: 1.3.6.1.4.1.99999.2.1.6
  name: vendor.disk.used_percent
  extract_value: '^([0-9]+[.]?[0-9]*)%?$'
  metric_type: gauge
```

Notes:

- The regular expression must contain at least one capture group. DataKit only uses the first group;
- Enclose regular expressions in single quotes and use `^` and `$` to match the complete value;
- Go regular expression syntax is used. Lookahead and lookbehind are not supported;
- A value is not reported when the regular expression does not match;
- Plain text String values should be configured as tags or metadata instead of time-series metric values.

##### Metric Types and Value Scaling {#profile-metric-type}

The following `metric_type` values are recommended for custom Profiles:

| Type | Use Case |
| --- | --- |
| `gauge` | Current values that can increase or decrease, such as utilization, temperature, capacity, and connection count |
| `monotonic_count` | Monotonically increasing cumulative values, such as byte, packet, and request counts |

`metric_type` can be configured on an individual `symbol`, or at the table metric root to apply to all its `symbols`. A value configured on an individual `symbol` takes precedence. When omitted, DataKit infers the type from the SNMP return type and uses `gauge` when it cannot infer a type.

Use `scale_factor` for numeric conversion. For example, if the device returns `5234` but the actual value is `52.34%`:

```yaml
- OID: 1.3.6.1.4.1.99999.1.2.0
  name: vendor.system.cpu_usage
  scale_factor: 0.01
  metric_type: gauge
```

`scale_factor` only affects the final reported metric value. For a String metric, use `extract_value` first and then apply `scale_factor`.

##### Report a Fixed Value of 1 {#profile-constant-value-one}

When a table represents a set of entities but has no suitable numeric column, use `constant_value_one` to report a fixed value of `1` for every row:

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

`constant_value_one` can only be used in table `symbols`, not scalar metrics. Do not configure an OID for this symbol, but configure its `name` and at least one OID-based `metric_tags.symbol` without `index_transform`, so DataKit can discover the table rows.

##### Tags {#profile-tags}

Use root-level `static_tags` to add fixed tags:

```yaml
static_tags:
  - "device_type:router"
  - "environment:production"
```

Root-level `metric_tags` generate dynamic tags from scalar OIDs and apply them to device metrics:

```yaml
metric_tags:
  - symbol:
      OID: 1.3.6.1.2.1.1.5.0
      name: sysName
    tag: snmp_host
```

Within a table, `metric_tags` can use `mapping` to convert raw values into readable tag values:

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

Use `match` and `tags` to generate multiple tags from one table column value:

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

For a raw value of `ethernet-12`, this generates `interface_type:ethernet` and `interface_number:12`. A non-empty `tags` mapping is required when `match` is configured. No tag is generated when the regular expression does not match.

When the metric table and tag source table use different index structures, use `index_transform` to extract and rebuild the index:

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

For a current index of `1.2.3.4.5.6.7.8`, the transformed index is `2.3.7.8`. Both `start` and `end` are zero-based, and the value at `end` is included. This option is only needed for cross-table association.

##### Metadata {#profile-metadata}

A custom Profile can add device metadata through `metadata.device`. Supported fields are `name`, `description`, `sys_object_id`, `location`, `serial_number`, `vendor`, `version`, `product_name`, `model`, `os_name`, `os_version`, `os_hostname`, and `type`. A field can obtain its value from a scalar OID or use a fixed `value`:

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

A field can define multiple candidate `symbols`. DataKit uses the first OID that returns a value. Use `match_pattern` and `match_value` to extract or replace text:

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

When `match_value` is omitted, the first capture group `$1` is used. If the regular expression does not match, DataKit tries the next candidate `symbol`.

After inheriting `_generic-if.yaml` or `generic-router.yaml`, common interface metadata is already included and usually does not need to be configured again.

##### Metric Naming {#profile-metric-naming}

Metric names are converted before reporting according to these rules:

1. If the name contains an underscore `_`, all dots `.` are replaced with underscores `_`;
1. If the name contains no underscore `_`, dots `.` are removed and the first letter after each dot is converted to uppercase;
1. A name without dots `.` remains unchanged.

Examples:

| Profile `name` | Reported Field Name |
| --- | --- |
| `sangfor.disk.used` | `sangforDiskUsed` |
| `sangfor.disk.used_percent` | `sangfor_disk_used_percent` |
| `vendor.interface.in_octets` | `vendor_interface_in_octets` |
| `cpu_usage` | `cpu_usage` |

The same conversion applies to tag keys but not tag values.

Use one naming style consistently within a Profile. A dot-separated name without underscores, such as `vendor.disk.used`, or an underscore name without dots, such as `vendor_disk_used`, can be used. Avoid mixing dots and underscores because all dots will then be converted to underscores. Existing built-in Profiles commonly use names such as `huawei.hwEntityTemperature`, which is reported as `huaweiHwEntityTemperature`.

##### Complete Example {#profile-example}

The following example inherits common router metrics and collects device metadata, a numeric scalar, String disk metrics, and an entity-presence metric:

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

##### Validation and Troubleshooting {#profile-troubleshooting}

Common issues:

- Profile not matched: Check the device's actual `sysObjectID`, wildcard range, and whether a more specific matching rule exists;
- YAML not loaded: Check the file extension, file name, indentation, and Profile validation errors in the DataKit logs;
- No OID data: Run `snmpget` or `snmpwalk` with the same SNMP version and authentication settings as DataKit;
- String metric not reported: Check that `extract_value` contains a capture group and matches the complete actual value;
- Only one table row reported: Add `metric_tags` that distinguish every row;
- Metric skipped: Confirm that the final value can be converted to a number, and check `metric_type` and `scale_factor`.

#### Zabbix format {#format-zabbix}

- Config

  ```toml
    [[inputs.snmp.zabbix_profiles]]
      profile_name = "xxx.yaml"
      ip_list = ["ip1", "ip2"]
      class = "server"
  
    [[inputs.snmp.zabbix_profiles]]
      profile_name = "yyy.xml"
      ip_list = ["ip3", "ip4"]
      class = "switch"
  
    # ...
  ```

  `profile_name` can be full path file name or only file name.
  If only file name, the path is *./conf.d/snmp/userprofiles/*

  profile_name can from Zabbix official, or from [community](https://github.com/zabbix/community-templates){:target="_blank"} .

  You can modify the yaml or xml.

- AutoDiscovery

    - Automatic discovery matches the collection rules in the imported multiple yaml configurations and performs collection.

    - Please try to configure according to class C. Configuring class B may be slower.

    - If automatic discovery fails to match yaml, it is because these yaml does not contain the manufacturer's signature code of the collected     device.

        - Add an oid message to the items of yaml to guide the automatic matching process.

          ```yaml
          zabbix_export:
            templates:
            - items:
              - snmp_oid: 1.3.6.1.4.1.2011.5.2.1.1.1.1.6.114.97.100.105.117.115.0.0.0.0
          ```

        - The oid to be added is obtained by executing the following command. .0.0.0.0 is added at the end to prevent the generation of useless     indicators.

          ```shell
          $ snmpwalk -v 2c -c public <ip> 1.3.6.1.2.1.1.2.0
          iso.3.6.1.2.1.1.2.0 = OID: iso.3.6.1.4.1.2011.2.240.12
          
          $ snmpgetnext -v 2c -c public <ip> 1.3.6.1.4.1.2011.2.240.12
          iso.3.6.1.4.1.2011.5.2.1.1.1.1.6.114.97.100.105.117.115 = STRING: "radius"
          ```

#### Prometheus format {#format-Prometheus}

- Config

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

    Please refer to the snmp.yml file of  Prometheus [snmp_exporter](https://github.com/prometheus/snmp_exporter){:target="_blank"}  for the profile.
    It is recommended to split [module](https://github.com/prometheus/snmp_exporter?tab=readme-ov-file#prometheus-configuration){:target="_blank"} of different classes into different .yml configurations.

    Prometheus profile allows you to configure a separate community name for a module.
    This community name takes precedence over the community name configured for the input.
  
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

- AutoDiscovery

  The SNMP collector can discovery instance through Consul service, and the service injection format can be found on [prom official website](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#consul_sd_config){:target="_blank"}。


### Configure SNMP device {#config-snmp}

When SNMP devices are in the default, the general SNMP protocol is closed, you need to enter the management interface to open manually. At the same time, it is necessary to select the protocol version and fill in the corresponding information according to the actual situation.

<!-- markdownlint-disable MD046 -->
???+ tip

    Some devices require additional configuration to release SNMP for security, which varies from device to device. For example, Huawei is a firewall, so it is necessary to check SNMP in "Enable Access Management" to release it. You can use the `snmpwalk` command to test whether the acquisition side and the device side are configured to connect successfully(These commands runs on the host which DataKit running on):
    
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

<!-- markdownlint-enable MD046 -->


#### SNMPv3 Example {#snmpv3-example}

We take an `snmpd` service on Linux as an example to demonstrate how to collecting SNMPv3 data.

- On an Ubuntu machine, install the `snmpd` service:

    ```shell
    sudo apt install snmp snmpd libsnmp-dev
    ```

- Prepare a simple *snmpd.conf* configuration as follows:

    ```conf title="my-snmpd.conf"
    # Set the UDP 161 port
    agentaddress udp:161

    # Define a user and related authentication configurations
    createUser snmpv3user1 SHA "authPassAgent1" AES "privPassAgent1"

    # Grant user access permissions (rouser: read-only, rwuser: read-write)
    rouser snmpv3user1 priv .1.3.6
    ```

- Stop the `snmpd` service first, then manually start the `snmpd` process:

    ```shell
    sudo /usr/sbin/snmpd -f -Lo -C \
      -p x.pid -Ddump,usm,acl,header,context,pdu,snmpv3 \
      -c my-snmpd.conf
    ```

- Use the `snmpwalk` command to test, which is expected to output extensive OID device information:

    ```shell
    snmpwalk -v3 -l authPriv \
      -u snmpv3user1 \
      -a SHA \
      -A "authPassAgent1" \
      -x AES \
      -X "privPassAgent1" \
      udp:127.0.0.1:161 .1.3.6.1.2.1.1
    ```

- If the `snmpwalk` command succeeds, enable the collector on DataKit to collect metrics via SNMPv3. The key configuration is as follows:

    ```toml title="conf.d/snmp/snmp.conf"
    specific_devices = ["127.0.0.1"] # Do not use "localhost" here
    snmp_version     = 3
    port             = 161

    v3_user                = "snmpv3user1"
    v3_auth_protocol       = "SHA" # MD5/SHA/SHA224/SHA256/SHA384/SHA512 or empty
    v3_auth_key            = "authPassAgent1"
    v3_priv_protocol       = "AES" # DES/AES/AES192/AES192C/AES256/AES256C or empty
    v3_priv_key            = "privPassAgent1"
    # v3_context_engine_id = "" # Optional
    # v3_context_name      = "" # Optional
    ```

## LLDP Network Topology Collection {#lldp-topology}

DataKit supports collecting LLDP (Link Layer Discovery Protocol) neighbor information from network devices via SNMP protocol for automatic network topology construction.

### What is LLDP {#what-is-lldp}

LLDP is a standardized link-layer protocol that allows network devices (such as switches and routers) to broadcast their identity and capability information to directly connected neighbor devices. By collecting LLDP data, you can:

- Automatically discover network topology relationships
- Understand physical connection status between devices
- Obtain neighbor device port, hostname, system description, and other information
- Build visualized network topology diagrams

### Enable LLDP Collection {#enable-lldp}

The recommended mode attaches LLDP neighbors to the `snmp_object` object as link metadata:

```toml
[[inputs.snmp]]
  ## Collect LLDP neighbors with object data and write topology links to
  ## the links field of snmp_object
  collect_topology = true
```

`collect_topology` is disabled by default. When enabled, DataKit queries the LLDP/CDP MIB during each object collection cycle and writes local-device, local-interface, remote-device, and remote-interface link metadata to the `links` field of `snmp_object` as a JSON array. LLDP is preferred when LLDP links are available; CDP is used only as a fallback.

This option only applies to the built-in Profile collection mode that reports `snmp_object`; it does not apply to `snmp_<class>` custom objects generated by user Profiles.

#### `links` Field {#topology-links-field}

The following is an annotated decoded LLDP example from the `links` field. Lines beginning with `//` describe the fields:

```jsonc
[
  {
    // Topology neighbor-record ID on the local device.
    // LLDP format: <device_namespace>:<management IP>:<lldpRemLocalPortNum>.<lldpRemIndex>;
    // CDP format: <device_namespace>:<management IP>:<cdpCacheIfIndex>.<cdpCacheDeviceIndex>.
    "id": "default:192.0.2.10:7.1",
    // Neighbor data source: lldp or cdp.
    "source_type": "lldp",
    // Collection integration name. It is currently fixed to snmp.
    "integration": "snmp",
    // Local link endpoint directly collected by DataKit.
    "local": {
      // Local device.
      "device": {
        // Resolved local device ID in the form <device_namespace>:<management IP>.
        "resolved_id": "default:192.0.2.10"
      },
      // Local interface.
      "interface": {
        // Resolved local interface ID in the form <local.device.resolved_id>:<ifIndex>.
        // It can be absent when an LLDP local port cannot be resolved to an IF-MIB interface.
        "resolved_id": "default:192.0.2.10:12",
        // LLDP lldpLocPortId; currently an empty string for a CDP link.
        "id": "82:a5:6e:a5:c9:01",
        // Local port ID type converted from lldpLocPortIdSubtype; usually absent from a CDP link.
        "id_type": "mac_address"
      }
    },
    // Remote endpoint discovered by the neighbor protocol.
    // It does not imply that DataKit directly collects the remote device.
    "remote": {
      // Remote device.
      "device": {
        // LLDP lldpRemChassisId or CDP cdpCacheDeviceId.
        "id": "01:00:00:00:01:02",
        // LLDP chassis ID type converted from lldpRemChassisIdSubtype; usually absent from a CDP link.
        "id_type": "mac_address",
        // LLDP lldpRemSysName or CDP cdpCacheSysName.
        "name": "switch-b",
        // LLDP lldpRemSysDesc or CDP cdpCacheVersion.
        "description": "remote switch",
        // Remote management IP. LLDP derives it from the remote-management-address table index;
        // CDP tries the primary, secondary, and cached management addresses. It is absent if not advertised.
        "ip_address": "10.250.0.6"
      },
      // Remote interface.
      "interface": {
        // LLDP lldpRemPortId or CDP cdpCacheDevicePort.
        "id": "Ethernet1/7",
        // LLDP remote port ID type; fixed to interface_name for CDP.
        "id_type": "interface_name",
        // LLDP lldpRemPortDesc; usually absent from a CDP link.
        "description": "remote uplink"
      }
    }
  }
]
```

LLDP device and port identifier types can include `mac_address`, `network_address`, `interface_name`, `interface_alias`, `port_component`, and `local`. Except for structurally required fields, names, descriptions, management IPs, and unresolved object IDs can be absent from the result.

The existing `enable_lldp` option is an independent LLDP collection path. It runs on `lldp_interval` and reports `snmp_lldp` logging data, and is retained for compatibility with existing configurations:

```toml
[[inputs.snmp]]
  enable_lldp = true
  lldp_interval = "10m"
```

The two options are independent and both default to `false`. Normally, enable only the output you need. If both are enabled, DataKit queries LLDP data on both schedules and produces both object-link and logging outputs.

### Device Configuration Requirements {#lldp-device-config}

Network devices need to:

- **Enable LLDP**

- **Configure SNMP access to LLDP MIB**


### Verify Configuration {#lldp-verify}

**Verify LLDP neighbors on device:**

```bash
# Huawei device
display lldp neighbor

# Cisco device
show lldp neighbor
```

**Verify SNMP can query LLDP data from DataKit host:**

```bash
# SNMPv2c verification
snmpwalk -v2c -c [COMMUNITY_STRING] [DEVICE_IP] 1.0.8802.1.1.2.1.4.1

# SNMPv3 verification
snmpwalk -v3 -u [USERNAME] -l authPriv \
  -a SHA -A [AUTH_PASSWORD] \
  -x AES -X [PRIV_PASSWORD] \
  [DEVICE_IP] 1.0.8802.1.1.2.1.4.1
```

## Metric {#metric}

For all of the following data collections, the global election tags will added automatically, we can add extra tags in `[inputs.{{.InputName}}.tags]` if needed:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

<!-- markdownlint-disable MD046 -->
???+ note

    All the following measurements and their metrics contain only some common fields, some device-specific fields, and some additional fields will be added according to different configurations and device models.
<!-- markdownlint-enable MD046 -->

<!-- markdownlint-disable MD024 -->
{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}

## Object {#objects}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{ end }}

{{ end }}

## Logging {#logging}

<!-- markdownlint-disable MD024 -->
{{ range $i, $m := .Measurements }}

{{if eq $m.Type "logging"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}
<!-- markdownlint-enable MD024 -->

<!-- markdownlint-disable MD013 -->
## FAQ {#faq}

### How dows DataKit find devices? {#faq-discover}

DataKit supports "Specified device mode" and "auto-discovery mode" two modes. The two modes can enabled at the same time.

In "specified device mode", DataKit communicates with the specified IP device using the SNMP protocol to know its current online status.

In "auto-discovery mode", DataKit sends SNMP packets to all address in the specified IP segment one by one, and if the response matches the corresponding profile, DataKit assumes that there is a SNMP device on that IP.

### Device Not Supported {#faq-not-support}

DataKit collects common baseline metrics from SNMP devices. If the collected data does not contain the required metrics, you may need to [add a custom Profile](snmp.md#advanced-custom-oid).

To do this, obtain the OID manual for the device model from the vendor's website.

### Can't see any metrics after configuration? {#faq-no-metrics}

<!-- markdownlint-enable MD013 -->

Try loosening ACLs/firewall rules for your devices.

Run `snmpwalk -O bentU -v 2c -c <COMMUNITY_STRING> <IP_ADDRESS>:<PORT> 1.3.6` from the host DataKit is running on. If you get a timeout without any response, there is likely something blocking DataKit from collecting metrics from your device.
