---
title     : 'Journald'
summary   : 'Collect systemd journal logs'
tags:
  - 'HOST'
  - 'LOG'
__int_icon      : 'icon/system'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

The Journald collector is used to collect logs from the systemd journal (journald) on Linux systems. It uses an external binary wrapper to interface with `libsystemd` and efficiently collects structured log entries from the journal.

## Prerequisites {#prerequisites}

- **Linux only**: Requires `systemd` and `journald`
- **libsystemd**: External binary requires `libsystemd` development libraries
- **Permissions**: DataKit needs read access to journal files (typically requires joining `systemd-journal` group)

### System Requirements Check {#system-requirements-check}

Before deploying the journald collector, verify your system meets the requirements:

Quick check with one-liner:

```bash
systemctl --version >/dev/null 2>&1 && journalctl -n 1 >/dev/null 2>&1 && echo "Systemd OK" || echo "Systemd not available"
```

Comprehensive pre-flight check script:

<!-- markdownlint-disable MD046 -->
??? info "*journald-prereq-check.sh*"

    ``` bash linenums="1"
    #!/bin/bash
    # journald-prereq-check.sh - Verify systemd requirements
    
    echo "=== Journald Collector Prerequisites Check ==="
    echo
    
    # 1. Check if systemctl exists
    echo -n "1. systemctl command: "
    if command -v systemctl >/dev/null 2>&1; then
        VERSION=$(systemctl --version | head -1)
        echo "✅ Found - $VERSION"
    else
        echo "❌ NOT FOUND - systemctl not installed"
        exit 1
    fi
    
    # 2. Check libsystemd library
    echo -n "2. libsystemd.so.0: "
    if ldconfig -p 2>/dev/null | grep -q "libsystemd.so.0"; then
        LIBPATH=$(ldconfig -p 2>/dev/null | grep "libsystemd.so.0" | head -1 | awk '{print $NF}')
        echo "✅ Found - $LIBPATH"
    else
        echo "❌ NOT FOUND - libsystemd.so.0 missing"
        exit 1
    fi
    
    # 3. Check journalctl access
    echo -n "3. journalctl access: "
    if journalctl -n 1 >/dev/null 2>&1; then
        echo "✅ OK - Can read journal"
    else
        echo "⚠️  LIMITED - journalctl exists but no read access"
    fi
    
    # 4. Check journal directories
    echo "4. Journal directories:"
    for dir in "/var/log/journal" "/run/log/journal"; do
        echo -n "   $dir: "
        if [ -d "$dir" ]; then
            if [ -r "$dir" ]; then
                echo "✅ Exists and readable"
            else
                echo "⚠️  Exists but NOT readable"
            fi
        else
            echo "❌ NOT FOUND"
        fi
    done
    
    # 5. Check systemd version
    echo -n "5. systemd version: "
    SYSTEMD_VERSION=$(systemctl --version | head -1 | grep -oP 'systemd \K\d+' || echo "0")
    if [ "$SYSTEMD_VERSION" -ge 205 ]; then
        echo "✅ v$SYSTEMD_VERSION (meets minimum v205)"
    else
        echo "⚠️  v$SYSTEMD_VERSION (older than recommended v205)"
    fi
    
    echo
    echo "=== Check Complete ==="
    ```
<!-- markdownlint-enable -->

Save as `journald-prereq-check.sh` and run:

```bash
chmod +x journald-prereq-check.sh
./journald-prereq-check.sh
```

Expected output:

``` txt
=== Journald Collector Prerequisites Check ===

1. systemctl command: ✅ Found - systemd 257 (257.3-1-arch)
2. libsystemd.so.0: ✅ Found - /usr/lib/libsystemd.so.0
3. journalctl access: ✅ OK - Can read journal
4. Journal directories:
   /var/log/journal: ✅ Exists and readable
   /run/log/journal: ✅ Exists and readable
5. systemd version: ✅ v257 (meets minimum v205)

=== Check Complete ===
```

Possible troubleshooting solutions:

| Issue | Solution |
|-------|----------|
| `systemctl: command not found` | Install systemd or use alternative log collection |
| `libsystemd.so.0: cannot open` | Install systemd-libs: `apt install libsystemd0` or `yum install systemd-libs` |
| `journalctl: no read access` | Add user to `systemd-journal` group: `usermod -aG systemd-journal $USER` |
| `/var/log/journal not found` | Enable persistent journal: `mkdir -p /var/log/journal && systemd-tmpfiles --create` |

