### `add_key()` {#fn-add-key}

函数原型：`add_key(key-name=required, key-value=required)`

函数说明：增加一个字段

函数参数

- `key-name`: 新增的 key 名称
- `key-value`：key 值（只能是 string/number/bool/nil 这几种类型）

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
