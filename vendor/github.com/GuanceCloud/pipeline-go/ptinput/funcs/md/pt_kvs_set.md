### `pt_kvs_set()` {#fn_pt_kvs_set}

函数原型：`fn pt_kvs_set(name: str, value: any, as_tag: bool = false, raw: bool = false) -> bool`

函数说明：往 Point 中添加 key 或修改 Point 中 key 的值

说明：

- 默认设置 field 时，`list` / `map` 会写成字符串
- 传入 `raw=true` 时，`list` 会优先保留为 Point 原生数组字段
- 传入 `raw=true` 时，`map` 会优先保留为 Point 原生字典字段
- 设置 tag 时，值会转换为字符串

函数参数：

- `name`: 待添加或修改的字段或标签的名
- `value`: 字段或者标签的值
- `as_tag`: 是否设置为标签
- `raw`: 设置 field 时是否保留 Point 原生数组/字典；为 `false` 时，`list` / `map` 会写成字符串

示例：

```python
kvs = {
    "a": 1,
    "b": 2
}

for k in kvs {
    pt_kvs_set(k, kvs[k])
}

nums = pt_kvs_get("nums")
nums = append(nums, 4)
pt_kvs_set("nums", nums, raw=true)

obj = {"a": 1}
pt_kvs_set("obj_raw", obj, raw=true)
pt_kvs_set("obj_str", obj, raw=false)
```
