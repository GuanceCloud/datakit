// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ipmi

const sampleCfg = `
[[inputs.ipmi]]
  ## If you have so many servers that 10 seconds can't finish the job.
  ## You can start multiple collectors.

  ## (Optional) Collect interval: (defaults to "10s").
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

  ## (Optional) Provide the hex key for the IMPI connection: (defaults to []string{}).
  ## If len(hex_keys)<len(ipmi_ips), will use hex_keys[0].
  ## (Example) hex_keys = ["XXXX"]
  # hex_keys = []

  ## (Optional) Schema Version: (defaults to [1]).input.go
  ## If len(metric_versions)<len(ipmi_ips), will use metric_versions[0].
  ## (Example) metric_versions = [2]
  metric_versions = [2]

  ## (Optional) Exec ipmitool timeout: (defaults to "5s").
  timeout = "5s"

  ## (Optional) Ipmi server drop warning delay: (defaults to "300s").
  ## (Example) drop_warning_delay = "300s"
  drop_warning_delay = "300s"

  ## Key words of current.
  ## (Example) regexp_current = ["current"]
  regexp_current = ["current"]

  ## Key words of voltage.
  ## (Example) regexp_voltage = ["voltage"]
  regexp_voltage = ["voltage"]

  ## Key words of power.
  ## (Example) regexp_power = ["pwr","power"]
  regexp_power = ["pwr","power"]

  ## Key words of temp.
  ## (Example) regexp_temp = ["temp"]
  regexp_temp = ["temp"]

  ## Key words of fan speed.
  ## (Example) regexp_fan_speed = ["fan"]
  regexp_fan_speed = ["fan"]

  ## Key words of usage.
  ## (Example) regexp_usage = ["usage"]
  regexp_usage = ["usage"]

  ## Key words of usage.
  ## (Example) regexp_count = []
  # regexp_count = []

  ## Key words of status.
  ## (Example) regexp_status = ["fan"]
  regexp_status = ["fan"]

[inputs.ipmi.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
