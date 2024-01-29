---
title     : 'Windows Event'
summary   : 'Collect event logs in Windows'
__int_icon      : 'icon/winevent'
dashboard :
  - desc  : 'N/A'
    path  : '-'
monitor   :
  - desc  : 'N/A'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Windows Event
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Windows Event Log Collection is used to collect applications, security, systems and so on.

## Configuration {#config}

### Preconditions {#requrements}

- Windows version >= Windows Server 2008 R2

### Collector Configuration {#input-config}

Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `windows_event.conf`. Examples are as follows:

```toml
{{.InputSample}}
```

After configuration, restart DataKit.

## Logging {#logging}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration through `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.windows_event.tags]
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
