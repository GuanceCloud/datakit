
# etcd
---

{{.AvailableArchs}}

---

The tcd collector can take many metrics from the etcd instance, such as the status of the etcd server and network, and collect the metrics to DataFlux to help you monitor and analyze various abnormal situations of etcd.

## Preconditions {#requirements}

- etcd version  >=3

- Open etcd, the default metrics interface is http://localhost:2379/metrics, or you can modify it in your configuration file.

## Configuration {#config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:


​    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap injection collector configuration](datakit-daemonset-deploy.md#configmap-setting).

## Measurements {#measurements}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
