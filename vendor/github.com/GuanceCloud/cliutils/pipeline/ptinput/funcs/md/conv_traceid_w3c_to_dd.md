### `conv_traceid_w3c_to_dd()`  {#fn-conv-traceid-w3c-to-dd}

函数原型：`fn conv_traceid_w3c_to_dd(key)`

函数说明：将 16 进制编码的 128-bit/64-bit  W3C Trace ID 字符串（长度 32 个字符或 16 个字符）转换为 10 进制编码的 64-bit DataDog Trace ID 字符串。

函数参数

- `key`: 待转换的 128-bit/64-bit Trace ID

示例：

```python

# script input:

"18962fdd9eea517f2ae0771ea69d6e16"

# script:

grok(_, "%{NOTSPACE:trace_id}")

conv_traceid_w3c_to_dd(trace_id)

# result:

{
    "trace_id": "3089600317904219670",
}

```
