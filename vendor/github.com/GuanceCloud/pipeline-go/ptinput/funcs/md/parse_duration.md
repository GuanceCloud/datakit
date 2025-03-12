### `parse_duration()` {#fn-parse-duration}

函数原型：`fn parse_duration(key: str)`

函数说明：如果 `key` 的值是一个 golang 的 duration 字符串（如 `123ms`），则自动将 `key` 解析成纳秒为单位的整数

目前 golang 中的 duration 单位如下：

- `ns` 纳秒
- `us/µs` 微秒
- `ms` 毫秒
- `s` 秒
- `m` 分钟
- `h` 小时

函数参数

- `key`: 待解析的字段

示例：

```python
# 假定 abc = "3.5s"
parse_duration(abc) # 结果 abc = 3500000000

# 支持负数：abc = "-3.5s"
parse_duration(abc) # 结果 abc = -3500000000

# 支持浮点：abc = "-2.3s"
parse_duration(abc) # 结果 abc = -2300000000

```

