### `load_json()` {#fn-load-json}

Function prototype: `fn load_json(val: str) nil|bool|float|map|list`

Function description: Convert the JSON string to one of map, list, nil, bool, float, and the value can be obtained and modified through the index expression.If deserialization fails, it also returns nil instead of terminating the script run.

Function parameters:

- `val`: Requires data of type string.

Example:

```python
# _: {"a":{"first": [2.2, 1.1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}
abc = load_json(_)

add_key(abc, abc["a"]["first"][-1])

abc["a"]["first"][-1] = 11

# Need to synchronize the data on the stack to point
add_key(abc, abc["a"]["first"][-1])

add_key(len_abc, len(abc))

add_key(len_abc, len(load_json(abc["a"]["ff"])))
```
