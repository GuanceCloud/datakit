
# Zipkin
---

{{.AvailableArchs}}

---

The Zipkin Agent embedded in Datakit is used to receive, calculate and analyze the data of Zipkin Tracing protocol.

## Zipkin Docs {#docs}

- [Quickstart](https://zipkin.io/pages/quickstart.html){:target="_blank"}
- [Docs](https://zipkin.io/pages/instrumenting.html){:target="_blank"}
- [Souce Code](https://github.com/openzipkin/zipkin){:target="_blank"}

## Configure Zipkin Agent {#config-agent}

=== "Host Installation"

    Go to the `conf.d/zipkin` directory under the DataKit installation directory, copy `zipkin.conf.sample` and name it `zipkin.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, restart DataKit.

=== "Kubernetes"

    At present, the collector can be turned on by [injecting the collector configuration in ConfigMap mode](datakit-daemonset-deploy.md#configmap-setting).

## Measurements {#measurements}

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "tracing"}}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}