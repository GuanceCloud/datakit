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

  ## Rootfs mount point for container/Kubernetes mode only
  ## DataKit uses this as the host root prefix when auto-prefixing absolute paths
  ## and preparing host-side systemd libraries (copy_node_libs).
  mount_dir = "/rootfs"

  ## Journal directory paths
  ## Host installation: use default paths
  ## Container/Kubernetes: DataKit auto-prefixes absolute paths with mount_dir.
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

  ## Host-side systemd library prepare:
  ## - Container/Kubernetes (Docker or Kubernetes): auto forced to true.
  ## - Non-container host: disabled by default. If enabled manually, set copy_node_libs_files explicitly.
  ## - In container/kubernetes mode, when copy_node_libs_files is empty, DataKit first copies
  ##   libsystemd.so* then runs "LD_LIBRARY_PATH=<dst> ldd libsystemd.so.0"
  ##   style dependency probing and copies missing .so files automatically.
  # copy_node_libs = true
  ## Optional override file list. If set, only these patterns/files are copied.
  # copy_node_libs_files = [
  #   "libsystemd.so*",
  #   "liblz4.so*",
  #   "libzstd.so*",
  #   "liblzma.so*",
  #   "libcap.so*",
  #   "libgcrypt.so*",
  #   "libgpg-error.so*",
  #   "libselinux.so*",
  #   "libmount.so*",
  #   "libblkid.so*",
  #   "libacl.so*",
  #   "libpcre2-8.so*",
  #   "libpcre.so*",
  # ]

  ## Additional arguments for external binary
  # args = []

  [inputs.journald.tags]
    # Add custom tags as needed
    # environment = "production"
    # cluster = "k8s-cluster-1"
`
