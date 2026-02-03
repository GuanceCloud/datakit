// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package hostchange

var sampleCfg = `
[[inputs.host_change]]
  ## Interval between collections
  interval = "1m"

  ## Enable user and group change detection
  [inputs.host_change.user_group]
    ## Whether to enable user and group change detection
    enabled = true

  ## Crontab change detection configuration
  # Collect files from /etc/crontab, /etc/cron.d/*, /var/spool/cron/crontabs/*
  [inputs.host_change.crontab]
    ## Whether to enable crontab change detection
    enabled = true

  ## File change detection configuration
  [inputs.host_change.file]
    ## Whether to enable file change detection
    enabled = false

    ## Files to monitor for changes
    # Notes:
    # 1. Only regular files are supported, directories are not allowed
    # 2. All paths must be absolute paths
    files = [
      # "/etc/passwd",
      # "/etc/group",
      # "/etc/sudoers"
    ]

    ## Files larger than this size will not compare full content.
    ## Default value: 262144 bytes (256KB)
    max_file_size = 262144 

    ## Paths to ignore when monitoring file changes
    ignore_paths = [
      # "/etc/ssh/sshd_config.d/*",
      # "/tmp/",
      # "*.tmp"
    ]

  ## Service change detection configuration
  [inputs.host_change.service]
    ## Whether to enable service change detection
    enabled = true

    ## Service types to monitor (systemd, sysvinit)
    # If empty, all service types will be monitored and systemd is preferred when both are available
    service_types = ["systemd"]

    ## Services to ignore (service names without .service suffix, supports regex)
    ignore_services = []

    ## Services to include (service names without .service suffix, supports regex)
    # If not empty, only services matching these patterns will be monitored
    # include_services = []

  ## Network configuration change detection configuration
  [inputs.host_change.network]
    ## Whether to enable network configuration change detection
    enabled = true

    ## Interfaces to ignore (interface names, supports wildcard)
    ignore_interfaces = [
      # "lo",
      # "docker*",
      # "veth*"
    ]

  [inputs.host_change.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
`
