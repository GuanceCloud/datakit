{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# {{.InputName}}

该采集器是网络拨测结果数据采集，所有拨测产生的数据，都以行协议方式，通过 `/v1/write/logging` 接口,上报DataFlux平台

## 前置条件

暂无

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## http task 结构示例

| 参数名          | type       | 必选 | 说明                                   |
| --------------- | ---------- | ---- | ------------------------               |
| type            | string     | Y    | 云拨测类型，可选选项`http`,`tcp`,`dns` |
| name            | string     | Y    | 任务名称                               |
| url             | string     | Y    | url                                    |
| method          | string     | Y    | url 请求方法                           |
| status          | string     | Y    | 任务状态，可选值`ok`,`stop`            |
| frequency       | string     | Y    | 任务频率                               |
| advance_options | struct     |      | 高级设置,[详见](#advance_options)      |
| success_when    | array      | Y    | 判定条件,[详见](#success_when)         |
| post_url        | string     | Y    | 拨测结果上报dataway地址                |

### advance_options 

| 参数名          | type       | 必选 | 说明                                            |
| --------------- | ---------- | ---- | ------------------------                        |
| request_options | struct     | N    | 云拨测http请求参数设置,[详见](#request_options) |
| request_body    | struct     | N    | 云拨测http请求体设置,[详见](#request_body)      |
| certificate     | struct     | N    | 相关证书设置,[详见](#certificate)               |
| proxy           | struct     | N    | 代理设置,[详见](#proxy)                         |
| secret          | struct     | N    | 隐私设置,[详见](#secret)                        |


### success_when

| 参数名          | type                                      | 必选 | 说明                                            |
| --------------- | ----------                                | ---- | ------------------------                        |
| body            | SuccessOption的array                      | N    | body内容相关判断条件设置,[详见](#SuccessOption) |
| response_time   | string                                    | N    | 最小响应时间设置，如 100ms                      |
| header          | dictory (value值为 SuccessOption的array ) | N    | 头信息的相关判断条件设置,[详见](#SuccessOption) |
| status_code     | SuccessOption的array                      | N    | 状态码的相关判断条件设置,[详见](#SuccessOption) |

### SuccessOption

| 参数名          | type       | 必选 | 说明                     |
| --------------- | ---------- | ---- | ------------------------ |
| is              | string     | N    | 等于                     |
| is_not          | string     | N    | 不等于                   |
| match_regex     | string     | N    | 匹配                     |
| not_match_regex | string     | N    | 不匹配                   |
| contains        | string     | N    | 包含                     |
| not_contains    | string     | N    | 不包含                   |

### request_options

| 参数名          | type       | 必选 | 说明                               |
| --------------- | ---------- | ---- | ------------------------           |
| follow_redirect | bool       | N    | 是否重定向设置                     |
| headers         | dictory    | N    | http 请求头信息设置                |
| cookies         | string     | N    | cookie 设置                        |
| auth            | struct     | N    | 用户名、密码认证设置,[详见](#auth) |

### auth

| 参数名          | type       | 必选 | 说明                     |
| --------------- | ---------- | ---- | ------------------------ |
| username        | string     | N    | 用户名                   |
| password        | string     | N    | 密码                     |

###  request_body

| 参数名          | type       | 必选 | 说明                                                                 |
| --------------- | ---------- | ---- | ------------------------                                             |
| body_type       | string     | N    | 示例： `text/plain` `application/json` `text/xml` `None` `text/html` |
| body            | string     | N    | http 请求体                                                          |

###  certificate

| 参数名                          | type       | 必选 | 说明                           |
| ---------------                 | ---------- | ---- | ------------------------       |
| ignore_server_certificate_error | bool       | N    | 忽略服务器证书报错,默认为false |
| private_key                     | string     | N    | 私钥                           |
| certificate                     | string     | N    | 证书                           |
| ca                              | string     | N    | ca 证书                        |

### proxy

| 参数名          | type       | 必选 | 说明                     |
| --------------- | ---------- | ---- | ------------------------ |
| url             | string     | N    | 代理url                  |
| headers         | dictory    | N    | 请求头参数，kv结构       |


### secret

| 参数名          | type       | 必选 | 说明                                     |
| --------------- | ---------- | ---- | ------------------------                 |
| not_save        | bool       | N    | 不保存响应内容，则设置为true,默认为false |

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
