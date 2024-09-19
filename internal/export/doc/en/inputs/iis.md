---
title     : 'IIS'
summary   : 'Collect IIS metrics'
__int_icon      : 'icon/iis'
dashboard :
  - desc  : 'IIS'
    path  : 'dashboard/en/iis'
monitor   :
  - desc  : 'IIS'
    path  : 'monitor/en/iis'
---

<!-- markdownlint-disable MD025 -->
# IIS
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Microsoft IIS collector

## Configuration {#config}

### Preconditions {#requirements}

Operating system requirements::

- Windows Vista and above (excluding Windows Vista)
- Windows Server 2008 R2 and above

### Collector Configuration {#input-config}

Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

```toml
{{.InputSample}} 
```

After configuration, restart DataKit.

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

``` toml
  [inputs.{{.InputName}}.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
    # ...
```

## Metric {#metric}

{{ range $i, $m := .Measurements }}

{{if or (eq $m.Type "metric") (eq $m.Type "")}}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}{{end}}

{{ end }}

## Log {#logging}

If you need to collect IIS logs, open the log-related configuration in the configuration, such as:

```toml
[inputs.{{.InputName}}.log]
    # Fill in the absolute path
    files = ["C:/inetpub/logs/LogFiles/W3SVC1/*"] 
```
