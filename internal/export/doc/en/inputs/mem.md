
# Memory
---

{{.AvailableArchs}}

---

Mem collector is used to collect system memory information, some general metrics such as total memory, used memory and so on. 

## Preconditions {#requirements}

None

## Configuration {#config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Support modifying configuration parameters as environment variables:
    
    | Environment Variable Name               | Corresponding Configuration Parameter Item | Parameter Example                                                     |
    | :---                     | ---              | ---                                                          |
    | `ENV_INPUT_MEM_TAGS`     | `tags`           | `tag1=value1,tag2=value2`; If there is a tag with the same name in the configuration file, it will be overwritten. |
    | `ENV_INPUT_MEM_INTERVAL` | `interval`       | `10s`                                                        |

## Measurements {#measurements}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.mem.tags]`:

``` toml
 [inputs.mem.tags]
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
