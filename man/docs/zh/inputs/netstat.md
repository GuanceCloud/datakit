
# NetStat

---

{{.AvailableArchs}}

---

Netstat 指标采集，包括 TCP/UDP 连接数、等待连接、等待处理请求等。

## 前置条件 {#precondition}

暂无

## 配置 {#input-config}

<!-- markdownlint-disable MD046 -->
=== "主机部署"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置技巧: 

    ``` toml
    ## (1) 配置关注的端口号
    [[inputs.netstat.addr_ports]]
      ports = ["80","443"]
    ```

    ``` toml
    # (2) 配置两组端口，加上不同的 tag，方便统计
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
    # (3) 服务器有多个网卡，只关心某几个网卡的情况
    [[inputs.netstat.addr_ports]]
      ports = ["1.1.1.1:80","2.2.2.2:80"]
    ```

    ``` toml
    # (4) 服务器有多个网卡，要求按每个网卡分别展示这个配置，会屏蔽掉 ports 的配置值
    [[inputs.netstat.addr_ports]]
      ports = ["1.1.1.1:80","2.2.2.2:80"] // 无效，被 ports_match 屏蔽
      ports_match = ["*:80","*:443"] // 有效
    ```

    配置好后，重启 DataKit 即可。

=== "Kubernetes"

    Kubernetes 中支持以环境变量的方式修改配置参数：


    | 环境变量名                          | 对应的配置参数项 | 参数示例 |
    |:-----------------------------     | ---            | ---   |
    | `ENV_INPUT_NETSTAT_TAGS`          | `tags`         | `tag1=value1,tag2=value2` 如果配置文件中有同名 tag，会覆盖它 |
    | `ENV_INPUT_NETSTAT_INTERVAL`      | `interval`     | `10s` |
    | `ENV_INPUT_NETSTAT_ADDR_PORTS`    | `ports`        | `["1.1.1.1:80","443"]` |
<!-- markdownlint-enable -->
---

## 指标集 {#measurements}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

不分端口号统计的指标集: `netstat` ，分端口号统计的指标集: `netstat_port` 。

{{ range $i, $m := .Measurements }}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
