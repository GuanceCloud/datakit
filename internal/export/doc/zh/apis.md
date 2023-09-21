
# DataKit API

---

本文档主要描述 DataKit 开放出来 HTTP API 接口。

## API 综述 {#intro}

DataKit 目前只支持 HTTP 接口，主要涉及数据写入，数据查询。

### 通过 API 获取远端 DataKit 版本号 {#api-get-dk-version}

有两种方式可获取版本号：

- 请求 DataKit ping 接口： `curl http://ip:9529/v1/ping`
- 在下述每个 API 请求的返回 Header 中，通过 `X-DataKit` 可获知当前请求的 DataKit 版本

## `/v1/write/:category` {#api-v1-write}

本 API 用于给 DataKit 上报各类数据（`category`），参数说明如下：

| 参数名                    | 类型   | 是否必选 | 默认值    | 说明                                                                                                                                          |
| --------------------      | ------ | -------- | --------- | --------------------------------------------------                                                                                            |
| `category`                | string | Y        | -         | 目前只支持 `metric,logging,rum,object,custom_object,keyevent`，以 `metric` 为例， 其 URL 应该写成 `/v1/write/metric`                          |
| `echo_line_proto`         | string | N        | -         | 给任意值（如 `true`）即返回行协议形式的点数据，默认不返回                                                                                     |
| `echo_json`               | string | N        | -         | 给任意值（如 `true`）即返回 JSON 格式的数据点，默认不返会，如果同时指定两种 echo，优先返回行协议形式的点数据                                  |
| `global_election_tags`    | string | N        | -         | 给任意值（如 `true`）即认为追加全局选举类 tag（[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)）                                     |
| `ignore_global_host_tags` | string | false    | -         | 给任意值（如 `true`）即认为忽略 DataKit 上的全局 tag（[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)）。`ignore_global_tags` 将弃用 |
| `input`                   | string | N        | `datakit` | 数据源名称                                                                                                                                    |
| `loose`                   | bool   | N        | true      | 宽松模式，对于一些不合规的行协议，DataKit 会尝试修复它们（[:octicons-tag-24: Version-1.4.11](changelog.md#cl-1.4.11)）                        |
| `strict`                  | bool   | N        | false     | 严格模式，对于一些不合规的行协议，API 直接报错，并告知具体的原因（[:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9)）                  |
| `precision`               | string | N        | `n`       | 数据精度（支持 `n/u/ms/s/m/h`）                                                                                                               |
| `source`                  | string | N        | -         | 仅仅针对 logging 支持指定该字段（即 `category` 为 `logging`）。如果不指定 `source`，则上传的日志数据不会执行 Pipeline 切割                    |
| `version`                 | string | N        | -         | 当前采集器的版本号                                                                                                                            |

HTTP body 支持行协议以及 JSON 俩种形式。关于数据结构（不管是行协议形式还是 JSON 形式）的约束，参见[这里](apis.md#lineproto-limitation)。

### 数据类型分类 {#category}

DataKit 中主要有如下数据类型（以简称字母序排列）：

| 简称 | 名称            | URL 表示                  | 说明               |
| ---- | ----            | ----                      | ---                |
| CO   | `custom_object` | `/v1/write/custom_object` | 自定义对象数据     |
| E    | `keyevent`      | `/v1/write/keyevent`      | Event 数据         |
| L    | `logging`       | `/v1/write/logging`       | 日志数据           |
| M    | `metric`        | `/v1/write/metric`        | 时序数据           |
| N    | `network`       | `/v1/write/network`       | 一般指 eBPF 数据   |
| O    | `object`        | `/v1/write/object`        | 对象数据           |
| P    | `profiling`     | `/v1/write/profiling`     | Profiling 数据     |
| R    | `rum`           | `/v1/write/rum`           | RUM 数据           |
| S    | `security`      | `/v1/write/security`      | 安全巡检数据       |
| T    | `tracing`       | `/v1/write/tracing`       | APM（Tracing）数据 |

不同的数据类型，其处理方式不一样，在观测云的用法也不尽相同。在 Datakit 的配置和使用过程中，有时候会穿插使用某个类型的不同形式（比如在 sinker 配置中用简写，在 API 请求中则用其 URL 表示）

### JSON Body 示例 {#api-json-example}

为便于行协议处理，所有数据上传 API 均支持 JSON 形式的 body。JSON body 参数说明

| 参数名        | 类型                        | 是否必选 | 默认值 | 说明                                                                          |
| ------------- | --------------------------- | -------- | ------ | ----------------------------------------------------------------------------- |
| `measurement` | `string`                    | 是       | 无     | 指标集名称                                                                    |
| `tags`        | `map[string]string`         | 否       | 无     | 标签列表                                                                      |
| `fields`      | `map[string]any-basic-type` | 是       | 无     | 行协议不能没有指标（field），只能是基础类型，不是是复合类型（如数组、字典等） |
| `time`        | `int64`                     | 否       | 无     | 如果不提供，则以 DataKit 的接收时间为准                                       |

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

注意事项：

- 如果是 JSON body，需在请求头上标注 `Content-Type: application/json`，否则当做普通行协议处理
- 目前 `any-basic-type` 指通俗意义上的 `int/float/bool/string`，不考虑不同编程语言差异
- 关于数值类型的 field，在 JSON 中，由于数值不区分 float/int，导致对于 `{"a" : 123}` 这段 JSON，目前难以判断其 int 还是 float，基于此，API 对数值处理，统一翻译成 float 类型。这种做法，可能造成存储上的类型冲突（如之前是用行协议 body，后面采用 JSON body）
    - 行协议中，对 int/float 有明显的标识，如 `123i` 为 int，而 `123` 为 float
- 相比行协议的 Body，JSON 形式的 body 性能较差，大概有 7~8 倍的差距。同等数据量前提下，粗略的 Benchmark 对比：

```shell
$ go test -bench=.
goos: darwin
goarch: amd64
pkg: gitlab.jiagouyun.com/cloudcare-tools/datakit/http
cpu: Intel(R) Core(TM) i5-8210Y CPU @ 1.60GHz
BenchmarkHandleWriteBody-4     582792  2135 ns/op
BenchmarkHandleJSONWriteBody-4  75606 15693 ns/op  # 明显 json-body 的单次开销更大
PASS
ok      gitlab.jiagouyun.com/cloudcare-tools/datakit/http       4.499s
```

### 日志示例 {#api-logging-example}

```http
POST /v1/write/logging?precision=n&input=my-sample-logger&ignore_global_tags=123 HTTP/1.1

nginx,tag1=a,tag2=b,filename=a.log f1=1i,f2=1.2,f3="abc",message="real-log-data",status="debug" 1620723870000000000
mysql,tag1=a,tag2=b,filename=b.log f1=1i,f2=1.2,f3="abc",message="other-log-data",status="info" 1620723870000000000
redis,tag1=a,tag2=b,filename=c.log f1=1i,f2=1.2,f3="abc",message="more-log-data",status="error" 1620723870000000000
```

- 行协议中的指标集名称（此处的 `nginx/mysql/redis`）会作为日志的 `source` 字段来存储。
- 原式日志数据存放在 `message` 字段上

### 时序数据示例 {#api-metric-example}

``` http
POST /v1/write/metric?precision=n&input=my-sample-logger&ignore_global_tags=123 HTTP/1.1

cpu,tag1=a,tag2=b f1=1i,f2=1.2,f3="abc" 1620723870000000000
mem,tag1=a,tag2=b f1=1i,f2=1.2,f3="abc" 1620723870000000000
net,tag1=a,tag2=b f1=1i,f2=1.2,f3="abc" 1620723870000000000
```

### 对象数据示例 {#api-object-example}

``` http
POST /v1/write/object?precision=n&input=my-sample-logger&ignore_global_tags=123 HTTP/1.1

redis,name=xxx,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
rds,name=yyy,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
slb,name=zzz,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
```

<!-- markdownlint-disable MD046 -->
???+ attention

    对象数据必须有 `name` 这个 tag，否则协议报错。

    对象数据最好有 `message` 字段，主要便于做全文搜索。
<!-- markdownlint-enable -->

### 自定义对象数据示例 {#api-custom-object-example}

自定义对象跟对象几乎一致，只是后者是 Datakit 自主采集的，前者是用户通过 Datakit API 创建的对象。

```http
POST /v1/write/custom_object?precision=n&input=my-sample-logger&ignore_global_tags=123 HTTP/1.1

redis,name=xxx,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
rds,name=yyy,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
slb,name=zzz,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
```

<!-- markdownlint-disable MD046 -->
???+ attention

    自定义对象数据必须有 `name` 这个 tag，否则协议报错
    
    自定义对象数据最好有 `message` 字段，主要便于做全文搜索
<!-- markdownlint-enable -->

### RUM {#api-rum}

参见 [RUM 文档](../integrations/rum.md)

## `/v1/ping` {#api-ping}

检测目标地址是否有 DataKit 运行，可获取 DataKit 启动时间以及版本信息。示例：

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

用于上报外部采集器的错误，示例：

``` http
POST /v1/lasterror HTTP/1.1
Content-Type: application/json

{
    "input":"redis",
    "source":"us-east-9xwha",
    "err_content":"Cache avalanche"
}
```

## `/v1/workspace` {#api-workspace}

查看工作空间信息及数据配额信息，示例：

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

使用 DQL 进行数据查询（只能查询该 DataKit 所在的工作空间的数据），示例：

``` http
POST /v1/query/raw HTTP/1.1
Content-Type: application/json

{
    "queries":[
        {
            "query": "cpu:(usage_idle) LIMIT 1",  # DQL 查询语句（必填）
            "conditions": "",                     # 追加 DQL 查询条件
            "max_duration": "1d",                 # 最大时间范围
            "max_point": 0,                       # 最大点数
            "time_range": [],                     #
            "orderby": [],                        #
            "disable_slimit": true,               # 禁用默认 SLimit，当为 true 时，将不添加默认 SLimit 值，否则会强制添加 SLimit 20
            "disable_multiple_field": true        # 禁用多字段。当为 true 时，只能查询单个字段的数据（不包括 time 字段）
        }
    ],
    "echo_explain":true
}
```

参数说明

| 名称                     | 说明                                                                                                                                                                                                                       |
| :----------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `conditions`             | 额外添加条件表达式，使用 DQL 语法，例如 `hostname="cloudserver01" OR system="ubuntu"`。与现有 `query` 中的条件表达式成 `AND` 关系，且会在最外层添加括号避免与其混乱                                                         |
| `disable_multiple_field` | 是否禁用多字段。当为 true 时，只能查询单个字段的数据（不包括 time 字段），默认为 `false`                                                                                                                                   |
| `disable_slimit`         | 是否禁用默认 SLimit，当为 true 时，将不添加默认 SLimit 值，否则会强制添加 SLimit 20，默认为 `false`                                                                                                                        |
| `echo_explain`           | 是否返回最终执行语句（返回 JSON 数据中的 `raw_query` 字段）                                                                                                                                                                |
| `highlight`              | 高亮搜索结果                                                                                                                                                                                                               |
| `limit`                  | 限制单个时间线返回的点数，将覆盖 DQL 中存在的 limit                                                                                                                                                                        |
| `max_duration`           | 限制最大查询时间，支持单位 `ns/us/ms/s/m/h/d/w/y` ，例如 `3d` 是 3 天，`2w` 是 2 周，`1y` 是 1 年。默认是 1 年，此参数同样会限制 `time_range` 参数                                                                         |
| `max_point`              | 限制聚合最大点数。在使用聚合函数时，如果聚合密度过小导致点数太多，则会以 `(end_time-start_time)/max_point` 得到新的聚合间隔将其替换                                                                                        |
| `offset`                 | 一般跟 limit 配置使用，用于结果分页                                                                                                                                                                                        |
| `orderby`                | 指定 `order by` 参数，内容格式为 `map[string]string` 数组，`key` 为要排序的字段名，`value` 只能是排序方式即 `asc` 和 `desc`，例如 `[ { "column01" : "asc" }, { "column02" : "desc" } ]`。此条会替换原查询语句中的 `order by` |
| `queries`                | 基础查询模块，包含查询语句和各项附加参数                                                                                                                                                                                   |
| `query`                  | DQL 查询语句（DQL [文档](../dql/define.md)）                                                                                                                                                                               |
| `search_after`           | 深度分页，第一次调用分页的时候，传入空列表：`"search_after": []`，成功后服务端会返回一个列表，客户端直接复用这个列表的值再次通过 `search_after` 参数回传给后续的查询即可                                                   |
| `slimit`                 | 限制时间线个数，将覆盖 DQL 中存在的 `slimit`                                                                                                                                                                                 |
| `time_range`             | 限制时间范围，采用时间戳格式，单位为毫秒，数组大小为 2 的 int，如果只有一个元素则认为是起始时间，会覆盖原查询语句中的查询时间区间                                                                                          |

返回数据示例：

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

创建或者更新对象的 `labels`

`request body` 说明

|           参数 | 描述                                                                          | 类型       |
| -------------: | ----------------------------------------------------------------------------- | ---------- |
| `object_class` | 表示 `labels` 所关联的 `object` 类型，如 `HOST`                               | `string`   |
|  `object_name` | 表示 `labels` 所关联的 `object` 名称，如 `host-123`                           | `string`   |
|          `key` | 表示 `labels` 所关联的 `object` 的具体字段名，如进程名字段 `process_name`     | `string`   |
|        `value` | 表示 `labels` 所关联的 `object` 的具体字段值，如进程名为 `systemsoundserverd` | `void`     |
|       `labels` | `labels` 列表，一个 `string` 数组                                             | `[]string` |

请求示例：

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

成功返回示例：

``` json
status_code: 200
{
    "content": {
        "_id": "375370265b0641xxxxxxxxxxxxxxxxxxxxxxxxxx"
    }
}
```

失败返回示例：

``` json
status_code: 500
{
    "errorCode":"some-internal-error"
}
```

## `/v1/object/labels` | `DELETE` {#api-delete-object-labels}

删除对象的 `labels`

`request body` 说明

|           参数 | 描述                                                                          | 类型     |
| -------------: | ----------------------------------------------------------------------------- | -------- |
| `object_class` | 表示 `labels` 所关联的 `object` 类型，如 `HOST`                               | `string` |
|  `object_name` | 表示 `labels` 所关联的 `object` 名称，如 `host-123`                           | `string` |
|          `key` | 表示 `labels` 所关联的 `object` 的具体字段名，如进程名字段 `process_name`     | `string` |
|        `value` | 表示 `labels` 所关联的 `object` 的具体字段值，如进程名为 `systemsoundserverd` | `void`   |

请求示例：

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

成功返回示例：

``` json
status_code: 200
{
    "content": {
        "msg": "delete success!"
    }
}
```

失败返回示例：

``` json
status_code: 500
{
    "errorCode": "some-internal-error"
}
```

## `/v1/pipeline/debug` | `POST` {#api-debug-pl}

提供远程调试 PL 的功能。

错误信息 `PlError` 结构：

```go
type Position struct {
    File string `json:"file"`
    Ln   int    `json:"ln"`
    Col  int    `json:"col"`
    Pos  int    `json:"pos"`
}

type PlError struct {
    PosChain []Position `json:"pos_chain"`
    Err      string     `json:"error"`
}
```

错误信息 JSON 示例：

```json
{
  "pos_chain": [
    { // 错误生成位置（脚本终止运行）
      "file": "xx.p",    // 文件名或文件路径
      "ln":   15,        // 行号
      "col":  29,        // 列号
      "pos":  576,       // 从 0 开始的字符在文本中绝对位置
    },
    ... ,
    { // 调用链的起点
      "file": "b.p",
      "ln":   1,
      "col":  1,
      "pos":  0,
    },
  ],
  "error": "error msg"
}
```

请求示例：

``` http
POST /v1/pipeline/debug
Content-Type: application/json

{
    "pipeline": {
      "<caregory>": {
        "<script_name>": <base64("pipeline-source-code")>
      }
    },
    "script_name": "<script_name>"
    "category": "<logging[metric, tracing, ...]>", # 日志类别传入日志文本，其他类别需要传入行协议文本
    "data": [ base64("raw-logging-data1"), ... ], # 可以是日志或者行协议
    "encode": "@data 的字符编码",         # 默认是 utf8 编码
    "benchmark": false,                  # 是否开启 benchmark
}
```

正常返回示例：

``` http
HTTP/1.1 200 OK

{
    "content": {
        "cost": "2.3ms",
        "benchmark": BenchmarkResult.String(), # 返回 benchmark 结果
        "pl_errors": [],                       # 脚本解析或检查时产生的 PlError 列表
        "plresults": [                         # 由于日志可能是多行的，此处会返回多个切割结果
            {
                "point": {
                  "name" : "可以是指标集名称、日志 source 等",
                  "tags": { "key": "val", "other-key": "other-val"},
                  "fields": { "f1": 1, "f2": "abc", "f3": 1.2 }
                  "time": 1644380607,   # Unix 时间戳（单位秒）, 前端可将其转成可读日期
                  "time_ns": 421869748, # 余下的纳秒时间，便于精确转换成日期，完整的纳秒时间戳为 1644380607421869748
                }
                "dropped": false,  # 是否在执行 pipeline 中将结果标记为待丢弃
                "run_error": null  # 如果没有错误，值为 null
            },
            {  another-result },
            ...
        ]
    }
}
```

错误返回示例：

``` http
HTTP Code: 400

{
    "error_code": "datakit.invalidCategory",
    "message": "invalid category"
}
```

## `/v1/dialtesting/debug` | `POST` {#api-debug-dt}

提供远程调试拨测的功能。

请求示例：

``` http
POST /v1/dialtesting/debug
Content-Type: application/json

{
    "task_type" : "http",//"http","tcp","icmp","websocket"
    "task" : {
        "name"               : "",
        "method"             : "",
        "url"                : "",
        "post_url"           : "",
        "cur_status"         : "",
        "frequency"          : "",
        "enable_traceroute"  : true, // true 代表勾选，tcp，icmp 才有用
        "success_when_logic" : "",
        "SuccessWhen"        : []*HTTPSuccess ,
        "tags"               : map[string]string ,
        "labels"             : []string,
        "advance_options"    : *HTTPAdvanceOption,
    }
}
```

正常返回示例：

``` http
HTTP/1.1 200 OK

{
    "content": {
        "cost": "2.3ms",
        "status": "success", # success/fail
        "error_msg": "",
        "traceroute":[
          {
              "total"    : 3,
              "failed"   : 0,
              "loss"     : 0,
              "avg_cost" : 0,
              "min_cost" : 2,
              "max_cost" : 3,
              "std_cost" : 33,
              "items" : [
                  {
                      "ip"            : "127.0.0.1",
                      "response_time" : 33
                  }
              ]
         }
        ]
    }
}
```

错误返回示例：

``` http
HTTP Code: 400

{
    "error_code": "datakit.invalidClass",
    "message": "invalid class"
}
```

## `/v1/sourcemap` | `PUT` {#api-sourcemap-upload}

[:octicons-tag-24: Version-1.12.0](changelog.md#cl-1.12.0)

上传 sourcemap 文件，该接口需要开启 [RUM 采集器](../integrations/rum.md)。

请求参数说明。

|           参数 | 描述                                                            | 类型     |
| -------------: | -------------------------------------------------------------- | -------- |
| `token` |`datakit.conf` 配置中的 `dataway` 地址中包含的 token                      | `string` |
| `app_id` | 用户访问应用唯一 ID 标识，如 `test-sourcemap`                            | `string` |
| `env` | 应用的部署环境，如 `prod`                                                  | `string` |
| `version` |应用的版本，如 `1.0.0`                                                 | `string` |
| `platform` |应用类型， 可选值 `web/miniapp/android/ios`, 默认 `web`                | `string` |

请求示例：

``` shell
curl -X PUT "http://localhost:9529/v1/sourcemap?app_id=test_sourcemap&env=production&version=1.0.0&token=tkn_xxxxx&platform=web" \
-F "file=@./sourcemap.zip" \
-H "Content-Type: multipart/form-data"
```

成功返回示例：

``` json
{
  "content": "uploaded to [/path/to/datakit/data/rum/web/test_sourcemap-production-1.0.0.zip]!",
  "errorMsg": "",
  "success": true
}
```

失败返回示例：

``` json
{
  "content": null,
  "errorMsg": "app_id not found",
  "success": false
}
```

## `/v1/sourcemap` | `DELETE` {#api-sourcemap-delete}

[:octicons-tag-24: Version-1.16.0](changelog.md#cl-1.16.0)

删除 sourcemap 文件，该接口需要开启 [RUM 采集器](../integrations/rum.md)。

请求参数说明。

|           参数 | 描述                                                            | 类型     |
| -------------: | -------------------------------------------------------------- | -------- |
| `token` |`datakit.conf` 配置中的 `dataway` 地址中包含的 token                      | `string` |
| `app_id` | 用户访问应用唯一 ID 标识，如 `test-sourcemap`                            | `string` |
| `env` | 应用的部署环境，如 `prod`                                                  | `string` |
| `version` |应用的版本，如 `1.0.0`                                                 | `string` |
| `platform` |应用类型， 可选值 `web/miniapp/android/ios`, 默认 `web`     | `string` |

请求示例：

``` shell
curl -X DELETE "http://localhost:9529/v1/sourcemap?app_id=test_sourcemap&env=production&version=1.0.0&token=tkn_xxxxx&platform=web"
```

成功返回示例：

``` json
{
  "content":"deleted [/path/to/datakit/data/rum/web/test_sourcemap-production-1.0.0.zip]!",
  "errorMsg":"",
  "success":true
}
```

失败返回示例：

``` json
{
  "content": null,
  "errorMsg": "delete sourcemap file [/path/to/datakit/data/rum/web/test_sourcemap-production-1.0.0.zip] failed: remove /path/to/datakit/data/rum/web/test_sourcemap-production-1.0.0.zip: no such file or directory",
  "success": false
}
```

## `/v1/sourcemap/check` | `GET` {#api-sourcemap-check}

[:octicons-tag-24: Version-1.16.0](changelog.md#cl-1.16.0)

验证 sourcemap 文件是否正确配置，该接口需要开启 [RUM 采集器](../integrations/rum.md)。

请求参数说明。

|           参数 | 描述                                                                          | 类型     |
| -------------: | ----------------------------------------------------------------------------- | -------- |
| `error_stack` | error 的堆栈信息                      | `string` |
| `app_id` | 用户访问应用唯一 ID 标识，如 `test-sourcemap`                            | `string` |
| `env` | 应用的部署环境，如 `prod`                           | `string` |
| `version` |应用的版本，如 `1.0.0`     | `string` |
| `platform` |应用类型， 可选值 `web/miniapp/android/ios`, 默认 `web`     | `string` |

请求示例：

``` shell
curl "http://localhost:9529/v1/sourcemap/check?app_id=test_sourcemap&env=production&version=1.0.0&error_stack=at%20test%20%40%20http%3A%2F%2Flocalhost%3A8080%2Fmain.min.js%3A1%3A48"
```

成功返回示例：

``` json
{
  "content": {
    "error_stack": "at test @ main.js:6:6",
    "original_error_stack": "at test @ http://localhost:8080/main.min.js:1:48"
  },
  "errorMsg": "",
  "success": true
}

```

失败返回示例：

``` json
{
  "content": {
    "error_stack": "at test @ http://localhost:8080/main.min.js:1:483",
    "original_error_stack": "at test @ http://localhost:8080/main.min.js:1:483"
  },
  "errorMsg": "fetch original source information failed, make sure sourcemap file [main.min.js.map] is valid",
  "success": false
}

```

## `/metrics` | `GET` {#api-metrics}

获取 Datakit 暴露的 Prometheus 指标。

## DataKit 数据结构约束 {#lineproto-limitation}

为规范观测云中的数据，现对 DataKit 采集的数据，做如下约束（不管是行协议还是 JSON 形式的数据），并对违反约束的数据将进行相应的处理。

1. Tag 和 Field 之间的 key 不允许重名，即同一个 key 不能在 Tag 和 Field 中同时出现，否则将丢弃重名的 Field
1. Tag 或 Field 内部不允许出现同名 key，即同一个 key 不能在 Tag/Field 中出现多次，对于同名 key，将仅保留其中一个
1. Tag 个数不超过 256 个，超过个数后，将按 key 排序，去掉多余的
1. Field 个数不超过 1024 个，超过个数后，将按 key 排序，去掉多余的
1. Tag/Field Key 长度不超过 256 字节，超过长度时，将进行截断处理
1. Tag Value 长度不超过 1024 字节，超过长度时，将进行截断处理
1. Field Value 不超过 32M(32x1024x1024) 字节，超过长度时，将进行截断处理
1. 除时序数据外，其它类数据中，均不允许在 Tag/Field key 中出现 `.` 字符

## 延伸阅读 {#more-reading}

- [API 访问设置](datakit-conf.md#config-http-server)
- [API 限流配置](datakit-conf.md#set-http-api-limit)
- [API 安全控制](../integrations/rum.md#security-setting)
