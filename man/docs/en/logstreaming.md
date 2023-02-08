
# logstreaming
---

{{.AvailableArchs}}

---

Start an HTTP Server, receive the log text data and report it to Guance Cloud.

HTTP URL is fixed as: `/v1/write/logstreaming`, that is, `http://Datakit_IP:PORT/v1/write/logstreaming`

Note: If DataKit is deployed in Kubernetes as a daemonset, it can be accessed as a Service at `http://datakit-service.datakit:9529`

## Configuration {#config}

=== "Host Installation"

    Go to the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```
    
    Once configured, [restart DataKit](datakit-service-how-to.md#manage-service).

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](datakit-daemonset-deploy.md#configmap-setting).

### Support Parameter {#args}

logstreaming supports adding parameters to the HTTP URL to manipulate log data. The list of parameters is as follows:

- `type`: Data format, currently only `influxdb` is supported.
  - When `type` is `inflxudb` (`/v1/write/logstreaming?type=influxdb`), the data itself is in row protocol format (default precision is `s`), and only built-in Tags will be added and nothing else will be done
- `source`: Identify the source of the data, that is, the measurement of the line protocol. Such as `nginx` or `redis` (`/v1/write/logstreaming?source=nginx`)
  - This value is not valid when `type` is `influxdb`
  - Default is `default`
- `service`: Add a service label field, such as (`/v1/write/logstreaming?service=nginx_service`）
  - Default to `source` parameter value.
- `pipeline`: Specify the pipeline name required for the data, such as `nginx.p`（`/v1/write/logstreaming?pipeline=nginx.p`）
- `tags`: Add custom tags, split by `,`, such as `key1=value1` and `key2=value2`（`/v1/write/logstreaming?tags=key1=value1,key2=value2`)

### Usage {#usage}

- Fluentd uses Influxdb Output [doc](https://github.com/fangli/fluent-plugin-influxdb){:target="_blank"}
- Fluentd uses HTTP Output [doc](https://docs.fluentd.org/output/http){:target="_blank"}
- Logstash uses Influxdb Output [doc](https://www.elastic.co/guide/en/logstash/current/plugins-outputs-influxdb.html){:target="_blank"}
- Logstash uses HTTP Output [doc](https://www.elastic.co/guide/en/logstash/current/plugins-outputs-http.html){:target="_blank"}

Simply configure Output Host as a logstreaming URL (`http://Datakit_IP:PORT/v1/write/logstreaming`）and add corresponding parameters.

## Measurements {#measurement}

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

{{$m.Desc}}

- tag

{{$m.TagsMarkdownTable}}

- metric list

{{$m.FieldsMarkdownTable}}

{{ end }}
