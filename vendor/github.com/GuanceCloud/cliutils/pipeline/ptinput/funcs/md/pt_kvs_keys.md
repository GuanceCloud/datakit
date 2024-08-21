### `pt_kvs_keys()` {#fn_pt_kvs_keys}

函数原型：`fn pt_kvs_keys(tags: bool = true, fields: bool = true) -> list`

函数说明：返回 Point 中的 key 列表

函数参数：

- `tags`: 是否包含所有标签的名
- `fields`: 是否包含所有字段的名

示例：

```python
for k in pt_kvs_keys() {
    if match("^prefix_", k) {
        pt_kvs_del(k)
    }
}
```
