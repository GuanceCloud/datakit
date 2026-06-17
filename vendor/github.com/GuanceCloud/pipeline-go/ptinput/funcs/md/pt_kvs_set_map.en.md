### `pt_kvs_set_map()` {#fn_pt_kvs_set_map}

Function prototype: `fn pt_kvs_set_map(values: map, include_keys: list|nil = nil, key_patterns: list|nil = nil, as_tag: bool = false, raw: bool = false) -> int`

Function description: Extract key/value pairs from a map and write them into the Point in batch. The return value is the number of keys written.

Notes:

- When both `include_keys` and `key_patterns` are empty or omitted, no keys are written
- `include_keys` matches exact keys only
- `key_patterns` supports `*` / `?` wildcard patterns
- By default, `list` / `map` field values are written as strings
- With `raw=true`, field values prefer native Point array/dict storage
- When setting tags, values are converted to strings

Parameters:

- `values`: map to write into the Point
- `include_keys`: exact key list to extract
- `key_patterns`: wildcard patterns to extract; supports `*` and `?`
- `as_tag`: whether to write values as tags
- `raw`: whether to preserve native Point array/dict values when writing fields

Example:

```python
fields = {
    "service": "api",
    "status": 200,
    "trace_id": "abc",
    "trace_span": "def",
}

count = pt_kvs_set_map(fields, include_keys=["service"], key_patterns=["trace_*"])
```
