# Defining Dialing Test Tasks Through Local JSON
---

In some cases, you may not be able to connect to SAAS's dialing task service. In this case, we can define the dialing task through the local json file.

## Configuration {#config}

### Configure Collector {#config-inputs}

=== "Host Installation"

    Go to the `conf.d/network` directory under the DataKit installation directory, copy `dialtesting.conf.sample` and name it `dialtesting.conf`. Examples are as follows:
    
    ```toml
    [[inputs.dialtesting]]
      server = "file://</path/to/your/local.json>"
    
      # Note: Taking Linux as an example, assuming your json directory is /some/path/my.json, then the
      # server should be written as file:///some/path/my.json
    
      # Note that the following tag suggestions are filled in one by one (do not modify the tag key here), so that the complete dialing test results can be displayed on the page.
      [inputs.dialtesting.tags] 
        country  = "<specify-datakit-country>"  # Countries where DataKit is deployed
        province = "<specify-datakit-province>" # Provices where DataKit is deployed
        city     = "<specify-datakit-city>"     # Cities where DataKit is deployed
        isp      = "<specify-datakit-ISP>"      # Specify the network service provider where DataKit is located
        region   = "<your-region>"              # You can specify a region name at will
    ```

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap injection collector configuration](datakit-daemonset-deploy.md#configmap-setting).

---

The specific country/region and ISP selection can be selected as shown in the following figure (note that you don't really create a new "self-built node", just provide an alternative source here):

<figure markdown>
![](https://static.guance.com/images/datakit/dialtesting-select-country-city-isp.png){ width="800" }
</figure>

### Configure the Dial Test Task {#config-task}

At present, the dialing test task supports four dialing test types, namely HTTP, TCP, ICMP and WEBSOCKET services. The JSON format is as follows:

```json
{
  "<Dial Test Type>": [
    {Dial test task 1},
    {Dial test task 2},
       ...
    {Dial test task n},
  ]
}
```

The following is a specific dialing test example:

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

>  After editing this JSON, it is recommended to find some（[online tools](https://www.json.cn/){:target="_blank"} or [this tool](https://jsonformatter.curiousconcept.com/#){:target="_blank"}）to verify that the JSON format is correct. If the JSON format is incorrect, the dialing test will not take effect.

After configuration, restart DataKit.

### Test Task Field Definition {#field-def}

The dialing task fields include "public fields" and "additional fields" for specific dialing tasks.

#### Public Field {#pub}

The public fields of dialing test tasks are defined as follows:

| Field                 | Type   | Whether Required | Description                                                                                 |
| :---                 | ---    | ---      | ---                                                                                  |
| `name`               | string | Y        | Dial test service name                                                                         |
| `status`             | string | Y        | Dial test service status, such as "OK"/"stop"                                                         |
| `frequency`          | string | Y        | Dial frequency                                                                             |
| `success_when_logic` | string | N        | The logical relationship between success_when conditions, such as "and"/"or", defaults to "and"                             |
| `success_when`       | object | Y        | See below for details                                                                             |
| `advance_options`    | object | N        | See below for details                                                                             |
| `post_url`           | string | N        | Send the dialing test result to the workspace pointed by the Token, and if it is not filled in, send it to the workspace where the current DataKit is located |

#### HTTP Dial Test {#http}

**Extra field**

| Field              | Type   | Whether Required | Description                                    |
| :---              | ---    | ---      | ---                                     |
| `method`          | string | Y        | HTTP request method                           |
| `url`             | string | Y        | Complete HTTP request address                   |


The overall JSON structure is as follows:

```
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

##### `success_when` Definition {#http-success-when}

The judging conditions used to define the success of dialing test mainly include the following aspects:

- HTTP request returns body judgment（`body`）

| Field              | Type   | Whether Required | Description                                      |
| :---              | ---    | ---      | ---                                       |
| `is`              | string | N        | Whether the returned body is equal to the specified field                  |
| `is_not`          | string | N        | Whether the returned body is not equal to the specified field                 |
| `match_regex`     | string | N        | Whether the returned body contains a substring of the matching regular expression   |
| `not_match_regex` | string | N        | Whether the returned body does not contain a substring of the matching regular expression|
| `contains`        | string | N        | Whether the returned body contains the specified substring             |
| `not_contains`    | string | N        | Whether the returned body does not contain the specified substring           |

eg.

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

Here, `body` can configure multiple verification rules, and the relationship between them is determined by "success_when_logic". When it is configured as `and`, **if any rule is verified, it will be considered that the current dialing test failed**; When it is configured to `or`, **if any rule is verified, it will be considered that the current dialing test is successful**. The default is an `and` relationship. The following verification rules all follow this judgment principle.

> Note that the regular is escaped correctly here, and the actual regular expression in the example is `\d\d.*`.

- HTTP request returns header judgment (`header`)

| Field              | Type   | Whether Required | Description                                                       |
| :---              | ---    | ---      | ---                                                        |
| `is`              | string | N        | The header returned specifies whether the field is equal to the specified value                     |
| `is_not`          | string | N        | The header returned specifies whether the field is not equal to the specified value.                   |
| `match_regex`     | string | N        | The header returned specifies whether the field contains a substring of the matching regular expression.   |
| `not_match_regex` | string | N        | The header returned specifies whether the field does not contain a substring of the matching regular expression. |
| `contains`        | string | N        | The header returned specifies whether the field contains the specified substring.             |
| `not_contains`    | string | N        | The header returned specifies whether the field does not contain the specified substring.           |

for example:

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

Because there may be decisions for multiple types of headers, validation for multiple headers can also be configured here:

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

- HTTP request returns status code (`status_code`)

| Field              | Type   | Whether Required | Description                                             |
| :---              | ---    | ---      | ---                                              |
| `is`              | string | N        | Whether the status code returned is equal to the specified field                   |
| `is_not`          | string | N        | Whether the status code returned is not equal to the specified field                 |
| `match_regex`     | string | N        | Whether the status code returned contains a substring of the matching regular expression   |
| `not_match_regex` | string | N        | Whether the status code returned does not contain a substring of the matching regular expression |
| `contains`        | string | N        | Whether the status code returned contains the specified substring             |
| `not_contains`    | string | N        | Whether the status code returned does not contain the specified substring           |

for example:

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

> For a certain URL dial test, its HTTP return is usually only one, so only one validation rule is generally configured here (although multiple array configurations are supported).

- HTTP request response time (`response_time`)

Only one time value can be filled in here. If the response time of the request is less than the specified value, the dialing test is judged to be successful, such as:

```json
"success_when": [
  {
    "response_time": "100ms"
  }
]
```

> Note that the time units specified here are `ns` (nanoseconds)/`us` (microseconds) /`ms` (milliseconds) /`s` (seconds) /`m` (minutes) /`h` (hours). For HTTP dial testing, `ms` units are generally used.

Several kinds of judgment basis listed above can be used in combination, and the relationship between them is determined by "success_when_logic". When it is configured as `and`, **if any rule is verified, it is considered that the current dialing test fails**; When it is configured to `or`, **if any rule is verified, it will be considered that the current dialing test is successful**; The default is an `and` relationship. Such as:

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

##### `advance_options` Definition {#http-advance-options}

Advanced options are mainly used to adjust specific dialing behavior, mainly in the following aspects:

- HTTP Request Option (`request_options`）

| Field              | Type              | Whether Required | Description                       |
| :---              | ---               | ---      | ---                        |
| `follow_redirect` | bool              | N        | Whether redirect jump is supported         |
| `headers`         | map[string]string | N        | Specify a set of headers on an HTTP request |
| `cookies`         | string            | N        | Specify the requested Cookie          |
| `auth`            | object            | N        | Specify the authentication method of the request         |

Among them, `auth` only supports ordinary username and password authentication, which is defined as follows:

| Field       | Type   | Whether Required | Description       |
| :---       | ---    | ---      | ---        |
| `username` | string | Y        | User name     |
| `password` | string | Y        | User name and password |

`request_options` example:

```json
"advance_options": {
  "request_options": {
    "auth": {
        "username": "zhangsan",
        "password": "fawaikuangtu"
      },
    "headers": {
      "X-Prison-Breaker": "zhangsan",
      "X-Prison-Break-Password": "fawaikuangtu"
    },
    "follow_redirect": false
  },
}
```

- HTTP Request Body（`request_body`）

| Field        | Type   | Whether Required | Description                                    |
| :---        | ---    | ---      | ---                                     |
| `body_type` | string | N        | Body type, that is, the value of the request header `Content-Type` |
| `body`      | string | N        | Request Body                               |

`request_body` example:

```json
"advance_options": {
  "request_body": {
    "body_type": "text/html",
    "body": "Fill in the request body, and pay attention to various complicated escapes here"
  }
}
```

- HTTP Request a Certificate (`certificate`)

| Field                              | Type   | Whether Required | Description             |
| :---                              | ---    | ---      | ---              |
| `ignore_server_certificate_error` | bool   | N        | Whether to ignore certificate errors |
| `private_key`                     | string | N        | key              |
| `certificate`                     | string | N        | Certificate             |
| `ca`                              | string | N        | Temporarily unused       |

`certificate` example:

```json
"advance_options": {
  "certificate": {
    "ignore_server_certificate_error": false,
    "private_key": "<your-private-key>",
    "certificate": "<your-certificate-key>"
  },
}
```

`private_key` example:

```
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

Here is an example of `certificate`:

```
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

Under Linux, this pair of keys can be generated by the following command:

```shell
openssl req -newkey rsa:2048 -x509 -sha256 -days 3650 -nodes -out example.crt -keyout example.key
```

- HTTP Request broker (`proxy`)

| Field      | Type              | Whether Required | Description                                 |
| :---      | ---               | ---      | ---                                  |
| `url`     | string            | N        | 代理的 The URL of the proxy, such as `http://1.2.3.4:4321` |
| `headers` | map[string]string | N        | Specify a set of headers on an HTTP request           |

`proxy` example:

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

#### TCP Dial Test {#tcp}

##### Extra Field {#tcp-extra}

| Field              | Type   | Whether Required | Description                                    |
| :---              | ---    | ---      | ---                                     |
| `host`          | string | Y        | TCP Host address                           |
| `port`             | string | Y        | TCP Port                    |
| `timeout`             | string | N        | TCP connection timeout                    |

The complete JSON structure is as follows:

```
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

##### `success_when` Definition {#tcp-success-when}

- TCP Response Time Determination (`response_time`)

`response_time` is an array object with the following parameters for each object:

| Field              | Type   | Whether Required | Description                                                       |
| :---              | ---    | ---      | ---                                                        |
| `target`          | string | Y        | Determining whether the response time is less than the value                     |
| `is_contain_dns`  | bool | N        | Indicates whether the response time includes DNS resolution time                     |


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

- Network hop count (`hops`)

`hops` is an array object with the following parameters for each object:

| Field              | Type   | Whether Required | Description                                      |
| :---              | ---    | ---      | ---                                       |
| `op`              | string | Y        | Compare relation, retrievable `eq(=),lt(<),leq(<=),gt(>),geq(>=)`|
| `target`          | float | Y        | Decision value                 |


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


#### ICMP Dial Test {#icmp}

##### Extra Field {#icmp-extra}

| Field              | Type   | Whether Required | Description                                    |
| :---              | ---    | ---      | ---                                     |
| `host`            | string | Y        | Host address                           |
| `packet_count`    | int |   N         | Number of ICMP packets sent  
| `timeout`             | string | N    | Connection timeout

The complete JSON structure is as follows:

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

##### `success_when` Definition {#icmp-success-when}

- ICMP packet loss rate (`packet_loss_percent`)

Fill in the specific value as an array object, and each object parameter is as follows:

| Field              | Type   | Whether Required | Description                                      |
| :---              | ---    | ---      | ---                                       |
| `op`              | string | Y        | Compare relationship retrievable `eq(=),lt(<),leq(<=),gt(>),geq(>=)`|
| `target`          | float | Y        | Decision value                 |

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

- ICMP response time (`response_time`)

Fill in the specific time as an array object, and each object parameter is as follows:

| Field              | Type   | Whether Required | Description                                      |
| :---              | ---    | ---      | ---                                       |
| `func`            | string | Y        | Statistical type, take the value `avg,min,max,std`|
| `op`              | string | Y        | Comparison relation, take value `eq(=),lt(<),leq(<=),gt(>),geq(>=)`|
| `target`          | string | Y        | Decision value                 |

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

- Network hop count (`hops`)

`hops` is an array object with the following parameters for each object:

| Field              | Type   | Whether Required | Description                                      |
| :---              | ---    | ---      | ---                                       |
| `op`              | string | Y        | Compare relationships, retrievable `eq(=),lt(<),leq(<=),gt(>),geq(>=)`|
| `target`          | float | Y        | Decision value                 |


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

- Number of packets grabbed (`packets`)

`packets` is an array object with the following parameters for each object:

| Field              | Type   | Whether Required | Description                                      |
| :---              | ---    | ---      | ---                                       |
| `op`              | string | Y        | Compare relationship retrievable `eq(=),lt(<),leq(<=),gt(>),geq(>=)`|
| `target`          | float | Y        | Decision value                 |


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

#### WEBSOCKET Dial Test {#ws}

##### Extra Field {#ws-extra}

| Field              | Type   | Whether Required | Description                                    |
| :---              | ---    | ---      | ---                                     |
| `url`          | string | Y        | Websocket connection address, such as ws://localhost:8080  |
| `message`       | string | Y        | Websocket message sent after successful connection                |

The complete JSON structure is as follows:

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

##### `success_when` Definition {#ws-success-when}

- Response time judgment (`response_time`)

`response_time` is an array object with the following parameters for each object:

| Field             | Type   | Whether Required | Description                              |
| :---             | ---    | ---      | ---                               |
| `target`         | string | Y        | Determining whether the response time is less than the value          |
| `is_contain_dns` | bool   | N        | Indicates whether the response time includes DNS resolution time |


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

- Return a message decision (`response_message`）

`response_message` is an array object with the following parameters for each object:

| Field              | Type   | Whether Required | Description                                                |
| :---              | ---    | ---      | ---                                                 |
| `is`              | string | N        | Whether the returned message is equal to the specified field                   |
| `is_not`          | string | N        | Whether the returned message is not equal to the specified field                 |
| `match_regex`     | string | N        | Whether the returned message contains a substring of the matching regular expression   |
| `not_match_regex` | string | N        | Whether the returned message does not contain a substring of the matching regular expression |
| `contains`        | string | N        | Whether the returned message contains the specified substring             |
| `not_contains`    | string | N        | Whether the returned message does not contain the specified substring           |

for example:

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

- Request to return header judgment（`header`）

`header` is a dictionary type object whose value for each object element is an array object with the following parameters:

| Field              | Type   | Whether Required | Description                                                       |
| :---              | ---    | ---      | ---                                                        |
| `is`              | string | N        | The header returned specifies whether the field is equal to the specified value                     |
| `is_not`          | string | N        | The header returned specifies whether the field is not equal to the specified value                   |
| `match_regex`     | string | N        | The header returned specifies whether the field contains a substring of the matching regular expression  |
| `not_match_regex` | string | N        | The header returned specifies whether the field does not contain a substring of the matching regular expression |
| `contains`        | string | N        | The header returned specifies whether the field contains the specified substring             |
| `not_contains`    | string | N        | The header returned specifies whether the field does not contain the specified substring           |

for example:

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

##### `advance_options` Definition {#ws-advance-options}

- Request option (`request_options`)

| Field              | Type              | Whether Required | Description                       |
| :---              | ---               | ---      | ---                        |
| `timeout` | string              | N        | Connection timeout         |
| `headers` | map[string]string | N        |  Specify a set of headers on request |

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

- Authentication information (`auth`)

Support for common user name and password authentication (Basic access authentication)。

| Field       | Type   | Whether Required | Description       |
| :---       | ---    | ---      | ---        |
| `username` | string | Y        | user name     |
| `password` | string | Y        | user name and password |

```json
"advance_options": {
  "auth": {
    "username": "admin",
    "password": "123456"
  },
}
```
