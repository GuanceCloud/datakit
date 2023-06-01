
# DataKit API

---

This document mainly describes the HTTP API interface opened by DataKit.

## API Summary {#intro}

At present, DataKit only supports HTTP interface, which mainly involves data writing and data query.

### Get the DataKit Version {#api-get-dk-version}

There are two ways to get the version number:

- Request DataKit ping interface: `curl http://ip:9529/v1/ping`
- In the return Header of each of the following API requests, you can know the DataKit version of the current request through `X-DataKit`.

## `/v1/write/:category` {#api-v1-write}

This API is used to report various `category` of data to DataKit, and the parameters are described as follows:

| Parameter                 | Type   | Required or not | Default Value | Description                                                                                                                                                                               |
| ---                       | ---    | ---             | ---           | ---                                                                                                                                                                                       |
| `category`                | string | Y               | no            | support `metric/logging/rum/object/custom_object/keyevent`. For example, the URL of `metric` is `/v1/write/metric`                                                                        |
| `echo_line_proto`         | string | N               | no            | Giving any value (such as `true`) returns line protocol format point data. default not echoed.                                                                                            |
| `echo_json`               | string | N               | -             | Giving any value (such as `true`) returns JSON format point data. default not echoed. If both echo enabled, preferred line protocol                                                       |
| `global_election_tags`    | string | N               | -             | Giving any value (such as `true`) to append global-election tags（[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)）                                                              |
| `ignore_global_host_tags` | string | false           | no            | Giving any value (such as `true`) is considered to ignore the global tag on DataKit（[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)）。`ignore_global_tags` would be abandoned. |
| `input`                   | string | N               | `datakit`     | Data source name                                                                                                                                                                          |
| `loose`                   | bool   | N               | true          | Loose mode, for some invalid POST(json or lineprotocol), DataKit would try to auto-fix them ([:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9)).                                   |
| `strict`                  | bool   | N               | false         | Strict mode, for some invalid POST(json or lineprotocol), DataKit would reject them and showing why([:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9)).                            |
| `precision`               | string | N               | `n`           | Data accuracy (supporting `n/u/ms/s/m/h`)                                                                                                                                                 |
| `source`                  | string | N               | no            | Specify this field only for logging support (that is, `category` is `logging`). If you do not specify `source`, the uploaded log data would not be cut by Pipeline.                       |
| `version`                 | string | N               | no            | The version number of the current collector                                                                                                                                               |

