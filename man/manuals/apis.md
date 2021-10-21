{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：全平台

# DataKit API 文档

本文档主要描述 DataKit 开放出来 HTTP API 接口。

## `/v1/write/:category`


写入日志数据，参数列表：

| 参数名               | 类型   | 是否必选 | 默认值    | 说明                                               |
| -------------------- | ------ | -------- | --------- | -------------------------------------------------- |
| `category`           | string | true     | 无        | 目前支持 `metric/logging/rum/object/custom_object` |
| `precision`          | string | false    | `n`       | 数据精度(支持 `n/u/ms/s/m/h`)                      |
| `input`              | string | false    | `datakit` | 数据源名称                                         |
| `ignore_global_tags` | string | false    | 无        | 任意给值即认为忽略 DataKit 上的全局 tag            |
| `version`            | string | false    | 无        | 当前采集器的版本号                                |

HTTP body 支持行协议以及 JSON 俩种形式。

### JSON Body 示例

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
    "measurement":"abc",
      "tags": {
        "t1":"b",
        "t2":"d"
      },
      "fields": {
        "f1": 123,
        "f2": 3.4,
        "f3": "strval"
      },
      "time":1624550216
  },
  {
    "measurement":"def",
    "tags": {
      "t1":"b",
      "t2":"d"
    },
    "fields": {
      "f1": 123,
      "f2": 3.4,
      "f3": "strval"
    },
    "time":1624550216
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

### 日志(logging)示例

```http
POST /v1/write/logging?precision=n&input=my-sample-logger&ignore_global_tags=123 HTTP/1.1

nginx,tag1=a,tag2=b,filename=a.log f1=1i,f2=1.2,f3="abc",message="real-log-data",status="debug" 1620723870000000000
mysql,tag1=a,tag2=b,filename=b.log f1=1i,f2=1.2,f3="abc",message="other-log-data",status="info" 1620723870000000000
redis,tag1=a,tag2=b,filename=c.log f1=1i,f2=1.2,f3="abc",message="more-log-data",status="error" 1620723870000000000
```

- 行协议中的指标集名称(此处的 `nginx/mysql/redis`) 会作为日志的 `source` 字段来存储。
- 原式日志数据存放在 `message` 字段上

### 时序数据(metric)示例

```http
POST /v1/write/metric?precision=n&input=my-sample-logger&ignore_global_tags=123 HTTP/1.1

cpu,tag1=a,tag2=b f1=1i,f2=1.2,f3="abc" 1620723870000000000
mem,tag1=a,tag2=b f1=1i,f2=1.2,f3="abc" 1620723870000000000
net,tag1=a,tag2=b f1=1i,f2=1.2,f3="abc" 1620723870000000000
```

### 对象数据(object)示例

```http
POST /v1/write/object?precision=n&input=my-sample-logger&ignore_global_tags=123 HTTP/1.1

redis,name=xxx,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
rds,name=yyy,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
slb,name=zzz,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
```

> 注意：

- 对象数据必须有 `name` 这个 tag，否则协议报错
- 对象数据最好有 `message` 字段，主要便于做全文搜索


### 自定义对象数据(custom_object)示例

```http
POST /v1/write/custom_object?precision=n&input=my-sample-logger&ignore_global_tags=123 HTTP/1.1

redis,name=xxx,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
rds,name=yyy,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
slb,name=zzz,tag2=b f1=1i,f2=1.2,f3="abc",message="xxx" 1620723870000000000
```

> 注意：

- 自定义对象数据必须有 `name` 这个 tag，否则协议报错
- 自定义对象数据最好有 `message` 字段，主要便于做全文搜索

### RUM

参见 [RUM 文档](rum)

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

## `/v1/lasterror`

用于上报外部采集器的错误

### 示例

```
POST /v1/lasterror HTTP/1.1
Content-Type: application/json

{ 
	"input":"redis",
	"err_content":"Cache avalanche"
}
```
## `/v1/workspace`

查看工作空间信息及数据配额信息

### 示例

```http
GET /v1/worksapce HTTP/1.1

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
## `/v1/query/raw`

使用 DQL 进行数据查询（只能查询该 DataKit 所在的工作空间的数据）

### 示例

```
POST /v1/query/raw HTTP/1.1
Content-Type: application/json

{
    "queries":[
        {
            "query": "cpu:(usage_idle) LIMIT 1",  # dql查询语句（必填）
            "conditions": "",                     # 追加dql查询条件
            "max_duration": "1d",                 # 最大时间范围
            "max_point": 0,                       # 最大点数
            "time_range": [],                     #
            "orderby: [],                         #
            "disable_slimit": true,               # 禁用默认SLimit，当为true时，将不添加默认SLimit值，否则会强制添加SLimit 20
            "disable_multiple_field": true        # 禁用多字段。当为true时，只能查询单个字段的数据（不包括time字段）
        }
    ],
    "echo_explain":true
}
```

参数说明：

| 名称                     | 说明                                                                                                                                                                                                                       |
| :----------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `queries`                | 基础查询模块，包含查询语句和各项附加参数                                                                                                                                                                                   |
| `echo_explain`           | 是否返回最终执行语句（返回 JSON 数据中的 `raw_query` 字段）                                                                                                                                                                |
| `query`                  | DQL查询语句（DQL [文档](https://www.yuque.com/dataflux/doc/fsnd2r)）                                                                                                                                                       |
| `conditions`             | 额外添加条件表达式，使用 DQL 语法，例如`hostname="cloudserver01" OR system="ubuntu"`。与现有 `query` 中的条件表达式成 `AND` 关系，且会在最外层添加括号避免与其混乱                                                         |
| `time_range`             | 限制时间范围，采用时间戳格式，单位为毫秒，数组大小为2的int，如果只有一个元素则认为是起始时间，会覆盖原查询语句中的查询时间区间                                                                                             |
| `max_duration`           | 限制最大查询时间，支持单位 `ns/us/ms/s/m/h/d/w/y` ，例如 `3d` 是3天，`2w` 是2周，`1y` 是1年。默认是1年，此参数同样会限制 `time_range` 参数                                                                                 |
| `orderby`                | 指定`order by`参数，内容格式为 `map[string]string` 数组，`key` 为要排序的字段名，`value` 只能是排序方式即 `asc` 和 `desc`，例如 `[ { "column01" : "asc" }, { "column02" : "desc" } ]`。此条会替换原查询语句中的 `order by` |
| `max_point`              | 限制聚合最大点数。在使用聚合函数时，如果聚合密度过小导致点数太多，则会以 `(end_time-start_time)/max_point` 得到新的聚合间隔将其替换                                                                                        |
| `disable_slimit`         | 是否禁用默认SLimit，当为true时，将不添加默认SLimit值，否则会强制添加SLimit 20，默认为 `false`                                                                                                                              |
| `disable_multiple_field` | 是否禁用多字段。当为true时，只能查询单个字段的数据（不包括time字段），默认为 `false`                                                                                                                                       |

返回数据：

```
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

## `/v1/object/labels` | `POST`

创建或者更新对象的 `labels`

请求示例:

`request body`说明

| 参数           | 描述                                                                          | 类型       |
| -------------: |  ---------------------------------------------------------------------------  | ---------  |
| `object_class` | 表示 `labels` 所关联的 `object` 类型，如 `HOST`                               | `string`   |
| `object_name`  | 表示 `labels` 所关联的 `object` 名称，如 `host-123`                           | `string`   |
| `key`          | 表示 `labels` 所关联的 `object` 的具体字段名，如进程名字段 `process_name`     | `string`   |
| `value`        | 表示 `labels` 所关联的 `object` 的具体字段值，如进程名为 `systemsoundserverd` | `void`     |
| `labels`       | `labels` 列表，一个 `string` 数组                                             | `[]string` |

```shell
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

成功返回示例:

```json
status_code: 200
{
	"content": {
		"_id": "375370265b0641xxxxxxxxxxxxxxxxxxxxxxxxxx"
	}
}
```

失败返回示例:

```json
status_code: 500
{
	"errorCode":""
}
```

## `/v1/object/labels` | `DELETE`

删除对象的 `labels`

请求示例:

`request body`说明

| 参数           | 描述                                                                          | 类型     |
| -------------: | ---------------------------------------------------------------------------   | -------  |
| `object_class` | 表示 `labels` 所关联的 `object` 类型，如 `HOST`                               | `string` |
| `object_name`  | 表示 `labels` 所关联的 `object`名称，如 `host-123`                            | `string` |
| `key`          | 表示 `labels` 所关联的 `object` 的具体字段名，如进程名字段 `process_name`     | `string` |
| `value`        | 表示 `labels` 所关联的 `object` 的具体字段值，如进程名为 `systemsoundserverd` | `void`   |


```shell
curl -XPOST "127.0.0.1:9529/v1/object/labels"  \
	-H 'Content-Type: application/json'  \
	-d'{
			"object_class": "host_processes",
			"object_name": "ubuntu20-dev_49392",
			"key": "host",
			"value": "ubuntu20-dev"
		}'
```

成功返回示例:

```json
status_code: 200
{
	"content": {
		"msg": "delete success!"
	}
}
```

失败返回示例:

```json
status_code: 500
{
	"errorCode":""
}
```

## DataKit 行协议约束

为规范观测云中的数据，现对经过 DataKit 的行协议数据，做如下约束：

- tags 和 fields 中的 key 不允许重名
- tags 内部或 fields 内部不允许出现同名 key
