{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# DataKit API 文档

本文档主要描述 DataKit 开放出来 HTTP API 接口。

## `/v1/write/:category`


写入日志数据，参数列表：

| 参数名               | 类型   | 是否必选 | 默认值    | 说明                                    |
| -----                | ----   | -------  | ----      | -----                                   |
| `category`           | string | true     | 无        | 目前只支持 `metric/logging/rum`         |
| `precision`          | string | false    | `n`       | 数据精度(支持 `n/u/ms/s/m/h`)           |
| `input`              | string | false    | `datakit` | 数据源名称                              |
| `ignore_global_tags` | string | false    | 无        | 任意给值即认为忽略 DataKit 上的全局 tag |

HTTP body 为行协议。

### 日志(logging)示例

```http
POST /v1/write/logging?precision=n&input=my-sample-logger&ignore_global_tags=123 HTTP/1.1

nginx,tag1=a,tag2=b f1=1i,f2=1.2,f3="abc" 1620723870000000000
mysql,tag1=a,tag2=b f1=1i,f2=1.2,f3="abc" 1620723870000000000
redis,tag1=a,tag2=b f1=1i,f2=1.2,f3="abc" 1620723870000000000
```

注意：行协议中的 measurement-name 会作为日志的 `source` 字段来存储。

### 时序数据(metric)示例

```http
POST /v1/write/metric?precision=n&input=my-sample-logger&ignore_global_tags=123 HTTP/1.1

cpu,tag1=a,tag2=b f1=1i,f2=1.2,f3="abc" 1620723870000000000
mem,tag1=a,tag2=b f1=1i,f2=1.2,f3="abc" 1620723870000000000
net,tag1=a,tag2=b f1=1i,f2=1.2,f3="abc" 1620723870000000000
```

### RUM 数据示例

```http
POST /v1/write/rum?precision=n&input=my-sample-rum&ignore_global_tags=true HTTP/1.1

error,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
resource,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
view,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
long_task,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
action,t1=tag1,t2=tag2 f1=1.0,f2=2i,f3="abc"
```

目前 RUM 接口只接收如下几种指标集名称：`error/resource/view/long_task/action`

> 注意：RUM 请求中，如果不指定 `input` 参数，默认用 `rum` 来命名。

## `/v1/ping`

检测目标地址是否有 DataKit 运行

### 示例

```http
GET /v1/ping HTTP/1.1

HTTP/1.1 200 OK

{
	"content":{
		"version":"1.1.6-rc0",
		"uptime":"1.022205003s"
	}
}
```

## `/v1/query/raw`

使用 DQL 进行数据查询（只能查询该 DataKit 所在的工作空间的数据）

### 示例

```http
POST /v1/query/raw HTTP/1.1
Content-Type: application/json

{
    "queries":[
        {
            "query": "cpu:(cpu_usage_idle) LIMIT 1"  # dql查询语句（必填）
        },
        {
            "query": "mem:(mem_usage_percent) LIMIT 1"
        }
    ],
    "echo_explain": false
}

HTTP/1.1 200 OK
Content-Type: application/json

{
    "content": [
        {
            "series": [
                {
                    "name": "cpu",
                    "columns": [
                        "time",
                        "cpu_usage_idle"
                    ],
                    "values": [
                        [
                            1608612960000,
                            99.59595959596913
                        ]
                    ]
                }
            ],
            "cost": "25.093363ms",
        },
        {
            "series": [
                {
                    "name": "mem",
                    "columns": [
                        "time",
                        "mem_usage_percent"
                    ],
                    "values": [
                        [
                            1608612960000,
                            99.59595959596913
                        ]
                    ]
                }
            ],
            "cost": "32.093363ms",
    ]
}
```

## `/v1/host/meta`

待补充
