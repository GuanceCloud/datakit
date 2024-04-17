### `cache_set()` {#fn-cache}

Function prototype: `fn cache_set(key: str, value: str, expiration: int) nil`

Function description: save key value pair to cache

Function parameters:

- `key`：key (required)
- `value`：value (required)
- `expiration`：expire time (default=100s)

Example:

```python
a = cache_set("a", "123")
a = cache_get("a")
add_key(abc, a)
```