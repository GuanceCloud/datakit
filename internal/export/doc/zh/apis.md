
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

本 API 用于给 DataKit 上报各类数据（`category`），URL 参数说明如下：

| 参数名                                                            | 类型   | 是否必选 | 默认值         | 说明                                                                                                                                                                                                      |
| ---                                                               | ---    | ---      | ---            | ---                                                                                                                                                                                                       |
| `category`                                                        | string | Y        | -              | 目前只支持 `metric,logging,rum,object,custom_object,keyevent`，以 `metric` 为例， 其 URL 应该写成 `/v1/write/metric`                                                                                      |
| `dry`                                                             | bool   | N        | false          | 测试模式，只是将 Point POST 给 Datakit，实际上并不上传到观测云（[:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)）                                                                              |
| `echo` [:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0) | string | N        | -              | 可选值 `lp/json/pbjson`，`lp` 表示在返回的 Body 中以行协议形式来表示上传的 Point，后面分别是[普通 JSON](apis.md#api-v1-write-body-json-protocol) 和[高级 JSON](apis.md#api-v1-write-body-pbjson-protocol) |
| `global_election_tags`                                            | bool   | N        | false          | 是否追加 *全局选举 tag* （[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)）                                                                                                                      |
| `ignore_global_host_tags`                                         | bool   | N        | -              | 是否忽略 DataKit 上的 *全局主机 tag* （[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)）                                                                                                         |
| `input`                                                           | string | N        | `datakit-http` | 数据源名称，该名称会在 Datakit monitor 上展示，便于调试（[:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)）                                                                                     |
| `loose`                                                           | bool   | N        | true           | 是否宽松模式，对于一些不合规的 Point，DataKit 会尝试修复它们（[:octicons-tag-24: Version-1.4.11](changelog.md#cl-1.4.11)）                                                                                |
| `precision`                                                       | string | N        |                | 数据精度（支持 `n/u/ms/s/m/h`）。如果参数不传入，则自动识别时间精度[:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)                                                                             |
| `source`                                                          | string | N        | -              | 如果不指定 `source`（或者对应的 *source.p* 不存在或无效），上传的 Point 数据不会执行 Pipeline                                                                                                             |
| `strict`                                                          | bool   | N        | false          | 严格模式，对于一些不合规的行协议，API 直接报错，并告知具体的原因（[:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9)）                                                                              |

<!-- markdownlint-disable MD046 -->
???+ attention

    - 以下参数已弃用 [:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)

        - `echo_line_proto` : 用 `echo` 参数替代
        - `echo_json`       : 用 `echo` 参数替代

    - 虽然多个参数都是 bool 类型，如果不需要开启对应的 option，不要传入 `false` 值，API 只会判断对应参数上是否有值，而不管其值内容。

    - 时间精度（`precision`）自动识别（[:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)）指根据传入的时间戳数值，猜测其可能的时间精度，数学意义上它不能保证正确，但是日常使用是足够的。比如对于时间戳 1716544492，其时间戳判断为秒，对 1716544492000 会判断为毫秒，等等。
    - 虽然目前协议上支持二进制格式以及 any 格式两种类型，但目前中心尚未支持这两种数据的写入。**特此注明**。
<!-- markdownlint-enable -->

### Body 说明 {#api-v1-write-body}

HTTP body 支持行协议以及两种 JSON 俩种形式。

#### 行协议 Body {#api-v1-write-body-line-protocol}

单条行协议形式如下：

```text
measurement,<tag-list> <field-list> timestamp
```

多条行协议之间，以换行分隔：

```text
measurement_1,<tag-list> <field-list> timestamp
measurement_2,<tag-list> <field-list> timestamp
```

其中：

- `measurement` 为指标集名字，它表示一组指标的集合名称，比如指标集名 `disk` 下可能有 `free/used/total` 等指标，
- `<tag-list>` 为一组 tag 列表，tag 之间以 `,` 分隔。单个 tag 形式为 `key=value`，此处 `value` 均被视为字符串。在行协中，`<tag-list>` 是**可选**的
- `<field-list>` 为一组指标列表，彼此支架以 `,` 分隔。在行协中，`<field-list>` 是**必填**的。单个指标形式为 `key=value`，`value` 形式视其类型而定，分别如下：
    - int 示例：`some_int=42i`，即在整数数值后面追加一个 `i` 表示
    - uint 示例：`some_uint=42u`，即在整数数值后面追加一个 `u` 表示
    - float 示例：`some_float_1=3.14,some_float_2=3`，此处 `some_float_2` 虽然是整数 3，但其仍被视为一个 float
    - string 示例：`some_string="hello world"`，字符串值需要在两端都加上 `"`
    - bool 示例：`some_true=T,some_false=F`，此处 `T/F` 还可以用 `t/f/true/false` 分别表示
    - 二进制示例：`some_binary="base-64-encode-string"b`，二进制数据（文本字节流 `[]byte` 等）需要 base64 编码才能在行协议中表示，它跟 string 的表示类似，只是在后面追加了一个 `b` 用来标识
    - 数组示例：`some_array=[1i,2i,3i]`，注意，数组内的类型只能是基础类型（`int/uint/float/boolean/string/[]byte`，**不含数组**），且其类型必须一致，形如 `invalid_array=[1i,3.14,"string"]` 这种数组目前是不支持的
- `timestamp` 为整数时间戳，默认情况下，Datakit  以纳秒单位来处理这个时间戳，如果原数据不是纳秒，需要通过请求参数 `precision` 来指定真实的时间戳精度。在行协议中，`timestamp` 是可选的，如果数据中不带时间戳，Datakit 以接受到的时间作为当前行协议时间。

这几个部分之间：

- `measurement` 和 `<tag-list>` 之间以 `,` 分隔
- `<tag-list>` 和 `<field-list>` 之间以单个空格分隔
- `<field-list>` 和 `timestamp` 之间以单个空格分隔
- 行协议中，如果头部有 `#`，视为注释，它实际上会被解析器忽略

下面是一些行协议简单示例：

```text
# 普通示例
some_measurement,host=my_host,region=my_region cpu_usage=0.01,memory_usage=1048576u 1710321406000000000

# 不含 tag 示例
some_measurement cpu_usage=0.01,memory_usage=1048576u 1710321406000000000

# 不含时间戳示例
some_measurement,host=my_host,region=my_region cpu_usage=0.01,memory_usage=1048576u

# 含所有基本类型
some_measurement,host=my_host,region=my_region float=0.01,uint=1048576u,int=42i,string="my-host",boolean=T,binary="aGVsbG8="b,array=[1.414,3.14] 1710321406000000000
```

一些字段名和字段值值的特殊的转义：

- `measurement` 中 `,` 需要转义
- tag key 和 field key 中的 `=`、`,` 和空格需要转义
- `measurement`、tag key 和 field key 中不允许出现换行（`\n`）
- tag value 中不允许出现换行（`\n`），field value 中的换行不需要转义
- field value 如果是 string，其中如果有 `"` 字符，也需要转义

#### JSON Body {#api-v1-write-body-json-protocol}

JSON 形式的 body 相比行协议，它无需做太多的转义，一个简单 JSON 格式如下：

```json
[
    {
        "measurement": "指标集名字",

        "tags": {
            "key": "value",
            "another-key": "value"
        },

        "fields": {
            "key": value,
            "another-key": value # 此处 value 可以是 number/bool/string/list 这几种
        },

        "time": unix-timestamp
    },

    {
        # another-point...
    }
]
```

以下是一个简单的 JSON 示例：

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
    "time": 1624550216000000000
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
      "f4": false,
      "f5": [1, 2, 3, 4],
      "f6": ["str1", "str2", "str3"]
    },
    "time": 1624550216000000000
  }
]
```

<!-- markdownlint-disable MD046 -->
???+ warning

    这种 JSON 结构虽然简单，但其有几个缺点：
    
    - 不能区分 int/uint/float 这几种数值类型，比如，对于所有的数值，JSON 默认都以 float 来处理，而对于数值 42，JSON 无法区分它是有符号还是无符号
    - 不支持表示二进制（`[]byte`）数据：虽然某些情况下，JSON 编码自动会将 `[]byte` 表示为 base64 字符串，但 JSON 自身并无二进制的类型表示
    - 它无法表示具体 field 的其它信息，比如单位、指标类型（gauge/count/...）等
<!-- markdownlint-enable -->

#### PB-JSON Body {#api-v1-write-body-pbjson-protocol}

[:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0) · [:octicons-beaker-24: Experimental](index.md#experimental)

由于简单 JSON 有其自身缺点，建议使用另一种 JSON 形式，其结构如下：

```json
[
  {
    "name": "point-1", # 指标集名字
    "fields": [...], # 具体字段列表，包括 Field 和 Tag
    "time": "1709523668830398000"
  },
  {
    # another point...
  }
]
```

其中单个 field 结构如下：

```json
{
  "key"    : "field-name",        # 字段名（必填）
  "x"      : <value>,             # 字段值，其类型视 x 而定（必填）
  "type"   : "<COUNT/GAUGE/...>", # 指标类型（选填）
  "unit"   : "<kb/s/...>"         # 指标单位（选填）
  "is_tag" : true/false           # 是否是 tag（选填）
}
```

此处 `x` 有几种选项，列表如下

- `b`：表示这个 `key` 的数值是 boolean 值
- `d`：表示这个 `key` 的数值是串字节流，可能是二进制（`[]byte`），在 JSON 中，它必须用 base64 编码
- `f`：表示这个 `key` 的数值是一个浮点类型（float64）
- `i`：表示这个 `key` 的数值是一个有符号整型（int64）
- `s`：表示这个 `key` 的数值是一个字符串类型（string）
- `u`：表示这个 `key` 的数值是一个无符号整型（uint64）
- `a`: 表示这个 `key` 数值是一个动态类型（`any`），目前它只支持数组。它有两个二级字段：
    - `@type`：字符串，值固定为 `type.googleapis.com/point.Array`
    - `arr`：对象数组，数组中每个元素形式为 `{"x": <value>}`，此处 `x` 就是上面几种基础类型（`f/i/u/s/d/b`），但不含 `a`。此处，每个元素的 `x` 必须一致

<!-- markdownlint-disable MD046 -->
???+ warning

    此处的 `i` 和 `u` 以及每个 Point 的 `time` 字段值在 JSON 中均为字符串表示
<!-- markdownlint-enable -->

以下是一个具体 JSON 的示例：

```json
[
  {
    "name": "abc",
    "fields": [
      {
        "key": "say",
        "s": "hello"
      },
      {
        "key": "some-flag",
        "b": false
      },
      {
        "key": "binary-data",
        "d": "aGVsbG8gd29ybGQ="
      },
      {
        "key": "int-arr",
        "a": {
          "@type": "type.googleapis.com/point.Array",
          "arr": [
            { "i": "1" },
            { "i": "2" },
            { "i": "3" }
          ]
        }
      },
      {
        "key": "large-int",
        "i": "1234567890"
      },
      {
        "key": "large-bytes",
        "u": "1234567890",
        "type": "COUNT",
        "unit": "kb"
      },
      {
        "key": "some-tag",
        "s": "v1",
        "is_tag": true
      },
      {
        "key": "pi",
        "f": 3.14
      }
    ],
    "time": "1709523668830398000"
  }
]
```

---

<!-- markdownlint-disable MD046 -->
???+ attention

    - 所有 Body，不管是行协议还是其它两种 JSON 格式，都是数组结构，即每次上传至少一个 Point
    - 对于 JSON 形式的 Body，必须在 Header 中标注 `Content-Type: application/json`，否则 Datakit 以行协议来解析
    - field 中数组支持要求 [:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0) 以上（含）版本才支持
    - 相比行协议的 Body，JSON 形式的 body 性能较差，大概有 7~8 倍的差距
<!-- markdownlint-enable -->

---

### 数据类型分类 {#category}

DataKit 中主要有如下数据类型（以简称字母序排列）：

| 简称 | 名称            | URL 表示                  | 说明               |
| ---  | ---             | ---                       | ---                |
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

---

### DataKit 数据结构约束 {#point-limitation}

1. 所有种类的 Point，如果缺少 measurement（或者 measurement 为空字符串），自动补全 `measurement` 值为 `__default`
1. 时序类 Point（M），field 中不允许有字符串值，Datakit 会自动丢弃它们
1. 非时序类 Point，tag key 和 field key 中不允许出现 `.` 字符，Datakit 会自动将其替换成 `_`
1. 日志类 Point（L），如果缺少 `status` 字段（即 tag 和 field 中都不存在），Datakit 会自动将其置为 `unknown`
1. 对象类 Point （O/CO），如果缺少 `name` 字段（即 tag 和 field 中都不存在），Datakit 会自动将其置为 `default`
1. Tag 和 Field 之间的 key 不允许重名，即同一个 key 不能在 Tag 和 Field 中同时出现，否则，具体哪个 key 的值被写入是未定义的
1. Tag 或 Field 内部不允许出现同名 key，即同一个 key 不能在 Tag/Field 中出现多次，对于同名 key，将仅保留其中一个，具体哪一个也是未定义的
1. Tag 个数不超过 256 个，超过个数后将截掉尾部多余的 Tag
1. Field 个数不超过 1024 个，超过个数后将截掉尾部多余的 Field
1. Tag/Field Key 长度不超过 256 字节，超过长度时，将进行截断处理
1. Tag Value 长度不超过 1024 字节，超过长度时，将进行截断处理
1. Field Value 为字符串或字节流时，其长度不能超过 32M(32x1024x1024) 字节，超过长度时，将进行截断处理
1. 如果 field value 是空值（`null/nil` 等），最终行为是未定义的

---

### 行协议报错分析 {#line-proto-parse-error}

[:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)

如果上报的行协议有误，Datakit API 将返回对应的错误码以及出错详情。

假定我们将如下行协议内容通过 HTTP POST 发送给 Datakit。此处行协议有俩处错误，第二条和第四条的 `t2` 缺少 tag 值。

```not-set
# path/to/some/file.data
some1,t1=1,t2=v2 f1=1i,f2=3
some2,t1=1,t2 f1=1i,f2=3
some3,t1=1,t2=v3 f1=1i,f2=3
some2,t1=1,t2 f1=1i,f2=
```

```shell
$ curl -s http://datakit-ip:9529/v1/write/logging --data-binary "@path/to/some/file.data"

