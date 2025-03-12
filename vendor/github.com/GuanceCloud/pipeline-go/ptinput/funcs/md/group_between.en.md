### `group_between()` {#fn-group-between}

Function prototype: `fn group_between(key: int, between: list, new_value: int|float|bool|str|map|list|nil, new_key)`

Function description: If the `key` value is within the specified range `between` (note: it can only be a single interval, such as `[0,100]`), a new field can be created and assigned a new value. If no new field is provided, the original field value will be overwritten

Example 1:

```python
# input data: {"http_status": 200, "code": "success"}

json(_, http_status)

# If the field http_status value is within the specified range, change its value to "OK"
group_between(http_status, [200, 300], "OK")

# result
# {
#     "http_status": "OK"
# }
```

Example 2:

```python
# input data: {"http_status": 200, "code": "success"}

json(_, http_status)

# If the value of the field http_status is within the specified range, create a new status field with the value "OK"
group_between(http_status, [200, 300], "OK", status)

# result
{
    "http_status": 200,
    "status": "OK"
}
```
