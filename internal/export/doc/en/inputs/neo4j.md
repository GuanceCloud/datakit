---
title     : 'Neo4j'
summary   : 'Collect Neo4j server metrics'
__int_icon      : 'icon/neo4j'
dashboard :
  - desc  : 'Neo4j'
    path  : 'dashboard/en/neo4j'
monitor   :
  - desc  : 'Neo4j'
    path  : 'monitor/en/neo4j'
---

<!-- markdownlint-disable MD025 -->
# Neo4j
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

Neo4j collector is used to collect metric data related to Neo4j, and currently it only supports data in Prometheus format. 

Already tested version:

- [x] Neo4j 5.11.0 enterprise
- [x] Neo4j 4.4.0 enterprise
- [x] Neo4j 3.4.0 enterprise
- [ ] Neo4j 3.3.0 enterprise this and versions earlier than this do not support
- [ ] Neo4j 5.11.0 community all community versions do not support

## Preconditions {#requirements}

- Install Neo4j server
  
See [official document](https://neo4j.com/docs/operations-manual/current/installation/){:target="_blank"}

- Verify correct installation

  Visit URL in browser `<ip>:7474` can open Neo4j manage UI.

- Open Neo4j Prometheus port
  
  Search Neo4j start config file, usually `/etc/neo4j/neo4j.conf`

  Add in the tail

  ```ini
  # Enable the Prometheus endpoint. Default is false.
  server.metrics.prometheus.enabled=true
  # The hostname and port to use as Prometheus endpoint.
  # A socket address is in the format <hostname>, <hostname>:<port>, or :<port>.
  # If missing, the port or hostname is acquired from server.default_listen_address.
  # The default is localhost:2004.
  server.metrics.prometheus.endpoint=0.0.0.0:2004
  ```

  See [official document](https://neo4j.com/docs/operations-manual/current/monitoring/metrics/expose/#_prometheus){:target="_blank"}
  
- Restart Neo4j

<!-- markdownlint-disable MD046 -->
???+ tip

    - To collect data, port `2004` need to be used. When collecting data remotely, need to be opened.
    - 0.0.0.0:2004 If it is a local collection, need be localhost:2004.
<!-- markdownlint-enable -->

## Config {#input-config}

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
