---
title: 'DataKit metrics'
summary: 'Collect DataKit metrics'
tags:
  - 'HOST'
__int_icon: 'icon/dk'
dashboard:
  - desc: 'DataKit dashboard'
    path: 'dashboard/en/dk'
  - desc: 'DataKit dial test built-in dashboard'
    path: 'dashboard/en/dialtesting'

monitor:
  - desc: 'N/A'
    path: '-'
---

{{.AvailableArchs}} · [:octicons-tag-24: Version-1.10.0](../datakit/changelog.md#cl-1.10.0)

---

This input collects DataKit runtime metrics, including runtime environment, CPU, memory usage, and core module metrics. The collected data can be used by the DataKit dashboard, Bug Report troubleshooting, and runtime metric archiving.

## Configuration {#config}

After DataKit starts, it exposes [Prometheus metrics](../datakit/datakit-metrics.md). The `dk` input starts by default and replaces the earlier `self` input.

Basic features:

- Collects DataKit CPU, memory, Goroutine, HTTP API, data upload, Pipeline, filter, disk cache, and other runtime metrics.
- Uses `interval` to adjust the collection interval and `metric_types` to limit collected metric types.
- Supports custom tags through `[inputs.dk.tags]`.

Starting from [:octicons-tag-24: Version-2.2.0](../datakit/changelog-2026.md#cl-2.2.0), `dk` collects all DataKit self-metrics except internally blocked metrics by default and no longer provides `metric_name_filter`. To keep only part of the metrics, use Pipeline filtering by metric name.

To disable DataKit self-metric collection, set the following in `dk.conf`:

```toml
[[inputs.dk]]
  enabled = false
```

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    To adjust the collection interval, metric types, self profiling, or other settings, go to the `conf.d/samples` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

    Configuration can also be adjusted through environment variables:
    
{{ CodeBlock .InputENVSample 4 }}

<!-- markdownlint-enable -->

## Metric {#metric}

DataKit exported Prometheus metrics, see [here](../datakit/datakit-metrics.md) for full metric list.

## Self Profiling {#self-profiling}

Starting from [:octicons-tag-24: Version-2.2.0](../datakit/changelog-2026.md#cl-2.2.0), `dk` supports threshold-triggered DataKit self profiling. This feature is disabled by default. It can collect CPU, heap, goroutine, and other profiles when DataKit CPU or memory reaches the configured thresholds.

For host installation, enable it in `dk.conf`:

```toml
[inputs.dk.self_profiling]
  enabled = true
```

For Kubernetes installation, enable it through an environment variable:

```shell
ENV_INPUT_DK_ENABLE_SELF_PROFILING=true
```

See `[inputs.dk.self_profiling]` in the sample configuration above for threshold, profile type, cache, and upload settings.
