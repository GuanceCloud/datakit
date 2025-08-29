---
title     : 'Prometheus Exporter'
summary   : 'Collect metrics exposed by Prometheus Exporter'
tags:
  - 'PROMETHEUS'
  - 'THIRD PARTY'
__int_icon      : 'icon/prometheus'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

{{.AvailableArchs}}

---

The PromV2 collector is an upgraded version of the Prom collector, which simplifies the configuration and enhances the collection performance.

## Configuration {#config}

<!-- markdownlint-disable MD046 -->
???+ attention

    PromV2 lacks many data modification configuration options and can only adjust the collected data through the Pipeline.
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Navigate to the *conf.d/{{.Catalog}}* directory in your DataKit installation path, copy *{{.InputName}}.conf.sample* and rename it to *{{.InputName}}.conf*. Example:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Currently enabled by injecting collector configuration via [ConfigMap](../datakit/datakit-daemonset-deploy.md#configmap-setting).

<!-- markdownlint-enable -->

