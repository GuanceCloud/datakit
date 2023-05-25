
# Promtail 数据接入

---

{{.AvailableArchs}}

---

启动一个 HTTP 端点监听并接收 Promtail 日志数据，上报到观测云。

## 配置 {#config}

已测试的版本:

- [x] 2.8.2
- [x] 2.0.0
- [x] 1.5.0
- [x] 1.0.0
- [x] 0.1.0

进入 DataKit 安装目录下的 `conf.d/log` 目录，复制 `promtail.conf.sample` 并命名为 `promtail.conf`。示例如下：

```toml
{{.InputSample}} 
```

### API 版本 {#API version}

对于 `v0.3.0` 及之前的版本需要配置 `legacy = true`，即使用 [`POST /api/prom/push`](https://grafana.com/docs/loki/latest/api/#post-apiprompush){:target="_blank"}，可以用 Legacy 版本 API 处理接收 Promtail 的日志数据。

之后的版本使用默认配置，即 `legacy = false`，即使用 [`POST /loki/api/v1/push`](https://grafana.com/docs/loki/latest/api/#post-lokiapiv1push){:target="_blank"}。

### 自定义标签 {#custom tags}

通过配置 `[inputs.promtail.tags]`，可以在日志数据中添加自定义标签，示例如下：

```toml
  [inputs.promtail.tags]
    some_tag = "some_value"
    more_tag = "some_other_value"
```

配置好后，重启 DataKit 即可。

### 支持参数 {#args}

Promtail 采集器支持在 HTTP URL 中添加参数。参数列表如下：

- `source`：标识数据来源。例如 `nginx` 或者 `redis`（`/v1/write/promtail?source=nginx`)，默认将 `source` 设为 `default`；
- `pipeline`：指定数据需要使用的 pipeline 名称，例如 `nginx.p`（`/v1/write/promtail?pipeline=nginx.p`）；
- `tags`：添加自定义 tag，以英文逗号 `,` 分割，例如 `key1=value1` 和 `key2=value2`（`/v1/write/promtail?tags=key1=value1,key2=value2`）。

## 最佳实践 {#best practice}

Promtail 的数据原本发送给 Loki，即 `/loki/api/v1/push`。将 Promtail 配置中的 `url` 修改为指向 Datakit，开启 Datakit 的 Promtail 采集器后，Promtail 会将其数据发送给 Datakit 的 Promtail 采集器。

Promtail 的配置示例如下：

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