HTTP body supports both line protocol and JSON. See [here](apis.md#lineproto-limitation) for constraints on data structures, whether in line protocol or JSON form.

### Categories {#category}

In DataKit, the main data types are as follows (listed in alphabetical order according to their abbreviations):

| Abbreviations | Name          | URL form                | Description            |
| ----          | ----          | ----                    | ---                    |
| CO            | custom_object | /v1/write/custom_object | customized object data |
| E             | keyevent      | /v1/write/keyevent      | Event data             |
| L             | logging       | /v1/write/logging       | Logging data           |
| M             | metric        | /v1/write/metric        | Time series            |
| N             | network       | /v1/write/network       | eBPF data              |
| O             | object        | /v1/write/object        | object data            |
| P             | profiling     | /v1/write/profiling     | Profiling data         |
| R             | rum           | /v1/write/rum           | RUM data               |
| S             | security      | /v1/write/security      | Security check data    |
| T             | tracing       | /v1/write/tracing       | APM(tracing) data      |

不同的数据类型，其处理方式不一样，在观测云的用法也不尽相同。在 Datait 的配置和使用过程中，有时候会穿插使用某个类型的不同形式（比如在 sinker 配置中用简写，在 API 请求中则用其 URL 表示）

### JSON Body Examples {#api-json-example}

To facilitate line protocol processing, all data upload APIs support body in JSON form.  JSON body parameter descriptions are as follows.

| Parameter        | Type                        | Required or not | Default Value | Description                                                                          |
| ------------- | --------------------------- | -------- | ------ | ----------------------------------------------------------------------------- |
| `measurement` | `string`                    | Yes       | None     | Measurement name                                                            |
| `tags`        | `map[string]string`         | No       | None     | Tag list                                                                      |
| `fields`      | `map[string]any-basic-type` | Yes       | None     | Line protocol can not be without a field, it can only be a base type, not a compound type (such as array, dictionary, etc.). |
| `time`        | `int64`                     | No       | None     | If it is not provided, the receiving time of DataKit shall prevail.                                      |

```json
[
  {
    "measurement": "abc",
    "tags": {
      "t1": "b",
      "t2": "d"
    },
    "fields": {
      "f1": 123,
      "f2": 3.4,
      "f3": "strval"
    },
    "time": 1624550216
  },
  {
    "measurement": "def",
    "tags": {
      "t1": "b",
      "t2": "d"
    },
    "fields": {
      "f1": 123,
      "f2": 3.4,
      "f3": "strval"
    },
    "time": 1624550216
  }
]
```

Notes:

- If it is a json body, mark  `Content-Type: application/json` on the request header, otherwise it would be treated as a normal line protocol.
- For now, `any-basic-type` refers to `int/float/bool/string` in the popular sense, regardless of the differences between different programming languages.
- As for the field of numeric type, in JSON, because numeric value does not distinguish float/int, it is difficult to judge whether it is int or float for `{"a" : 123}` JSON at present. Based on this, API treats numeric value and translates it into float type uniformly. This approach may cause type conflicts on storage (for example, line protocol body is used before, and JSON body is used later).
  - In the line protocol, int/float is clearly identified, such as `123i` for int and  `123` for float.
- Compared with the line protocol body, the performance of JSON body is poor, which is about 7 ~ 8 times different. On the premise of the same amount of data, a rough Benchmark comparison:

```shell
$ go test -bench=.
goos: darwin
goarch: amd64
pkg: gitlab.jiagouyun.com/cloudcare-tools/datakit/http
cpu: Intel(R) Core(TM) i5-8210Y CPU @ 1.60GHz
BenchmarkHandleWriteBody-4     582792  2135 ns/op
BenchmarkHandleJSONWriteBody-4  75606 15693 ns/op  # Obviously, the single overhead of json-body is higher
PASS
ok      gitlab.jiagouyun.com/cloudcare-tools/datakit/http       4.499s
```

### Logging Example {#api-logging-example}

```http
POST /v1/write/logging?precision=n&input=my-sample-logger&ignore_global_tags=123 HTTP/1.1

nginx,tag1=a,tag2=b,filename=a.log f1=1i,f2=1.2,f3="abc",message="real-log-data",status="debug" 1620723870000000000
mysql,tag1=a,tag2=b,filename=b.log f1=1i,f2=1.2,f3="abc",message="other-log-data",status="info" 1620723870000000000
redis,tag1=a,tag2=b,filename=c.log f1=1i,f2=1.2,f3="abc",message="more-log-data",status="error" 1620723870000000000
```

- The metric set name in the line protocol ( `nginx/mysql/redis`here) is stored as the `source` field of the log.
- The original log data is stored in the `message` field.
  
### Metric Data Example {#api-metric-example}

``` http
POST /v1/write/metric?precision=n&input=my-sample-logger&ignore_global_tags=123 HTTP/1.1

cpu,tag1=a,tag2=b f1=1i,f2=1.2,f3="abc" 1620723870000000000
mem,tag1=a,tag2=b f1=1i,f2=1.2,f3="abc" 1620723870000000000
net,tag1=a,tag2=b f1=1i,f2=1.2,f3="abc" 1620723870000000000
```

### Object Data Example {#api-object-example}

``` http
POST /v1/write/object?precision=n&input=my-sample-logger&ignore_global_tags=123 HTTP/1.1

redis,name=xxx,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
rds,name=yyy,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
slb,name=zzz,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
```

<!-- markdownlint-disable MD046 -->
???+ attention

    Object data must have the tag  `name` , otherwise the protocol will report an error.
    
    Object data should have a `message` field, which is mainly convenient for full-text search.

<!-- markdownlint-enable -->

### Custom Object Data Sample {#api-custom-object-example}

Custom objects are almost identical to objects, except that the latter are collected autonomously by DataKit, and the former are objects created by users through the datakit API.

```http
POST /v1/write/custom_object?precision=n&input=my-sample-logger&ignore_global_tags=123 HTTP/1.1

redis,name=xxx,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
rds,name=yyy,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
slb,name=zzz,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
```

<!-- markdownlint-disable MD046 -->
???+ attention

    Custom object data must have the tag `name` , otherwise the protocol will report an error.
    
    It would be better  to have a `message` field for custom object data, which is mainly convenient for full-text search.

<!-- markdownlint-enable -->

### RUM {#api-rum}

See [the document RUM](rum.md).

## `/v1/ping` {#api-ping}

Detect whether there is DataKit running at the target address, and obtain the startup time and version information of DataKit. Example:

``` http
GET /v1/ping HTTP/1.1

HTTP/1.1 200 OK

{
    "content":{
        "version":"1.1.6-rc0",
        "uptime":"1.022205003s"
    }
}
```

## `/v1/lasterror` {#api-lasterror}

Used to report errors of external collectors, for example:

``` http
POST /v1/lasterror HTTP/1.1
Content-Type: application/json

{
    "input":"redis",
    "err_content":"Cache avalanche"
}
```

## `/v1/workspace` {#api-workspace}

View workspace information and data quota information, for example:

``` http
GET /v1/workspace HTTP/1.1

HTTP/1.1 200 OK

{
  "content":
  [
    {
      "token": {
        "ws_uuid": "wksp_2dc431d6693711eb8ff97aeee04b54af",
        "bill_state": "normal",
        "ver_type": "pay",
        "token": "tkn_2dc438b6693711eb8ff97aeee04b54af",
        "db_uuid": "ifdb_c0fss9qc8kg4gj9bjjag",
        "status": 0,
        "creator": "",
        "expire_at": -1,
        "create_at": 0,
        "update_at": 0,
        "delete_at": 0
      },
      "data_usage": {
        "data_metric": 96966,
        "data_logging": 3253,
        "data_tracing": 2868,
        "data_rum": 0,
        "is_over_usage": false
      }
    }
  ]
}
```

## `/v1/query/raw` {#api-raw-query}

Use DQL to query data (only the data of the workspace where the DataKit is located), for example:

``` http
POST /v1/query/raw HTTP/1.1
Content-Type: application/json

{
    "queries":[
        {
            "query": "cpu:(usage_idle) LIMIT 1",  # dql query statement (required)
            "conditions": "",                     # append dql query criteria
            "max_duration": "1d",                 # maximum time range
            "max_point": 0,                       # maximum points
            "time_range": [],                     #
            "orderby": [],                        #
            "disable_slimit": true,               # Disable the default SLimit. When true, no default SLimit value will be added, otherwise SLimit 20 will be forced to be added
            "disable_multiple_field": true        # Disable multiple fields. When true, only the data of a single field can be queried (excluding the time field)
        }
    ],
    "echo_explain":true
}
```

Parameter description:

| Name                     | Description                                                                                                                                                                                                                       |
| :----------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `conditions`             | Add additional conditional expressions, using DQL syntax, such as `hostname="cloudserver01" OR system="ubuntu"`. It has an `AND` relationship with the conditional expression in the existing `query` , and parentheses are added at the outermost layer to avoid confusion.                                                  |
| `disable_multiple_field` | Whether to disable multiple fields. When true, only the data of a single field (excluding the time field) can be queried, and the default is `false`.                                                                                                                                   |
| `disable_slimit`         | Whether to disable the default SLimit, when true, the default SLimit value will not be added, otherwise, SLimit 20 will be forced to be added, and the default is `false`.                                                                                                                        |
| `echo_explain`           | Whether to return the final execution statement (returns the `raw_query` field in the JSON data).                                                                                                                                                                |
| `highlight`              | Highlight search results                                                                                                                |
| `limit`                  | Limiting the number of points returned by a single timeline will override the limit existing in DQL.                                                                                                                                                                        |
| `max_duration`           | Limit the maximum query time and support units  `ns/us/ms/s/m/h/d/w/y`, for example,  `3d` is 3 days, `2w` is 2 weeks, and `1y` is 1 year. The default is 1 year, which also restricts the `time_range` parameter.                                                                         |
| `max_point`              | Limit the maximum number of aggregated points. When using an aggregate function, if the aggregation density is too low, resulting in too many points, it is replaced by a new aggregation interval `(end_time-start_time)/max_point`.                                                                                        |
| `offset`                 | Typically used with limit configuration for result paging.                                                                                                                      |
| `orderby`                | Specify `order by`, content format is `map[string]string`, `key` s the field name to be sorted, `value` can only be sorted by `asc` and `desc`, such as `[ { "column01" : "asc" }, { "column02" : "desc" } ]`This article replaces the `order by` in the original query statement. |
| `queries`                | Basic query module, including query statements and various additional parameters.                                                                                                                                                                                   |
| `query`                  | DQL query statement（DQL [document](../dql/define.md)）                                                                                                                                                                               |
| `search_after`           | Deep paging, when calling paging for the first time, an empty list is passed in: `"search_after": []`. After success, the server will return a list, and the client can directly reuse the value of this list and pass it back to the subsequent query through the  `search_after` parameter.                                                   |
| `slimit`                 | Limiting the number of timelines will override slime existing in DQL.                                                                                                                                                                                 |
| `time_range`             | Limit the time range, adopt timestamp format, unit is millisecond, array size is 2 int, if there is only one element, it is considered as the start time, which will overwrite the query time interval in the original query statement.                                                                                          |

Return data example:

``` http
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
                        "usage_idle"
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
            "raw_query": "SELECT \"usage_idle\" FROM \"cpu\" LIMIT 1",
        }
    ]
}
```

## `/v1/object/labels` | `POST` {#api-object-labels}

Creat or update the `labels` of objects:

`request body` description

|           Parameter | Description                                                                          | Type       |
| -------------: | ----------------------------------------------------------------------------- | ---------- |
| `object_class` | represent the `object` type associated with `labels`, such as `HOST`                               | `string`   |
|  `object_name` | represent the `object` type associated with `labels`, such as `host-123`                           | `string`   |
|          `key` | represents the specific field name of the `object` to which `labels` is associated, such as the field `process_name`     | `string`   |
|        `value` | represents the specific field value of the `object` to which `labels` is associated, such as the process name field `systemsoundserverd` | `void`     |
|       `labels` | `labels` list, a `string` array                                             | `[]string` |

Example of request:

``` shell
curl -XPOST "127.0.0.1:9529/v1/object/labels" \
    -H 'Content-Type: application/json'  \
    -d'{
            "object_class": "host_processes",
            "object_name": "ubuntu20-dev_49392",
            "key": "host",
            "value": "ubuntu20-dev",
            "labels": ["l1","l2"]
        }'
```

Example of successful return:

``` json
status_code: 200
{
    "content": {
        "_id": "375370265b0641xxxxxxxxxxxxxxxxxxxxxxxxxx"
    }
}
```

Example of failing return:

``` json
status_code: 500
{
    "errorCode":"some-internal-error"
}
```

## `/v1/object/labels` | `DELETE` {#api-delete-object-labels}

Delete the `labels` of objects

`request body` description

|           Parameter | Description                                                                          | Type     |
| -------------: | ----------------------------------------------------------------------------- | -------- |
| `object_class` | represent the `object` type associated with `labels`, such as `HOST`                               | `string` |
|  `object_name` | represent the `object` type associated with `labels`, such as `host-123`                            | `string` |
|          `key` | represent the specific field name of the `object` to which `labels` is associated, such as the field `process_name`     | `string` |
|        `value` | represent the specific field value of the `object` to which `labels` is associated, such as the process name field `systemsoundserverd` | `void`   |

Example of request:

``` shell
curl -XPOST "127.0.0.1:9529/v1/object/labels"  \
    -H 'Content-Type: application/json'  \
    -d'{
            "object_class": "host_processes",
            "object_name": "ubuntu20-dev_49392",
            "key": "host",
            "value": "ubuntu20-dev"
        }'
```

Example of successful return:

``` json
status_code: 200
{
    "content": {
        "msg": "delete success!"
    }
}
```

Example of failing return:

``` json
status_code: 500
{
    "errorCode": "some-internal-error"
}
```

## `/v1/pipeline/debug` | `POST` {#api-debug-pl}

Providing the function of remote debugging PL.

Example of request:

``` http
POST /v1/pipeline/debug
Content-Type: application/json

{
    "pipeline": base64("pipeline-source-code"),
    "script_name": "<script_name>"
    "category": "<logging[metric, tracing, ...]>", # Log categories pass in log text, while other categories need to pass in row protocol text
    "data": [ base64("raw-logging-data1"), ... ], # It can be a log or a line protocol
    "encode": "@data 的字符编码",         # The default utf8 encode
    "benchmark": false,                  # Whether to turn on benchmark
}
```

Example of successful return:

``` http
HTTP/1.1 200 OK

{
    "content": {
        "cost": "2.3ms",
        "benchmark": BenchmarkResult.String(), # return benchmark results
        "error_msg": "",
        "plresults": [ # Since the log may be multi-line, multiple cutting results will be returned here.
            {
                "measurement" : "Metrics set name, typically log source",
                "tags": { "key": "val", "other-key": "other-val"},
                "fields": { "f1": 1, "f2": "abc", "f3": 1.2 },
                "time": 1644380607, # Unix time stamp (in seconds), which can be converted into a readable date by the front end.
                "time_ns": 421869748, # The remaining nanosecond time is easy to accurately convert into a date, and the complete nanosecond timestamp is 1644380607421869748,
                "dropped": false, # Whether to mark the result as to be discarded in the execution pipeline
                "error":""
            },
            {  another-result },
            ...
        ]
    }
}
```

Example of failing return:

``` http
HTTP Code: 400

{
    "error_code": "datakit.invalidCategory",
    "message": "invalid category"
}
```

## `/v1/dialtesting/debug` | `POST` {#api-debug-dt}

Providing the ability to debug dialtesting remotely.

Example of request: ：

``` http
POST /v1/dialtesting/debug
Content-Type: application/json

{
    "task_type":"http",//"http","tcp","icmp","websocket"
    "task": {
        "name":"",
        "method":"",
        "url":"",
        "post_url":"",
        "cur_status":"",
        "frequency":"",
        "enable_traceroute":true,//true代表勾选，tcp，icmp才有用
        "success_when_logic":"",
        "SuccessWhen":[]*HTTPSuccess ,
        "tags":map[string]string ,
        "labels":[]string,
        "advance_options":*HTTPAdvanceOption,
    }
}
```

Example of successful return:

``` http
HTTP/1.1 200 OK

{
    "content": {
        "cost": "2.3ms",
        "status": "success",//  success/fail
        "error_msg": "",
        "traceroute":[{"total":3,"failed":0,"loss":0,"avg_cost":0,"min_cost":2,"max_cost":3,"std_cost":33,"items":[{"ip":"127.0.0.1","response_time":33}]}],
    }
}
```

Example of failing return:

``` http
HTTP Code: 400

{
    "error_code": "datakit.invalidClass",
    "message": "invalid class"
}
```

## `/metrics` | `GET` {#api-metrics}

Get Datakit Prometheus metrics.

## DataKit Data Structure Constraint {#lineproto-limitation}

In order to standardize the data of Guance Cloud, the data collected by DataKit is constrained as follows (whether it is data in line protocol or JSON form), and the data that violates the constraints will be processed accordingly.

1. The key between Tag and Field does not allow duplicate names, that is, the same key cannot appear in both Tag and Field, otherwise the Field with duplicate names will be discarded.
2. key with the same name is not allowed inside Tag or Field, that is, the same key cannot appear more than once in Tag/Field, and only one key with the same name will be kept.
3. Tag number does not exceed 256. After the number exceeds, it will be sorted by key, and the redundant ones will be removed.
4. The number of Field does not exceed 1024. After the number exceeds, it will be sorted by key, and the redundant ones will be removed.
5. Tag/Field Key length does not exceed 256 bytes, and when it exceeds the length, it will be truncated.
6. Tag Value length does not exceed 1024 bytes, and truncation is performed when it exceeds the length.
7. The Field Value does not exceed 32M (32x1024x1024) bytes, and when it exceeds the length, it will be truncated.
8. `.` character is not allowed in Tag/Field key in any class of data except time series data.

## Extended Reading {#more-reading}

- [API Access Settings](datakit-conf.md#config-http-server)
- [API Current Limiting Configuration](datakit-conf.md#set-http-api-limit)
- [API Security Control](rum.md#security-setting)
