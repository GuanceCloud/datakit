### `delete()` {#fn-delete}

Function prototype: `fn delete(src: map[string]any, key: str)`

Function description: Delete the key in the JSON map

```python

# input
# {"a": "b", "b":[0, {"c": "d"}], "e": 1}

# script
j_map = load_json(_)

delete(j_map["b"][-1], "c")

delete(j_map, "a")

add_key("j_map", j_map)

# result:
# {
#   "j_map": "{\"b\":[0,{}],\"e\":1}",
# }
```
