# DataKit API
---

This document mainly describes the HTTP API interfaces exposed by DataKit.

## API Overview {#intro}

Currently, DataKit only supports HTTP interfaces, mainly involving data writing and data querying.

## Data Writing APIs {#write-apis}

### `/v1/write/:category` {#api-v1-write}

This API is used to report various types of data (`category`) to DataKit and has several different usage methods:

- **Sending Line Protocol Data**

```shell
curl -X POST -d '<YOUR-LINEPROTOCOL-DATA>' http://localhost:9529/v1/write/metric
```

- **Sending Ordinary JSON Data**

```shell
curl -X POST -H "Content-Type: application/json" -d '<YOUR-JSON-DATA>' http://localhost:9529/v1/write/metric
```

- **Sending PBJSON Data**

```shell
curl -X POST -H "Content-Type: application/pbjson; proto=com.guance.Point" -d '<YOUR-PBJSON-DATA>' http://localhost:9529/v1/write/metric
```

The complete description of URL parameters is as follows:

> In the following `curl` examples, `category` is exemplified by `metric`, and the `Content-Type` header is omitted.

**`category`**

- Type: string
- Required: No
- Default Value: -
- Description: Currently, only `metric,logging,object,network,custom_object,security,rum` are supported. Taking `metric` as an example, the URL should be written as `/v1/write/metric`.
- Examples:
    - `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/metric"`
    - `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/logging"`
    - `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/object"`
    - `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/network"`
    - `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/custom_object"`
    - `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/security"`
    - `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/rum"`

**`dry`** [:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)

- Type: bool
- Required: No
- Default Value: false
- Description: Test mode. It only POSTs the Point to DataKit but does not actually upload it to <<<custom_key.brand_name>>>.
- Example: `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/metric&dry=true"`

**`echo`** [:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)

