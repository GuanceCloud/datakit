### `lowercase()` {#fn-lowercase}

函数原型：`fn lowercase(key: str)`

函数说明：将已提取 key 中内容转换成小写

函数参数

- `key`: 指定已提取的待转换字段名

示例：

```python
# 待处理数据：{"first": "HeLLo","second":2,"third":"aBC","forth":true}

# 处理脚本
json(_, first) lowercase(first)

# 处理结果
{
    "first": "hello"
}
```

