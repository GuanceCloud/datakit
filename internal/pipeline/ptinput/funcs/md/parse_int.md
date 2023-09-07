### `parse_int()` {#fn-parse-int}

函数原型：`fn parse_int(val: int, base: int) str`

函数说明：将数值的字符串表示转换为数值。

参数：

- `val`: 待转换的字符串
- `base`: 进制，范围 0，或 2 到 36；值为 0 时根据字符串前缀判断进制。

示例：

```python
# script0
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

# script1
a = "6a60b39fd95aaf71" 
b = parse_int(a, 16)            # base 16
if b != 7665324064912355185 {
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


# script2
a = "0x6a60b39fd95aaf71" 
b = parse_int(a, 0)            # the true base is implied by the string's 
if b != 7665324064912355185 {
    add_key(abc, b)
} else {
    c = format_int(b, 16)
    if "0x"+c != a {
        add_key(abc, c)
    } else {
        add_key(abc, "ok")
    }
}


# result
'''
{
    "abc": "ok"
}
'''
```