- Type: enum
- Required: No
- Default Value: -
- Description: Optional values are `lp/json/pbjson`. `lp` means representing the uploaded Point in the line protocol form in the returned Body. The following are [ordinary JSON](apis.md#api-v1-write-body-json-protocol) and [PBJSON](apis.md#api-v1-write-body-pbjson-protocol) respectively.
- Example: `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/metric&echo=pbjson"`

**`encoding`** [:octicons-tag-24: Version-1.62.0](changelog.md#cl-1.62.0)

- Type: string
- Required: No
- Default Value: -
- Description: Supports four compression methods: `gzip`, `deflate`, `br`, and `zstd`. If this parameter is passed in, DataKit will automatically decompress the request body.
- Example: `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/metric&encoding=gzip"`

**`global_election_tags`** [:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)

- Type: bool
- Required: No
- Default Value: false
- Description: Whether to append [global election tags](datakit-conf.md#set-global-tag).
- Example: `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/metric&global_election_tags=true"`

**`ignore_global_host_tags`** [:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)

- Type: bool
- Required: No
- Default Value: false
- Description: Whether to ignore the [global host tags](datakit-conf.md#set-global-tag) on DataKit. By default, the data written through this interface will carry the global host tags.
- Example: `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/metric&ignore_global_host_tags=true"`

**`input`** [:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)

- Type: string
- Required: No
- Default Value: `datakit-http`
- Description: The name of the data source, which will be displayed on the DataKit monitor for debugging convenience.
- Example: `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/metric&input=my-data-source"`

**`precision`** [:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)

- Type: enum
- Required: No
- Default Value: -
- Description: Data precision (supports `n/u/ms/s/m/h`). If the parameter is not passed in, the timestamp precision will be automatically recognized.
- Example: `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/metric&precision=ms"`

**`source`**

- Type: string
- Required: No
- Default Value: -
- Description: If `source` is not specified (or the corresponding *source.p* does not exist or is invalid), the uploaded Point data will not execute the Pipeline.
- Example: `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/metric&source=my-data-source-name"`

**`strict`** [:octicons-tag-24: Version-1.5.9](changelog.md#cl-1.5.9)

- Type: bool
- Required: No
- Default Value: false
- Description: Strict mode. For some non - compliant line protocols, the API will directly report an error and inform the specific reason.
- Example: `curl -X POST -d '<YOUR-DATA>' "http://localhost:9529/v1/write/metric&strict=true"`

<!-- markdownlint-disable MD046 -->
???+ warning

    - The following parameters have been deprecated [:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)
        - `echo_line_proto`: Replaced by the `echo` parameter.
        - `echo_json`: Replaced by the `echo` parameter.
    - Although multiple parameters are of the bool type, if you do not need to enable the corresponding option, do not pass in the `false` value. The API will only determine whether the corresponding parameter has a value, regardless of its content.
    - The automatic recognition of timestamp precision (`precision`) [:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0) means guessing the possible timestamp precision based on the incoming timestamp value. Mathematically, it cannot guarantee correctness, but it is sufficient for daily use. For example, for the timestamp 1716544492, its timestamp is judged as seconds, and for 1716544492000, it will be judged as milliseconds, and so on.
    - If there is no time in the data point, the timestamp of the machine where DataKit is located will be used as the standard.
    - Although the protocol currently supports binary format and any format, the central system does not yet support writing these two types of data. **Specially noted here**.
<!-- markdownlint-enable -->

#### Body Description {#api-v1-write-body}

The HTTP body supports the line protocol and two JSON forms.

##### Line Protocol Body {#api-v1-write-body-line-protocol}

The form of a single line protocol is as follows:

```text
measurement,<tag-list> <field-list> timestamp
```

Multiple line protocols are separated by line breaks:

```text
measurement_1,<tag-list> <field-list> timestamp
measurement_2,<tag-list> <field-list> timestamp
```

Among them:

- `measurement` is the name of the metric set, which represents the collective name of a set of metrics. For example, under the metric set name `disk`, there may be metrics such as `free/used/total`.
- `<tag-list>` is a list of tags, separated by `,`. A single tag is in the form of `key = value`, and here `value` is regarded as a string. In the line protocol, `<tag-list>` is **optional**.
- `<field-list>` is a list of metrics, separated by `,`. In the line protocol, `<field-list>` is **required**. A single metric is in the form of `key = value`, and the form of `value` depends on its type, as follows:
    - int example: `some_int = 42i`, that is, append an `i` after the integer value to indicate it.
    - uint example: `some_uint = 42u`, that is, append an `u` after the integer value to indicate it.
    - float example: `some_float_1 = 3.14,some_float_2 = 3`. Here, although `some_float_2` is the integer 3, it is still regarded as a float.
    - string example: `some_string = "hello world"`, the string value needs to be enclosed in `"` at both ends.
    - bool example: `some_true = T,some_false = F`. Here, `T/F` can also be represented by `t/f/true/false` respectively.
    - binary example: `some_binary = "base - 64 - encode - string"b`. Binary data (text byte stream `[]byte`, etc.) needs to be base64 - encoded to be represented in the line protocol. It is similar to the representation of a string, but with a `b` appended at the end to identify it.
    - array example: `some_array = [1i,2i,3i]`. Note that the types in the array can only be basic types (`int/uint/float/boolean/string/[]byte`, **excluding arrays**), and their types must be consistent. Arrays like `invalid_array = [1i,3.14,"string"]` are currently not supported.
- `timestamp` is an integer timestamp. By default, DataKit processes this timestamp in nanoseconds. If the original data is not in nanoseconds, the actual timestamp precision needs to be specified through the request parameter `precision`. In the line protocol, `timestamp` is optional. If there is no timestamp in the data, DataKit uses the received time as the current line protocol time.

The relationships between these parts are as follows:

- `measurement` and `<tag-list>` are separated by `,`.
- `<tag-list>` and `<field-list>` are separated by a single space.
- `<field-list>` and `timestamp` are separated by a single space.
- In the line protocol, if there is a `#` at the beginning, it is regarded as a comment and will actually be ignored by the parser.

The following are some simple examples of line protocols:

```text
# Ordinary example
some_measurement,host = my_host,region = my_region cpu_usage = 0.01,memory_usage = 1048576u 1710321406000000000
# Example without tags
some_measurement cpu_usage = 0.01,memory_usage = 1048576u 1710321406000000000
# Example without timestamp
some_measurement,host = my_host,region = my_region cpu_usage = 0.01,memory_usage = 1048576u
# Containing all basic types
some_measurement,host = my_host,region = my_region float = 0.01,uint = 1048576u,int = 42i,string = "my - host",boolean = T,binary = "aGVsbG8="b,array = [1.414,3.14] 1710321406000000000
```

Some special escapes for field names and field values:

- The `,` in `measurement` needs to be escaped.
- The `=`, `,`, and spaces in tag keys and field keys need to be escaped.
- No line breaks (`\n`) are allowed in `measurement`, tag keys, and field keys.
- No line breaks (`\n`) are allowed in tag values, and line breaks in field values do not need to be escaped.
- If the field value is a string and contains the `"` character, it also needs to be escaped.

##### JSON Body {#api-v1-write-body-json-protocol}

Compared with the line protocol, the JSON-formatted body does not require too many escapes. A simple JSON format is as follows:

```json
[
    {
        "measurement": "measurement name",
        "tags": {
            "key": "value",
            "another - key": "value"
        },
        "fields": {
            "key": value,
            "another - key": value // Here, value can be number/bool/string/list
        },
        "time": unix - timestamp
    },
    {
        # another - point...
    }
]
```

The following is a simple JSON example:

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
      "f3": "strval",
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

    Although this JSON structure is simple, it has several disadvantages:

    - It cannot distinguish between int/uint/float numerical types. For example, for all numerical values, JSON defaults to treating them as floats. For the value 42, JSON cannot distinguish whether it is signed or unsigned.
    - It does not support representing binary (`[]byte`) data. Although in some cases, JSON encoding will automatically represent `[]byte` as a base64 - encoded string, JSON itself has no binary type representation.
    - It cannot represent other information of specific fields, such as units, metric types (gauge/count/...), etc.
<!-- markdownlint-enable -->

##### PBJSON Body {#api-v1-write-body-pbjson-protocol}

[:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0) · [:octicons-beaker-24: Experimental](index.md#experimental)

Due to the shortcomings of simple JSON, it is recommended to use another JSON form with the following structure:

```json
[
  {
    "name": "point-1", // Name of the metric set
    "fields": [...], // Specific field list, including Field and Tag
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
  "key"    : "field-name",        // Field name (required)
  "x"      : <value>,             // Field value, and its type depends on x (required)
  "type"   : "<COUNT/GAUGE/...>", // Metric type (optional)
  "unit"   : "<kb/s/...>",         // Metric unit (optional)
  "is_tag" : true/false           // Whether it is a tag (optional)
}
```

Here, there are several options for "x", listed as follows:

- "b": Indicates that the value of this "key" is a boolean value.
- "d": Indicates that the value of this "key" is a byte stream, which may be binary (`[]byte`). In JSON, it must be encoded in base64.
- "f": Indicates that the value of this "key" is a floating - point type (float64).
- "i": Indicates that the value of this "key" is a signed integer type (int64).
- "s": Indicates that the value of this "key" is a string type (string).
- "u": Indicates that the value of this "key" is an unsigned integer type (uint64).
- "a": Indicates that the value of this "key" is a dynamic type ("any"), and currently it only supports arrays. It has two secondary fields:
    - "@type": A string with a fixed value of "type.googleapis.com/point.Array".
    - "arr": An array of objects. Each element in the array is in the form of `{"x": <value>}`, where "x" represents the above - mentioned basic types ("f/i/u/s/d/b"), excluding "a". Here, the "x" of each element must be the same.

<!-- markdownlint-disable MD046 -->
???+ warning

    The values of "i" and "u" here, as well as the "time" field value of each Point, are represented as strings in JSON.
<!-- markdownlint-enable -->

The following is a specific JSON example:

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
???+ warning

    - All request bodies, whether in line protocol or the other two JSON formats, are in array structure, meaning that at least one data point (Point) must be uploaded each time.
    - For JSON - formatted request bodies, the following annotations must be made in the request header (Header), otherwise DataKit will parse it as a line protocol:
          - JSON: `Content-Type: application/json`
          - PBJSON: `Content-Type: application/pbjson; proto=com.guance.Point`
    - The support for arrays in fields requires a version of 1.30.0 or above (including 1.30.0) [:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0).
    - Compared with the line - protocol request body, the performance of the JSON - formatted request body is relatively poor, with approximately a 7 - to 8 - fold difference.
<!-- markdownlint-enable -->

---

#### Data Type Classification {#category}

The following are the main data types in DataKit (listed in alphabetical order of abbreviations):

 | Abbreviation | Name            | URL Representation        | Description                   |
 | ---          | ---             | ---                       | ---                           |
 | CO           | `custom_object` | `/v1/write/custom_object` | Custom object data            |
 | E            | `keyevent`      | `/v1/write/keyevent`      | Event data                    |
 | L            | `logging`       | `/v1/write/logging`       | Log data                      |
 | M            | `metric`        | `/v1/write/metric`        | Time - series data            |
 | N            | `network`       | `/v1/write/network`       | Generally refers to eBPF data |
 | O            | `object`        | `/v1/write/object`        | Object data                   |
 | P            | `profiling`     | `/v1/write/profiling`     | Profiling data                |
 | R            | `rum`           | `/v1/write/rum`           | RUM data                      |
 | S            | `security`      | `/v1/write/security`      | Security inspection data      |
 | T            | `tracing`       | `/v1/write/tracing`       | APM (Tracing) data            |

---

#### DataKit Data Structure Constraints {#point-limitation}

1. For all types of Points, if the measurement is missing (or the measurement is an empty string), the value of `measurement` will be automatically filled with `__default`.
1. For time - series Points (M), string values are not allowed in fields, and DataKit will automatically discard them.
1. For non - time - series Points, the `.` character is not allowed in tag keys and field keys, and DataKit will automatically replace it with `_`.
1. For log - type Points (L), if the `status` field is missing (i.e., it does not exist in either tags or fields), DataKit will automatically set it to `unknown`.
1. For object - type Points (O/CO), if the `name` field is missing (i.e., it does not exist in either tags or fields), DataKit will automatically set it to `default`.
1. The keys between Tags and Fields are not allowed to have the same name. That is, the same key cannot appear in both Tags and Fields simultaneously. Otherwise, which key's value will be written is undefined.
1. Duplicate keys are not allowed within Tags or Fields. That is, the same key cannot appear multiple times in Tags/Fields. For duplicate keys, only one of them will be retained, and which one is retained is also undefined.
1. The number of Tags cannot exceed 256. If it exceeds, the extra Tags at the end will be truncated.
1. The number of Fields cannot exceed 1024. If it exceeds, the extra Fields at the end will be truncated.
1. The length of Tag/Field Keys cannot exceed 256 bytes. If it exceeds, it will be truncated.
1. The length of Tag Values cannot exceed 1024 bytes. If it exceeds, it will be truncated.
1. When the Field Value is a string or byte stream, its length cannot exceed 32M (32x1024x1024) bytes. If it exceeds, it will be truncated.
1. If the field value is a null value (`null/nil`, etc.), the final behavior is undefined.

---

#### Line Protocol Error Analysis {#line-proto-parse-error}

[:octicons-tag-24: Version-1.30.0](changelog.md#cl-1.30.0)

If the reported line protocol is incorrect, the DataKit API will return the corresponding error code and error details.

Suppose we send the following line - protocol content to DataKit via HTTP POST. There are two errors in this line protocol: the `t2` in the second and fourth lines lacks a tag value.

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
  "message": "invalid lineprotocol: unable to parse'some2,t1=1,t2 f1=1i,f2=3'(pos: 29): missing tag value\nunable to parse'some2,t1=1,t2 f1=1i,f2='(pos: 82): missing tag value\nwith 2 point parse ok, 2 points failed. Origin data: \"some1,t1=1,t2=v2 f1=1i,f2=3\\nsome2,t1=1,t2 f1=1i,f2=3\\nsome3,t1=1,t2=v3 f1=1i,f2=3\\nsome2,t1=1,t2 f1=1i,f2=\\n\""
}
```
<!-- markdownlint-disable MD046 -->
???+ tip

    To better display the JSON in the request result, you can use the tool [jq](https://jqlang.github.io/jq/download/){:target="_blank"}. For example, for the complex `message` field above, you can directly extract the plain text through jq:

    ```shell
    $ curl -s http://datakit-ip:9529/v1/write/logging --data-binary "@path/to/some/file.data" | jq -r.message
    invalid lineprotocol: unable to parse'some2,t1=1,t2 f1=1i,f2=3'(pos: 29): missing tag value
    unable to parse'some2,t1=1,t2 f1=1i,f2='(pos: 82): missing tag value
    with 2 point parse ok, 2 points failed. Origin data: "some1,t1=1,t2=v2 f1=1i,f2=3\nsome2,t1=1,t2 f1=1i,f2=3\nsome3,t1=1,t2=v3 f1=1i,f2=3\nsome2,t1=1,t2 f1=1i,f2=\n"
    ```
<!-- markdownlint-enable -->

Here, the `message` is expanded as follows:

```not-set
invalid lineprotocol: unable to parse'some2,t1=1,t2 f1=1i,f2=3'(pos: 29): missing tag value
unable to parse'some2,t1=1,t2 f1=1i,f2='(pos: 82): missing tag value
with 2 point parse ok, 2 points failed. Origin data: "some1,t1=1,t2=v2 f1=1i,f2=3\nsome2,t1=1,t2 f1=1i,f2=3\nsome3,t1=1,t2=v3 f1=1i,f2=3\nsome2,t1=1,t2 f1=1i,f2=\n"
```

Interpretation of the `message`:

- Since there are two errors, there are two `unable to parse...` in the returned information. After each error, the position offset (`pos`) of this line - protocol in the original data is attached to facilitate error - checking.
- The number of successfully and failed parsed points is displayed in the returned error message.
- `Origin data...` attaches the original HTTP Body (if it contains binary, it will be displayed in hexadecimal form like `\x00\x32\x54...`).

In the DataKit logs, if the line protocol is incorrect, the relevant content in this `message` will also be recorded.

#### Verifying Uploaded Data {#review-post-point}

Regardless of the method (`lp`/`pbjson`/`json`) used to write data, DataKit will *attempt to correct the data*. These corrections may not be as expected, but we can use the `echo` parameter to view the final data:

<!-- markdownlint-disable MD046 -->
=== "PBJSON Form (`echo=pbjson`)"

    Compared with the other two methods, through the [PBJSON](apis.md#api-v1-write-body-pbjson-protocol) method, you can know the details and reasons for the correction. If the Point structure is automatically corrected, the Point will carry a `warns` field to indicate the reason for the correction of this Point.

    For example, in log data, field keys are not allowed to have the `.` character. DataKit will automatically convert it to `_`. At this time, the `warns` information will be additionally included in the viewed JSON:

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

=== "Ordinary JSON (`echo=json`)"

    See [Ordinary JSON Format](apis.md#api-v1-write-body-json-protocol).

=== "Line Protocol (`echo=lp`)"

    See [Line Protocol Format](apis.md#api-v1-write-body-line-protocol).
<!-- markdownlint-enable -->

---

#### `/v1/object/labels` | `POST` {#api-object-labels}

Create or update the `labels` of an object.  Description of the `request body`:

 | Parameter      | Description                                                                                                                 | Type       |
 | ---:           | ---                                                                                                                         | ---        |
 | `object_class` | Represents the type of the `object` associated with `labels`, such as `HOST`                                                | `string`   |
 | `object_name`  | Represents the name of the `object` associated with `labels`, such as `host-123`                                            | `string`   |
 | `key`          | Represents the specific field name of the `object` associated with `labels`, such as the process name field `process_name`  | `string`   |
 | `value`        | Represents the specific field value of the `object` associated with `labels`, such as the process name `systemsoundserverd` | `void`     |
 | `labels`       | List of `labels`, an array of `string`                                                                                      | `[]string` |

Request example:

``` shell
curl -XPOST "http://localhost:9529/v1/object/labels" \
    -H 'Content-Type: application/json'  \
    -d'{
            "object_class": "host_processes",
            "object_name": "ubuntu20-dev_49392",
            "key": "host",
            "value": "ubuntu20-dev",
            "labels": ["l1","l2"]
        }'
```

Successful return example:

``` json
status_code: 200
{
    "content": {
        "_id": "375370265b0641xxxxxxxxxxxxxxxxxxxxxxxxxx"
    }
}
```

Failed return example:

``` json
status_code: 500
{
    "errorCode":"some-internal-error"
}
```

### `/v1/object/labels` {#api-delete-object-labels}

Delete the `labels` of an object.  Description of the `request body`:

 | Parameter      | Description                                                                                                                 | Type     |
 | ---:           | ---                                                                                                                         | ---      |
 | `object_class` | Represents the type of the `object` associated with `labels`, such as `HOST`                                                | `string` |
 | `object_name`  | Represents the name of the `object` associated with `labels`, such as `host-123`                                            | `string` |
 | `key`          | Represents the specific field name of the `object` associated with `labels`, such as the process name field `process_name`  | `string` |
 | `value`        | Represents the specific field value of the `object` associated with `labels`, such as the process name `systemsoundserverd` | `void`   |

Request example:

``` shell
curl -XDELETE "http://localhost:9529/v1/object/labels"  \
    -H 'Content-Type: application/json'  \
    -d'{
            "object_class": "host_processes",
            "object_name": "ubuntu20-dev_49392",
            "key": "host",
            "value": "ubuntu20-dev"
        }'
```

Request ok example:

``` json
status_code: 200
{
    "content": {
        "msg": "delete success!"
    }
}
```

Request fail example:

``` json
status_code: 500
{
    "errorCode": "some-internal-error"
}
```

## Tool related APIs {#tools-apis}

### `PUT /v1/sourcemap` {#api-sourcemap-upload}

[:octicons-tag-24: Version-1.12.0](changelog.md#cl-1.12.0)

Upload the sourcemap file. This interface requires the [RUM Collector](../integrations/rum.md) to be enabled.

Description of request parameters:

| Parameter  | Description                                                                                          | Type     |
| ---:       | ---                                                                                                  | ---      |
| `token`    | The token included in the `dataway` address in the `datakit.conf` configuration                      | `string` |
| `app_id`   | The unique ID of the application accessed by the user, such as `test-sourcemap`                      | `string` |
| `env`      | The deployment environment of the application, such as `prod`                                        | `string` |
| `version`  | The version of the application, such as `1.0.0`                                                      | `string` |
| `platform` | The type of the application. Optional values are `web/miniapp/android/ios`, and the default is `web` | `string` |

Request example:

``` shell
curl -X PUT "http://localhost:9529/v1/sourcemap?app_id=test_sourcemap&env=production&version=1.0.0&token=tkn_xxxxx&platform=web" \
-F "file=@./sourcemap.zip" \
-H "Content-Type: multipart/form-data"
```

Successful return example:

``` json
{
  "content": "uploaded to [/path/to/datakit/data/rum/web/test_sourcemap-production-1.0.0.zip]!",
  "errorMsg": "",
  "success": true
}
```

Failed return example:

``` json
{
  "content": null,
  "errorMsg": "app_id not found",
  "success": false
}
```

### `DELETE /v1/sourcemap` {#api-sourcemap-delete}

[:octicons-tag-24: Version-1.16.0](changelog.md#cl-1.16.0)

Delete the sourcemap file. This interface requires the [RUM Collector](../integrations/rum.md) to be enabled.

Description of request parameters:

| Parameter  | Description                                                                                          | Type     |
| ---:       | ---                                                                                                  | ---      |
| `token`    | The token included in the `dataway` address in the `datakit.conf` configuration                      | `string` |
| `app_id`   | The unique ID of the application accessed by the user, such as `test-sourcemap`                      | `string` |
| `env`      | The deployment environment of the application, such as `prod`                                        | `string` |
| `version`  | The version of the application, such as `1.0.0`                                                      | `string` |
| `platform` | The type of the application. Optional values are `web/miniapp/android/ios`, and the default is `web` | `string` |

Request example:

``` shell
curl -X DELETE "http://localhost:9529/v1/sourcemap?app_id=test_sourcemap&env=production&version=1.0.0&token=tkn_xxxxx&platform=web"
```

Successful return example:

``` json
{
  "content":"deleted [/path/to/datakit/data/rum/web/test_sourcemap-production-1.0.0.zip]!",
  "errorMsg":"",
  "success":true
}
```

Failed return example:

``` json
{
  "content": null,
  "errorMsg": "delete sourcemap file [/path/to/datakit/data/rum/web/test_sourcemap-production-1.0.0.zip] failed: remove /path/to/datakit/data/rum/web/test_sourcemap-production-1.0.0.zip: no such file or directory",
  "success": false
}
```

### `/v1/sourcemap/check` {#api-sourcemap-check}

[:octicons-tag-24: Version-1.16.0](changelog.md#cl-1.16.0)

Verify whether the sourcemap file is correctly configured. This interface requires the [RUM Collector](../integrations/rum.md) to be enabled.

Description of request parameters:

| Parameter     | Description                                                                                          | Type     |
| ---:          | ---                                                                                                  | ---      |
| `error_stack` | The stack information of the error                                                                   | `string` |
| `app_id`      | The unique ID of the application accessed by the user, such as `test-sourcemap`                      | `string` |
| `env`         | The deployment environment of the application, such as `prod`                                        | `string` |
| `version`     | The version of the application, such as `1.0.0`                                                      | `string` |
| `platform`    | The type of the application. Optional values are `web/miniapp/android/ios`, and the default is `web` | `string` |

Request example:

``` shell
curl "http://localhost:9529/v1/sourcemap/check?app_id=test_sourcemap&env=production&version=1.0.0&error_stack=at%20test%20%40%20http%3A%2F%2Flocalhost%3A8080%2Fmain.min.js%3A1%3A48"
```

Successful return example:

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

Failed return example:

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

### `/v1/global/host/tags` {#api-global-host-tags-get}

Get global-host-tags. Request example:

``` shell
curl http://localhost:9529/v1/global/host/tags
```

Successful return example:

``` json
status_code: 200
Response: {
    "host-tags": {
        "h": "h",
        "host": "host-name"
    }
}
```

### `/v1/global/host/tags` {#api-global-host-tags-post}

Create or update global-host-tags.  Request example:

``` shell
curl -X POST "http://localhost:9529/v1/global/host/tags?tag1=v1&tag2=v2"
```

Successful return example:

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

After a successful modification, if it is in the host mode, the modified content will be persisted to the configuration file `datakit.conf`.

### `/v1/global/host/tags` {#api-global-host-tags-delete}

Delete some global-host-tags.  Request example:

``` shell
curl -X DELETE "http://localhost:9529/v1/global/host/tags?tags=tag1,tag3"
```

Successful return example:

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

After a successful modification, if it is in the host mode, the modified content will be persisted to the configuration file `datakit.conf`.

### `/v1/global/election/tags` {#api-global-election-tags-get}

Get global-election-tags.  Request example:

``` shell
curl http://localhost:9529/v1/global/election/tags
```

Successful return example:

``` json
status_code: 200
Response: {
    "election-tags": {
        "e": "e"
    }
}
```

### `/v1/global/election/tags` {#api-global-election-tags-post}

Create or update global-election-tags.  Request example:

``` shell
curl -X POST "http://localhost:9529/v1/global/election/tags?tag1=v1&tag2=v2"
```

Successful return example:

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

After a successful modification, if it is in the host mode, the modified content will be persisted to the configuration file `datakit.conf`.

When the global `global-election-enable = false` prohibits the execution of this command, the failed return example is:

``` json
status_code: 500
Response: {
    "message": "Can't use this command when global-election is false."
}
```

### `/v1/global/election/tags` {#api-global-election-tags-delete}

Delete some global-election-tags.  Request example:

``` shell
curl -X DELETE "http://localhost:9529/v1/global/election/tags?tags=tag1,tag3"
```

Successful return example:

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

After a successful modification, if it is in the host mode, the modified content will be persisted to the configuration file `datakit.conf`.

When the global `global-election-enable = false` prohibits the execution of this command, the failed return example is:

``` json
status_code: 500
Response: {
    "message": "Can't use this command when global-election is false."
}
```

### `/v1/ping` {##api-get-dk-version}

```shell
$ curl "http://localhost:9529/v1/ping"

{
  "content": {
    "version": "1.72.0",
    "uptime": "41m44.632183515s",
    "host": "centos",
    "commit": "db3ce3b914"
  }
}
```

In addition, in the returned Header of each of the following API requests, you can obtain the version number of the DataKit for the current request through `X-DataKit`.

### `/v1/pipeline/debug` {#api-debug-pl}

Provide the function of remotely debugging PL. The structure of the error message `PlError`:

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

An example of the error message in JSON:

```json
{
  "pos_chain": [
    { // The position where the error occurred (the script stops running)
      "file": "xx.p",    // File name or file path
      "ln":   15,        // Line number
      "col":  29,        // Column number
      "pos":  576,       // The absolute position of the character in the text starting from 0
    },
   ... ,
    { // The starting point of the call chain
      "file": "b.p",
      "ln":   1,
      "col":  1,
      "pos":  0,
    }
  ],
  "error": "error msg"
}
```

Request example:

``` http
curl -XPOST -H "Content-Type: application/json" http://localhost:9529/v1/pipeline/debug -d'{
    "pipeline": {
      "<caregory>": {
        "<script_name>": <base64("pipeline-source-code")>
      }
    },
    "script_name": "<script_name>",
    "category": "<logging[metric, tracing, ...]>", # Log category, log text should be passed in. For other categories, line protocol text should be passed in
    "data": [ base64("raw-logging-data1"), ... ], # It can be log or line protocol
    "data_type": "application/line-protocol",
    "encode": "@data's character encoding",         // The default is utf8 encoding
    "benchmark": false                    // Whether to enable benchmark
}
```

Normal return example:

``` http
HTTP/1.1 200 OK

{
    "content": {
        "cost": "2.3ms",
        "benchmark": BenchmarkResult.String(), # Returns the benchmark result
        "pl_errors": [],                       // The list of PlError generated during script parsing or checking
        "plresults": [                         // Since the log may be multi-line, multiple segmentation results will be returned here
            {
                "point": {
                  "name" : "It can be the name of the metric set, the log source, etc.",
                  "tags": { "key": "val", "other-key": "other-val"},
                  "fields": { "f1": 1, "f2": "abc", "f3": 1.2 }
                  "time": 1644380607,   // Unix timestamp (in seconds), the front-end can convert it to a readable date
                  "time_ns": 421869748, // The remaining nanoseconds for accurate date conversion. The complete nanosecond timestamp is 1644380607421869748
                },
                "dropped": false,  // Whether the result is marked as to be discarded during the execution of the pipeline
                "run_error": null  // If there is no error, the value is null
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

### `/v1/dialtesting/debug` {#api-debug-dt}

Provide the function of remotely debugging dial testing, and the prohibited network access can be controlled through [environment variables](../integrations/dialtesting.md#env).

Request example:

``` http
curl -XPOST -H "Content-Type: application/json" http://localhost:9529/v1/dialtesting/debug -d'{
    "task_type" : "http",//"http","tcp","icmp","websocket","multi"
    "task" : {
        "name"               : "",
        "method"             : "",
        "url"                : "",
        "post_url"           : "",
        "cur_status"         : "",
        "frequency"          : "",
        "enable_traceroute"  : true, // true represents checked, only valid for tcp and icmp
        "success_when_logic" : "",
        "SuccessWhen"        : []*HTTPSuccess ,
        "tags"               : map[string]string ,
        "labels"             : []string,
        "advance_options"    : *HTTPAdvanceOption,
    },
    "variables": {
      "variable_uuid": {
       "name": "token",
       "value": "token"
      }
    }
}'
```

Normal return example:

``` http
HTTP/1.1 200 OK

{
    "content": {
        "cost": "2.3ms",
        "status": "success", # success/fail
        "error_msg":"",
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
        ],
        "fields": {
          "config_vars: "",
          "url": "",
          "task": "",
          "post_script_variables": "{\"a\":1}"
        }
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

## Information Query APIs {#query-apis}

### `/v1/env_variable` {#api-env-variable}

[:octicons-tag-24: Version-1.72.0](changelog.md#cl-1.72.0)

Get the list of RUM environment variables. Description of request parameters:

 | Parameter | Description                                           | Type     |
 | ---:      | ---                                                   | ---      |
 | `app_id`  | The unique ID of the application accessed by the user | `string` |


``` shell
curl "http://localhost:9529/v1/env_variable?app_id=app_id"
```

Successful return example:

``` json
status_code: 200
Response: {
  "content": {
    "R.app.key": "value"
  }
}
```

### `/metrics` {#api-metrics}

Get the Prometheus metrics exposed by DataKit. Request example:

```shell
curl http://localhost:9529/metrics
```

### `/v1/lasterror` {#api-lasterror}

Used to report errors of external collectors. Example:

``` http
POST /v1/lasterror HTTP/1.1
Content-Type: application/json

{
  "input":"redis",
  "source":"us-east-9xwha",
  "err_content":"Cache avalanche"
}
```

### `/v1/query/raw` {#api-raw-query}

Use DQL for data querying (only data in the workspace where this DataKit is located can be queried). Example:

``` shell
curl -XPOST "http://localhost:9529/v1/query/raw" \
    -H 'Content-Type: application/json'  \
    -d'{
    "queries":[
        {
            "query": "cpu:(usage_idle) LIMIT 1",
            "conditions": "",
            "max_duration": "1d",
            "max_point": 0,
            "time_range": [],
            "orderby": [],
            "disable_slimit": true,
            "disable_multiple_field": true
        }
    ],
    "echo_explain":true
}'
```

<!--
```shell
curl --data-binary @/path/to/dql.json -H "Content-Type:application/json" http://localhost:9529/v1/query/raw
```
-->

Parameter description:

 | Name                     | Required Parameter | Description                                                                                                                                                                                                                                                                                                                                                  |
 | :---                     | ---                | ---                                                                                                                                                                                                                                                                                                                                                          |
 | `queries`                | Y                  | Basic query module, including query statements and various additional parameters                                                                                                                                                                                                                                                                             |
 | `query`                  | Y                  | DQL query statement (DQL [Documentation](../dql/define.md))                                                                                                                                                                                                                                                                                                  |
 | `conditions`             | N                  | Additional conditional expressions, using DQL syntax, such as `hostname="cloudserver01" OR system="ubuntu"`. It has an `AND` relationship with the conditional expressions in the existing `query`, and parentheses will be added to the outermost layer to avoid confusion                                                                                  |
 | `disable_multiple_field` | N                  | Whether to disable multiple fields. When it is `true`, only data of a single field (excluding the `time` field) can be queried, and the default is `false`                                                                                                                                                                                                   |
 | `disable_slimit`         | N                  | Whether to disable the default SLimit. When it is `true`, the default SLimit value will not be added; otherwise, SLimit 20 will be forced to be added, and the default is `false`                                                                                                                                                                            |
 | `echo_explain`           | N                  | Whether to return the final executed statement (the `raw_query` field in the returned JSON data)                                                                                                                                                                                                                                                             |
 | `highlight`              | N                  | Highlight search results                                                                                                                                                                                                                                                                                                                                     |
 | `limit`                  | N                  | Limit the number of points returned by a single timeline, which will overwrite the `limit` in the DQL                                                                                                                                                                                                                                                        |
 | `max_duration`           | N                  | Limit the maximum query time, supporting units `ns/us/ms/s/m/h/d/w/y`, for example, `3d` means 3 days, `2w` means 2 weeks, and `1y` means 1 year. The default is 1 year, and this parameter also limits the `time_range` parameter                                                                                                                           |
 | `max_point`              | N                  | Limit the maximum number of aggregated points. When using aggregation functions, if the aggregation density is too small resulting in too many points, a new aggregation interval will be obtained by `(end_time - start_time)/max_point` and replaced                                                                                                       |
 | `offset`                 | N                  | Generally used in conjunction with `limit` for result pagination                                                                                                                                                                                                                                                                                             |
 | `orderby`                | N                  | Specify the `order by` parameter. The content format is an array of `map[string]string`, where the `key` is the field name to be sorted, and the `value` can only be the sorting method, i.e., `asc` and `desc`, for example, `[ { "column01" : "asc" }, { "column02" : "desc" } ]`. This item will replace the `order by` in the original query statement |
 | `search_after`           | N                  | Deep pagination. When calling pagination for the first time, pass in an empty list: `"search_after": []`. After success, the server will return a list, and the client can directly reuse the value of this list and pass it back to subsequent queries through the `search_after` parameter                                                                 |
 | `slimit`                 | N                  | Limit the number of timelines, which will overwrite the `slimit` in the DQL                                                                                                                                                                                                                                                                                  |
 | `time_range`             | N                  | Limit the time range, using timestamp format, with a unit of milliseconds. It is an `int` array with a size of 2. If there is only one element, it is considered the start time and will overwrite the query time range in the original query statement                                                                                                      |

## Further Reading {#more-reading}

<font size=3>
<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>API Access Settings</u>: Modify DataKit HTTP API settings</font>](datakit-conf.md#config-http-server)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>API Rate Limiting</u>: Limit DataKit HTTP API traffic</font>](datakit-conf.md#set-http-api-limit)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>API Security</u>: Instructions on some security issues involved in the configuration of DataKit</font>](datakit-conf.md#public-apis)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Blacklist Rules</u>: The blacklist operation mechanism on the DataKit side</font>](datakit-filter.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Tags</u>: Commonly used data tags in DataKit collection</font>](common-tags.md)
</div>

<div class="grid cards" markdown>
- [<font color="coral"> :fontawesome-solid-arrow-right-long: &nbsp; <u>Extending DataKit</u>: Modify DataKit from the source code</font>](development.md)
</div>
</font>
