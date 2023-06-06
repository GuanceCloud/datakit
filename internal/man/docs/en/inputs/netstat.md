
# NetStat
---

{{.AvailableArchs}}

---

Netstat metrics collection, including TCP/UDP connections, waiting for connections, waiting for requests to be processed, and so on.

## Preconditions {#precondition}

None

## Configuration {#input-config}

=== "Host deployment"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    After configuration, restart DataKit.

=== "Kubernetes"

    Kubernetes supports modifying configuration parameters in the form of environment variables:


    | Environment Variable Name                          | Corresponding Configuration Parameter Item | Parameter Example |
    |:-----------------------------     | ---            | ---   |
    | `ENV_INPUT_NETSTAT_TAGS`          | `tags`         | `tag1=value1,tag2=value2`; If there is a tag with the same name in the configuration file, it will be overwritten. |
    | `ENV_INPUT_NETSTAT_INTERVAL`      | `interval`     | `10s` |
    | `ENV_INPUT_NETSTAT_ADDR_PORTS`    | `ports`        | `["1.1.1.1:80","443"]` |

---

## Measurements {#measurements}

For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.netstat.tags]`:

``` toml
 [inputs.netstat.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

Measurements for statistics regardless of port number: `netstat` ; Measurements for statistics by port number: `netstat_port`.

{{ range $i, $m := .Measurements }}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
