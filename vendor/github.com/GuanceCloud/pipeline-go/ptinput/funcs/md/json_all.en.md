### `json_all()` {#fn-json-all}

Function prototype: `fn json_all(input, include_keys: list = nil, key_patterns: list = nil)`

Function description: Extract top-level scalar fields from JSON and write them to the current Point by top-level key.

Function parameters:

- `input`: The JSON to expand. It can be the original text (`_`) or an extracted key
- `include_keys`: Extract only the specified top-level keys. Keys must match exactly
- `key_patterns`: Extract only top-level keys that match wildcard patterns. `*` and `?` are supported

Notes:

- Only top-level strings, numbers, and booleans are written. Top-level objects, arrays, and `null` are not written as fields
- Object fields and array elements are not expanded recursively, so `name.first` and `children[0]` are not extracted
- Top-level scalar array elements can be matched with `[index]`; array elements that are objects or arrays are not expanded
- `include_keys` uses exact matching. Use `key_patterns` for wildcard matching
- When both `include_keys` and `key_patterns` are empty or omitted, no fields are extracted

Example:

```python
# input data:
# {
#   "service": "api",
#   "name": {"first": "Tom", "last": "Anderson"},
#   "age": 37,
#   "trace_id": "abc"
# }

# script:
json_all(_, include_keys=["service", "age"], key_patterns=["trace_*"])

# result contains:
# {
#   "service": "api",
#   "age": 37,
#   "trace_id": "abc"
# }
```

Filter by exact path:

```python
json_all(_, include_keys=["service", "age"])
```

Filter by wildcard:

```python
json_all(_, key_patterns=["trace_*"])
```
