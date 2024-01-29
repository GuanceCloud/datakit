---
title     : 'Hardware Sensors data collection'
summary   : 'Collect hardware temperature indicators through Sensors command'
__int_icon      : 'icon/sensors'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Hardware temperature Sensors
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

## Configuration {#config}

Computer chip temperature data acquisition using the `lm-sensors` command (currently only support `Linux` operating system).

### Preconditions {#requrements}

- Run the install command `apt install lm-sensors -y`
- Run the scan command `sudo sensors-detect` enter `Yes` for each question
- After running the scan, you will see 'service kmod start' to load the scanned sensors, which may vary depending on your operating system.

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    After configuration, restart DataKit.

=== "Kubernetes"

<!-- markdownlint-enable -->

The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).

### Metric {#metric}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

```toml
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
