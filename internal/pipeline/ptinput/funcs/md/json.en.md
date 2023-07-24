### `json()` {#fn-json}

Function prototype: `fn json(input: str, json_path, newkey, trim_space: bool = true)`

Function description: Extract the specified field in JSON and name it as a new field.

Function parameters:

- `input`: The JSON to be extracted can be the original text (`_`) or a `key` after the initial extraction
- `json_path`: JSON path information
- `newkey`：Write the data to the new key after extraction
- `trim_space`: Delete the leading and trailing blank characters in the extracted characters, the default value is true
- `delete_after_extract`: After extract delete the extracted info from input. Only map key and map value are deletable, list(array) are not supported. Default is `false'.

```python
# Directly extract the x.y field in the original input json, and name it as a new field abc
json(_, x.y, abc)

# For a `key` that has been extracted, extract `x.y` again, and the extracted field name is `x.y`
json(key, x.y) 
```

Example 1:

```python
# input data: 
# {"info": {"age": 17, "name": "zhangsan", "height": 180}}

# script:
json(_, info, "zhangsan")
json(zhangsan, name)
json(zhangsan, age, "age")

# result:
{
  "age": 17,
  "message": "{\"info\": {\"age\": 17, \"name\": \"zhangsan\", \"height\": 180}}",
  "name": "zhangsan",
  "zhangsan": "{\"age\":17,\"height\":180,\"name\":\"zhangsan\"}"
}
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

# script:
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
    
# script:
json(_, .[0].nets[-1])
```

Example 4:

```python
# input data:
{"item": " not_space ", "item2":{"item3": [123]}}

# script:
json(_, item2.item3, item, delete_after_extract = true)

# result:
{
  "item": "[123]",
  "message": "{\"item\":\" not_space \",\"item2\":{}}",
}
```


Example 5:

```python
# input data:
{"item": " not_space ", "item2":{"item3": [123]}}

# If you try to remove a list element it will fail the script check.
# Script:
json(_, item2.item3[0], item, delete_after_extract = true)


# test command:
# datakit pipeline -P j2.p -T '{"item": " not_space ", "item2":{"item3": [123]}}'
# report error:
# [E] j2.p:1:54: does not support deleting elements in the list
```
