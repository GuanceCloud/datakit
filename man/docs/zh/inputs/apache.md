
# Apache
---

{{.AvailableArchs}}

---

Apache 采集器可以从 Apache 服务中采集请求数、连接数等，并将指标采集到观测云，帮助监控分析 Apache 各种异常情况。

## 前置条件 {#requirements}

- Apache 版本 >= `2.4.6 (Unix)`。已测试版本:
    - [x] 2.4.56
    - [x] 2.4.54
    - [x] 2.4.41
    - [x] 2.4.38
    - [x] 2.4.29
    - [x] 2.4.6

- 一般发行版 Linux 会自带 Apache,如需下载[参见](https://httpd.apache.org/download.cgi){:target="_blank"};
- 默认配置路径: `/etc/apache2/apache2.conf`, `/etc/apache2/httpd.conf`, `/usr/local/apache2/conf/httpd.conf`;
- 开启 Apache `mod_status`，在 Apache 配置文件中添加:

```xml
<Location /server-status>
SetHandler server-status

Order Deny,Allow
Deny from all
Allow from [YOUR_IP]
</Location>
```

- 重启 Apache

```shell
sudo apachectl restart
```

## 配置 {#config}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：
    
    ```toml
    {{ CodeBlock .InputSample 4 }}
    ```

    配置好后，[重启 DataKit](datakit-service-how-to.md#manage-service) 即可。

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

## 指标集 {#measurements}

以下所有数据采集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[inputs.{{.InputName}}.tags]` 指定其它标签：

``` toml
 [inputs.{{.InputName}}.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
  # ...
```

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}` {#{{$m.Name}}}

{{$m.Desc}}

- 标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}

## 日志采集 {#logging}

如需采集 Apache 的日志，可在 {{.InputName}}.conf 中 将 `files` 打开，并写入 Apache 日志文件的绝对路径。比如：

```toml
[[inputs.apache]]
  ...
  [inputs.apache.log]
    files = [
      "/var/log/apache2/error.log",
      "/var/log/apache2/access.log"
    ]
```

开启日志采集以后，默认会产生日志来源（`source`）为 `apache` 的日志。

<!-- markdownlint-disable MD046 -->
???+ attention

    必须将 DataKit 安装在 Apache 所在主机才能采集 Apache 日志
<!-- markdownlint-enable -->

## 日志 Pipeline 功能切割字段说明 {#pipeline}

- Apache 错误日志切割

错误日志文本示例：

``` log
[Tue May 19 18:39:45.272121 2021] [access_compat:error] [pid 9802] [client ::1:50547] AH01797: client denied by server configuration: /Library/WebServer/Documents/server-status
```

切割后的字段列表如下：

| 字段名   | 字段值                | 说明                         |
| ---      | ---                   | ---                          |
| `status` | `error`               | 日志等级                     |
| `pid`    | `9802`                | 进程 id                      |
| `type`   | `access_compat`       | 日志类型                     |
| `time`   | `1621391985000000000` | 纳秒时间戳（作为行协议时间） |

- Apache 访问日志切割

访问日志文本示例:

``` log
127.0.0.1 - - [17/May/2021:14:51:09 +0800] "GET /server-status?auto HTTP/1.1" 200 917
```

切割后的字段列表如下：

| 字段名         | 字段值                | 说明                         |
| ---            | ---                   | ---                          |
| `status`       | `info`                | 日志等级                     |
| `ip_or_host`   | `127.0.0.1`           | 请求方 IP 或者 host          |
| `http_code`    | `200`                 | http status code             |
| `http_method`  | `GET`                 | http 请求类型                |
| `http_url`     | `/`                   | http 请求 URL                |
| `http_version` | `1.1`                 | http version                 |
| `time`         | `1621205469000000000` | 纳秒时间戳（作为行协议时间） |
