---
title     : 'Couchbase'
summary   : 'Collect Couchbase server metrics'
__int_icon      : 'icon/couchbase'
dashboard :
  - desc  : 'Couchbase dashboard'
    path  : 'dashboard/en/couchbase'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Couchbase
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

The Couchbase collector can take metrics from the Couchbase server.


Already tested version:

- [x] Couchbase enterprise-7.2.0
- [x] Couchbase community-7.2.0

## Configuration {#config}

### Preconditions {#requirements}

- Install Couchbase server
  
[official document - CentOS/RHEL install](https://docs.couchbase.com/server/current/install/install-intro.html){:target="_blank"}

[official document - Debian/Ubuntu install](https://docs.couchbase.com/server/current/install/ubuntu-debian-install.html){:target="_blank"}

[official document - Windows install](https://docs.couchbase.com/server/current/install/install-package-windows.html){:target="_blank"}

- Verify correct installation

  Visit URL in browser `<ip>:8091` can open Couchbase manage UI.

<!-- markdownlint-disable MD046 -->
???+ tip
    - To collect data, several ports `8091` `9102` `18091` `19102` need to be used. When collecting data remotely, these ports need to be opened.
<!-- markdownlint-enable -->

### Collector Configuration {#input-conifg}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Can be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting) or [Config ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) .

    Can also be turned on by environment variables, (needs to be added as the default collector in ENV_DEFAULT_ENABLED_INPUTS):
    
{{ CodeBlock .InputENVSample 4 }}

<!-- markdownlint-enable -->

### TLS config {#tls}

TLS need Couchbase enterprise

[official document - configure-server-certificates](https://docs.couchbase.com/server/current/manage/manage-security/configure-server-certificates.html){:target="_blank"}

[official document - configure-client-certificates](https://docs.couchbase.com/server/current/manage/manage-security/configure-client-certificates.html){:target="_blank"}

## Metric {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- Tag

{{$m.TagsMarkdownTable}}

- Metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
