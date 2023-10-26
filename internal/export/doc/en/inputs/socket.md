
# TCP/UDP
---

{{.AvailableArchs}}

---

The socket collector is used to collect UDP/TCP port information.

## Preconditions {#requrements}

UDP metrics require the operating system to have `nc` programs.

<!-- markdownlint-disable MD046 -->
???+ attention

    The socket collector are suitable for collecting local network TCP/UDP service. For public network, [Dialtesting](dialtest.md) is recommended. If the URLs point to localhost, please turn off the election flag(`election: false`).
<!-- markdownlint-enable -->

## Configuration {#config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, restart DataKit.

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->

## Measurements {#requrements}

For all of the following measurements, the `proto/dest_host/dest_port` global tag is appended by default, or other tags can be specified in the configuration by `[inputs.socket.tags]`:

``` toml
 [inputs.socket.tags]
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
