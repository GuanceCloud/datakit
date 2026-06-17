### `json_all()` {#fn-json-all}

函数原型：`fn json_all(input, include_keys: list = nil, key_patterns: list = nil)`

函数说明：提取 JSON 顶层的标量字段，并按顶层 key 写入当前 Point。

函数参数：

- `input`: 待展开的 JSON，可以是原始文本（`_`）或已经提取出的某个 key
- `include_keys`: 只提取指定顶层 key，key 需要完整匹配
- `key_patterns`: 只提取匹配通配符的顶层 key，支持 `*` 和 `?`

说明：

- 只写入顶层字符串、数字、布尔值；顶层对象、数组和 `null` 不会作为字段写入
- 不会递归展开对象字段或数组元素，例如 `name.first` 和 `children[0]` 不会被提取
- 顶层数组中的标量元素可通过 `[index]` 匹配；数组元素如果是对象或数组则不会继续展开
- `include_keys` 是精确匹配；如果需要通配符匹配，请使用 `key_patterns`
- `include_keys` 和 `key_patterns` 均为空或未传时，不提取任何字段

示例：

```python
# 待处理数据：
# {
#   "service": "api",
#   "name": {"first": "Tom", "last": "Anderson"},
#   "age": 37,
#   "trace_id": "abc"
# }

# 处理脚本：
json_all(_, include_keys=["service", "age"], key_patterns=["trace_*"])

# 处理结果包含：
# {
#   "service": "api",
#   "age": 37,
#   "trace_id": "abc"
# }
```

按路径筛选：

```python
json_all(_, include_keys=["service", "age"])
```

按通配符筛选：

```python
json_all(_, key_patterns=["trace_*"])
```
