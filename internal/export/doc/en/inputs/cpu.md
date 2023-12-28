---
title     : 'CPU'
summary   : 'Collect metric of cpu'
__int_icon      : 'icon/cpu'
dashboard :
  - desc  : 'CPU'
    path  : 'dashboard/en/cpu'
monitor   :
  - desc  : 'host detection library'
    path  : 'monitor/en/host'
---

<!-- markdownlint-disable MD025 -->
# CPU
<!-- markdownlint-enable -->
---

{{.AvailableArchs}}

---

The CPU collector is used to collect the CPU utilization rate of the system.

## Configuration {#config}

After successfully installing and starting DataKit, the CPU collector will be enabled by default without the need for manual activation.

<!-- markdownlint-disable MD046 -->

=== "host installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, [restart Datakit](../datakit/datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    Kubernetes supports modifying configuration parameters in the form of environment variables:
    
    | Environment Variable Name                   | Corresponding Configuration Parameter Item              | Parameter Example                                                                              |
    | :---                                        | ---                           | ---                                                                                   |
    | `ENV_INPUT_CPU_PERCPU`                      | `percpu`                      | `true/false`                                                                          |
    | `ENV_INPUT_CPU_ENABLE_TEMPERATURE`          | `enable_temperature`          | `true/false`                                                                          |
    | `ENV_INPUT_CPU_TAGS`                        | `tags`                        | `tag1=value1,tag2=value2` If there is a tag with the same name in the configuration file, it will be overwritten.                          |
    | `ENV_INPUT_CPU_INTERVAL`                    | `interval`                    | `10s`                                                                                 |
    | `ENV_INPUT_CPU_DISABLE_TEMPERATURE_COLLECT` | `disable_temperature_collect` | `false/true`. Any string is considered ` true `, and if it is not defined, it is ` false `.                     |
    | `ENV_INPUT_CPU_ENABLE_LOAD5S`               | `enable_load5s`               | `false/true`. Any string is considered ` true `, and if it is not defined, it is ` false `. |

<!-- markdownlint-enable -->

---

## Metric {#metric}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration through `[inputs.cpu.tags]`:

``` toml
 [inputs.cpu.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- Metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
