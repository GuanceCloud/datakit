---
title     : 'Solr'
summary   : 'Collect Solr metrics'
__int_icon      : 'icon/solr'
dashboard :
  - desc  : 'Solr'
    path  : 'dashboard/en/solr'
monitor   :
  - desc  : 'Solr'
    path  : 'monitor/en/solr'
---

<!-- markdownlint-disable MD025 -->
# Solr
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Solr collector, which collects statistics of Solr Cache, Request Times, and so on.

## Configuration {#config}

### Preconditions {#requrements}

DataKit uses the Solr Metrics API to collect metrics data and supports Solr 7.0 and above. Available for Solr 6.6, but the indicator data is incomplete.

Already tested version:

- [x] 8.11.2
- [x] 7.0.0

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/db` directory under the DataKit installation directory, copy `solr.conf.sample` and name it  `solr.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, restart DataKit.

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

---

To collect Solr's log, open `files` in Solr.conf and write to the absolute path of the Solr log file. For example:

```toml
[inputs.solr.log]
    # fill in the absolute path
    files = ["/path/to/demo.log"]
```

## Metric {#metric}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}

## Log Collection {#logging}

Example of cutting logs:

```log
2013-10-01 12:33:08.319 INFO (org.apache.solr.core.SolrCore) [collection1] webapp.reporter
```

Cut fields:

| Field Name | Field Value                   |
| ---------- | ----------------------------- |
| Reporter   | webapp.reporter               |
| status     | INFO                          |
| thread     | org.apache.solr.core.SolrCore |
| time       | 1380630788319000000           |
