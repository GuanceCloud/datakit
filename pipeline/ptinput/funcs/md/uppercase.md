### `uppercase()` {#fn-uppercase}

函数原型：`fn uppercase(key: str)`

函数说明：将已提取 key 中内容转换成大写

函数参数

- `key`: 指定已提取的待转换字段名，将 `key` 内容转成大写

示例：

```python
# 待处理数据：{"first": "hello","second":2,"third":"aBC","forth":true}

# 处理脚本
json(_, first) uppercase(first)

# 处理结果
{
   "first": "HELLO"
}
```

