---
title     : '健康检查'
summary   : '定期检查主机进程和网络健康状况'
__int_icon      : 'icon/healthcheck'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# 健康检查
<!-- markdownlint-enable -->

[:octicons-tag-24: Version-1.24.0](../datakit/changelog.md#cl-1.24.0)

---

{{.AvailableArchs}}

---

健康检查采集器可以定期去监控主机的进程和网络（如 TCP 和 HTTP）的健康状况，如果不符合健康要求，DataKit 会收集相应的信息，并上报指标数据。

## 配置 {#config}

### 采集器配置 {#input-config}

<!-- markdownlint-disable MD046 -->

=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    支持以环境变量的方式修改配置参数（只在 Datakit 以 K8s DaemonSet 方式运行时生效，主机部署的 Datakit 不支持此功能）：

    | 环境变量名                             | 对应的配置参数项    | 参数示例                                                     |
    | :---                                 | ---              | ---                                                          |
    | `ENV_INPUT_HEALTHCHECK_INTERVAL`     | `interval`       | `5m`                                               |
    | `ENV_INPUT_HEALTHCHECK_PROCESS`      | `process`        | `[{"names":["nginx","mysql"],"min_run_time":"10m"}]`|
    | `ENV_INPUT_HEALTHCHECK_TCP`          | `tcp`            | `[{"host_ports":["10.100.1.2:3369","192.168.1.2:6379"],"connection_timeout":"3s"}]`|
    | `ENV_INPUT_HEALTHCHECK_HTTP`         | `http`           | `[{"http_urls":["http://local-ip:port/path/to/api?arg1=x&arg2=y"],"method":"GET","expect_status":200,"timeout":"30s","ignore_insecure_tls":false,"headers":{"Header1":"header-value-1","Hedaer2":"header-value-2"}}]`                                               |
    | `ENV_INPUT_HEALTHCHECK_TAGS`         | `tags`           | `{"some_tag":"some_value","more_tag":"some_other_value"}`|

<!-- markdownlint-enable -->

## 指标 {#metric}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

```toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

{{if eq $m.Type "metric"}}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 字段列表

{{$m.FieldsMarkdownTable}}
{{end}}

{{ end }}