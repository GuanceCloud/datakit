### `gjson()` {#fn-gjson}

Function prototype: `fn gjson(input, json_path: str, newkey: str)`

Function description: Extract specified fields from JSON, rename them as new fields, and ensure they are arranged in the original order.

Function parameters:

- `input`: The JSON to be extracted can either be the original text (`_`) or a specific `key` after the initial extraction.
- `json_path`: JSON path information
- `newkey`: Write the data to the new key after extraction

```python
# Directly extract the field x.y from the original input JSON and rename it as a new field abc.
gjson(_, "x.y", "abc")

# Extract the x.y field from a previously extracted key, and name the extracted field as x.y.
gjson(key, "x.y")

# Extract arrays, where `key` and `abc` are arrays.
gjson(key, "1.abc.2")
```

Example 1:

```python
# input data:
# {"info": {"age": 17, "name": "zhangsan", "height": 180}}

# script:
gjson(_, "info", "zhangsan")
gjson(zhangsan, "name")
gjson(zhangsan, "age", "age")

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
gjson(_, "name")
gjson(name, "first")
```

Example 3:

```python
# input data:
#    [
#            {"first": "Dale", "last": "Murphy", "age": 44, "nets": ["ig", "fb", "tw"]},
#            {"first": "Roger", "last": "Craig", "age": 68, "nets": ["fb", "tw"]},
#            {"first": "Jane", "last": "Murphy", "age": 47, "nets": ["ig", "tw"]}
#    ]
    
# scripts for JSON list:
gjson(_, "0.nets.1")
```