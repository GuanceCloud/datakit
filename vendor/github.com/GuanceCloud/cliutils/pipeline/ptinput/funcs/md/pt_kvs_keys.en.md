### `pt_kvs_keys()` {#fn_pt_kvs_keys}

Function prototype: `fn pt_kvs_keys(tags: bool = true, fields: bool = true) -> list`

Function description: Return the key list in Point

Function parameters:

- `tags`: Whether to include the names of all tags
- `fields`: Whether to include the names of all fields

Example:

```python
for k in pt_kvs_keys() {
    if match("^prefix_", k) {
        pt_kvs_del(k)
    }
}
```
