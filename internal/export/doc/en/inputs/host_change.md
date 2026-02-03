---
title     : 'HostChange'
summary   : 'Monitor host configuration changes and report event data'
tags:
  - 'HOST'
__int_icon: ''
dashboard :
  - desc  : 'None'
    path  : '-'
monitor   :
  - desc  : 'None'
    path  : '-'
---

{{.AvailableArchs}}

---

HostChange collector monitors various configuration changes on Linux hosts, builds change event data, and reports it to the <<<custom_key.brand_name>>> platform.

> **Note**: This collector only supports Linux operating systems and does not support Windows systems.

## Feature Description {#features}

The HostChange collector supports the following change detection features:

| Feature Module | Description |
|----------------|-------------|
| User and Group Changes | Monitor `/etc/passwd`, `/etc/shadow`, `/etc/group`, `/etc/gshadow` files to detect user and group creation, deletion, attribute modifications, and membership changes |
| Crontab Changes | Monitor `/etc/crontab`, `/etc/cron.d/*`, `/var/spool/cron/crontabs/*` files to detect scheduled task changes |
| File Content Changes | Monitor content changes of specified files with diff comparison support |
| Service Changes | Monitor `systemd` or `sysvinit` service creation, deletion, attribute modifications, and status changes |
| Network Configuration Changes | Monitor network interface, DNS configuration, routing configuration, and firewall rule changes |

## Configuration {#config}

<!-- markdownlint-disable MD046 -->

=== "Host Installation"

    Navigate to the *conf.d/samples* directory under the DataKit installation directory, copy *{{.InputName}}.conf.sample* and rename it to *{{.InputName}}.conf*. An example is as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service) to apply the changes.

=== "Kubernetes"

    Currently, you can enable the collector by [injecting collector configuration through ConfigMap](../datakit/datakit-daemonset-deploy.md#configmap-setting).

<!-- markdownlint-enable -->

## Change Events {#change-event}

All data collection will append a global tag named `host` (with the value being the hostname where DataKit is located). You can also specify other tags in the configuration via `[inputs.{{.InputName}}.tags]`:

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

### Event Field Description {#event-fields}

<!-- markdownlint-disable MD024 -->

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "keyevent"}}

### `{{$m.Name}}`

{{$m.Desc}}

{{$m.MarkdownTable}}
{{end}}

{{ end }}

### Change Event Types {#event-types}

- User and Group Change Events

| Change ID | Description |
|-----------|-------------|
| `host_change_01_01` | Create user |
| `host_change_01_02` | Delete user |
| `host_change_01_03` | Modify user attributes |
| `host_change_01_04` | Create group |
| `host_change_01_05` | Delete group |
| `host_change_01_06` | Modify group attributes |
| `host_change_01_07` | Add user to group |
| `host_change_01_08` | Remove user from group |

- Crontab Change Events

| Change ID | Description |
|-----------|-------------|
| `host_change_02_01` | Crontab task change |

- File Change Events

| Change ID | Description |
|-----------|-------------|
| `host_change_03_01` | File content change |

- Service Change Events

| Change ID | Description |
|-----------|-------------|
| `host_change_04_01` | Create service |
| `host_change_04_02` | Delete service |
| `host_change_04_03` | Modify service |
| `host_change_04_04` | Service status change |

- Network Configuration Change Events

| Change ID | Description |
|-----------|-------------|
| `host_change_05_01` | Network interface change |
| `host_change_05_02` | DNS configuration change |
| `host_change_05_03` | Routing configuration change |
| `host_change_05_04` | Firewall rule change |
