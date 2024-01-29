---
title     : 'Profiling'
summary   : 'Collect application runtime performance data'
__int_icon: 'icon/profiling'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Profiling
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Profile supports collecting dynamic performance data of applications running in different language environments such as Java/Python, and helps users to view performance problems of CPU, memory and IO.

## Configuration {#config}

At present, DataKit collects profiling data in two ways:

- Push mode: the DataKit Profile service needs to be opened, and the client actively pushes data to the DataKit

- Pull method: currently only [Go](profile-go.md) support, need to manually configure relevant information

### DataKit Configuration {#datakit-config}
<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/profile` directory under the DataKit installation directory, copy `profile.conf.sample` and name it `profile.conf`. The configuration file is described as follows:
    
    ```shell
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).
<!-- markdownlint-enable -->
## Profiling {#profiling}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
