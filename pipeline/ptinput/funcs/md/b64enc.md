### `b64enc()` {#fn-b64enc}

函数原型：`fn b64enc(key: str)`

函数说明：对指定字段上获取的字符串数据进行 base64 编码

函数参数

- `key`: 待提取字段

示例：

```python
# 待处理数据 {"str": "hello, world"}
json(_, `str`)
b64enc(`str`)

# 处理结果
# {
#   "str": "aGVsbG8sIHdvcmxk"
# }
```
