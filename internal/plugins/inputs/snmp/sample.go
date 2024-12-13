// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmp

// nolint:lll
const sampleCfg = `
[[inputs.snmp]]
  ## Filling in specific device IP address, example ["10.200.10.240", "10.200.10.241"].
  ## And you can use auto_discovery and specific_devices at the same time.
  ## If you don't want to specific device, you don't need provide this.
  # specific_devices = ["***"] # SNMP Device IP.

  ## Filling in autodiscovery CIDR subnet, example ["10.200.10.0/24", "10.200.20.0/24"].
  ## If you don't want to enable autodiscovery feature, you don't need provide this.
  # auto_discovery = ["***"] # Used in autodiscovery mode only, ignore this in other cases.

  ## Consul server url for consul discovery
  ## We can discovery snmp instance from consul services
  # consul_discovery_url = "http://127.0.0.1:8500"

  ## Consul token, optional.
  # consul_token = "<consul token>"

  ## Instance ip key name. ("IP" case sensitive)
  # instance_ip_key = "IP"

  ## Witch task will collect, according to consul service filed "Address"
  ## [] mean collect all, optional, default to []
  # exporter_ips = ["<ip1>", "<ip2>"...]

  ## Consul TLS connection config, optional.
  # ca_certs = ["/opt/tls/ca.crt"]
  # cert = "/opt/tls/client.crt"
  # cert_key = "/opt/tls/client.key"
  # insecure_skip_verify = true

  ## SNMP protocol version the devices using, fill in 2 or 3.
  ## If you using the version 1, just fill in 2. Version 2 supported version 1.
  ## This is must be provided.
  snmp_version = 2

  ## SNMP port in the devices. Default is 161. In most cases, you don't need change this.
  ## This is optional.
  # port = 161

  ## Password in SNMP v2, enclose with single quote. Only worked in SNMP v2.
  ## If you are using SNMP v2, this is must be provided.
  ## If you are using SNMP v3, you don't need provide this.
  # v2_community_string = "***"

  ## Authentication stuff in SNMP v3.
  ## If you are using SNMP v2, you don't need provide this.
  ## If you are using SNMP v3, this is must be provided.
  # v3_user = "***"
  # v3_auth_protocol = "***"
  # v3_auth_key = "***"
  # v3_priv_protocol = "***"
  # v3_priv_key = "***"
  # v3_context_engine_id = "***"
  # v3_context_name = "***"

  ## Number of workers used to collect and discovery devices concurrently. Default is 100.
  ## Modifying it based on device's number and network scale.
  ## This is optional.
  # workers = 100

  ## Number of max OIDs during walk(default 1000)
  # max_oids = 1000

  ## Interval between each auto discovery in seconds. Default is "1h".
  ## Only worked in auto discovery feature.
  ## This is optional.
  # discovery_interval = "1h"

  ## Collect metric interval, default is 10s. (optional)
  # metric_interval = "10s"

  ## Collect object interval, default is 5m. (optional)
  # object_interval = "5m"

  ## Filling in excluded device IP address, example ["10.200.10.220", "10.200.10.221"].
  ## Only worked in auto discovery feature.
  ## This is optional.
  # discovery_ignored_ip = []

  ## Set true to enable election
  # election = true

  ## Device Namespace. Default is "default".
  # device_namespace = "default"

  ## Picking the metric data only contains the field's names below.
  # enable_picking_data = true # Default is "false", which means collecting all data.
  # status = ["sysUpTimeInstance", "tcpCurrEstab", "ifAdminStatus", "ifOperStatus", "cswSwitchState"]
  # speed = ["ifHCInOctets", "ifHCInOctetsRate", "ifHCOutOctets", "ifHCOutOctetsRate", "ifHighSpeed", "ifSpeed", "ifBandwidthInUsageRate", "ifBandwidthOutUsageRate"]
  # cpu = ["cpuUsage"]
  # mem = ["memoryUsed", "memoryUsage", "memoryFree"]
  # extra = []

  ## The matched tags would be dropped.
  # tags_ignore = ["Key1","key2"]

  ## The regexp matched tags would be dropped.
  # tags_ignore_regexp = ["^key1$","^(a|bc|de)$"]

  ## Zabbix profiles
  # [[inputs.snmp.zabbix_profiles]]
    ## Can be full path file name or only file name.
    ## If only file name, the path is "./conf.d/snmp/userprofiles/
    ## Suffix can be .yaml .yml .xml
    # profile_name = "xxx.yaml"
    ## ip_list is optional
    # ip_list = ["ip1", "ip2"]
    ## Device class, Best to use the following words:
    ## access_point, firewall, load_balancer, pdu, printer, router, sd_wan, sensor, server, storage, switch, ups, wlc, net_device
    # class = "server"

  # [[inputs.snmp.zabbix_profiles]]
    # profile_name = "yyy.xml"
    # ip_list = ["ip3", "ip4"]
    # class = "switch"

  # ...

  ## Prometheus snmp_exporter profiles, 
  ## If module mapping different class, can disassemble yml file.
  # [[inputs.snmp.prom_profiles]]
    # profile_name = "xxx.yml"
    ## ip_list useful when xxx.yml have 1 module 
    # ip_list = ["ip1", "ip2"]
    # class = "net_device"

  # ...

  ## Prometheus consul discovery module mapping.  ("type"/"isp" case sensitive)
  # [[inputs.snmp.module_regexps]]
    # module = "vpn5"
    ## There is an and relationship between step regularization
    # step_regexps = [["type", "vpn"],["isp", "CT"]]

  # [[inputs.snmp.module_regexps]]
    # module = "switch"
    # step_regexps = [["type", "switch"]]

  # ...
    
  ## Field key or tag key mapping. Do NOT edit.
  [inputs.snmp.key_mapping]
    CNTLR_NAME = "unit_name"
    DISK_NAME = "unit_name"
    ENT_CLASS = "unit_class"
    ENT_NAME = "unit_name"
    FAN_DESCR = "unit_desc"
    IF_OPERS_TATUS = "unit_status"
    IFADMINSTATUS = "unit_status"
    IFALIAS = "unit_alias"
    IFDESCR = "unit_desc"
    IFNAME = "unit_name"
    IFOPERSTATUS = "unit_status"
    IFTYPE = "unit_type"
    PSU_DESCR = "unit_desc"
    SENSOR_LOCALE = "unit_locale"
    SNMPINDEX = "snmp_index"
    SNMPVALUE = "snmp_value"
    TYPE = "unit_type"
    SENSOR_INFO = "unit_desc"
    ## We can add more mapping below
    # dev_fan_speed = "fanSpeed"
    # dev_disk_size = "diskTotal
  
  ## Reserved oid-key mappings. Do NOT edit.
  [inputs.snmp.oid_keys]
    "1.3.6.1.2.1.1.3.0" = "netUptime"
    "1.3.6.1.2.1.25.1.1.0" = "uptime"
    "1.3.6.1.2.1.2.2.1.13" = "ifInDiscards"
    "1.3.6.1.2.1.2.2.1.14" = "ifInErrors"
    "1.3.6.1.2.1.31.1.1.1.6" = "ifHCInOctets"
    "1.3.6.1.2.1.2.2.1.19" = "ifOutDiscards"
    "1.3.6.1.2.1.2.2.1.20" = "ifOutErrors"
    "1.3.6.1.2.1.31.1.1.1.10" = "ifHCOutOctets"
    "1.3.6.1.2.1.31.1.1.1.15" = "ifHighSpeed"
    "1.3.6.1.2.1.2.2.1.8" = "ifNetStatus"
    ## We can add more oid-key mapping below

  # [inputs.snmp.tags]
    # tag1 = "val1"
    # tag2 = "val2"

  [inputs.snmp.traps]
    enable = true
    bind_host = "0.0.0.0"
    port = 9162
    stop_timeout = 3    # stop timeout in seconds.
`
