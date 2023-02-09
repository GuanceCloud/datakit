
# Solr
---

{{.AvailableArchs}}

---

Solr collector, which collects statistics of solr cache, request times, and so on.

## Preconditions {#requrements}

DataKit uses the Solr Metrics API to collect metrics data and supports Solr 7.0 and above. Available for Solr 6.6, but the indicator data is incomplete.

## Configuration {#config}

=== "Host Installation"

    Go to the `conf.d/db` directory under the DataKit installation directory, copy `solr.conf.sample` and name it  `solr.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    


​    
    After configuration, restart DataKit.

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](datakit-daemonset-deploy.md#configmap-setting).

## Measurements {#measurements}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.solr.tags]`:

``` toml
 [inputs.solr.tags]
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

To collect Solr's log, open `files` in Solr.conf and write to the absolute path of the Solr log file. For example:

```toml
[inputs.solr.log]
    # fill in the absolute path
    files = ["/path/to/demo.log"]
```

Example of cutting logs:

```
2013-10-01 12:33:08.319 INFO (org.apache.solr.core.SolrCore) [collection1] webapp.reporter
```

Cut fields:

| Field Name   | Field Value                        |
| -------- | ----------------------------- |
| Reporter | webapp.reporter               |
| status   | INFO                          |
| thread   | org.apache.solr.core.SolrCore |
| time     | 1380630788319000000           |
