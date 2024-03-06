---
title: 'DataKit own metrics collection'
summary: 'Collect DataKit's own operational metrics'
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

<!-- markdownlint-disable MD025 -->
# DataKit Metrics
<!-- markdownlint-enable -->

---

{{.AvailableArchs}} · [:octicons-tag-24: Version-1.10.0](../datakit/changelog.md#cl-1.10.0)

---

This Input used to collect Datakit exported metrics, such as runtime/CPU/memory and various other metrics of each modules.

## Configuration {#config}

After Datakit startup, it will expose a lot of [Prometheus metrics](datakit-metrics.md), and the input `dk` can scrap
these metrics.

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

    Can also be turned on by environment variables, (needs to be added as the default collector in ENV_DEFAULT_ENABLED_INPUTS):
    
{{ CodeBlock .InputENVSample 4 }}

<!-- markdownlint-enable -->

## Metric {#metric}

Datakit exported Prometheus metrics, see [here](../datakit/datakit-metrics.md) for full metric list.
