### `format_int()` {#fn-format-int}

函数原型：`fn format_int(val: int, base: int) str`

函数说明：将数值转换为指定进制的数值字符串。

参数：

- `val`: 待转换的整数
- `base`: 进制，范围 2 到 36；进制大于 10 时使用小写字母 a 到 z 表示 10 及以后的数值。

示例：

```python
# script0
a = 7665324064912355185
b = format_int(a, 16)
if b != "6a60b39fd95aaf71" {
    add_key(abc, b)
} else {
    add_key(abc, "ok")
}

# result
'''
{
    "abc": "ok"
}
'''

# script1
a = "7665324064912355185"
b = format_int(parse_int(a, 10), 16)
if b != "6a60b39fd95aaf71" {
    add_key(abc, b)
} else {
    add_key(abc, "ok")
}

# result
'''
{
    "abc": "ok"
}
'''
```
