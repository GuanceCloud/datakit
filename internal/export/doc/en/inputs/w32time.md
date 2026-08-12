---
title     : 'W32Time'
summary   : 'Collect Windows Time Service metrics'
tags:
  - 'WINDOWS'
  - 'TIME'
__int_icon: 'icon/windows'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---


{{.AvailableArchs}}

---

The W32Time collector monitors the local Windows Time Service through Windows Performance Data Helper counters. It does not execute or parse `w32tm` output.

## Configuration {#config}

### Preconditions {#requirements}

- Windows Server 2016 or later, Windows 10 Anniversary Update or later, or Windows 11
- 64-bit DataKit for Windows (`windows/amd64`)
- The Windows Time (`W32Time`) service is registered; no separate installation of `w32tm` or another program is required

When the `W32Time` service is not running, the collector still reports `service_running=0`. Other time synchronization fields are reported only when the service is running and the corresponding performance counters are available.

### Collector Configuration {#input-config}

Go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. The configuration is as follows:

```toml
{{.InputSample}}
```

After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

## Metric {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.MarkdownTable}}

{{ end }}
