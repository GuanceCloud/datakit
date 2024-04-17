### `cache_set()` {#fn-cache-set}

函数原型：`fn cache_set(key: str, value: str, expiration: int) nil`

函数说明：将键值对保存到 cache 中

参数：

- `key`：键（必填）
- `value`：值（必填）
- `expiration`：过期时间（默认=100s）

示例：

```python
a = cache_set("a", "123")
a = cache_get("a")
add_key(abc, a)
```
