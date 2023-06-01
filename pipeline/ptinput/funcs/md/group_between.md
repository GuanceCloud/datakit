### `group_between()` {#fn-group-between}

函数原型：`fn group_between(key: int, between: list, new_value: int|float|bool|str|map|list|nil, new_key)`

函数说明：如果 `key` 值在指定范围 `between` 内（注意：只能是单个区间，如 `[0,100]`），则可创建一个新字段，并赋予新值。若不提供新字段，则覆盖原字段值

示例一：

```python
# 待处理数据：{"http_status": 200, "code": "success"}

json(_, http_status)

# 如果字段 http_status 值在指定范围内，则将其值改为 "OK"
group_between(http_status, [200, 300], "OK")

# 处理结果
{
    "http_status": "OK"
}
```

示例二：

```python
# 待处理数据：{"http_status": 200, "code": "success"}

json(_, http_status)

# 如果字段 http_status 值在指定范围内，则新建 status 字段，其值为 "OK"
group_between(http_status, [200, 300], "OK", status)

# 处理结果
{
    "http_status": 200,
    "status": "OK"
}
```
