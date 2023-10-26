### `group_in()` {#fn-group-in}

Function prototype: `fn group_in(key: int|float|bool|str, range: list, new_value: int|float|bool|str|map|list|nil, new-key = "")`

Function description: If the `key` value is in the list `in`, a new field can be created and assigned the new value. If no new field is provided, the original field value will be overwritten

Example:

```python
# If the field log_level value is in the list, change its value to "OK"
group_in(log_level, ["info", "debug"], "OK")

# If the field http_status value is in the specified list, create a new status field with the value "not-ok"
group_in(log_level, ["error", "panic"], "not-ok", status)
```
