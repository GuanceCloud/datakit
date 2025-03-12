### `decode()` {#fn-decode}

函数原型：`fn decode(text: str, text_encode: str)`

函数说明：把 text 变成 UTF8 编码，以处理原始日志为非 UTF8 编码的问题。目前支持的编码为 utf-16le/utf-16be/gbk/gb18030（这些编码名只能小写）

```python
decode("wwwwww", "gbk")

# Extracted data(drop: false, cost: 33.279µs):
# {
#   "message": "wwwwww",
# }
```