{
  "error_code": "datakit.invalidLinePoint",
  "message": "invalid lineprotocol: unable to parse 'some2,t1=1,t2 f1=1i,f2=3'(pos: 29): missing tag value\nunable to parse 'some2,t1=1,t2 f1=1i,f2='(pos: 82): missing tag value\nwith 2 point parse ok, 2 points failed. Origin data: \"some1,t1=1,t2=v2 f1=1i,f2=3\\nsome2,t1=1,t2 f1=1i,f2=3\\nsome3,t1=1,t2=v3 f1=1i,f2=3\\nsome2,t1=1,t2 f1=1i,f2=\\n\""
}
```

<!-- markdownlint-disable MD046 -->
???+ tips

    为了更好展示请求结果 中的 JSON，可以用工具 [jq](https://jqlang.github.io/jq/download/){:target="_blank"}，比如上面的复杂 `message` 字段，可以直接通过 jq 提取出纯文本：

    ```shell
    $ curl -s http://datakit-ip:9529/v1/write/logging --data-binary "@path/to/some/file.data" | jq -r .message
    invalid lineprotocol: unable to parse 'some2,t1=1,t2 f1=1i,f2=3'(pos: 29): missing tag value
    unable to parse 'some2,t1=1,t2 f1=1i,f2='(pos: 82): missing tag value
    with 2 point parse ok, 2 points failed. Origin data: "some1,t1=1,t2=v2 f1=1i,f2=3\nsome2,t1=1,t2 f1=1i,f2=3\nsome3,t1=1,t2=v3 f1=1i,f2=3\nsome2,t1=1,t2 f1=1i,f2=\n"
    ```
<!-- markdownlint-enable -->

此处 `message` 展开为：

```not-set
invalid lineprotocol: unable to parse 'some2,t1=1,t2 f1=1i,f2=3'(pos: 29): missing tag value
unable to parse 'some2,t1=1,t2 f1=1i,f2='(pos: 82): missing tag value
with 2 point parse ok, 2 points failed. Origin data: "some1,t1=1,t2=v2 f1=1i,f2=3\nsome2,t1=1,t2 f1=1i,f2=3\nsome3,t1=1,t2=v3 f1=1i,f2=3\nsome2,t1=1,t2 f1=1i,f2=\n"
```

`message` 解读：

- 由于有两处错误，故返回的信息中有俩个 `unable to parse...`。在每个错误后面，会附上本条行协议所在原始数据的位置偏移（`pos`），便于排错。
- 在返回的错误信息中会展示解析成功和失败的点数
- `Origin data...` 附上了原始的 HTTP Body（如果其中带二进制，则会以类似 `\x00\x32\x54...` 等 16 进制形式展示）

在 Datakit 日志中，如果行协议有误，也会记录这里 `message` 中的相关内容。

### 验证上传的数据 {#review-post-point}

通过 `echo` 参数可以回看上传的数据，用于调试数据是否按照预期的方式处理：

<!-- markdownlint-disable MD046 -->
=== "高级 JSON 形式（`ebco=pbjson`）"

    以[高级 JSON 方式](apis.md#api-v1-write-body-pbjson-protocol)展示。如果 Point 结构被自动纠正，JSON 中具体 Point 上会带一个 `warns` 字段，以表示这个 Point 被纠正的原因。
    
    比如，日志数据中，不允许字段的 Key 带 `.` 字段，Datakit 会自动将其转换成 `_`，此时回看的 JSON 中会额外带上 `warns` 信息：


    ```json
    [
       {
           "name": "...",
           "fields": [...],
           "time": "...",
           "warns": [
             {
                 "type": "dot_in_key",
                 "message": "invalid field key `some.field`: found `.'"
             }
           ]
       }
    ]
    ```

=== "普通 JSON（`echo=json`）"

    参见[普通 JSON 格式](apis.md#api-v1-write-body-json-protocol)

=== "行协议（`echo=lp`）"

    参见[行协议格式](apis.md#api-v1-write-body-line-protocol)
<!-- markdownlint-enable -->

---

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
        "ws_uuid": "wksp_2dc431d6693xxxxxxxxxxxxxxxxxxxxx",
        "bill_state": "normal",
        "ver_type": "pay",
        "token": "tkn_2dc438bxxxxxxxxxxxxxxxxxxxxxxxxx",
        "db_uuid": "ifdb_c0fsxxxxxxxxxxxxxxxx",
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

| 名称                     | 说明                                                                                                                                                                                                                         |
| :---                     | ---                                                                                                                                                                                                                          |
| `conditions`             | 额外添加条件表达式，使用 DQL 语法，例如 `hostname="cloudserver01" OR system="ubuntu"`。与现有 `query` 中的条件表达式成 `AND` 关系，且会在最外层添加括号避免与其混乱                                                          |
| `disable_multiple_field` | 是否禁用多字段。当为 true 时，只能查询单个字段的数据（不包括 time 字段），默认为 `false`                                                                                                                                     |
| `disable_slimit`         | 是否禁用默认 SLimit，当为 true 时，将不添加默认 SLimit 值，否则会强制添加 SLimit 20，默认为 `false`                                                                                                                          |
| `echo_explain`           | 是否返回最终执行语句（返回 JSON 数据中的 `raw_query` 字段）                                                                                                                                                                  |
| `highlight`              | 高亮搜索结果                                                                                                                                                                                                                 |
| `limit`                  | 限制单个时间线返回的点数，将覆盖 DQL 中存在的 limit                                                                                                                                                                          |
| `max_duration`           | 限制最大查询时间，支持单位 `ns/us/ms/s/m/h/d/w/y` ，例如 `3d` 是 3 天，`2w` 是 2 周，`1y` 是 1 年。默认是 1 年，此参数同样会限制 `time_range` 参数                                                                           |
| `max_point`              | 限制聚合最大点数。在使用聚合函数时，如果聚合密度过小导致点数太多，则会以 `(end_time-start_time)/max_point` 得到新的聚合间隔将其替换                                                                                          |
| `offset`                 | 一般跟 limit 配置使用，用于结果分页                                                                                                                                                                                          |
| `orderby`                | 指定 `order by` 参数，内容格式为 `map[string]string` 数组，`key` 为要排序的字段名，`value` 只能是排序方式即 `asc` 和 `desc`，例如 `[ { "column01" : "asc" }, { "column02" : "desc" } ]`。此条会替换原查询语句中的 `order by` |
| `queries`                | 基础查询模块，包含查询语句和各项附加参数                                                                                                                                                                                     |
| `query`                  | DQL 查询语句（DQL [文档](../dql/define.md)）                                                                                                                                                                                 |
| `search_after`           | 深度分页，第一次调用分页的时候，传入空列表：`"search_after": []`，成功后服务端会返回一个列表，客户端直接复用这个列表的值再次通过 `search_after` 参数回传给后续的查询即可                                                     |
| `slimit`                 | 限制时间线个数，将覆盖 DQL 中存在的 `slimit`                                                                                                                                                                                 |
| `time_range`             | 限制时间范围，采用时间戳格式，单位为毫秒，数组大小为 2 的 int，如果只有一个元素则认为是起始时间，会覆盖原查询语句中的查询时间区间                                                                                            |

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

| 参数           | 描述                                                                          | 类型       |
| ---:           | ---                                                                           | ---        |
| `object_class` | 表示 `labels` 所关联的 `object` 类型，如 `HOST`                               | `string`   |
| `object_name`  | 表示 `labels` 所关联的 `object` 名称，如 `host-123`                           | `string`   |
| `key`          | 表示 `labels` 所关联的 `object` 的具体字段名，如进程名字段 `process_name`     | `string`   |
| `value`        | 表示 `labels` 所关联的 `object` 的具体字段值，如进程名为 `systemsoundserverd` | `void`     |
| `labels`       | `labels` 列表，一个 `string` 数组                                             | `[]string` |

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

| 参数           | 描述                                                                          | 类型     |
| ---:           | ---                                                                           | ---      |
| `object_class` | 表示 `labels` 所关联的 `object` 类型，如 `HOST`                               | `string` |
| `object_name`  | 表示 `labels` 所关联的 `object` 名称，如 `host-123`                           | `string` |
| `key`          | 表示 `labels` 所关联的 `object` 的具体字段名，如进程名字段 `process_name`     | `string` |
| `value`        | 表示 `labels` 所关联的 `object` 的具体字段值，如进程名为 `systemsoundserverd` | `void`   |

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

提供远程调试拨测的功能，可通过[环境变量](../integrations/dialtesting.md#env)来控制禁拨网络。

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
| ---: | --- | --- |
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

| 参数       | 描述                                                    | 类型     |
| ---:       | ---                                                     | ---      |
| `token`    | `datakit.conf` 配置中的 `dataway` 地址中包含的 token    | `string` |
| `app_id`   | 用户访问应用唯一 ID 标识，如 `test-sourcemap`           | `string` |
| `env`      | 应用的部署环境，如 `prod`                               | `string` |
| `version`  | 应用的版本，如 `1.0.0`                                  | `string` |
| `platform` | 应用类型， 可选值 `web/miniapp/android/ios`, 默认 `web` | `string` |

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

| 参数          | 描述                                                    | 类型     |
| ---:          | ---                                                     | ---      |
| `error_stack` | error 的堆栈信息                                        | `string` |
| `app_id`      | 用户访问应用唯一 ID 标识，如 `test-sourcemap`           | `string` |
| `env`         | 应用的部署环境，如 `prod`                               | `string` |
| `version`     | 应用的版本，如 `1.0.0`                                  | `string` |
| `platform`    | 应用类型， 可选值 `web/miniapp/android/ios`, 默认 `web` | `string` |

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

## `/v1/global/host/tags` | `GET` {#api-global-host-tags-get}

获取 global-host-tags。

请求示例：

``` shell
curl 127.0.0.1:9529/v1/global/host/tags
```

成功返回示例：

``` json
status_code: 200
Response: {
    "host-tags": {
        "h": "h",
        "host": "host-name"
    }
}
```

## `/v1/global/host/tags` | `POST` {#api-global-host-tags-post}

创建或者更新 global-host-tags。

请求示例：

``` shell
curl -X POST "127.0.0.1:9529/v1/global/host/tags?tag1=v1&tag2=v2"
```

成功返回示例：

``` json
status_code: 200
Response: {
    "dataway-tags": {
        "e": "e",
        "h": "h",
        "tag1": "v1",
        "tag2": "v2",
        "host": "host-name"
    },
    "election-tags": {
        "e": "e"
    },
    "host-tags": {
        "h": "h",
        "tag1": "v1",
        "tag2": "v2",
        "host": "host-name"
    }
}
```

修改成功后，如果是主机模式下，修改内容会持久化到配置文件 `datakit.conf` 中。

## `/v1/global/host/tags` | `DELETE` {#api-global-host-tags-delete}

删除部分 global-host-tags。

请求示例：

``` shell
curl -X DELETE "127.0.0.1:9529/v1/global/host/tags?tags=tag1,tag3"
```

成功返回示例：

``` json
status_code: 200
Response: {
    "dataway-tags": {
        "e": "e",
        "h": "h",
        "host": "host-name"
    },
    "election-tags": {
        "e": "e"
    },
    "host-tags": {
        "h": "h",
        "host": "host-name"
    }
}
```

修改成功后，如果是主机模式下，修改内容会持久化到配置文件 `datakit.conf` 中。

## `/v1/global/election/tags` | `GET` {#api-global-election-tags-get}

获取 global-election-tags。

请求示例：

``` shell
curl 127.0.0.1:9529/v1/global/election/tags
```

成功返回示例：

``` json
status_code: 200
Response: {
    "election-tags": {
        "e": "e"
    }
}
```

## `/v1/global/election/tags` | `POST` {#api-global-election-tags-post}

创建或者更新 global-election-tags。

请求示例：

``` shell
curl -X POST "127.0.0.1:9529/v1/global/election/tags?tag1=v1&tag2=v2"
```

成功返回示例：

``` json
status_code: 200
Response: {
    "dataway-tags": {
        "e": "e",
        "h": "h",
        "tag1": "v1",
        "tag2": "v2",
        "host": "host-name"
    },
    "election-tags": {
        "tag1": "v1",
        "tag2": "v2",
        "e": "e"
    },
    "host-tags": {
        "h": "h",
        "host": "host-name"
    }
}
```

修改成功后，如果是主机模式下，修改内容会持久化到配置文件 `datakit.conf` 中。

当全局 `global-election-enable = false` 禁止执行本指令，失败返回示例：

``` json
status_code: 500
Response: {
    "message": "Can't use this command when global-election is false."
}
```

## `/v1/global/election/tags` | `DELETE` {#api-global-election-tags-delete}

删除部分 global-election-tags。

请求示例：

``` shell
curl -X DELETE "127.0.0.1:9529/v1/global/election/tags?tags=tag1,tag3"
```

成功返回示例：

``` json
status_code: 200
Response: {
    "dataway-tags": {
        "e": "e",
        "h": "h",
        "host": "host-name"
    },
    "election-tags": {
        "e": "e"
    },
    "host-tags": {
        "h": "h",
        "host": "host-name"
    }
}
```

修改成功后，如果是主机模式下，修改内容会持久化到配置文件 `datakit.conf` 中。

当全局 `global-election-enable = false` 禁止执行本指令，失败返回示例：

``` json
status_code: 500
Response: {
    "message": "Can't use this command when global-election is false."
}
```

## 延伸阅读 {#more-reading}

- [API 访问设置](datakit-conf.md#config-http-server)
- [API 限流配置](datakit-conf.md#set-http-api-limit)
- [API 安全控制](../integrations/rum.md#security-setting)
