---
title     : 'Log Forward'
summary   : 'Collect log data in Pod through sidecar method'
__int_icon      : 'icon/logfwd'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# {{.InputName}}
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

## Introduction {#intro}

Logfwdserver will turn on the websocket function, which is used together with logfwd, and is responsible for receiving and processing the data sent by logfwd.

See [here](logfwd.md) for the use of logfwd.

## Configuration {#config}

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->
