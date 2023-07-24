### `json()` {#fn-json}

函数原型：`fn json(input: str, json_path, newkey, trim_space: bool = true, delete_after_extract = false)`

函数说明：提取 JSON 中的指定字段，并可将其命名成新的字段。

参数：

- `input`: 待提取 JSON，可以是原始文本（`_`）或经过初次提取之后的某个 `key`
- `json_path`: JSON 路径信息
- `newkey`：提取后数据写入新 key
- `trim_space`: 删除提取出的字符中的空白首尾字符，默认值为 `true`
- `delete_after_extract`: 在提取结束后删除当前对象，在重新序列化后回写待提取对象；只能应用于 map 的 key 与 value 的删除，不能用于删除 list 的元素；默认值为 `false`，不进行任何操作

```python
# 直接提取原始输入 JSON 中的 x.y 字段，并可将其命名成新字段 abc
json(_, x.y, abc)

# 已提取出的某个 `key`，对其再提取一次 `x.y`，提取后字段名为 `x.y`
json(key, x.y) 
```

示例一：

```python
# 待处理数据：
# {"info": {"age": 17, "name": "zhangsan", "height": 180}}

# 处理脚本：
json(_, info, "zhangsan")
json(zhangsan, name)
json(zhangsan, age, "age")

# 处理结果：
{
  "age": 17,
  "message": "{\"info\": {\"age\": 17, \"name\": \"zhangsan\", \"height\": 180}}",
  "name": "zhangsan",
  "zhangsan": "{\"age\":17,\"height\":180,\"name\":\"zhangsan\"}"
}
```

示例二：

```python
# 待处理数据：
#    data = {
#        "name": {"first": "Tom", "last": "Anderson"},
#        "age":37,
#        "children": ["Sara","Alex","Jack"],
#        "fav.movie": "Deer Hunter",
#        "friends": [
#            {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
#            {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
#            {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
#        ]
#    }

# 处理脚本：
json(_, name)
json(name, first)
```

示例三：

```python
# 待处理数据：
#    [
#            {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
#            {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
#            {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
#    ]
    
# 处理脚本，json 数组处理：
json(_, .[0].nets[-1])
```

示例四：

```python
# 待处理数据：
{"item": " not_space ", "item2":{"item3": [123]}}

# 处理脚本：
json(_, item2.item3, item, delete_after_extract = true)

# 输出：
{
  "item": "[123]",
  "message": "{\"item\":\" not_space \",\"item2\":{}}",
}
```


示例五：

```python
# 待处理数据：
{"item": " not_space ", "item2":{"item3": [123]}}

# 处理脚本：
# 如果尝试删除列表元素将无法通过脚本检查
json(_, item2.item3[0], item, true, true)

# 本地测试命令：
# datakit pipeline -P j2.p -T '{"item": " not_space ", "item2":{"item3": [123]}}'
# 报错：
# [E] j2.p:1:37: does not support deleting elements in the list
```
