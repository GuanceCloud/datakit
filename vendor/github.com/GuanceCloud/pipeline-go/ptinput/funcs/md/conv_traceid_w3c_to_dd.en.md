### `conv_traceid_w3c_to_dd()`  {#fn-conv-traceid-w3c-to-dd}

Function prototype: `fn conv_traceid_w3c_to_dd(key)`

Function description: Convert a hex-encoded 128-bit/64-bit W3C Trace ID string(length 32 characters or 16 characters) to a decimal-encoded 64-bit DataDog Trace ID string.

Function parameters:

- `key`: 128-bit/64-bit Trace ID to convert

Example:

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
