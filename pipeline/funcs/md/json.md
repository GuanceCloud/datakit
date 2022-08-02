
### `json()` {#fn-json}

函数原型：`json(input=required, jsonPath=required, newkey=required, trim_space=optional)`

函数说明：提取 json 中的指定字段，并可将其命名成新的字段。

参数:

- `input`: 待提取 json，可以是原始文本（`_`）或经过初次提取之后的某个 `key`
- `jsonPath`: json 路径信息
- `newkey`：提取后数据写入新 key
- `trim_space`: 删除提取出的字符中的空白首尾字符，默认值为 true

```python
# 直接提取原始输入 json 中的x.y字段，并可将其命名成新字段abc
json(_, x.y, abc)

# 已提取出的某个 `key`，对其再提取一次 `x.y`，提取后字段名为 `x.y`
json(key, x.y) 
```

示例一:

```python
# 待处理数据: {"info": {"age": 17, "name": "zhangsan", "height": 180}}

# 处理脚本
json(_, info, "zhangsan")
json(zhangsan, name)
json(zhangsan, age, "年龄")

# 处理结果
{
    "message": "{\"info\": {\"age\": 17, \"name\": \"zhangsan\", \"height\": 180}}
    "zhangsan": {
        "age": 17,
        "height": 180,
        "name": "zhangsan"
    }
}
```

示例二:

```python
# 待处理数据
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

# 处理脚本
json(_, name) json(name, first)
```

示例三:

```python
# 待处理数据
#    [
#            {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
#            {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
#            {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
#    ]
    
# 处理脚本, json数组处理
json(_, [0].nets[-1])
```
