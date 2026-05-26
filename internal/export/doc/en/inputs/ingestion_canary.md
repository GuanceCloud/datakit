---
title     : 'Ingestion Canary'
summary   : 'Detect data ingestion availability and latency by automatically generating and querying probe data'
tags:
  - 'AVAILABILITY'
  - 'LATENCY'
  - 'MONITORING'
__int_icon      : 'icon/ingestion_canary'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

{{.AvailableArchs}}

---

This collector is used to detect data ingestion availability and latency. It automatically generates probe data (metrics, logs, traces) and then verifies data collection success through DQL queries, measuring the latency from data sending to queryable.

## Requirements {#requirements}

- DataWay must be configured for data reporting and DQL queries
- If `result_workspace` is configured, ensure the workspace URL is accessible

## Configuration {#config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting).

<!-- markdownlint-enable -->

## Measurements {#measurements}

This collector generates two types of data:

1. **Probe Data**: Probe data points (metrics, logs, traces) used to test data ingestion availability
2. **Result Metric**: Test result metrics containing latency and test status

Probe data points do not carry global tags, only include the data point's own fields and tags, plus tags specified in configuration.

### Probe Data {#probe-data}

{{ range $i, $m := .Measurements }}
{{ if and (ne $m.Name "ingestion_canary_result") (eq (printf "%v" $m.Cat) "metric") }}
#### `{{ $m.Name }}` (Metric)

{{ if $m.Desc }}{{ $m.Desc }}{{ end }}

{{ $m.MarkdownTable }}

{{ end }}
{{ end }}

{{ range $i, $m := .Measurements }}
{{ if and (ne $m.Name "ingestion_canary_result") (eq (printf "%v" $m.Cat) "logging") }}
#### `{{ $m.Name }}` (Logging)

{{ if $m.Desc }}{{ $m.Desc }}{{ end }}

{{ $m.MarkdownTable }}

{{ end }}
{{ end }}

{{ range $i, $m := .Measurements }}
{{ if and (ne $m.Name "ingestion_canary_result") (eq (printf "%v" $m.Cat) "tracing") }}
#### `{{ $m.Name }}` (Tracing)

{{ if $m.Desc }}{{ $m.Desc }}{{ end }}

{{ $m.MarkdownTable }}

{{ end }}
{{ end }}

### Result Metric {#result-metric}

{{ range $i, $m := .Measurements }}
{{ if eq $m.Name "ingestion_canary_result" }}
#### `{{ $m.Name }}`

{{ if $m.Desc }}{{ $m.Desc }}{{ end }}

{{ $m.MarkdownTable }}

{{ end }}
{{ end }}

## CLI Tool {#cli-tool}

In addition to the collector mode, a CLI tool is provided for one-time testing:

```bash
# Use default configuration
datakit tool --ingestion-canary

# Specify storage index for logging data
datakit tool --ingestion-canary --ingestion-canary-index my_index
```

**Options:**

- `--ingestion-canary`: Enable ingestion canary test tool
- `--ingestion-canary-index`: Specify storage index for logging data, default is "default" (only applies to logging data)

**Description:**

The tool generates one round of probe data (metrics, logs, traces), sends it to DataWay, then continuously queries until data is found or user interrupts, and outputs latency for each data type. The tool runs continuously with 10 second interval between rounds.
