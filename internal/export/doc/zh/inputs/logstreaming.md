---
title     : 'Log Streaming'
summary   : '通过 HTTP 上报日志数据'
__int_icon      : 'icon/logstreaming'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Log Streaming
<!-- markdownlint-enable -->

---

{{.AvailableArchs}}

---

启动一个 HTTP Server，接收日志文本数据，上报到观测云。HTTP URL 固定为：`/v1/write/logstreaming`，即 `http://Datakit_IP:PORT/v1/write/logstreaming`

> 注：如果 DataKit 以 DaemonSet 方式部署在 Kubernetes 中，可以使用 Service 方式访问，地址为 `http://datakit-service.datakit:9529`

## 配置 {#config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](../datakit/datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

### 支持参数 {#args}

Log-Streaming 支持在 HTTP URL 中添加参数，对日志数据进行操作。参数列表如下：

- `type`：数据格式，目前只支持 `influxdb` 和 `firelens`。
    - 当 `type` 为 `inflxudb` 时（`/v1/write/logstreaming?type=influxdb`），说明数据本身就是行协议格式（默认 precision 是 `s`），将只添加内置 Tags，不再做其他操作
    - 当 `type` 为 `firelens` 时 (`/v1/write/logstreaming?type=firelens`)，数据格式应是 JSON 格式的多条日志
    - 当此值为空时，会对数据做分行和 Pipeline 等处理
- `source`：标识数据来源，即行协议的 measurement。例如 `nginx` 或者 `redis`（`/v1/write/logstreaming?source=nginx`）
    - 当 `type` 是 `influxdb` 时，此值无效
    - 默认为 `default`
- `service`：添加 service 标签字段，例如（`/v1/write/logstreaming?service=nginx_service`）
    - 默认为 `source` 参数值。
- `pipeline`：指定数据需要使用的 pipeline 名称，例如 `nginx.p`（`/v1/write/logstreaming?pipeline=nginx.p`）
- `tags`：添加自定义 tag，以英文逗号 `,` 分割，例如 `key1=value1` 和 `key2=value2`（`/v1/write/logstreaming?tags=key1=value1,key2=value2`）

#### FireLens 数据源类型 {#firelens}

该类型数据中的 `log`, `source`, `date` 字段将会特殊处理，数据示例：

```json
[
  {
    "date": 1705485197.93957,
    "container_id": "xxxxxxxxxxx-xxxxxxx",
    "container_name": "nginx_demo",
    "source": "stdout",
    "log": "127.0.0.1 - - [19/Jan/2024:11:49:48 +0800] \"GET / HTTP/1.1\" 403 162 \"-\" \"curl/7.81.0\"",
    "ecs_cluster": "Cluster_demo"
  },
  {
    "date": 1705485197.943546,
    "container_id": "f68a9aeb3d64493595e89f8821fa3f86-4093234565",
    "container_name": "javatest",
    "source": "stdout",
    "log": "2024/01/19 11:49:48 [error] 1316#1316: *1 directory index of \"/var/www/html/\" is forbidden, client: 127.0.0.1, server: _, request: \"GET / HTTP/1.1\", host: \"localhost\"",
    "ecs_cluster": "Cluster_Demo"
  }
]
```

在提取出列表中的两条日志后，其中 `log` 将作为数据的 `message` 字段，`date` 将转换为日志的时间，`source` 将被重命名为 `firelens_source`。

### 使用方式 {#usage}

- Fluentd 使用 Influxdb Output [文档](https://github.com/fangli/fluent-plugin-influxdb){:target="_blank"}
- Fluentd 使用 HTTP Output [文档](https://docs.fluentd.org/output/http){:target="_blank"}
- Logstash 使用 Influxdb Output [文档](https://www.elastic.co/guide/en/logstash/current/plugins-outputs-influxdb.html){:target="_blank"}
- Logstash 使用 HTTP Output [文档](https://www.elastic.co/guide/en/logstash/current/plugins-outputs-http.html){:target="_blank"}

只需要将 Output Host 配置为 Log-Streaming URL（`http://Datakit_IP:PORT/v1/write/logstreaming`）并添加对应参数即可。

## 日志 {#logging}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
