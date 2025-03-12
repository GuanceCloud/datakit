### `pt_kvs_set()` {#fn_pt_kvs_set}

Function prototype: `fn pt_kvs_set(name: str, value: any, as_tag: bool = false) -> bool`

Function description: Add a key to a Point or modify the value of a key in a Point

Function parameters:

- `name`: The name of the field or label to be added or modified
- `value`: The value of a field or label
- `as_tag`: Set as tag or not

Example:

```python
kvs = {
    "a": 1,
    "b": 2
}

for k in kvs {
    pt_kvs_set(k, kvs[k])
}
```
