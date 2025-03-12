### `gjson()` {#fn-gjson}

函数原型：`fn gjson(input, json_path: str, newkey: str)`

函数说明：提取 JSON 中的指定字段，可将其命名成新的字段，并保证按原始顺序排列

参数：

- `input`: 待提取 JSON，可以是原始文本（`_`）或经过初次提取之后的某个 `key`
- `json_path`: JSON 路径信息
- `newkey`：提取后数据写入新 key

```python
# 直接提取原始输入 JSON 中的 x.y 字段，并可将其命名成新字段 abc
gjson(_, "x.y", "abc")

# 已提取出的某个 `key`，对其再提取一次 `x.y`，提取后字段名为 `x.y`
gjson(key, "x.y") 

# 提取数组，`key` 和 `abc` 均为数组类型
gjson(key, "1.abc.2")
```

示例一：

```python
# 待处理数据：
# {"info": {"age": 17, "name": "zhangsan", "height": 180}}

# 处理脚本：
gjson(_, "info", "zhangsan")
gjson(zhangsan, "name")
gjson(zhangsan, "age", "age")

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
gjson(_, "name")
gjson(name, "first")
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
gjson(_, "0.nets.1")
```