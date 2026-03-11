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

- Kubernetes node journal collection

```toml
[[inputs.journald]]
  paths = ["/rootfs/var/log/journal", "/rootfs/run/log/journal"]
  tail_only = true
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

1. Test with `journalctl`:

```bash
journalctl -n 10
```

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
