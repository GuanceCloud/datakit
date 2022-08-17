### `query_refer_table()` {#fn-query-refer-table}

函数原型：`query_refer_table(table_name=requierd, key=required, value=required)`

函数说明：通过指定的 key 查询外部引用表，并将查询结果的首行的所有列追加到 field 中。

参数:

- `table_name`: 待查找的表名
- `key`: 列名
- `value`: 列对应的值

示例:

```python
# 从输入中提取 表名，列名，列值
json(_, table)
json(_, key)
json(_, value)

# 查询并追加当前列的数据，默认作为 field 添加到数据中
query_refer_table(table, key, value)

```

示例结果:

```json
{
  "col": "ab",
  "col2": 1234,
  "col3": 123,
  "col4": true,
  "key": "col2",
  "message": "{\"table\": \"table_abc\", \"key\": \"col2\", \"value\": 1234.0}",
  "status": "unknown",
  "table": "table_abc",
  "time": "2022-08-16T15:02:14.158452592+08:00",
  "value": 1234
}
```