## Configuration {#config}

### Collector Configuration {#collector-configuration}

After successfully installing and starting DataKit, enable the Journald collector by copying the configuration file:

<!-- markdownlint-disable MD046 -->

=== "Host Installation"

    Go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting).
<!-- markdownlint-enable -->

### Configuration Options {#config-options}

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `paths` | []string | `["/var/log/journal", "/run/log/journal"]` | Journal directory paths |
| `units` | []string | `[]` | Filter by systemd unit names (supports glob patterns, e.g., `*.service`) |
| `priorities` | []string | `[]` | Filter by priority levels: `emerg`, `alert`, `crit`, `err`, `warning`, `notice`, `info`, `debug` |
| `exclude_fields` | []string | `[]` | Journal fields to exclude from collection (e.g., `_BOOT_ID`, `_MACHINE_ID`) |
| `tail_only` | bool | `true` | Only collect new entries (skip historical logs on startup) |
| `max_entries_per_batch` | int | `1000` | Maximum number of entries to collect per batch |
| `save_cursor` | bool | `true` | Persist read position to resume after restart |
| `cursor_file` | string | `/usr/local/datakit/cache/journald.pos` | Path to store cursor position |
| `mount_dir` | string | `"/rootfs"` | Rootfs mount directory used in container/Kubernetes mode only. DataKit uses this prefix for absolute `paths` and as source root for host-side library prepare |
| `copy_node_libs` | bool | `false` (auto forced to `true` in container or Kubernetes mode) | Whether to copy host-side dynamic libraries from mount dir into DataKit-managed `external-libs` before starting the external collector. In container or Kubernetes environments (`datakit.Docker || config.IsKubernetes()`), DataKit auto-enables this |
| `copy_node_libs_files` | []string | `[]` | Dynamic library file names or glob patterns to copy. If configured, only these are copied. If empty in container/Kubernetes auto mode, DataKit first copies `libsystemd.so*`, then runs `LD_LIBRARY_PATH=/usr/local/datakit/externals/systemd-libs ldd libsystemd.so.0`-style dependency probing and copies missing `.so` automatically. If empty outside container/Kubernetes mode while `copy_node_libs=true`, startup fails with configuration error |

## Log Fields {#logging}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}

{{ end }}

## Common Use Cases {#use-cases}

- Collect logs from specific services

```toml
[[inputs.journald]]
  units = ["nginx.service", "mysql.service", "docker.service"]
  priorities = ["err", "crit", "alert", "emerg"]
  tail_only = true
```

- Exclude verbose fields

```toml
[[inputs.journald]]
  exclude_fields = [
    "_BOOT_ID",
    "_MACHINE_ID",
    "__MONOTONIC_TIMESTAMP",
    "_AUDIT_SESSION",
    "_AUDIT_LOGINUID",
  ]
```

- Kubernetes node journal collection (auto mode)

```toml
[[inputs.journald]]
  paths = ["/var/log/journal", "/run/log/journal"]
  tail_only = true
```

Notes:

- The collector resolves candidate directories in configuration order and tries to open the first readable journal directory first
- In container or Kubernetes environments (`datakit.Docker || config.IsKubernetes()`), DataKit auto-enables journald rootfs mode
- In container/Kubernetes mode, absolute paths are automatically prefixed with `mount_dir` (default `"/rootfs"`)
- If the configured path is a journal root such as `<mount_dir>/var/log/journal`, the collector automatically descends into the machine-id subdirectory before opening it
- In containerized node environments such as kind or k3d, validate `logger` and `journalctl` inside the node container rather than on the outer host

- Kubernetes node journal collection with host-side systemd library prepare

```toml
[[inputs.journald]]
  mount_dir = "/rootfs"
  paths = ["/var/log/journal", "/run/log/journal"]
  tail_only = true
  copy_node_libs = true
  copy_node_libs_files = [
    "libsystemd.so*",
    "liblz4.so*",
    "libzstd.so*",
    "liblzma.so*",
    "libcap.so*",
    "libgcrypt.so*",
    "libgpg-error.so*",
    "libselinux.so*",
    "libmount.so*",
    "libblkid.so*",
    "libacl.so*",
    "libpcre2-8.so*",
    "libpcre.so*",
  ]
```

- Collect all logs (debugging)

