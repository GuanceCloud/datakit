### `pt_kvs_set()` {#fn_pt_kvs_set}

函数原型：`fn pt_kvs_set(name: str, value: any, as_tag: bool = false) -> bool`

函数说明：往 Point 中添加 key 或修改 Point 中 key 的值

函数参数：

- `name`: 待添加或修改的字段或标签的名
- `value`: 字段或者标签的值
- `as_tag`: 是否设置为标签

示例：

```python
kvs = {
    "a": 1,
    "b": 2
}

for k in kvs {
    pt_kvs_set(k, kvs[k])
}
```
