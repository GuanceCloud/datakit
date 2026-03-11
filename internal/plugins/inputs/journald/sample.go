// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package journald

var sampleConfig = `
# Collect systemd journal logs using external binary
[[inputs.journald]]
  ## Name of the collector
  name = 'journald'

  ## Run as daemon (required for journald collection)
  daemon = true

  http_endpoint = "http://localhost:9529"
  log_level = "info"
  log_path = "/usr/local/datakit/externals/journald.log"

  ## Path to datakit-journald binary
  ## Default: searches in /usr/local/datakit/externals/datakit-journald and ./externals/datakit-journald
  # cmd = "/usr/local/datakit/externals/datakit-journald"

  ## Interval to check external process (for non-daemon mode)
  # interval = "10s"

  ## Journal directory paths
  ## Host installation: use default paths
  ## Kubernetes: use /rootfs prefixed paths (e.g., /rootfs/var/log/journal)
  paths = [
    "/var/log/journal",      # Persistent storage
    "/run/log/journal",      # Runtime storage
  ]

  ## Filter by systemd unit names (supports glob patterns)
  ## Empty = all units
  # units = ["*.service", "docker.service", "kubelet.service"]

  ## Filter by priority levels
  ## Levels: emerg(0), alert(1), crit(2), err(3), warning(4), notice(5), info(6), debug(7)
  ## Empty = all priorities
  # priorities = ["err", "warning", "crit", "alert", "emerg"]

  ## Field selection - collect all by default, exclude specific fields
  exclude_fields = [
    "_BOOT_ID",
    "_MACHINE_ID",
    "__MONOTONIC_TIMESTAMP",
  ]

  ## Collection behavior
  ## tail_only=true: Only collect new entries (cursor not needed)
  ## tail_only=false: Read from last position (cursor required)
  tail_only = true
  max_entries_per_batch = 1000

  ## Cursor management (only used when tail_only=false)
  # save_cursor = true
  # cursor_file = "/usr/local/datakit/cache/journald.cursor"

  ## Environment variables for external binary
  # envs = [
  #   "LD_LIBRARY_PATH=/usr/local/datakit/externals:$LD_LIBRARY_PATH",
  # ]

  ## Additional arguments for external binary
  # args = []

  [inputs.journald.tags]
    # Add custom tags as needed
    # environment = "production"
    # cluster = "k8s-cluster-1"
`
