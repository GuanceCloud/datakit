### `cover()` {#fn-cover}

函数原型：`fn cover(key: str, range: list)`

函数说明：对指定字段上获取的字符串数据，按范围进行数据脱敏处理

函数参数

- `key`: 待提取字段
- `range`: 脱敏字符串的索引范围（`[start,end]`） start 和 end 均支持负数下标，用来表达从尾部往前追溯的语义。区间合理即可，end 如果大于字符串最大长度会默认成最大长度

示例：

```python
# 待处理数据 {"str": "13789123014"}
json(_, `str`)
cover(`str`, [8, 9])

# 待处理数据 {"abc": "13789123014"}
json(_, abc)
cover(abc, [2, 4])
```
