### `pt_kvs_set_map()` {#fn_pt_kvs_set_map}

函数原型：`fn pt_kvs_set_map(values: map, include_keys: list|nil = nil, key_patterns: list|nil = nil, as_tag: bool = false, raw: bool = false) -> int`

函数说明：从 map 中批量提取 key/value 并写入 Point，返回成功写入的数量。

说明：

- `include_keys` 和 `key_patterns` 均为空或未传时，不写入任何 key
- `include_keys` 中的元素只按精确 key 匹配
- `key_patterns` 中的元素支持 `*` 和 `?` 通配
- 默认设置 field 时，`list` / `map` 会写成字符串
- 传入 `raw=true` 时，设置 field 会优先保留 Point 原生数组/字典
- 设置 tag 时，值会转换为字符串

函数参数：

- `values`: 待写入 Point 的 map
- `include_keys`: 待提取的精确 key 列表
- `key_patterns`: 待提取的 key 通配模式列表；支持 `*` 和 `?`
- `as_tag`: 是否设置为标签
- `raw`: 设置 field 时是否保留 Point 原生数组/字典

示例：

```python
fields = {
    "service": "api",
    "status": 200,
    "trace_id": "abc",
    "trace_span": "def",
}

count = pt_kvs_set_map(fields, include_keys=["service"], key_patterns=["trace_*"])
```
