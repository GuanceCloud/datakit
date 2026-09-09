// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmp

// nolint:lll
const sampleCfg = `
[[inputs.snmp]]
  ## SNMP version: 1, 2 (v2c), or 3.
  snmp_version = 2
  # port = 161

  ## Community string, required for SNMP v1/v2c.
  # v2_community_string = ""

  ## SNMP v3 credentials; authentication and privacy settings must match the device.
  # v3_user              = ""
  # v3_auth_protocol     = "" # MD5/SHA/SHA224/SHA256/SHA384/SHA512 or empty
  # v3_auth_key          = ""
  # v3_priv_protocol     = "" # DES/AES/AES192/AES192C/AES256/AES256C or empty
  # v3_priv_key          = ""
  # v3_context_engine_id = "" # optional
  # v3_context_name      = "" # optional

  ## Namespace used in device identity; cannot be overridden by custom tags.
  # device_namespace = "default"

  ## Device IPs, e.g. ["10.200.10.240"]; can be combined with auto_discovery.
  ## With Zabbix/Prometheus profiles, use their ip_list instead.
  # specific_devices = [""]

  ## Autodiscovery subnets in CIDR notation, e.g. ["10.200.10.0/24"].
  # auto_discovery       = [""]
  # discovery_interval   = "1h"
  # discovery_ignored_ip = [] # Device IPs excluded from autodiscovery.

  ## Consul discovery with Prometheus profiles and module_regexps below.
  # consul_discovery_url = "http://127.0.0.1:8500"
  # consul_token         = "<consul token>"  # Optional.
  # instance_ip_key      = "IP"              # Device IP metadata key (case-sensitive).
  # exporter_ips         = ["<ip1>", "<ip2>"] # Filter by service Address; defaults to [] (all).
  ## TLS settings for Consul.
  # ca_certs             = ["/opt/tls/ca.crt"]
  # cert                 = "/opt/tls/client.crt"
  # cert_key             = "/opt/tls/client.key"
  # insecure_skip_verify = true

  ## Metric collection interval.
  # metric_interval = "10s"

  ## Device object collection; optional topology links follow object_interval.
  ## Topology uses built-in Profiles, preferring LLDP with CDP as a fallback.
  # object_interval  = "5m"
  # collect_topology = false

  ## Collect LLDP neighbors as snmp_lldp logs, independently of object topology.
  ## Enable with collect_topology only if both outputs are needed.
  # enable_lldp   = false
  # lldp_interval = "10m"

  ## Collection concurrency and request sizes.
  # workers              = 100  # Concurrent discovery and collection workers.
  # max_oids             = 1000 # Maximum OIDs allowed per Get call.
  # oid_batch_size       = 5    # OIDs requested per Get/GetBulk call.
  # bulk_max_repetitions = 10   # Maximum repetitions per GetBulk call.

  ## Enable election.
  # election = true

  ## Optional filtering for built-in Profiles; collect data containing any listed field.
  # enable_picking_data = true # Defaults to false (collect all metric data).
  # status = ["sysUpTimeInstance", "tcpCurrEstab", "ifAdminStatus", "ifOperStatus", "cswSwitchState"]
  # speed = ["ifHCInOctets", "ifHCInOctetsRate", "ifHCOutOctets", "ifHCOutOctetsRate", "ifHighSpeed", "ifSpeed", "ifBandwidthInUsageRate", "ifBandwidthOutUsageRate"]
  # cpu = ["cpuUsage"]
  # mem = ["memoryUsed", "memoryUsage", "memoryFree"]
  # extra = []

  ## Drop tags by exact key or regular expression.
  # tags_ignore        = ["Key1","key2"]
  # tags_ignore_regexp = ["^key1$","^(a|bc|de)$"]

  ## Zabbix profiles.
  # [[inputs.snmp.zabbix_profiles]]
    ## Full path or file name under ./conf.d/snmp/userprofiles/ (.yaml, .yml, or .xml).
    # profile_name = "xxx.yaml"
    ## Optional device IPs.
    # ip_list = ["ip1", "ip2"]
    ## Recommended device classes:
    ## access_point, firewall, load_balancer, pdu, printer, router, sd_wan, sensor, server, storage, switch, ups, wlc, net_device
    # class = "server"

  # [[inputs.snmp.zabbix_profiles]]
    # profile_name = "yyy.xml"
    # ip_list = ["ip3", "ip4"]
    # class = "switch"

  # ...

  ## Prometheus snmp_exporter profiles; split files by device class if needed.
  # [[inputs.snmp.prom_profiles]]
    # profile_name = "xxx.yml"
    ## ip_list applies only to single-module profiles.
    # ip_list = ["ip1", "ip2"]
    # class = "net_device"

  # ...

  ## Map Consul services to Prometheus modules; metadata keys are case-sensitive.
  # [[inputs.snmp.module_regexps]]
    # module = "vpn5"
    ## All regular expressions must match (AND).
    # step_regexps = [["type", "vpn"],["isp", "CT"]]

  # [[inputs.snmp.module_regexps]]
    # module = "switch"
    # step_regexps = [["type", "switch"]]

  # ...

  ## Field and tag key mappings for Zabbix/Prometheus profiles. Do NOT edit existing entries.
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
    ## Add custom mappings below.
    # dev_fan_speed = "fanSpeed"
    # dev_disk_size = "diskTotal"

  ## OID-to-key mappings for Zabbix/Prometheus profiles. Do NOT edit existing entries.
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
    ## Add custom OID-to-key mappings below.

  # [inputs.snmp.tags]
    # tag1 = "val1"
    # tag2 = "val2"

  ## Enable to receive device traps and report them as logs.
  [inputs.snmp.traps]
    enable = false
    bind_host = "0.0.0.0"
    port = 9162
    stop_timeout = 3    # Shutdown timeout in seconds.
    # source = "traps"
`
