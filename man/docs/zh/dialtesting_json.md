# 通过本地 JSON 定义拨测任务
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

    目前可以通过 [ConfigMap 方式注入采集器配置](datakit-daemonset-deploy.md#configmap-setting)来开启采集器。
<!-- markdownlint-enable -->

---

具体的国家/地域以及 ISP 选择，可按照下图所示方式来选择（注意，不要真的新建「自建节点」，此处只是提供一个可供选择的来源）：

<figure markdown>
![](https://static.guance.com/images/datakit/dialtesting-select-country-city-isp.png){ width="800" }
</figure>

### 配置拨测任务 {#config-task}

目前拨测任务支持四种拨测类型，即 HTTP, TCP, ICMP, WEBSOCKET 服务，JSON 格式如下：

```json
{
  "<拨测类型>": [
    {拨测任务1},
    {拨测任务2},
       ...
    {拨测任务n},
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
            {"response_time": {
                "is_contain_dns": true,
                "target": "10ms"
            }}
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
| `status`             | string | Y        | 拨测服务状态，如 "OK"/"stop"                                                         |
| `frequency`          | string | Y        | 拨测频率                                                                             |
| `success_when_logic` | string | N        | success_when条件之间的逻辑关系，如"and"/"or",默认为"and"                             |
| `success_when`       | object | Y        | 详见下文                                                                             |
| `advance_options`    | object | N        | 详见下文                                                                             |
| `post_url`           | string | N        | 将拨测结果发往该 Token 所指向的工作空间，如果不填写，则发给当前 DataKit 所在工作空间 |

#### HTTP 拨测 {#http}

额外字段

| 字段              | 类型   | 是否必须 | 说明                                    |
| :---              | ---    | ---      | ---                                     |
| `method`          | string | Y        | HTTP 请求方法                           |
| `url`             | string | Y        | 完整的 HTTP 请求地址                    |

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
      "advance_options": ...
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

此处 `body` 可以配置多个验证规则，由"success_when_logic"确定他们之间的关系，配置为`and`时，**任何一个规则验证不过，则认为当前拨测失败**；配置为`or`时，**任何一个规则验证通过，则认为当前拨测成功**；默认是 `and` 的关系。下面的验证规则，均遵循这一判定原则。

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

以上列举的几种判定依据，可以组合使用，由"success_when_logic"确定他们之间的关系，配置为`and`时，**任何一个规则验证不过，则认为当前拨测失败**；配置为`or`时，**任何一个规则验证通过，则认为当前拨测成功**；默认是 `and` 的关系。如：

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

`request_body` 示例：

```json
"advance_options": {
  "request_body": {
    "body_type": "text/html",
    "body": "填写好请求体，此处注意各种复杂的转义"
  }
}
```

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

#### TCP 拨测 {#tcp}

##### 额外字段 {#tcp-extra}

| 字段              | 类型   | 是否必须 | 说明                                    |
| :---              | ---    | ---      | ---                                     |
| `host`          | string | Y        | TCP 主机地址                           |
| `port`             | string | Y        | TCP 端口                    |
| `timeout`             | string | N        | TCP 连接超时时间                    |

完整 JSON 结构如下:

``` json
{
    "TCP": [
        {
            "name": "tcp-test",
            "host": "www.baidu.com",
            "port": "80",
            "timeout": "10ms",
            "enable_traceroute": true,
            "post_url": "https://<your-dataway-host>?token=<your-token>",
            "status": "OK",
            "frequency": "10s",
            "success_when_logic": "and",
            "success_when": [
                {
                    "response_time":[ 
                        {
                            "is_contain_dns": true,
                            "target": "10ms"
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

完整 JSON 结构如下:

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

完整 JSON 结构如下:

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

`header`为一个字典类型对象，其每个对象元素的值为为一个数组对象，相应参数如下：

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
