### `hash()` {#fn_hash}

函数原型： `fn hash(text: str, method: str) -> str`

函数说明： 计算文本的 hash

函数参数：

- `text`: 输入文本
- `method`: hash 算法，允许的值包含 `md5`，`sha1`，`sha256`，`sha512`

示例：

```python
pt_kvs_set("md5sum", hash("abc", "sha1"))
```
