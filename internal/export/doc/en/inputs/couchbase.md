---
title     : 'Couchbase'
summary   : 'Collect Couchbase server metrics'
__int_icon      : 'icon/couchbase'
dashboard :
  - desc  : 'Couchbase dashboard'
    path  : 'dashboard/en/couchbase'
monitor   :
  - desc  : 'null'
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

## Config {#config}

### Preconditions {#requirements}

- Install Couchbase server
  
[official document - CentOS/RHEL install](https://docs.couchbase.com/server/current/install/install-intro.html){:target="_blank"}

[official document - Debian/Ubuntu install](https://docs.couchbase.com/server/current/install/ubuntu-debian-install.html){:target="_blank"}

[official document - Windows install](https://docs.couchbase.com/server/current/install/install-package-windows.html){:target="_blank"}

- Verify correct installation

  Visit URL in browser `<ip>:8091` can open Couchbase manage UI.


???+ tip

  To collect data, several ports `8091` `9102` `18091` `19102` need to be used. 
  When collecting data remotely, these ports need to be opened.

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    Once configured, [restart DataKit](datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    
    The configuration parameters can be adjusted by the following environment variables:

    | Environment Variable Name       | Parameter Item | Parameter example                                                |
    | :-----------------------------        | ---               | ---                                                     |
    | `ENV_INPUT_COUCHBASE_INTERVAL`        | `interval`        | `"30s"` (`"10s"` ~ `"60s"`)                             |
    | `ENV_INPUT_COUCHBASE_TIMEOUT`         | `timeout`         | `"5s"`  (`"5s"` ~ `"30s"`)                              |
    | `ENV_INPUT_COUCHBASE_SCHEME`          | `scheme`          | `"http"` or `"https"`                                   |
    | `ENV_INPUT_COUCHBASE_HOST`            | `host`            | `"127.0.0.1"`                                           |
    | `ENV_INPUT_COUCHBASE_PORT`            | `port`            | `8091` or `18091`                                       |
    | `ENV_INPUT_COUCHBASE_ADDITIONAL_PORT` | `additional_port` | `9102` or `19102`                                       |
    | `ENV_INPUT_COUCHBASE_USER`            | `user`            | `"Administrator"`                                       |
    | `ENV_INPUT_COUCHBASE_PASSWORD`        | `password`        | `"123456"`                                              |
    | `ENV_INPUT_COUCHBASE_TLS_OPEN`        | `tls_open`        | `true` or `false`                                       |
    | `ENV_INPUT_COUCHBASE_TLS_CA`          | `tls_ca`          | `""`                                                    |
    | `ENV_INPUT_COUCHBASE_TLS_CERT`        | `tls_cert`        | `"/var/cb/clientcertfiles/travel-sample.pem"`           |
    | `ENV_INPUT_COUCHBASE_TLS_KEY`         | `tls_key`         | `"/var/cb/clientcertfiles/travel-sample.key"`           |
    | `ENV_INPUT_COUCHBASE_TAGS`            | `tags`            | `tag1=value1,tag2=value2`                               |
    | `ENV_INPUT_COUCHBASE_ELECTION`        | `election`        | `true` or `false`                                       |

    The collector can also be turned on by [ConfigMap injection collector configuration](datakit-daemonset-deploy.md#configmap-setting).

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
