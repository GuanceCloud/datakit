
# Swap
---

{{.AvailableArchs}}

---

The swap collector is used to collect the usage of the host swap memory.

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
    
    | Environment Variable Name                | Corresponding Configuration Parameter Item | Parameter Example                                                     |
    | :---                      | ---              | ---                                                          |
    | `ENV_INPUT_SWAP_TAGS`     | `tags`           | `tag1=value1,tag2=value2`. If there is a tag with the same name in the configuration file, it will be overwritten |
    | `ENV_INPUT_SWAP_INTERVAL` | `interval`       | `10s`                                                        |

## Measurements {#measurements}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.swap.tags]`:

``` toml
 [inputs.swap.tags]
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
