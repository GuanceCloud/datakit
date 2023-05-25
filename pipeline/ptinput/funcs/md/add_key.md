### `add_key()` {#fn-add-key}

函数原型：`fn add_key(key, value)`

函数说明：往 point 中增加一个字段

函数参数

- `key`: 新增的 key 名称
- `value`：作为 key 的值

示例：

```python
# 待处理数据：{"age": 17, "name": "zhangsan", "height": 180}

# 处理脚本
add_key(city, "shanghai")

# 处理结果
{
    "age": 17,
    "height": 180,
    "name": "zhangsan",
    "city": "shanghai"
}
```
