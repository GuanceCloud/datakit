### `get_key()` {#fn-get-key}

函数原型：`fn get_key(key)`

函数说明：从输入 point 中读取 key 的值，而不是堆栈上的变量的值

函数参数

- `key_name`: key 的名称

示例：

```python
add_key("city", "shanghai")

# 此处可以直接通过 city 访问获取 point 中的同名 key 的值
if city == "shanghai" {
  add_key("city_1", city)
}

# 由于赋值的右结合性，先获取 key 为 "city" 的值，
# 而后创建名为 city 的变量
city = city + " --- ningbo" + " --- " +
    "hangzhou" + " --- suzhou ---" + ""

# get_key 从 point 中获取 "city" 的值
# 存在名为 city 的变量，则无法直接从 point 中获取
if city != get_key("city") {
  add_key("city_2", city)
}

# 处理结果
"""
{
  "city": "shanghai",
  "city_1": "shanghai",
  "city_2": "shanghai --- ningbo --- hangzhou --- suzhou ---"
}
"""
```
