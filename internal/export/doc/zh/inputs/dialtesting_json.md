---
title     : '自定义拨测任务'
summary   : '自定义拨测采集器来定制拨测任务'
tags:
  - '拨测'
  - '网络'
__int_icon      : 'icon/dialtesting'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---


某些情况下，可能不能连接 SAAS 的拨测任务服务，此时，我们可以通过本地的 JSON 文件来定义拨测任务。

## 配置 {#config}

### 配置采集器 {#config-inputs}

<!-- markdownlint-disable MD046 -->
=== "主机安装"

    进入 DataKit 安装目录下的 `conf.d/network` 目录，复制 `dialtesting.conf.sample` 并命名为 `dialtesting.conf`。示例如下：

    ```toml
    [[inputs.dialtesting]]
      server = "file://</path/to/your/local.json>"
    
      # 注意：以 Linux 为例，假定你的 json 目录为 /some/path/my.json，那么此处的
      # server 应该写成 file:///some/path/my.json
    
      # 注意，以下 tag 建议都一一填写（不要修改这里的 tag key），便于在页面上展示完整的拨测结果
      [inputs.dialtesting.tags] 
        country  = "<specify-datakit-country>"  # DataKit 部署所在的国家
        province = "<specify-datakit-province>" # DataKit 部署所在的省份
        city     = "<specify-datakit-city>"     # DataKit 部署所在的城市
        isp      = "<specify-datakit-ISP>"      # 指定 DataKit 所在的网络服务商
        region   = "<your-region>"              # 可随意指定一个 region 名称
    ```

