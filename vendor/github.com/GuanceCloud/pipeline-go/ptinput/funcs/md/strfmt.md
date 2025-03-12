### `strfmt()` {#fn-strfmt}

函数原型：`fn strfmt(key, fmt: str, args ...: int|float|bool|str|list|map|nil)`

函数说明：对已提取 `arg1, arg2, ...` 指定的字段内容根据 `fmt` 进行格式化，并把格式化后的内容写入 `key` 字段中

函数参数

- `key`: 指定格式化后数据写入字段名
- `fmt`: 格式化字符串模板
- `args`：可变参数，可以是多个已提取的待格式化字段名

示例：

```python
# 待处理数据：{"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}

# 处理脚本
json(_, a.second)
json(_, a.thrid)
cast(a.second, "int")
json(_, a.forth)
strfmt(bb, "%v %s %v", a.second, a.thrid, a.forth)
```
