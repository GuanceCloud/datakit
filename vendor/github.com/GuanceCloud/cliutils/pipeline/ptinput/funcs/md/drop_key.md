### `drop_key()` {#fn-drop-key}

函数原型：`fn drop_key(key)`

函数说明：删除已提取字段

函数参数：

- `key`: 待删除字段名

示例：

```python
# data = "{\"age\": 17, \"name\": \"zhangsan\", \"height\": 180}"

# 处理脚本
json(_, age,)
json(_, name)
json(_, height)
drop_key(height)

# 处理结果
{
    "age": 17,
    "name": "zhangsan"
}
```

