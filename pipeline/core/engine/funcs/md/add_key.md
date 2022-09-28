### `add_key()` {#fn-add-key}

函数原型：`add_key(key_name=required, key_name=optional)`

函数说明：往 point 中增加一个字段

函数参数

- `key_name`: 新增的 key 名称
- `key_name`：key 值（只能是 string/int/float/bool/nil 这几种类型）, 未填写时尝试从堆栈中获取同名变量的值

示例:

```python
# 待处理数据: {"age": 17, "name": "zhangsan", "height": 180}

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
