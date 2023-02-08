
# TDengine
---

{{.AvailableArchs}}

---

TDEngine is a high-performance, distributed, SQL-enabled time series Database (Database). Familiarize yourself with the [basic concepts of TDEngine](https://docs.taosdata.com/concept/){:target="_blank"} before opening the collector.

TDengine collector needs to connect `taos_adapter` can work normally, taosAdapter from TDengine v2.4. 0.0 version comes to becoming a part of TDengine server software, this paper is mainly a detailed introduction of measurement.

## Configuration  {#config}

=== "Host Installation"


    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart DataKit](datakit-service-how-to.md#manage-service).


=== "Kubernetes"

    At present, the collector can be turned on by [injecting the collector configuration in ConfigMap mode](datakit-daemonset-deploy.md#configmap-setting).


### TdEngine Dashboard {#td-dashboard}

    At present, Guance Cloud has provided a built-in TDEngine dashboard, and you can select the TDEngine dashboard in ***Guance Cloud*** -- ***Scene***--***New Dashboard***.


## Measurement{#td-metrics}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}

> - Some tables in the database do not have the `ts` field, and Datakit uses the current collection time.
