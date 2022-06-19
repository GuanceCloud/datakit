
### `nullif()` {#fn-nullif}

函数原型：`nullif(key=required, value=required)`

函数说明：若已提取 `key` 指定的字段内容等于 `value` 值，则删除此字段

函数参数

- `key`: 指定字段
- `value`: 目标值

示例:

```python
# 待处理数据: {"first": 1,"second":2,"third":"aBC","forth":true}

# 处理脚本
json(_, first) json(_, second) nullif(first, "1")

# 处理结果
{
    "second":2
}
```

> 注：该功能可通过 `if/else` 语义来实现：

```python
if first == "1" {
	drop_key(first)
}
```

