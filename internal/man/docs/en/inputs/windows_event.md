
# Windows Events
---

{{.AvailableArchs}}

---

Windows Event Log Collection is used to collect applications, security, systems and so on.

## Preconditions {#requrements}

- Windows version >= Windows Server 2008 R2

## Configuration {#config}

Go to the `conf.d/windows` directory under the DataKit installation directory, copy `windows_event.conf.sample` and name it `windows_event.conf`. Examples are as follows:

```toml
{{.InputSample}}
```

After configuration, restart DataKit.

## Measurement {#measurements}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration through `[inputs.windows_event.tags]`:

``` toml
 [inputs.windows_event.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}


### `{{$m.Name}}`

-  tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }} 

 

