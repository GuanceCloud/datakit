### `decode()` {#fn-decode}

Function prototype: `fn decode(text: str, text_encode: str)`

Function description: Convert text to UTF8 encoding to deal with the problem that the original log is not UTF8 encoded. Currently supported encodings are utf-16le/utf-16be/gbk/gb18030 (these encoding names can only be lowercase)

Example:

```python
decode("wwwwww", "gbk")

# Extracted data(drop: false, cost: 33.279µs):
# {
#   "message": "wwwwww",
# }
```
