---
title     : 'External'
summary   : 'Start external program for collection'
__int_icon      : 'icon/external'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# External
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

External Collector launch outside program to collecting data.

## Configuration {#config}

### Preconditions {#requirements}

- The command and its running environment must have complete dependencies. For example, if Python is used to start an external Python script, the `import` package and other dependencies required for the script to run must be prepared.

### Input configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap injection collector configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->