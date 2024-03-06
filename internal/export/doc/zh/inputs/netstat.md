---
title     : 'NetStat'
summary   : '采集网卡流量指标数据'
__int_icon      : 'icon/netstat'
dashboard :
  - desc  : 'NetStat'
    path  : 'dashboard/zh/netstat'
monitor   :
  - desc  : 'NetStat'
    path  : 'monitor/zh/netstat'
---

<!-- markdownlint-disable MD025 -->
# NetStat
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

NetStat 指标展示，包括 Tcp 连接数、等待连接、等待处理请求、Udp Socket 连接等。

## 配置 {#config}

<!-- markdownlint-disable MD046 -->
=== "主机部署"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置技巧：

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

    可通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting) 或 [配置 ENV_DATAKIT_INPUTS](../datakit/datakit-daemonset-deploy.md#env-setting) 开启采集器。

    也支持以环境变量的方式修改配置参数（需要在 ENV_DEFAULT_ENABLED_INPUTS 中加为默认采集器）：

{{ CodeBlock .InputENVSampleZh 4 }}

<!-- markdownlint-enable -->
---

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

不分端口号统计的指标集：`netstat`，分端口号统计的指标集：`netstat_port`。

{{ range $i, $m := .Measurements }}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
