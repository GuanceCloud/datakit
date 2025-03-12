### `b64dec()` {#fn-b64dec}

函数原型：`fn b64dec(key: str)`

函数说明：对指定字段上获取的字符串数据进行 base64 解码

函数参数

- `key`: 待提取字段

示例：

```python
# 待处理数据 {"str": "aGVsbG8sIHdvcmxk"}
json(_, `str`)
b64enc(`str`)

# 处理结果
# {
#   "str": "hello, world"
# }
```
