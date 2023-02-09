### `json()` {#fn-json}

Function prototype: `fn json(input: str, json_path, newkey, trim_space: bool = true)`

Function description: Extract the specified field in json and name it as a new field.

Function parameters:

- `input`: The json to be extracted can be the original text (`_`) or a `key` after the initial extraction
- `json_path`: json path information
- `newkey`：Write the data to the new key after extraction
- `trim_space`: Delete the leading and trailing blank characters in the extracted characters, the default value is true

```python
# Directly extract the x.y field in the original input json, and name it as a new field abc
json(_, x.y, abc)

# For a `key` that has been extracted, extract `x.y` again, and the extracted field name is `x.y`
json(key, x.y) 
```

Example 1:

```python
# input data: {"info": {"age": 17, "name": "zhangsan", "height": 180}}

# script
json(_, info, "zhangsan")
json(zhangsan, name)
json(zhangsan, age, "年龄")

# result
# {
#     "message": "{\"info\": {\"age\": 17, \"name\": \"zhangsan\", \"height\": 180}}
#     "zhangsan": {
#         "age": 17,
#         "height": 180,
#         "name": "zhangsan"
#     }
# }
```

Example 2:

```python
# input data:
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

# script
json(_, name) 
json(name, first)
```

Example 3:

```python
# input data:
#    [
#            {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
#            {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
#            {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
#    ]
    
# script
json(_, [0].nets[-1])
```
