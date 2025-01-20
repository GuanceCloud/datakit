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

Datakit supports all of the above versions.

### Choosing v1/v2c version {#config-v2}

If you choose v1/v2c version, you need to provide `community string`, AKA `community name/community string/unencrypted password`, which is required for authentication when interacting with an SNMP device. In addition, some devices will be distinguished into `read-only community name` and `read-write community name`. As the name implies:

- `Read-only community name`: The device will only provide internal metrics data to that party, and cannot modify some internal configurations (this is enough for DataKit).
- `Read-write community name`: The provider has the permission to query the internal metrics data of the equipment and modify some configurations.

### Choosing v3 version {#config-v3}

If you choose v3 version, you need to provide `username`, `authentication algorithm/password`, `encryption algorithm/password`, `context`, etc. Each device is different and should be filled in as same as configuration in SNMP device.

## Configuration {#config}

### Input Configuration {#config-input}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
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

???+ attention

    1. If the `inputs.snmp.tags` configured above duplicates the key in the original fields with the same name, it will be overwritten by the original data.
    2. The IP address (required in specified device mode)/segment (required in auto-discovery mode) of the device, the version number of the SNMP protocol and the corresponding authentication fields are required.
    3. "Specified device mode" and "auto-discovery mode", the two modes can coexist, but the SNMP protocol version number and the corresponding authentication fields must be the same among devices.

<!-- markdownlint-enable -->

### Multiple configuration formats {#configuration-formats}

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

<!-- markdownlint-enable -->

## Metric {#metric}

For all of the following data collections, the global election tags will added automatically, we can add extra tags in `[inputs.{{.InputName}}.tags]` if needed:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

<!-- markdownlint-disable MD046 -->
???+ attention

    All the following measurements and their metrics contain only some common fields, some device-specific fields, and some additional fields will be added according to different configurations and device models.
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD024 -->
{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}

## Object {#objects}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "object"}}

### `{{$m.Name}}`

{{$m.Desc}}

- Tags

{{$m.TagsMarkdownTable}}

- Metrics

{{$m.FieldsMarkdownTable}}
{{ end }}

{{ end }}
<!-- markdownlint-disable MD013 -->
## FAQ {#faq}

### :material-chat-question: How dows Datakit find devices? {#faq-discover}

Datakit supports "Specified device mode" and "auto-discovery mode" two modes. The two modes can enabled at the same time.

In "specified device mode", Datakit communicates with the specified IP device using the SNMP protocol to know its current online status.

In "auto-discovery mode", Datakit sends SNMP packets to all address in the specified IP segment one by one, and if the response matches the corresponding profile, Datakit assumes that there is a SNMP device on that IP.

### :material-chat-question: I can't find metrics I'm looking for in [Guance](https://console.guance.com/){:target="_blank"}, what should I do?  {#faq-not-support}

Datakit collects generic base-line metrics from all devices. If you can't find the metric you want, you can [write a custom profile](snmp.md#advanced-custom-oid).

To archiving this, you probably needs to download the device's OID manual from its official website.

### :material-chat-question: Why I can't see any metrics in [Guance](https://console.guance.com/){:target="_blank"} after I completed configuration? {#faq-no-metrics}

<!-- markdownlint-enable -->

Try loosening ACLs/firewall rules for your devices.

Run `snmpwalk -O bentU -v 2c -c <COMMUNITY_STRING> <IP_ADDRESS>:<PORT> 1.3.6` from the host Datakit is running on. If you get a timeout without any response, there is likely something blocking Datakit from collecting metrics from your device.

