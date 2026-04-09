### `pt_kvs_get()` {#fn_pt_kvs_get}

Function prototype: `fn pt_kvs_get(name: str, raw: bool = false) -> any`

Function description: Return the value of the specified key in Point

Notes:

- By default, values are returned in plain form; list/map values are converted to strings
- With `raw=true`, array and map values are preserved as `list` / `map`, so they can be used by script operations such as `append()`, `len()`, and `in`
- Tags are always returned as strings
- If the value was already stored as a JSON string earlier in the script, the returned value is still a string

Function parameters:

- `name`: Key name
- `raw`: whether to return the Point native value; when `false`, list/map values are returned as plain strings

Example:

```python
host = pt_kvs_get("host")

nums = pt_kvs_get("nums", raw=true)
nums = append(nums, 4)
pt_kvs_set("nums", nums, raw=true)

obj = pt_kvs_get("obj", raw=true)
obj_str = pt_kvs_get("obj", raw=false)
```
