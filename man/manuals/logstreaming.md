{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# {{.InputName}}

启动一个 HTTP Server，接收日志文本数据，上报到观测云。

HTTP URL 固定为：`/v1/write/logstreaming`，即 `http://Datakit_IP:PORT/v1/write/logstreaming`

注：如果 DataKit 以 daemonset 方式部署在 Kubernetes 中，可以使用 Service 方式访问，地址为 `http://datakit-service.datakit:9529`

## 配置

进入 DataKit 安装目录下的 `conf.d/log` 目录，复制 `logstreaming.conf.sample` 并命名为 `logstreaming.conf`。示例如下：

``` toml
{{.InputSample}} 
```

配置好后，重启 DataKit 即可。

### 支持参数

logstreaming 支持在 HTTP URL 中添加参数，对日志数据进行操作。参数列表如下：

- `type`：数据格式，目前只支持 `influxdb`。
  - 当 `type` 为 `inflxudb` 时（`/v1/write/logstreaming?type=influxdb`），说明数据本身就是行协议格式（默认 precision 是 `s`），将只添加内置 Tags，不再做其他操作
  - 当此值为空时，会对数据做分行和 pipeline 等处理
- `source`：标识数据来源，即行协议的 measurement。例如 `nginx` 或者 `redis`（`/v1/write/logstreaming?source=nginx`）
  - 当 `type` 是 `influxdb` 时，此值无效
  - 默认为 `default`
- `service`：添加 service 标签字段，例如（`/v1/write/logstreaming?service=nginx_service`）
  - 默认为 `source` 参数值。
- `pipeline`：指定数据需要使用的 pipeline 名称，例如 `nginx.p`（`/v1/write/logstreaming?pipeline=nginx.p`）


### 使用方式

- Fluentd 使用 Influxdb Output [文档](https://github.com/fangli/fluent-plugin-influxdb)
- Fluentd 使用 HTTP Output [文档](https://docs.fluentd.org/output/http)
- Logstash 使用 Influxdb Output [文档](https://www.elastic.co/guide/en/logstash/current/plugins-outputs-influxdb.html)
- Logstash 使用 HTTP Output [文档](https://www.elastic.co/guide/en/logstash/current/plugins-outputs-http.html)

只需要将 Output Host 配置为 logstreaming URL （`http://Datakit_IP:PORT/v1/write/logstreaming`）并添加对应参数即可。

## 指标集

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }} 
