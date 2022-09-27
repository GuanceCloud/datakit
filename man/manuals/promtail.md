{{.CSS}}
# Promtail
---

{{.AvailableArchs}}

---

启动一个 HTTP 端点监听并接收 promtail 日志数据，上报到观测云。

## 配置 {#config}

进入 DataKit 安装目录下的 `conf.d/log` 目录，复制 `promtail.conf.sample` 并命名为 `promtail.conf`。示例如下：

```toml
{{.InputSample}} 
```

### API 版本 {#API version}

通过配置 `legacy = true`，可以用 legacy 版本 API 处理接收到的 promtail
日志数据。详见 [POST /api/prom/push](https://grafana.com/docs/loki/latest/api/#post-apiprompush)
及 [POST /loki/api/v1/push](https://grafana.com/docs/loki/latest/api/#post-lokiapiv1push) 。

### 自定义标签 {#custom tags}

通过配置 `[inputs.promtail.tags]`，可以在日志数据中添加自定义标签，示例如下：

```toml
  [inputs.promtail.tags]
    some_tag = "some_value"
    more_tag = "some_other_value"
```

配置好后，重启 DataKit 即可。

### 支持参数 {#args}

promtail 采集器支持在 HTTP URL 中添加参数。参数列表如下：

- `source`：标识数据来源。例如 `nginx` 或者 `redis`（`/v1/write/promtail?source=nginx`)，默认将 `source` 设为 `default`；
- `pipeline`：指定数据需要使用的 pipeline 名称，例如 `nginx.p`（`/v1/write/promtail?pipeline=nginx.p`）；
- `tags`：添加自定义 tag，以英文逗号 `,` 分割，例如 `key1=value1` 和 `key2=value2`（`/v1/write/promtail?tags=key1=value1,key2=value2`）。

## 最佳实践 {#best practice}

promtail 的数据原本主要发送给 loki，即 `/loki/api/v1/push`，其配置样例如下：

```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://localhost:3100/loki/api/v1/push

scrape_configs:
  - job_name: system
    static_configs:
      - targets:
          - localhost
        labels:
          job: varlogs
          __path__: /var/log/*log
```

在开启 promtail 采集器后，可以配置 promtail 让其将数据发送给 Datakit 的 promtail 采集器：

```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://localhost:9529/v1/write/promtail    # 发送到 promtail 采集器监听的端点

scrape_configs:
  - job_name: system
    static_configs:
      - targets:
          - localhost
        labels:
          job: varlogs
          __path__: /var/log/*log
```

## 指标集 {#measurement}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }} 
