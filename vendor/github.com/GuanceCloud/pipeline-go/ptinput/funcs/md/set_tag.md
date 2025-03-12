### `set_tag()` {#fn-set-tag}

函数原型：`fn set_tag(key, value: str)`

函数说明：对指定字段标记为 tag 输出，设置为 tag 后，其他函数仍可对该变量操作。如果被置为 tag 的 key 是已经切割出来的 field，那么它将不会在 field 中出现，这样可以避免切割出来的 field key 跟已有数据上的 tag key 重名

函数参数

- `key`: 待标记为 tag 的字段
- `value`: 可以为字符串字面量或者变量

```python
# in << {"str": "13789123014"}
set_tag(str)
json(_, str)          # str == "13789123014"
replace(str, "(1[0-9]{2})[0-9]{4}([0-9]{4})", "$1****$2")
# Extracted data(drop: false, cost: 49.248µs):
# {
#   "message": "{\"str\": \"13789123014\", \"str_b\": \"3\"}",
#   "str#": "137****3014"
# }
# * 字符 `#` 仅为 datakit --pl <path> --txt <str> 输出展示时字段为 tag 的标记

# in << {"str_a": "2", "str_b": "3"}
json(_, str_a)
set_tag(str_a, "3")   # str_a == 3
# Extracted data(drop: false, cost: 30.069µs):
# {
#   "message": "{\"str_a\": \"2\", \"str_b\": \"3\"}",
#   "str_a#": "3"
# }


# in << {"str_a": "2", "str_b": "3"}
json(_, str_a)
json(_, str_b)
set_tag(str_a, str_b) # str_a == str_b == "3"
# Extracted data(drop: false, cost: 32.903µs):
# {
#   "message": "{\"str_a\": \"2\", \"str_b\": \"3\"}",
#   "str_a#": "3",
#   "str_b": "3"
# }
```
