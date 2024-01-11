---
title     : 'Disk S.M.A.R.T'
summary   : 'Collect disk metrics through smartctl'
__int_icon      : 'icon/smartctl'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Disk S.M.A.R.T
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Data collection of computer hard disk running state.

## Configuration {#config}

### Preconditions {#requrements}

Installing smartmontools

- Linux: `sudo apt install smartmontools -y`

	If the solid state drive is nvme compliant, it is recommended to install nvme-cli for more nvme information: `sudo apt install nvme-cli -y`

- MacOS: `brew install smartmontools -y`
- WinOS: download [Windows version](https://www.smartmontools.org/wiki/Download#InstalltheWindowspackage){:target="_blank"}

### Collector Configuration {#input-config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, restart DataKit.

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](../datakit/datakit-daemonset-deploy.md#configmap-setting).

## Metric {#metric}

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
