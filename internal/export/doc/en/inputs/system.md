
# System
---

{{.AvailableArchs}}

---

The system collector collects system load, uptime, the number of CPU cores, and the number of users logged in.

## Preconditions {#requrements}

None

## Configuration {#config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, restart DataKit.

=== "Kubernetes"

    Modifying configuration parameters as environment variables is supported:
    
    | Environment variable name              | Corresponding configuration parameter item | Parameter example                                                     |
    | :---                    | ---              | ---                                                          |
    | `ENV_INPUT_SYSTEM_TAGS` | `tags`           | `tag1=value1,tag2=value2`. If there is a tag with the same name in the configuration file, it will be overwritten. |
    | `ENV_INPUT_SYSTEM_INTERVAL` | `interval` | `10s` |

---

## Measurements {#measurements}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration through `[inputs.system.tags]`:

``` toml
 [inputs.system.tags]
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

## FAQ {#faq}

### Why no `cpu_total_usage`? {#no-cpu}

Some CPU acquisition features are not supported on some platforms, such as macOS.