=== "Kubernetes"

    目前可以通过 [ConfigMap 方式注入采集器配置](../datakit/datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

---

### 配置拨测任务 {#config-task}

目前拨测任务支持五种拨测类型，即 HTTP, TCP, ICMP, WEBSOCKET, GRPC 服务，JSON 格式如下：

```json
{
  "<拨测类型>": [
    {拨测任务 1},
    {拨测任务 2},
       ...
    {拨测任务 n},
  ]
}
```

下面是一个具体的拨测示例：

```json
{
  "HTTP": [
    {
      "name": "baidu-json-test",
      "method": "GET",
      "url": "http://baidu.com",
      "post_url": "https://<your-dataway-host>?token=<your-token>",
      "status": "OK",
      "schedule_type": "frequency",
      "frequency": "10s",
      "success_when_logic": "and",
      "success_when": [
        {
          "response_time": "1000ms",
          "header": {
            "Content-Type": [
              {
                "contains": "html"
              }
            ]
          },
          "status_code": [
            {
              "is": "200"
            }
          ]
        }
      ],
      "advance_options": {
        "protocol": "auto",
        "request_options": {
          "auth": {}
        },
        "request_body": {}
      },
      "update_time": 1645065786362746
    }
  ],
  "TCP": [
    {
      "name": "tcp-test",
      "host": "www.baidu.com",
      "port": "80",
      "status": "OK",
      "frequency": "10s",
      "success_when_logic": "or",
      "success_when": [
        {
          "response_time": [
            {
              "is_contain_dns": true,
              "target": "10ms"
            }
          ]
        }
      ],
      "update_time": 1641440314550918
    }
  ]
}
```

> 编辑完这个 JSON 后，建议找一些[在线工具](https://www.json.cn/){:target="_blank"}验证下 JSON 格式是不是正确。如果 JSON 格式不对，那么会导致拨测不生效。

配置好后，重启 DataKit 即可。

### 拨测任务字段定义 {#field-def}

拨测任务字段包括「公共字段」和具体拨测任务的「额外字段」。

#### 公共字段 {#pub}

拨测任务公共字段定义如下：

| 字段                 | 类型   | 是否必须 | 说明                                                                                 |
| :---                 | ---    | ---      | ---                                                                                  |
| `name`               | string | Y        | 拨测服务名称                                                                         |
| `status`             | string | Y        | 拨测服务状态，如 `OK/stop`                                                           |
| `frequency`          | string | Y        | 拨测频率                                                                             |
| `schedule_type`      | string | Y        | 拨测任务执行类型，如 `frequency` 或 `crontab`，默认为 `frequency` |
| `crontab`            | string | N        | `crontab` 表达式，如 `*/10 * * * *`，仅当 `schedule_type` 为 `crontab` 时生效 |
| `success_when_logic` | string | N        | `success_when` 条件之间的逻辑关系，如 `and/or`，默认为 `and`                         |
| `success_when`       | object | Y        | 详见下文                                                                             |
| `advance_options`    | object | N        | 详见下文                                                                             |
| `post_url`           | string | N        | 将拨测结果发往该 Token 所指向的工作空间，如果不填写，则发给当前 DataKit 所在工作空间 |
| `tags_info`           | string | N        | 拨测任务自定义标签，如： `t1,t2` |
| `workspace_language`  | string | N        | 当前工作空间语言，如：`zh`，`en` |

#### HTTP 拨测 {#http}

额外字段

| 字段              | 类型   | 是否必须 | 说明                                    |
| :---              | ---    | ---      | ---                                     |
| `method`          | string | Y        | HTTP 请求方法                           |
| `url`             | string | Y        | 完整的 HTTP 请求地址                    |
| `post_script`     | string | N        | Pipeline 脚本，用于结果判断和提取变量                    |

总体的 JSON 结构如下：

``` json
{
  "HTTP": [
    {
      "name": "baidu-json-test",
      "method": "GET",
      "url": "http://baidu.com",
      "post_url": "https://<your-dataway-host>?token=<your-token>",
      "status": "OK",
      "frequency": "10s",
      "success_when_logic": "and",
      "success_when": ...,
      "advance_options": ...,
      "post_script":...,
    },
    {
      ... another HTTP dialtesting
    }
  ]
}
```

##### `success_when` 定义 {#http-success-when}

用来定义拨测成功与否的判定条件，主要有如下几个方面：

- HTTP 请求返回 body 判断（`body`）

| 字段              | 类型   | 是否必须 | 说明                                      |
| :---              | ---    | ---      | ---                                       |
| `is`              | string | N        | 返回的 body 是否等于该指定字段                   |
| `is_not`          | string | N        | 返回的 body 是否不等于该指定字段                 |
| `match_regex`     | string | N        | 返回的 body 是否含有该匹配正则表达式的子字符串   |
| `not_match_regex` | string | N        | 返回的 body 是否不含有该匹配正则表达式的子字符串 |
| `contains`        | string | N        | 返回的 body 是否含有该指定的子字符串             |
| `not_contains`    | string | N        | 返回的 body 是否不含有该指定的子字符串           |

如：

```json
"success_when": [
  {
    "body": [
      {
        "match_regex": "\\d\\d.*",
      }
    ]
  }
]
```

此处 `body` 可以配置多个验证规则，由 `success_when_logic` 确定他们之间的关系，配置为 `and` 时，**任何一个规则验证不过，则认为当前拨测失败**；配置为 `or` 时，**任何一个规则验证通过，则认为当前拨测成功**；默认是 `and` 的关系。下面的验证规则，均遵循这一判定原则。

> 注意，此处正则要正确转义，示例中的实际正则表达式是 `\d\d.*`。

- HTTP 请求返回 Header 判断（`header`）

| 字段              | 类型   | 是否必须 | 说明                                                       |
| :---              | ---    | ---      | ---                                                        |
| `is`              | string | N        | 返回的 header 指定字段是否等于该指定值                     |
| `is_not`          | string | N        | 返回的 header 指定字段是否不等于该指定值                   |
| `match_regex`     | string | N        | 返回的 header 指定字段是否含有该匹配正则表达式的子字符串   |
| `not_match_regex` | string | N        | 返回的 header 指定字段是否不含有该匹配正则表达式的子字符串 |
| `contains`        | string | N        | 返回的 header 指定字段是否含有该指定的子字符串             |
| `not_contains`    | string | N        | 返回的 header 指定字段是否不含有该指定的子字符串           |

如：

```json
"success_when": [
  {
    "header": {
       "Content-Type": [
         {
           "contains": "html"
         }
       ]
    }
  }
]
```

由于可能存在多种类型 Header 的判定，此处也能配置多种 Header 的检验：

```json
"success_when": [
  {
    "header": {
       "Content-Type": [
         {
           "contains": "some-header-value"
         }
       ],

       "Etag": [
         {
           "match_regex": "\\d\\d-.*"
         }
       ]
    }
  }
]
```

- HTTP 请求返回状态码（`status_code`）

| 字段              | 类型   | 是否必须 | 说明                                             |
| :---              | ---    | ---      | ---                                              |
| `is`              | string | N        | 返回的 status code 是否等于该指定字段                   |
| `is_not`          | string | N        | 返回的 status code 是否不等于该指定字段                 |
| `match_regex`     | string | N        | 返回的 status code 是否含有该匹配正则表达式的子字符串   |
| `not_match_regex` | string | N        | 返回的 status code 是否不含有该匹配正则表达式的子字符串 |
| `contains`        | string | N        | 返回的 status code 是否含有该指定的子字符串             |
| `not_contains`    | string | N        | 返回的 status code 是否不含有该指定的子字符串           |

如：

```json
"success_when": [
  {
    "status_code": [
      {
        "is": "200"
      }
    ]
  }
]
```

> 对于一个确定的 URL 拨测，一般其 HTTP 返回就一个，故此处一般只配置一个验证规则（虽然支持数组配置多个）。

- HTTP 请求响应时间（`response_time`）

此处只能填写一个时间值，如果请求的响应时间小于该指定值，则判定拨测成功，如：

```json
"success_when": [
  {
    "response_time": "100ms"
  }
]
```

> 注意，此处指定的时间单位有 `ns`（纳秒）/`us`（微秒）/`ms`（毫秒）/`s`（秒）/`m`（分钟）/`h`（小时）。对 HTTP 拨测而言，一般使用 `ms` 单位。

以上列举的几种判定依据，可以组合使用，由 `success_when_logic` 确定他们之间的关系，配置为 `and` 时，**任何一个规则验证不过，则认为当前拨测失败**；配置为 `or` 时，**任何一个规则验证通过，则认为当前拨测成功**；默认是 `and` 的关系。如：

```json
"success_when": [
  {
    "response_time": "1000ms",
    "header": { HTTP header 相关判定 },
    "status_code": [ HTTP 状态码相关判定 ],
    "body": [ HTTP body 相关判定 ]
  }
]
```

##### `advance_options` 定义 {#http-advance-options}

高级选项主要用来调整具体的拨测行为，主要有如下几个方面：

- HTTP 请求选项（`request_options`）

| 字段              | 类型              | 是否必须 | 说明                       |
| :---              | ---               | ---      | ---                        |
| `follow_redirect` | bool              | N        | 是否支持重定向跳转         |
| `headers`         | map[string]string | N        | HTTP 请求时指定一组 Header |
| `cookies`         | string            | N        | 指定请求的 Cookie          |
| `auth`            | object            | N        | 指定请求的认证方式         |

其中 `auth` 只支持普通的用户名密码认证，定义如下：

| 字段       | 类型   | 是否必须 | 说明       |
| :---       | ---    | ---      | ---        |
| `username` | string | Y        | 用户名     |
| `password` | string | Y        | 用户名密码 |

`request_options` 示例：

```json
"advance_options": {
  "request_options": {
    "auth": {
        "username": "张三",
        "password": "fawaikuangtu"
      },
    "headers": {
      "X-Prison-Breaker": "张三",
      "X-Prison-Break-Password": "fawaikuangtu"
    },
    "follow_redirect": false
  },
}
```

- HTTP 请求 Body（`request_body`）

| 字段        | 类型   | 是否必须 | 说明                                    |
| :---        | ---    | ---      | ---                                     |
| `body_type` | string | N        | Body 类型，即请求头 `Content-Type` 的值 |
| `body`      | string | N        | 请求 Body                               |
| `form`      | map[string]string | N        | `Content-Type` 为 `multipart/form-data` 时，请求 Body 的表单数据                     |

`request_body` 示例：

```json
"advance_options": {
  "request_body": {
    "body_type": "multipart/form-data",
    "body": "填写好请求体，此处注意各种复杂的转义",
    "form": {
      "name": "file",
    }
  }
}
```

- HTTP 超时时间（`request_timeout`）

HTTP 请求的超时时间，默认是 "60s"，即 60 秒。

`request_timeout` 示例：

```json
"advance_options": {
  "request_timeout": "60s",
}
```

- HTTP 协议版本（`protocol`）

HTTP 请求协议版本，默认是 `auto`。可选值如下：

| 值            | 说明                                                                                                                                  |
| :---          | :---                                                                                                                                  |
| `auto`        | 默认模式。`http://` 走 HTTP/1.1；`https://` 通过 `TLS/ALPN` 自动协商，服务端支持 HTTP/2 就走 HTTP/2，否则回退 HTTP/1.1。不会自动尝试 h2c，也不会尝试 HTTP/3。 |
| `http/1.1`    | 强制 HTTP/1.1，不会尝试 HTTP/2。                                                                                                      |
| `http/2`      | 启用 HTTP/2，但允许回退。对 `https://` 会通过 `ALPN` 尝试 HTTP/2，不支持时回退 HTTP/1.1；对普通 `http://` 本质上仍是 HTTP/1.1，不会走 h2c。 |
| `http/2-only` | 强制 HTTP/2。`https://` 要求 `ALPN` 协商到 `h2`；`http://` 会按 h2c prior knowledge 直接发明文 HTTP/2。服务端不支持 HTTP/2 时任务失败。 |
| `http/3`      | 使用 HTTP/3，也就是 `QUIC` + TLS。需要服务端支持 HTTP/3；当前不支持代理配置。                                                           |

`protocol` 示例：

```json
"advance_options": {
  "protocol": "http/3"
}
```

> 注意：HTTP/3 不支持代理配置。如果 `protocol` 配置为 `http/3`，请不要同时配置 `proxy`。

- HTTP 请求证书（`certificate`）

| 字段                              | 类型   | 是否必须 | 说明             |
| :---                              | ---    | ---      | ---              |
| `ignore_server_certificate_error` | bool   | N        | 是否忽略证书错误 |
| `private_key`                     | string | N        | key              |
| `certificate`                     | string | N        | 证书             |
| `ca`                              | string | N        | 暂时未使用       |

`certificate` 示例：

```json
"advance_options": {
  "certificate": {
    "ignore_server_certificate_error": false,
    "private_key": "<your-private-key>",
    "certificate": "<your-certificate-key>"
  },
}
```

`private_key` 示例：

``` not-set
-----BEGIN PRIVATE KEY-----
MIIxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxNn+/x
9WKHZvRf3lbLY7GAR/emacU=
-----END PRIVATE KEY-----
```

下面是 `certificate` 示例：

```not-set
-----BEGIN CERTIFICATE-----
MIIxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxWDR/+
InEHyg==
-----END CERTIFICATE-----
```

在 Linux 下，可通过如下命令生成这对 key：

```shell
openssl req -newkey rsa:2048 -x509 -sha256 -days 3650 -nodes -out example.crt -keyout example.key
```

- HTTP 请求代理（`proxy`）

| 字段      | 类型              | 是否必须 | 说明                                 |
| :---      | ---               | ---      | ---                                  |
| `url`     | string            | N        | 代理的 URL，如 `http://1.2.3.4:4321` |
| `headers` | map[string]string | N        | HTTP 请求时指定一组 Header           |

`proxy` 示例：

```json
"advance_options": {
  "request_options": {
    "proxy": {
      "url": "http://1.2.3.4:4321",
      "headers": {
        "X-proxy-header": "my-proxy-foobar"
      }
    }
  },
}
```

##### `post_script` 定义 {#post_script}

`post_script` 是一个 [Pipeline](../pipeline/use-pipeline/index.md) 脚本，用于对拨测结果进行判断，并从中提取变量。

##### 注入变量 {#inject-variables}

为了便于 `post_script` 处理 HTTP 响应并对检测结果进行判定和变量提取，编写脚本时可使用一些预定义的变量。具体如下：

- `response`: 响应对象

| 字段              | 类型   | 说明                                    |
| :---              | ---    | ---                                     |
| `status_code`     | number | 响应状态码                           |
| `header`          | json | 响应头，格式为 `{"header1": ["value1", "value2"]}`|
| `body`            | string | 响应内容|

- `result`： 检测结果

| 字段              | 类型   | 说明                                    |
| :---              | ---    | ---                                     |
| `is_failed`     | bool | 是否失败                           |
| `error_message`     | string | 失败原因                           |

- `vars`: 提取变量

JSON 对象，key 为变量名，value 为变量值， 如 `vars["token"] = "123"`

##### 示例 {#example}

```javascript

body = load_json(response["body"])

if body["code"] == "200" {
  result["is_failed"] = false
  vars["token"] = body["token"]
} else {
  result["is_failed"] = true
  result["error_message"] = body["message"]
}

```

上面的脚本中，首先使用 `load_json` 将响应内容解析为 JSON 对象，然后判断响应状态码是否为 200，如果为 200，则将响应内容中的 token 提取出来，并设置到 `vars` 中，否则将 `result` 的 `is_failed` 设置为 true， 且 `error_message` 设置为响应内容中的 message 。

#### TCP 拨测 {#tcp}

##### 额外字段 {#tcp-extra}

| 字段              | 类型   | 是否必须 | 说明                                    |
| :---              | ---    | ---      | ---                                     |
| `host`          | string | Y        | TCP 主机地址                           |
| `port`             | string | Y        | TCP 端口                    |
| `timeout`             | string | N        | TCP 连接超时时间                    |
| `message`       | string | N        | TCP 发送的消息                |

完整 JSON 结构如下：

``` json
{
    "TCP": [
        {
            "name": "tcp-test",
            "host": "www.baidu.com",
            "port": "80",
            "message": "hello",
            "timeout": "10ms",
            "enable_traceroute": true,
            "post_url": "https://<your-dataway-host>?token=<your-token>",
            "status": "OK",
            "frequency": "60s",
            "success_when_logic": "and",
            "success_when": [
                {
                    "response_time":[ 
                        {
                            "is_contain_dns": true,
                            "target": "10ms"
                        }
                    ],
                    "response_message": [
                        {
                            "is": "hello"
                        }
                    ],
                    "hops": [
                        {
                            "op": "eq",
                            "target": 20
                        }
                    ]
                }
            ]
        }
    ]
}
```

##### `success_when` 定义 {#tcp-success-when}

- TCP 响应时间判断 (`response_time`)

`response_time` 为一个数组对象，每个对象参数如下：

| 字段              | 类型   | 是否必须 | 说明                                                       |
| :---              | ---    | ---      | ---                                                        |
| `target`          | string | Y        | 判定响应时间是否小于该值                     |
| `is_contain_dns`  | bool | N        | 指明响应时间是否包含 DNS 解析时间                     |

```json
"success_when": [
  {
    "response_time": [
      {
        "is_contain_dns": true,
        "target": "10ms"
      }
    ]
  }
]
```

- 返回消息判定（`response_message`）

`response_message` 为一个数组对象，每个对象参数如下：

| 字段              | 类型   | 是否必须 | 说明                                                |
| :---              | ---    | ---      | ---                                                 |
| `is`              | string | N        | 返回的 message 是否等于该指定字段                   |
| `is_not`          | string | N        | 返回的 message 是否不等于该指定字段                 |
| `match_regex`     | string | N        | 返回的 message 是否含有该匹配正则表达式的子字符串   |
| `not_match_regex` | string | N        | 返回的 message 是否不含有该匹配正则表达式的子字符串 |
| `contains`        | string | N        | 返回的 message 是否含有该指定的子字符串             |
| `not_contains`    | string | N        | 返回的 message 是否不含有该指定的子字符串           |

如：

```json
"success_when": [
  {
    "response_message": [
      {
        "is": "reply",
      }
    ]
  }
]
```

- 网络跳数 (`hops`)

`hops` 为一个数组对象，每个对象参数如下：

| 字段              | 类型   | 是否必须 | 说明                                      |
| :---              | ---    | ---      | ---                                       |
| `op`              | string | Y        | 比较关系，可取值 `eq(=),lt(<),leq(<=),gt(>),geq(>=)`|
| `target`          | float | Y        | 判定值                 |

```json
"success_when": [
  {
    "hops": [
      {
        "op": "eq",
        "target": 20
      }
    ]
  }
]
```

#### ICMP 拨测 {#icmp}

##### 额外字段 {#icmp-extra}

| 字段              | 类型   | 是否必须 | 说明                                    |
| :---              | ---    | ---      | ---                                     |
| `host`            | string | Y        | 主机地址                           |
| `packet_count`    | int |   N         | 发送 ICMP 包的次数  
| `timeout`             | string | N    | 连接超时时间

完整 JSON 结构如下：

``` json
{
    "ICMP": [
        {
            "name": "icmp-test",
            "host": "www.baidu.com",
            "timeout": "10ms",
            "packet_count": 3,
            "enable_traceroute": true,
            "post_url": "https://<your-dataway-host>?token=<your-token>",
            "status": "OK",
            "frequency": "10s",
            "success_when_logic": "and",
            "success_when": [
                {
                    "response_time": [
                        {
                            "func": "avg",
                            "op": "leq",
                            "target": "50ms"
                        }
                    ],
                    "packet_loss_percent": [
                        {
                            "op": "leq",
                            "target": 20
                        }
                    ],
                    "hops": [
                        {
                            "op": "eq",
                            "target": 20
                        }
                    ],
                    "packets": [
                        {
                            "op": "geq",
                            "target": 1
                        }
                    ]
                }
            ]
        }
    ]
}
```

##### `success_when` 定义 {#icmp-success-when}

- ICMP 丢包率 (`packet_loss_percent`)

填写具体的值，为一个数组对象，每个对象参数如下：

| 字段              | 类型   | 是否必须 | 说明                                      |
| :---              | ---    | ---      | ---                                       |
| `op`              | string | Y        | 比较关系，可取值 `eq(=),lt(<),leq(<=),gt(>),geq(>=)`|
| `target`          | float | Y        | 判定值                 |

```json
"success_when": [
  {
    "packet_loss_percent": [
      {
        "op": "leq",
        "target": 20
      }
    ]
  }
]
```

- ICMP 响应时间 (`response_time`)

填写具体的时间，为一个数组对象，每个对象参数如下：

| 字段              | 类型   | 是否必须 | 说明                                      |
| :---              | ---    | ---      | ---                                       |
| `func`            | string | Y        | 统计类型，可取值 `avg,min,max,std`|
| `op`              | string | Y        | 比较关系，可取值 `eq(=),lt(<),leq(<=),gt(>),geq(>=)`|
| `target`          | string | Y        | 判定值                 |

```json
"success_when": [
  {
     "response_time": [
        {
          "func": "avg",
          "op": "leq",
          "target": "50ms"
        }
      ],
  }
]
```

- 网络跳数 (`hops`)

`hops` 为一个数组对象，每个对象参数如下：

| 字段              | 类型   | 是否必须 | 说明                                      |
| :---              | ---    | ---      | ---                                       |
| `op`              | string | Y        | 比较关系，可取值 `eq(=),lt(<),leq(<=),gt(>),geq(>=)`|
| `target`          | float | Y        | 判定值                 |

```json
"success_when": [
  {
    "hops": [
      {
        "op": "eq",
        "target": 20
      }
    ]
  }
]
```

- 抓包数 (`packets`)

`packets` 为一个数组对象，每个对象参数如下：

| 字段              | 类型   | 是否必须 | 说明                                      |
| :---              | ---    | ---      | ---                                       |
| `op`              | string | Y        | 比较关系，可取值 `eq(=),lt(<),leq(<=),gt(>),geq(>=)`|
| `target`          | float | Y        | 判定值                 |

```json
"success_when": [
  {
    "packets": [
      {
        "op": "eq",
        "target": 20
      }
    ]
  }
]
```

#### WEBSOCKET 拨测 {#ws}

##### 额外字段 {#ws-extra}

| 字段              | 类型   | 是否必须 | 说明                                    |
| :---              | ---    | ---      | ---                                     |
| `url`          | string | Y        | Websocket 连接地址，如 ws://localhost:8080  |
| `message`       | string | Y        | Websocket 连接成功后发送的消息                |

完整 JSON 结构如下：

```json
{
    "WEBSOCKET": [
        {
            "name": "websocket-test",
            "url": "ws://localhost:8080",
            "message": "hello",
            "post_url": "https://<your-dataway-host>?token=<your-token>",
            "status": "OK",
            "frequency": "10s",
            "success_when_logic": "and",
            "success_when": [
                {
                    "response_time": [
                        {
                            "is_contain_dns": true,
                            "target": "10ms"
                        }
                    ],
                    "response_message": [
                        {
                            "is": "hello1"
                        }
                    ],
                    "header": {
                        "status": [
                            {
                                "is": "ok"
                            }
                        ]
                    }
                }
            ],
            "advance_options": {
                "request_options": {
                    "timeout": "10s",
                    "headers": {
                        "x-token": "aaaaaaa",
                        "x-header": "111111"
                    }
                },
                "auth": {
                    "username": "admin",
                    "password": "123456"
                }
            }
        }
    ]
}
```

##### `success_when` 定义 {#ws-success-when}

- 响应时间判断 (`response_time`)

`response_time` 为一个数组对象，每个对象参数如下：

| 字段             | 类型   | 是否必须 | 说明                              |
| :---             | ---    | ---      | ---                               |
| `target`         | string | Y        | 判定响应时间是否小于该值          |
| `is_contain_dns` | bool   | N        | 指明响应时间是否包含 DNS 解析时间 |

```json
"success_when": [
  {
    "response_time": [
      {
        "is_contain_dns": true,
        "target": "10ms"
      }
    ]
  }
]
```

- 返回消息判定（`response_message`）

`response_message` 为一个数组对象，每个对象参数如下：

| 字段              | 类型   | 是否必须 | 说明                                                |
| :---              | ---    | ---      | ---                                                 |
| `is`              | string | N        | 返回的 message 是否等于该指定字段                   |
| `is_not`          | string | N        | 返回的 message 是否不等于该指定字段                 |
| `match_regex`     | string | N        | 返回的 message 是否含有该匹配正则表达式的子字符串   |
| `not_match_regex` | string | N        | 返回的 message 是否不含有该匹配正则表达式的子字符串 |
| `contains`        | string | N        | 返回的 message 是否含有该指定的子字符串             |
| `not_contains`    | string | N        | 返回的 message 是否不含有该指定的子字符串           |

如：

```json
"success_when": [
  {
    "response_message": [
      {
        "is": "reply",
      }
    ]
  }
]
```

- 请求返回 Header 判断（`header`）

`header` 为一个字典类型对象，其每个对象元素的值为为一个数组对象，相应参数如下：

| 字段              | 类型   | 是否必须 | 说明                                                       |
| :---              | ---    | ---      | ---                                                        |
| `is`              | string | N        | 返回的 header 指定字段是否等于该指定值                     |
| `is_not`          | string | N        | 返回的 header 指定字段是否不等于该指定值                   |
| `match_regex`     | string | N        | 返回的 header 指定字段是否含有该匹配正则表达式的子字符串   |
| `not_match_regex` | string | N        | 返回的 header 指定字段是否不含有该匹配正则表达式的子字符串 |
| `contains`        | string | N        | 返回的 header 指定字段是否含有该指定的子字符串             |
| `not_contains`    | string | N        | 返回的 header 指定字段是否不含有该指定的子字符串           |

如：

```json
"success_when": [
  {
    "header": {
       "Status": [
         {
           "is": "ok"
         }
       ]
    }
  }
]
```

##### `advance_options` 定义 {#ws-advance-options}

- 请求选项 (`request_options`)

| 字段              | 类型              | 是否必须 | 说明                       |
| :---              | ---               | ---      | ---                        |
| `timeout` | string              | N        | 连接超时时间         |
| `headers` | map[string]string | N        |  请求时指定一组 Header |

```json
"advance_options": {
  "request_options": {
    "timeout": "30ms",
    "headers": {
      "X-Token": "xxxxxxxxxx"
    }
  },
}
```

- 认证信息 (`auth`)

支持普通的用户名和密码认证(Basic access authentication)。

| 字段       | 类型   | 是否必须 | 说明       |
| :---       | ---    | ---      | ---        |
| `username` | string | Y        | 用户名     |
| `password` | string | Y        | 用户名密码 |

```json
"advance_options": {
  "auth": {
    "username": "admin",
    "password": "123456"
  },
}
```

#### GRPC 拨测 {#grpc}

##### 额外字段 {#grpc-extra}

| 字段              | 类型   | 是否必须 | 说明                                    |
| :---              | ---    | ---      | ---                                     |
| `server`          | string | Y        | gRPC 服务器地址，如 `localhost:50051`   |
| `post_script`     | string | N        | Pipeline 脚本，用于结果判断   |

> 注意：gRPC 拨测仅支持 unary RPC（一元 RPC），不支持 streaming RPC（流式 RPC）。

完整 JSON 结构如下：

```json
{
  "GRPC": [
    {
      "name": "grpc-test",
      "server": "localhost:50051",
      "post_url": "https://<your-dataway-host>?token=<your-token>",
      "status": "OK",
      "frequency": "5m",
      "success_when_logic": "and",
      "success_when": [
        {
          "body": [
            {
              "contains": "success"
            }
          ],
          "response_time": "500ms"
        }
      ],
      "advance_options": {
        "request_options": {
          "request_timeout": "10s",
          "metadata": {
            "x-token": "test-token"
          },
          "proto_files": {
            "protofiles": {
               "greeter.proto": "syntax = \"proto3\";\n\npackage greeter;\n\noption go_package = \"datakittest/grpc/pb\";\n\nimport \"pb/common.proto\";\n\nservice Greeter {\n  rpc SayHello (HelloRequest) returns (common.result) {}\n}\n\nmessage HelloRequest {\n  string name = 1;\n}",
               "pb/common.proto": "syntax = \"proto3\";\n\npackage common;\n\noption go_package = \"datakittest/grpc/pb\";\n\nmessage result {\n    int32 code = 1;\n    string msg = 2;\n}"
            },
            "full_method": "greeter.Greeter/SayHello",
            "request": "{\"name\": \"world\"}"
          }
        },
        "certificate": {
          "ignore_server_certificate_error": false
        }
      },
      "post_script": "..."
    }
  ]
}
```

##### `success_when` 定义 {#grpc-success-when}

- gRPC 响应体判断 (`body`)

`body` 为一个数组对象，每个对象参数如下：

| 字段              | 类型   | 是否必须 | 说明                                                |
| :---              | ---    | ---      | ---                                                 |
| `is`              | string | N        | 返回的 body 是否等于该指定字段                      |
| `is_not`          | string | N        | 返回的 body 是否不等于该指定字段                    |
| `match_regex`     | string | N        | 返回的 body 是否含有该匹配正则表达式的子字符串      |
| `not_match_regex` | string | N        | 返回的 body 是否不含有该匹配正则表达式的子字符串    |
| `contains`        | string | N        | 返回的 body 是否含有该指定的子字符串                |
| `not_contains`    | string | N        | 返回的 body 是否不含有该指定的子字符串              |

如：

```json
"success_when": [
  {
    "body": [
      {
        "contains": "success"
      }
    ]
  }
]
```

- gRPC 响应时间判断 (`response_time`)

填写具体的时间值，如果请求的响应时间小于该指定值，则判定拨测成功，如：

```json
"success_when": [
  {
    "response_time": "1000ms"
  }
]
```

> 注意，此处指定的时间单位有 `ns`（纳秒）/`us`（微秒）/`ms`（毫秒）/`s`（秒）/`m`（分钟）/`h`（小时）。对 gRPC 拨测而言，一般使用 `ms` 或 `s` 单位。

##### `advance_options` 定义 {#grpc-advance-options}

- 请求选项 (`request_options`)

| 字段              | 类型              | 是否必须 | 说明                                                       |
| :---              | ---               | ---      | ---                                                        |
| `request_timeout` | string            | N        | 请求超时时间，默认 30s                                      |
| `metadata`        | map[string]string | N        | gRPC 请求的元数据（metadata）                               |
| `proto_files`     | object            | N        | 通过 proto 文件发现方法（与 `reflection`、`health_check` 三选一） |
| `reflection`      | object            | N        | 通过 gRPC 反射发现方法（与 `proto_files`、`health_check` 三选一） |
| `health_check`    | object            | N        | 使用 gRPC 健康检查服务（与 `proto_files`、`reflection` 三选一） |

`proto_files` 对象定义：

| 字段          | 类型                | 是否必须 | 说明                                    |
| :---          | ---                 | ---      | ---                                     |
| `protofiles`  | map[string]string   | Y        | proto 文件内容，key 为文件引用路径（主文件为文件名，被 import 的文件必须与 import 语句中的路径一致），value 为文件内容 |
| `full_method` | string              | Y        | 完整方法名，格式为 `ServiceName/MethodName` |
| `request`     | string              | N        | JSON 格式的请求体                        |

`reflection` 对象定义：

| 字段          | 类型   | 是否必须 | 说明                                    |
| :---          | ---    | ---      | ---                                     |
| `full_method` | string | Y        | 完整方法名，格式为 `ServiceName/MethodName` |
| `request`     | string | N        | JSON 格式的请求体                        |

`health_check` 对象定义：

| 字段      | 类型   | 是否必须 | 说明                    |
| :---      | ---    | ---      | ---                     |
| `service` | string | N        | 要检查的服务名称，为空时表示检查整个服务        |

`request_options` 示例：

```json
"advance_options": {
  "request_options": {
    "request_timeout": "10s",
    "metadata": {
      "x-token": "test-token",
    },
    "proto_files": {
      "protofiles": {
        "greeter.proto": "syntax = \"proto3\";\n\npackage greeter;\n\noption go_package = \"datakittest/grpc/pb\";\n\nimport \"pb/common.proto\";\n\nservice Greeter {\n  rpc SayHello (HelloRequest) returns (common.result) {}\n}\n\nmessage HelloRequest {\n  string name = 1;\n}",
        "pb/common.proto": "syntax = \"proto3\";\n\npackage common;\n\noption go_package = \"datakittest/grpc/pb\";\n\nmessage result {\n    int32 code = 1;\n    string msg = 2;\n}"
      },
      "full_method": "greeter.Greeter/SayHello",
      "request": "{\"name\": \"world\"}"
    }
  }
}
```

或使用反射：

> 注意：使用反射方式需要 gRPC 服务器端启用 [gRPC Server Reflection](https://github.com/grpc/grpc/blob/master/doc/server-reflection.md){:target="_blank"} 服务。

```json
"advance_options": {
  "request_options": {
    "reflection": {
      "full_method": "greeter.Greeter/SayHello",
      "request": "{\"name\": \"world\"}"
    }
  }
}
```

或使用健康检查：

> 注意：使用健康检查方式需要 gRPC 服务器端实现 [gRPC Health Checking Protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md){:target="_blank"} 服务。

```json
"advance_options": {
  "request_options": {
    "health_check": {
      "service": "my-service"
    }
  }
}
```

- 证书配置 (`certificate`)

| 字段                              | 类型   | 是否必须 | 说明                                    |
| :---                              | ---    | ---      | ---                                     |
| `ignore_server_certificate_error` | bool   | N        | 是否跳过服务器证书验证（不验证服务器证书的有效性）                  |
| `private_key`                     | string | N        | 客户端私钥（用于 mTLS）                 |
| `certificate`                     | string | N        | 客户端证书（用于 mTLS）                 |
| `ca`                              | string | N        | CA 证书（用于验证服务器证书）           |

`certificate` 示例：

```json
"advance_options": {
  "certificate": {
    "ignore_server_certificate_error": false,
    "ca": "<your-ca-cert>",
    "private_key": "<your-private-key>",
    "certificate": "<your-certificate>"
  }
}
```

- 安全选项 (`secret`)

| 字段                  | 类型 | 是否必须 | 说明                     |
| :---                  | ---  | ---      | ---                      |
| `not_save`            | bool | N        | 是否不保存响应体内容     |

`secret` 示例：

```json
"advance_options": {
  "secret": {
    "not_save": true
  }
}
```

##### `post_script` 定义 {#grpc-post-script}

`post_script` 是一个 [Pipeline](../pipeline/use-pipeline/index.md) 脚本，用于对拨测结果进行判断。

注入变量

为了便于 `post_script` 处理 gRPC 响应并对检测结果进行判定，编写脚本时可使用一些预定义的变量。具体如下：

- `response`: 响应对象

| 字段      | 类型   | 说明         |
| :---      | ---    | ---          |
| `body`    | string | 响应内容（JSON 格式字符串） |

- `result`： 检测结果

| 字段           | 类型   | 说明     |
| :---           | ---    | ---      |
| `is_failed`    | bool   | 是否失败 |
| `error_message` | string | 失败原因 |

示例

```javascript

body = load_json(response["body"])

if body["message"] == "Hello world" {
  result["is_failed"] = false
} else {
  result["is_failed"] = true
  result["error_message"] = "Unexpected response"
}

```

上面的脚本中，首先使用 `load_json` 将响应内容（`response["body"]`）解析为 JSON 对象，然后判断响应中的 message 字段是否为 "Hello world"，如果是，则将 `result` 的 `is_failed` 设置为 false，否则将 `result` 的 `is_failed` 设置为 true，且 `error_message` 设置为错误信息。

#### 多步拨测 {#multi}

[:octicons-tag-24: Version-1.68.0](../datakit/changelog-2025.md#cl-1.68.0)

多步拨测允许您定义一系列按顺序执行的拨测步骤，支持 HTTP 请求和等待操作，并可以在步骤之间传递变量。

额外字段

| 字段        | 类型   | 是否必须 | 说明                                      |
| :---        | ---    | ---      | ---                                       |
| `steps`     | array  | Y        | 拨测步骤列表，至少包含一个步骤            |

总体的 JSON 结构如下：

``` json
{
  "MULTI": [
    {
      "name": "multi-step-test",
      "status": "OK",
      "post_url": "https://<your-dataway-host>?token=<your-token>",
      "frequency": "10s",
      "steps": [
        {
          "type": "http",
          "name": "step1",
          "task": "{\"name\": \"step1\", \"method\": \"GET\", \"url\": \"http://api.example.com/resource\", \"post_script\": \"vars[\\\"token\\\"] = \\\"token_value\\\"\"}",
          "allow_failure": false,
          "retry": {
            "retry": 3,
            "interval": 1000
          },
          "extracted_vars": [
            {
              "name": "token",
              "field": "token",
              "secure": false
            }
          ]
        },
        {
          "type": "wait",
          "name": "step2",
          "value": 5
        },
        {
          "type": "http",
          "name": "step3",
          "task": "{\"name\": \"step3\", \"method\": \"POST\", \"url\": \"http://127.0.0.1:9000?token={{`{{token}}`}}\", \"post_script\":\"result[\\\"is_failed\\\"]=true\"}",
          "allow_failure": false
        }
      ]
    }
  ]
}
```

##### `steps` 定义 {#multi-steps}

`steps` 是一个拨测步骤数组，每个步骤可以是以下两种类型之一：

- HTTP 步骤 (`type: "http"`)

   | 字段              | 类型   | 是否必须 | 说明                                      |
   | :---              | ---    | ---      | ---                                       |
   | `type`            | string | Y        | 步骤类型，必须为 `http`                   |
   | `name`            | string | Y        | 步骤名称                                  |
   | `task`            | string | Y        | HTTP 拨测任务的 JSON 字符串               |
   | `allow_failure`   | bool   | N        | 是否允许步骤失败，默认 `false`            |
   | `retry`           | object | N        | 重试配置，详见下文                        |
   | `extracted_vars`  | array  | N        | 提取的变量列表，用于后续步骤      |

- 等待步骤 (`type: "wait"`)

   | 字段              | 类型   | 是否必须 | 说明                                      |
   | :---              | ---    | ---      | ---                                       |
   | `type`            | string | Y        | 步骤类型，必须为 `wait`                   |
   | `name`            | string | Y        | 步骤名称                                  |
   | `value`           | int    | Y        | 等待时间，单位为秒                        |
   | `allow_failure`   | bool   | N        | 是否允许步骤失败，默认 `false`            |

##### `retry` 定义 {#multi-retry}

重试配置用于设置当步骤失败时的重试策略。

| 字段              | 类型   | 是否必须 | 说明                                      |
| :---              | ---    | ---      | ---                                       |
| `retry`           | int    | Y        | 重试次数，必须在 0-5 之间                 |
| `interval`        | int    | Y        | 重试间隔，单位为毫秒，必须在 0-5000 之间  |

##### `extracted_vars` 定义 {#multi-extracted-vars}

从 HTTP `post_script` 中定义的 `vars` 中提取变量，用于后续步骤的模板替换。

| 字段              | 类型   | 是否必须 | 说明                                      |
| :---              | ---    | ---      | ---                                       |
| `name`            | string | Y        | 变量名称，用于后续步骤的模板替换          |
| `field`           | string | Y        | `vars` 中定义的变量名    |
| `secure`          | bool   | N        | 是否为安全变量，安全变量不会在结果中显示  |

### 模板函数使用说明 {#template-func}

[:octicons-tag-24: Version-1.80.0](../datakit/changelog.md#cl-1.80.0)

在定义拨测任务时，允许在字符串中嵌入模板变量。

#### 函数列表 {#template-func-list}

目前支持以下模板函数：

- `timestamp` 函数：生成时间戳

  返回当前 UTC 时间的时间戳，支持多种时间单位（`s`、`ms`、`us`、`ns`），默认返回 `0`（当单位参数无效时）。

  使用示例：

  ```go
  {{"{{ timestamp \"ns\" }}"}} //1754379375400790109
  ```

- `date` 函数：格式化时间字符串

  根据指定格式返回当前本地时间的字符串，支持两种定义格式（`rfc3339`、`iso8601`）。

  使用示例：

  ```go
  {{"{{ date \"rfc3339\" }}"}} // 2006-01-02T15:04:05Z07:00
  {{"{{ date \"iso8601\" }}"}} // 2006-01-02T15:04:05Z
  ```

- `urlencode` 函数：URL 编码

  对输入的字符串进行 URL 编码（符合 `application/x-www-form-urlencoded` 规范），用于处理 URL 中的特殊字符。

  使用示例：

  ```go
  {{"{{ urlencode \"hello world\" }}"}} // hello+world
  ```
  
#### 函数嵌套使用说明 {#template-func-description}

模板函数支持嵌套使用，即一个函数的返回值可以作为另一个函数的参数。

使用示例：

```go
{{"{{ urlencode (date \"iso8601\") }}"}} // 2006-01-02T15%3A04%3A05Z
```
