---
title     : 'NetStat'
summary   : 'Collect NIC traffic metrics data'
__int_icon      : 'icon/netstat'
dashboard :
  - desc  : 'NetStat'
    path  : 'dashboard/en/netstat'
monitor   :
  - desc  : 'NetStat'
    path  : 'monitor/en/netstat'
---

<!-- markdownlint-disable MD025 -->
# NetStat
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

Netstat metrics collection, including TCP/UDP connections, waiting for connections, waiting for requests to be processed, and so on.

## Config {#config}

### Collector Configuration {#input-config}

<!-- markdownlint-disable MD046 -->
=== "Host deployment"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Configuration Tips：

    ``` toml
    ## (1) Configure the ports of interest.
    [[inputs.netstat.addr_ports]]
      ports = ["80","443"]
    ```

    ``` toml
    # (2) Configure two groups of ports with different tags for easy statistics.
    [[inputs.netstat.addr_ports]]
      ports = ["80","443"]
      [inputs.netstat.addr_ports.tags]
        service = "http"

    [[inputs.netstat.addr_ports]]
        ports = ["9529"]
        [inputs.netstat.addr_ports.tags]
            service = "datakit"
    ```

    ``` toml
    # (3) The server has multiple NICs and only cares about certain ones.
    [[inputs.netstat.addr_ports]]
      ports = ["1.1.1.1:80","2.2.2.2:80"]
    ```

    ``` toml
    # (4) The server has multiple NICs, and the requirement to show this configuration on a per NIC basis will mask the ports configuration value.
    [[inputs.netstat.addr_ports]]
      ports = ["1.1.1.1:80","2.2.2.2:80"] // Invalid, masked by ports_match.
      ports_match = ["*:80","*:443"] // Valid.
    ```

    After configuration, restart DataKit.

=== "Kubernetes"

    Kubernetes supports modifying configuration parameters in the form of environment variables:


    | Environment Variable Name                          | Corresponding Configuration Parameter Item | Parameter Example |
    |:-----------------------------     | ---            | ---   |
    | `ENV_INPUT_NETSTAT_TAGS`          | `tags`         | `tag1=value1,tag2=value2`; If there is a tag with the same name in the configuration file, it will be overwritten. |
    | `ENV_INPUT_NETSTAT_INTERVAL`      | `interval`     | `10s` |
    | `ENV_INPUT_NETSTAT_ADDR_PORTS`    | `ports`        | `["1.1.1.1:80","443"]` |
<!-- markdownlint-enable -->
---

## Metric {#metric}

For all the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.{{.InputName}}.tags]`:

``` toml
 [inputs.{{.InputName}}.tags]
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
