# DataKit API

---

This document primarily describes the HTTP API interfaces provided by DataKit.

## API Overview {#intro}

DataKit currently only supports HTTP interfaces, mainly involving data writing and data querying.

### Obtaining the Remote DataKit Version Number {#api-get-dk-version}

There are two ways to obtain the version number:

- Request the DataKit ping interface: `curl http://ip:9529/v1/ping`
- In the response header of each of the following API requests, the current DataKit version for the request can be identified through `X-DataKit`

## `/v1/write/:category` | `POST` {#api-v1-write}

This API is used to upload(`POST`) various data (`category`) to DataKit. The URL parameters are explained as follows:

**`category`**

- Type: string
- Required: N
- Default value: -
- Description: Currently only supports `metric,logging,rum,object,custom_object,keyevent`, for example `metric`, the URL should be written as `/v1/write/metric`

**`dry`** [:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)

- Type: bool
- Required: N
- Default value: false
- Description: Test mode, just POST Point to Datakit, not actually uploaded to the Guance Cloud

**`echo`** [:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)

- Type: enum
- Required: N
- Default value: -
- Description: Optional values `lp/json/pbjson`, `lp` indicates that the uploaded Point is represented in line protocol format in the returned Body, followed by [normal JSON](apis.md#api-v1-write-body-json-protocol) and [PB-JSON](apis.md#api-v1-write-body-pbjson-protocol)

**`encoding`** [:octicons-tag-24: Version-1.60.0](changelog.md#cl-1.60.0)

- Type: string
- Required: N
- Default value: -
- Description: Supports `gzip`, `deflate`, `br`, and `zstd` compression methods. If this parameter is passed, DataKit will automatically decompress the request body.

**`global_election_tags`** [:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)

- Type: bool
- Required: N
- Default value: false
- Description: Whether to append [global election tags](datakit-conf.md#set-global-tag)

**`ignore_global_host_tags`** [:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)

- Type: bool
- Required: N
- Default value: false
- Description: Whether to ignore the [global host tags](datakit-conf.md#set-global-tag) on DataKit. By default, the data written by this interface will carry the global host tag

**`input`** [:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)

- Type: string
- Required: N
- Default value: `datakit-http`
- Description: Data source name, which will be displayed on the Datakit monitor for debugging

**`loose`** [:octicons-tag-24: Version-1.4.11](changelog.md#cl-1.4.11)

- Type: bool
- Required: N
- Default value: true
- Description: Whether to be in loose mode, for some non-compliant Points, DataKit will try to fix them

**`precision`** [:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)

- Type: enum
- Required: N
- Default value: -
- Description: Data precision (supports `n/u/ms/s/m/h`). If the parameter is not passed in, the timestamp precision will be automatically recognized

**`source`**

- Type: string
- Required: N
- Default value: -
- Description: If `source` is not specified (or the corresponding *source.p* does not exist or is invalid), the uploaded Point data will not execute the Pipeline

**`strict`** [:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9)

- Type: bool
- Required: N
- Default value: false
- Description: Strict mode, for some non-compliant line protocols, the API directly reports an error and tells the specific reason

<!-- markdownlint-disable MD046 -->
???+ attention

    - The following parameters have been deprecated:

        - `echo_line_proto` : Replace with the `echo` parameter
        - `echo_json`       : Replace with the `echo` parameter

    - Although multiple parameters are of boolean type, if you do not need to enable the corresponding option, do not pass a `false` value. The API will only check for the presence of a value on the parameter and will not consider its content.

    - Automatic recognition of time precision (`precision`) refers to guessing the likely time granularity based on the input timestamp value. While it cannot be mathematically guaranteed to be correct, it is sufficient for everyday use. For example, for a timestamp like 1716544492, it is interpreted as seconds, whereas 1716544492000 would be interpreted as milliseconds, and so on.

    - If the data point does not include a timestamp, the timestamp of the machine where Datakit is located will be used.
    - Although the current protocol supports both binary format and any format types, Kodo **has not yet** supported the upload of these two types of data.
<!-- markdownlint-enable -->

### Body Description {#api-v1-write-body}

The HTTP body supports line protocol as well as two forms of JSON.

#### Line Protocol Body {#api-v1-write-body-line-protocol}

A single line protocol format is as follows:

```text
measurement,<tag-list> <field-list> timestamp
```

Multiple line protocols are separated by newlines:

```text
measurement_1,<tag-list> <field-list> timestamp
measurement_2,<tag-list> <field-list> timestamp
```

Where:

- `measurement` is the name of the measurement set, which represents a collection of metrics, such as the `disk` measurement set, which may include metrics like `free/used/total`, etc.
- `<tag-list>` is a list of tags separated by `,`. A single tag is in the form of `key=value`, and the `value` is considered a string. In the line protocol, `<tag-list>` is **optional**
- `<field-list>` is a list of metrics separated by `,`. In the line protocol, `<field-list>` is **required**. A single metric is in the form of `key=value`, and the `value` format depends on its type, as follows:
    - int Example: `some_int=42i`, which means appending an `i` after the integer value to indicate
    - uint Example: `some_uint=42u`, which means appending a `u` after the integer value to indicate
    - float Example: `some_float_1=3.14,some_float_2=3`, where `some_float_2` is an integer 3, but it is still considered a float
    - string Example: `some_string="hello world"`, string values need to be enclosed in quotes on both ends
    - bool Example: `some_true=T,some_false=F`, where `T/F` can also be represented as `t/f/true/false` respectively
    - Binary Example: `some_binary="base-64-encode-string"b`, binary data (text byte stream `[]byte`, etc.) needs to be base64 encoded to be represented in the line protocol, similar to string representation, but with a `b` appended at the end to identify
    - Array Example: `some_array=[1i,2i,3i]`, note that the type within the array can only be a basic type (`int/uint/float/boolean/string/[]byte`, **excluding arrays**), and the types must be consistent, such as `invalid_array=[1i,3.14,"string"]` which is currently unsupported
- `timestamp` is an integer timestamp, by default, Datakit processes this timestamp in nanoseconds, if the original data is not in nanoseconds, the actual timestamp precision needs to be specified through the request parameter `precision`. In the line protocol, `timestamp` is optional, if the data does not include a timestamp, Datakit takes the time it receives as the current line protocol time.

The parts between them are:

- `measurement` and `<tag-list>` are separated by `,`
- `<tag-list>` and `<field-list>` are separated by a single space
- `<field-list>` and `timestamp` are separated by a single space
- In the line protocol, if there is a `#` at the beginning, it is considered a comment and will actually be ignored by the parser

Here are some simple line protocol examples:

```text
# Normal example
some_measurement,host=my_host,region=my_region cpu_usage=0.01,memory_usage=1048576u 1710321406000000000

# No tag example
some_measurement cpu_usage=0.01,memory_usage=1048576u 1710321406000000000

# No timestamp example
some_measurement,host=my_host,region=my_region cpu_usage=0.01,memory_usage=1048576u

# All basic types included
some_measurement,host=my_host,region=my_region float=0.01,uint=1048576u,int=42i,string="my-host",boolean=T,binary="aGVsbG8="b,array=[1.414,3.14] 1710321406000000000
```

Some special escapes for field names and field values are:

- `measurement` needs to escape `,`
- Tag key and field key need to escape `=`,`,` and spaces
- `measurement`, tag key, and field key must not contain newlines (`\n`)
- Tag value must not contain newlines (`\n`), newlines in field values do not need to be escaped
- If the field value is a string, if there is a `"` character, it also needs to be escaped

#### JSON Body {#api-v1-write-body-json-protocol}

Compared to the line protocol, the JSON format does not require much escaping, a simple JSON format is as follows:

```json
[
    {
        "measurement": "metric set name",

        "tags": {
            "key": "value",
            "another-key": "value"
        },

        "fields": {
            "key": value,
            "another-key": value # Here the value can be number/bool/string/list
        },

        "time": unix-timestamp
    },

    {
        # another-point...
    }
]
```

Here is a simple JSON example:

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

    Although this JSON structure is simple, it has several drawbacks:

    - It cannot distinguish between int/uint/float numeric types. For example, for all numbers, JSON defaults to handling them as floats, and for the number 42, JSON cannot distinguish whether it is signed or unsigned.
    - It does not support representing binary data (`[]byte`): Although in some cases, JSON encoding will automatically represent `[]byte` as a base64 string, JSON itself has no binary type representation.
    - It cannot represent other information for specific fields, such as units, metric types (gauge/count/...), etc.
<!-- markdownlint-enable -->

#### PB-JSON Body {#api-v1-write-body-pbjson-protocol}

[:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0) · [:octicons-beaker-24: Experimental](index.md#experimental)

Due to the inherent shortcomings of simple JSON, it is recommended to use another JSON format, which has the following structure:

```json
[
  {
    "name": "point-1", # Metric set name
    "fields": [...], # List of specific fields, including Field and Tag
    "time": "1709523668830398000"
  },
  {
    # another point...
  }
]
```

The structure of a single field is as follows:

```json
{
  "key"    : "field-name",        # Field name (required)
  "x"      : <value>,             # Field value, the type depends on x (required)
  "type"   : "<COUNT/GAUGE/...>", # Metric type (optional)
  "unit"   : "<kb/s/...>"         # Metric unit (optional)
  "is_tag" : true/false           # Whether it is a tag (optional)
}
```

Here `x` has several options, listed as follows:

- `b`: Indicates that the value of this `key` is a boolean
- `d`: Indicates that the value of this `key` is a stream of bytes, which may be binary (`[]byte`), in JSON, it must be base64 encoded
- `f`: Indicates that the value of this `key` is a floating-point type (float64)
- `i`: Indicates that the value of this `key` is a signed integer (int64)
- `s`: Indicates that the value of this `key` is a string type (string)
- `u`: Indicates that the value of this `key` is an unsigned integer (uint64)
- `a`: Indicates that the value of this `key` is a dynamic type (`any`), currently it only supports arrays. It has two secondary fields:
    - `@type`: String, the value is fixed as `type.googleapis.com/point.Array`
    - `arr`: An array of objects, each element in the array is in the form of `{"x": <value>}`, where `x` is one of the above basic types ( `f/i/u/s/d/b` ), but not `a`. Here, the `x` of each element must be consistent

<!-- markdownlint-disable MD046 -->
???+ warning

    The `i` and `u` here and the `time` field value of each Point are represented as strings in JSON
<!-- markdownlint-enable -->

Here is a specific JSON example:

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

    - All Bodies, whether line protocol or the other two JSON formats, are array structures, that is, at least one Point is uploaded each time
    - For JSON format Bodies, it is necessary to mark `Content-Type: application/json` in the Header, otherwise Datakit will parse it as line protocol
    - Array support in the field requires version 1.30.0 or above (inclusive)
    - Compared to the line protocol Body, the performance of the JSON format Body is relatively poor, with a gap of about 7-8 times
<!-- markdownlint-enable -->

---

### Data Type Classification {#category}

In DataKit, there are mainly the following data types (listed in alphabetical order of abbreviation):

| Abbreviation | Name            | URL Representation        | Description                   |
| ---          | ---             | ---                       | ---                           |
| CO           | `custom_object` | `/v1/write/custom_object` | Custom object data            |
| E            | `keyevent`      | `/v1/write/keyevent`      | Event data                    |
| L            | `logging`       | `/v1/write/logging`       | Log data                      |
| M            | `metric`        | `/v1/write/metric`        | Time series data              |
| N            | `network`       | `/v1/write/network`       | Generally refers to eBPF data |
| O            | `object`        | `/v1/write/object`        | Object data                   |
| P            | `profiling`     | `/v1/write/profiling`     | Profiling data                |
| R            | `rum`           | `/v1/write/rum`           | RUM data                      |
| S            | `security`      | `/v1/write/security`      | Security inspection data      |
| T            | `tracing`       | `/v1/write/tracing`       | APM (Tracing) data            |

---

### DataKit Data Structure Constraints {#point-limitation}

1. For all types of Points, if the measurement is missing (or the measurement is an empty string), the `measurement` value will be automatically filled with `__default`
2. For time series Points (M), strings are not allowed in the field, and Datakit will automatically discard them
3. For non-time series Points, `.` characters are not allowed in tag keys and field keys, Datakit will automatically replace them with `_`
4. For log Points (L), if the `status` field is missing (i.e., it does not exist in both tags and fields), Datakit will automatically set it to `unknown`
5. For object Points (O/CO), if the `name` field is missing (i.e., it does not exist in both tags and fields), Datakit will automatically set it to `default`
6. Tag and Field keys must not have the same name, that is, the same key cannot appear in both Tags and Fields, otherwise, it is undefined which key's value will be written
7. The same key cannot appear multiple times within Tags or Fields, that is, the same key cannot appear multiple times in Tags/Fields, and only one of them will be retained, which one is also undefined
8. The number of Tags does not exceed 256, and the excess Tags will be truncated
9. The number of Fields does not exceed 1024, and the excess Fields will be truncated
10. The length of Tag/Field Key does not exceed 256 bytes, and it will be truncated if it exceeds the length
11. The length of the Tag Value does not exceed 1024 bytes, and it will be truncated if it exceeds the length
12. When the Field Value is a string or byte stream, its length must not exceed 32M (32x1024x1024) bytes, and it will be truncated if it exceeds the length
13. If the field value is a null value (`null/nil`, etc.), the final behavior is undefined

---

### Line Protocol Error Analysis {#line-proto-parse-error}

[:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)

If the line protocol uploaded is incorrect, the Datakit API will return the corresponding error code and error details.

Assuming we send the following line protocol content to Datakit via HTTP POST. There are two errors in this line protocol, the second and fourth `t2` are missing tag values.

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

    To better display the JSON in the request result, you can use the tool [jq](https://jqlang.github.io/jq/download/){:target="_blank"}, for example, the complex `message` field above can be directly extracted into pure text through jq:

    ```shell
    $ curl -s http://datakit-ip:9529/v1/write/logging --data-binary "@path/to/some/file.data" | jq -r .message
    invalid lineprotocol: unable to parse 'some2,t1=1,t2 f1=1i,f2=3'(pos: 29): missing tag value
    unable to parse 'some2,t1=1,t2 f1=1i,f2='(pos: 82): missing tag value
    with 2 point parse ok, 2 points failed. Origin data: "some1,t1=1,t2=v2 f1=1i,f2=3\nsome2,t1=1,t2 f1=1i,f2=3\nsome3,t1=1,t2=v3 f1=1i,f2=3\nsome2,t1=1,t2 f1=1i,f2=\n"
    ```
<!-- markdownlint-enable -->

Here, `message` expands to:

```not-set
invalid lineprotocol: unable to parse 'some2,t1=1,t2 f1=1i,f2=3'(pos: 29): missing tag value
unable to parse 'some2,t1=1,t2 f1=1i,f2='(pos: 82): missing tag value
with 2 point parse ok, 2 points failed. Origin data: "some1,t1=1,t2=v2 f1=1i,f2=3\nsome2,t1=1,t2 f1=1i,f2=3\nsome3,t1=1,t2=v3 f1=1i,f2=3\nsome2,t1=1,t2 f1=1i,f2=\n"
```

`message` interpretation:

- Since there are two errors, there are two `unable to parse...` in the returned information. Behind each error, the offset position of this line protocol in the original data ( `pos` ) will be attached to facilitate troubleshooting.
- The returned error message will show the number of points that were parsed successfully and failed.
- `Origin data...` attaches the original HTTP Body (if it contains binary, it will be displayed in hexadecimal form such as `\x00\x32\x54...`, etc.)

In the Datakit logs, if the line protocol is incorrect, the content of `message` here will also be recorded.

### Verifying Uploaded Data {#review-post-point}

No matter which method (`lp`/`pbjson`/`json`) is used to upload data, DataKit will *attempt to make some corrections* to the data. These corrections may not be as expected, but we can use the `echo` parameter to view the final data:

<!-- markdownlint-disable MD046 -->
=== "PB-JSON(`echo=pbjson`)"

    Compared to the other two methods, using the [PB-JSON](apis.md#api-v1-write-body-pbjson-protocol) method allows you to see the details and reasons for the correction. If the Point structure is automatically corrected, the Point will have a `warns` field to indicate the reason for the correction.

    For example, in logging data, field keys are not allowed to contain `.` characters. DataKit will automatically convert them to `_`. In this case, the JSON(pbjson) response body will include additional `warns` information:
    
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

=== "Regular JSON (`echo=json`)"

    See [regular JSON format](apis.md#api-v1-write-body-json-protocol)

=== "Line protocol (`echo=lp`)"

    See [line protocol format](apis.md#api-v1-write-body-line-protocol)
<!-- markdownlint-enable -->

---

## `/v1/ping` {#api-ping}

Detects whether DataKit is running at the target address and can obtain the DataKit start time and version information. Example:

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

Used to report errors from external collectors. Example:

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

View workspace information and data quota information. Example:

``` http
GET /v1/workspace HTTP/1.1

HTTP/1.1 200 OK

{
  "content": [
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

Use DQL to query data (only data from the workspace where this DataKit is located can be queried), example:

``` http
POST /v1/query/raw HTTP/1.1
Content-Type: application/json

{
    "queries":[
        {
            "query": "cpu:(usage_idle) LIMIT 1",  # DQL query statement (required)
            "conditions": "",                     # Additional DQL query conditions
            "max_duration": "1d",                 # Maximum time range
            "max_point": 0,                       # Maximum number of points
            "time_range": [],                     #
            "orderby": [],                        #
            "disable_slimit": true,               # Disable the default SLimit, when set to true, no default SLimit value will be added, otherwise SLimit 20 will be forcibly added
            "disable_multiple_field": true        # Disable multiple fields. When set to true, only data from a single field (excluding the time field) can be queried
        }
    ],
    "echo_explain":true
}
```

Parameter Description

| Name                     | Description                                                                                                                                                                                          |
| :---                     | ---                                                                                                                                                           |
| `conditions`             | Additional conditional expressions using DQL syntax, for example `hostname="cloudserver01" OR system="ubuntu"`. It has an `AND` relationship with the existing `query` conditions and will be enclosed in parentheses to avoid confusion with them. |
| `disable_multiple_field` | Whether to disable multiple fields. When set to true, only data from a single field (excluding the time field) can be queried, default is `false`.                                                                            |
| `disable_slimit`         | Whether to disable the default SLimit, when set to true, no default SLimit value will be added, otherwise SLimit 20 will be forcibly added, default is `false`.                                                                       |
| `echo_explain`           | Whether to return the final execution statement (returned in the `raw_query` field of the JSON data).                                                                                                               |
| `highlight`              | Highlight search results.                                                                                                                                                                         |
| `limit`                  | Limit the number of points returned by a single timeline, which will override the limit in DQL.                                                                                                      |
| `max_duration`           | Limit the maximum query time, supports units `ns/us/ms/s/m/h/d/w/y`, for example, `3d` is 3 days, `2w` is 2 weeks, `1y` is 1 year. The default is 1 year, and this parameter also limits the `time_range` parameter.        |
| `max_point`              | Limit the maximum number of aggregated points. When using aggregate functions, if the aggregation density is too small and results in too many points, it will replace it with a new aggregation interval of `(end_time-start_time)/max_point`. |
| `offset`                 | Generally used in conjunction with limit configuration for result pagination.                                                                                                                            |
| `orderby`                | Specify `order by` parameters, content format is an array of `map[string]string`, `key` is the name of the field to be sorted, and `value` can only be the sorting method, i.e., `asc` and `desc`, for example `[ { "column01" : "asc" }, { "column02" : "desc" } ]`. This will replace the original query statement's `order by`. |
| `queries`                | Basic query module, including query statements and various additional parameters.                                                                                                                                      |
| `query`                  | DQL query statement (DQL [documentation](../dql/define.md)).                                                                                                                                                                          |
| `search_after`           | Deep pagination, for the first call to pagination, pass in an empty list: `"search_after": []`, after a successful server response, the client directly reuses the value of this list through the `search_after` parameter for subsequent queries.                      |
| `slimit`                 | Limit the number of timelines, which will override the `slimit` in DQL.                                                                                                                                                                      |
| `time_range`             | Limit the time range, in timestamp format, unit is milliseconds, an array of int with a size of 2, if there is only one element, it is considered the start time, which will override the time range interval in the original query statement.                             |

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

Create or update object `labels`.

`request body` description

| Parameter       | Description | Type    |
| ---:            | ---         | ---     |
| `object_class`  | Indicates the type of `object` associated with `labels`, such as `HOST` | `string` |
| `object_name`   | Indicates the name of the `object` associated with `labels`, such as `host-123` | `string` |
| `key`           | Indicates the specific field name of the `object` associated with `labels`, such as the process name field `process_name` | `string` |
| `value`         | Indicates the specific field value of the `object` associated with `labels`, such as the process name `systemsoundserverd` | `void`   |
| `labels`        | `labels` list, an array of `string` | `[]string` |

Request example:

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

Success return example:

``` json
status_code: 200
{
    "content": {
        "_id": "375370265b0641xxxxxxxxxxxxxxxxxxxxxxxxxx"
    }
}
```

Failure return example:

``` json
status_code: 500
{
    "errorCode":"some-internal-error"
}
```

## `/v1/object/labels` | `DELETE` {#api-delete-object-labels}

Delete object `labels`.

`request body` description

| Parameter       | Description | Type    |
| ---:            | ---         | ---     |
| `object_class`  | Indicates the type of `object` associated with `labels`, such as `HOST` | `string` |
| `object_name`   | Indicates the name of the `object` associated with `labels`, such as `host-123` | `string` |
| `key`           | Indicates the specific field name of the `object` associated with `labels`, such as the process name field `process_name` | `string` |
| `value`         | Indicates the specific field value of the `object` associated with `labels`, such as the process name `systemsoundserverd` | `void`   |

Request example:

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

Success return example:

``` json
status_code: 200
{
    "content": {
        "msg": "delete success!"
    }
}
```

Failure return example:

```json
status_code: 500
{
    "errorCode": "some-internal-error"
}
```

## `/v1/pipeline/debug` | `POST` {#api-debug-pl}

Provides the functionality of remote debugging of the Pipeline.

Error information `PlError` structure:

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

Error information JSON example:

```json
{
  "pos_chain": [
    { // Error generation location (script termination)
      "file": "xx.p",    // File name or file path
      "ln":   15,        // Line number
      "col":  29,        // Column number
      "pos":  576,       // Absolute character position from 0 in the text
    },
    ... ,
    { // Starting point of the call chain
      "file": "b.p",
      "ln":   1,
      "col":  1,
      "pos":  0,
    },
  ],
  "error": "error msg"
}
```

Request example:

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
    "category": "<logging[metric, tracing, ...]>", # Log category, pass in log text, other categories need to pass in line protocol text
    "data": [ base64("raw-logging-data1"), ... ], # Can be log or line protocol
    "encode": "@data's character encoding",         # Default is utf8 encoding
    "benchmark": false,                  # Whether to enable benchmark
}
```

Normal return example:

``` http
HTTP/1.1 200 OK

{
    "content": {
        "cost": "2.3ms",
        "benchmark": BenchmarkResult.String(), # Return benchmark results
        "pl_errors": [],                       # PlError list generated during script parsing or inspection
        "plresults": [                         # Since logs may be multi-line, multiple cutting results will be returned here
            {
                "point": {
                  "name" : "Can be metric set name, log source, etc.",
                  "tags": { "key": "val", "other-key": "other-val"},
                  "fields": { "f1": 1, "f2": "abc", "f3": 1.2 }
                  "time": 1644380607,   # Unix timestamp (unit seconds), the front end can convert it to a readable date
                  "time_ns": 421869748, # The remaining nanoseconds time, which is convenient for accurately converting to a date, the complete nanosecond timestamp is 1644380607421869748
                }
                "dropped": false,  # Whether the result is marked for discard during pipeline execution
                "run_error": null  # If there is no error, the value is null
            },
            {  another-result },
            ...
        ]
    }
}
```

Error return example:

``` http
HTTP Code: 400

{
    "error_code": "datakit.invalidCategory",
    "message": "invalid category"
}
```

## `/v1/dialtesting/debug` | `POST` {#api-debug-dt}

Provides remote debugging functionality for dial testing, which can control the prohibition of network dialing through [environment variables](../integrations/dialtesting.md#env).

Request example:

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
        "enable_traceroute"  : true, // true represents checking, only useful for tcp, icmp
        "success_when_logic" : "",
        "SuccessWhen"        : []*HTTPSuccess ,
        "tags"               : map[string]string ,
        "labels"             : []string,
        "advance_options"    : *HTTPAdvanceOption,
    }
}
```

Normal return example:

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

Error return example:

``` http
HTTP Code: 400

{
    "error_code": "datakit.invalidClass",
    "message": "invalid class"
}
```

## `/v1/sourcemap` | `PUT` {#api-sourcemap-upload}

[:octicons-tag-24: Version-1.12.0](changelog.md#cl-1.12.0)

Upload sourcemap files, this interface requires the [RUM collector](../integrations/rum.md) to be enabled.

Request parameter description.

| Parameter      | Description | Type    |
| ---:           | ---         | ---     |
| `token`        | The token contained in the `dataway` address in `datakit.conf` configuration | `string` |
| `app_id`       | The unique ID identifier for user access to the application, such as `test-sourcemap` | `string` |
| `env`          | The deployment environment of the application, such as `prod` | `string` |
| `version`      | The version of the application, such as `1.0.0` | `string` |
| `platform`     | The type of application, optional values `web/miniapp/android/ios`, default `web` | `string` |

Request example:

``` shell
curl -X PUT "http://localhost:9529/v1/sourcemap?app_id=test_sourcemap&env=production&version=1.0.0&token=tkn_xxxxx&platform=web"  \
-F "file=@./sourcemap.zip" \
-H "Content-Type: multipart/form-data"
```

Success return example:

``` json
{
  "content": "uploaded to [/path/to/datakit/data/rum/web/test_sourcemap-production-1.0.0.zip]!",
  "errorMsg": "",
  "success": true
}
```

Failure return example:

``` json
{
  "content": null,
  "errorMsg": "app_id not found",
  "success": false
}
```

## `/v1/sourcemap` | `DELETE` {#api-sourcemap-delete}

[:octicons-tag-24: Version-1.16.0](changelog.md#cl-1.16.0)

Delete sourcemap files; this endpoint requires the [RUM collector](../integrations/rum.md) to be enabled.

Request parameter description:

| Parameter | Description | Type    |
| ---: | --- | --- |
| `token` | The token included in the `dataway` address in `datakit.conf` configuration | `string` |
| `app_id` | The unique ID identifier for user access to the application, such as `test-sourcemap` | `string` |
| `env` | The deployment environment of the application, such as `prod` | `string` |
| `version` | The version of the application, such as `1.0.0` | `string` |
| `platform` | The type of application, with optional values `web/miniapp/android/ios`, defaulting to `web` | `string` |

Request example:

``` shell
curl -X DELETE "http://localhost:9529/v1/sourcemap?app_id=test_sourcemap&env=production&version=1.0.0&token=tkn_xxxxx&platform=web"
```

Success return example:

``` json
{
  "content":"deleted [/path/to/datakit/data/rum/web/test_sourcemap-production-1.0.0.zip]!",
  "errorMsg":"",
  "success":true
}
```

Failure return example:

``` json
{
  "content": null,
  "errorMsg": "delete sourcemap file [/path/to/datakit/data/rum/web/test_sourcemap-production-1.0.0.zip] failed: remove /path/to/datakit/data/rum/web/test_sourcemap-production-1.0.0.zip: no such file or directory",
  "success": false
}
```

## `/v1/sourcemap/check` | `GET` {#api-sourcemap-check}

[:octicons-tag-24: Version-1.16.0](changelog.md#cl-1.16.0)

Verify if the sourcemap files are correctly configured; this endpoint requires the [RUM collector](../integrations/rum.md) to be enabled.

Request parameter description:

| Parameter | Description | Type    |
| ---: | --- | --- |
| `error_stack` | The error stack information | `string` |
| `app_id` | The unique ID identifier for user access to the application, such as `test-sourcemap` | `string` |
| `env` | The deployment environment of the application, such as `prod` | `string` |
| `version` | The version of the application, such as `1.0.0` | `string` |
| `platform` | The type of application, with optional values `web/miniapp/android/ios`, defaulting to `web` | `string` |

Request example:

``` shell
curl "http://localhost:9529/v1/sourcemap/check?app_id=test_sourcemap&env=production&version=1.0.0&error_stack=at%20test%20%40%20http%3A%2F%2Flocalhost%3A8080%2Fmain.min.js%3A1%3A48"
```

Success return example:

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

Failure return example:

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

Retrieve the Prometheus metrics exposed by Datakit.

## `/v1/global/host/tags` | `GET` {#api-global-host-tags-get}

Retrieve global-host-tags.

Request example:

``` shell
curl 127.0.0.1:9529/v1/global/host/tags
```

Success return example:

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

Create or update global-host-tags.

Request example:

``` shell
curl -X POST "127.0.0.1:9529/v1/global/host/tags?tag1=v1&tag2=v2"
```

Success return example:

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

After successful modification, if in host mode, the changes will be persisted to the configuration file `datakit.conf`.

## `/v1/global/host/tags` | `DELETE` {#api-global-host-tags-delete}

Delete some global-host-tags.

Request example:

``` shell
curl -X DELETE "127.0.0.1:9529/v1/global/host/tags?tags=tag1,tag3"
```

Success return example:

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

After successful modification, if in host mode, the changes will be persisted to the configuration file `datakit.conf`.

## `/v1/global/election/tags` | `GET` {#api-global-election-tags-get}

Retrieve global-election-tags.

Request example:

``` shell
curl 127.0.0.1:9529/v1/global/election/tags
```

Success return example:

``` json
status_code: 200
Response: {
    "election-tags": {
        "e": "e"
    }
}
```

## `/v1/global/election/tags` | `POST` {#api-global-election-tags-post}

Create or update global-election-tags.

Request example:

``` shell
curl -X POST "127.0.0.1:9529/v1/global/election/tags?tag1=v1&tag2=v2"
```

Success return example:

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

After successful modification, if in host mode, the changes will be persisted to the configuration file `datakit.conf`.

When the global `global-election-enable = false`, this command is prohibited, and the failure return example is:

``` json
status_code: 500
Response: {
    "message": "Can't use this command when global-election is false."
}
```

## `/v1/global/election/tags` | `DELETE` {#api-global-election-tags-delete}

Delete some global-election-tags.

Request example:

``` shell
curl -X DELETE "127.0.0.1:9529/v1/global/election/tags?tags=tag1,tag3"
```

Success return example:

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

After successful modification, if in host mode, the changes will be persisted to the configuration file `datakit.conf`.

When the global `global-election-enable = false`, this command is prohibited, and the failure return example is:

``` json
status_code: 500
Response: {
    "message": "Can't use this command when global-election is false."
}
```

## Further Reading {#more-reading}

- [API Access Settings](datakit-conf.md#config-http-server)
- [API Rate Limiting Configuration](datakit-conf.md#set-http-api-limit)
- [API Security Controls](../integrations/rum.md#security-setting)