```toml
[[inputs.journald]]
  tail_only = false
  max_entries_per_batch = 500
  exclude_fields = []
```

## Troubleshooting {#troubleshooting}

### Permission errors {#troubleshoot-permissions}

Ensure DataKit has read access to journal files:

```bash
# Add datakit user to systemd-journal group
sudo usermod -aG systemd-journal datakit

# Restart DataKit
sudo systemctl restart datakit
```

### No logs collected {#troubleshoot-no-logs}

1. Verify journald is running:

```bash
systemctl status systemd-journald
```

1. Check journal files exist:

```bash
ls -la /var/log/journal/
ls -la /run/log/journal/
```

1. If `journalctl` is available in the current environment, use it for extra validation; if the container does not ship `journalctl`, rely on the DataKit compatibility warning and probe result directly:

```bash
journalctl -n 10
```

If startup logs report `reason=unsupported-format`, the collector runtime is older than the target journal file format. In this case DataKit keeps the journald collector inactive and logs a warning instead of collecting partial or misleading results.

This can happen in Kubernetes whenever DataKit collects journal files from the node while the container image ships an older `libsystemd` than the host journal format requires. Typical symptoms are:

- If `journalctl` is installed inside the Pod, it may report `unsupported feature`
- DataKit starts, but the journald collector stays inactive after the compatibility warning

In container or Kubernetes environments (`datakit.Docker || config.IsKubernetes()`), DataKit already auto-enables host-side `systemd` library prepare. If you need this behavior on non-container hosts, enable:

```toml
[[inputs.journald]]
  copy_node_libs = true
```

When enabled, DataKit copies dynamic libraries from candidate system library directories under `mount_dir` (default `"/rootfs"`) into its own `external-libs` directory, then prepends that directory to `LD_LIBRARY_PATH` automatically.

Copy behavior details:

- If `copy_node_libs_files` is configured and non-empty, DataKit copies only that list.
- If `copy_node_libs_files` is empty in container/Kubernetes auto mode, DataKit first copies `libsystemd.so*`, then probes missing dependencies with `ldd libsystemd.so.0` under the copied library path, and copies the missing `.so` files automatically.
- If `copy_node_libs_files` is empty on non-container and non-Kubernetes hosts while `copy_node_libs=true`, DataKit reports a configuration error and keeps the collector inactive.
- If library prepare fails while `copy_node_libs` is enabled, the journald collector stays inactive (other DataKit collectors are not affected).

After the collector opens the journal successfully, it also logs the effective `libsystemd` path in external `journald.log`, for example:

```text
loaded libsystemd paths: [/usr/local/datakit/externals/systemd-libs/libsystemd.so.0.35.0]
```

Constraints:

- The host `libsystemd` is not guaranteed to be compatible with the journald external binary currently shipped in DataKit
- If the host `libsystemd` is too old, the external binary may fail during dynamic linking because of missing symbols or version mismatches
- If the host `libsystemd` is newer, it may still fail later with `unsupported feature` when reading journal files
- Therefore, `copy_node_libs` is only a preparation mechanism, not a guarantee that the copied libraries are compatible; the final result still needs to be verified from startup logs and probe results

Do not point `LD_LIBRARY_PATH` at the entire host `/usr/lib64` directory. That can also pull incompatible glibc components into the collector process and create a less predictable failure mode.

If startup logs contain:

```text
resolved journal directory: target=...
opening journal from directory: ...
```

the collector is using directory-based journal opening, which is the recommended path for live journals. Avoid configuring individual `.journal` files as the primary input path.

### Cursor file issues {#troubleshoot-cursor}

If the cursor file becomes corrupted (e.g., after host reboot), the collector automatically falls back to tail mode and creates a new cursor. To manually reset:

```bash
# Remove cursor file
rm /usr/local/datakit/cache/journald.pos

# Restart DataKit
sudo systemctl restart datakit
```

### High memory usage {#troubleshoot-memory}

Default batch size is 1000 entries. If memory usage is a concern, reduce the batch size:

```toml
[[inputs.journald]]
  max_entries_per_batch = 100
```

## Related Documentation {#related}

- [DataKit Logging Overview](../datakit/datakit-logging.md)
- [Logging Pipeline Configuration](../pipeline/pipeline.md)
- [Systemd Journal Documentation](https://www.freedesktop.org/software/systemd/man/systemd-journald.service.html){:target="_blank"}
