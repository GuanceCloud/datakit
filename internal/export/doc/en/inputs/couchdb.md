---
title     : 'CouchDB'
summary   : 'Collect CouchDB server metrics'
__int_icon      : 'icon/couchdb'
dashboard :
  - desc  : 'CouchDB'
    path  : 'dashboard/en/couchdb'
monitor   :
  - desc  : 'CouchDB'
    path  : 'monitor/en/couchdb'
---

<!-- markdownlint-disable MD025 -->
# CouchDB
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

CouchDB collector is used to collect metric data related to CouchDB, and currently it only supports data in Prometheus format.

Already tested version:

- [x] CouchDB 3.3.2
- [x] CouchDB 3.2
- [ ] CouchDB 3.1 this and versions earlier than this do not support

## Configuration {#config}

### Preconditions {#requirements}

- Install CouchDB server
  
See [official document](https://docs.couchdb.org/en/stable/install/index.html){:target="_blank"}

- Verify correct installation

  Visit URL in browser `<ip>:5984/_utils/` can open CouchDB manage UI.

- Open CouchDB Prometheus port
  
  Search CouchDB start config file, usually `/opt/couchdb/etc/local.ini`

  ```ini
  [prometheus]
  additional_port = false
  bind_address = 127.0.0.1
  port = 17986
  ```

  Change as

  ```ini
  [prometheus]
  additional_port = true
  bind_address = 0.0.0.0
  port = 17986
  ```

  See [official document](https://docs.couchdb.org/en/stable/config/misc.html#configuration-of-prometheus-endpoint){:target="_blank"}
  
- Restart CouchDB

<!-- markdownlint-disable MD046 -->
???+ tip

    - To collect data, several ports `5984` `17986` need to be used. When collecting data remotely, these ports need to be opened.
    - bind_address = 127.0.0.1 If it is a local collection, there is no need to modify it.
<!-- markdownlint-enable -->

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

## Metric {#metric}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- Tag

{{$m.TagsMarkdownTable}}

- Metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
