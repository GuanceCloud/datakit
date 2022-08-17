### `mquery_refer_table()` {#fn-mquery-refer-table}

函数原型：`mquery_refer_table(table_name=requierd, keys=required, values=required)`

函数说明：通过指定多个 key 查询外部引用表，并将查询结果的首行的所有列追加到 field 中。

参数:

- `table_name`: 待查找的表名
- `keys`: 多个列名构成的列表
- `values`: 每个列对应的值

示例:

```python
json(_, table)
json(_, key)
json(_, value)

# 查询并追加当前列的数据，默认作为 field 添加到数据中
mquery_refer_table(table, values=[value, false], keys=[key, "col4"])
```

示例结果:

```json
{
  "col": "ab",
  "col2": 1234,
  "col3": 1235,
  "col4": false,
  "key": "col2",
  "message": "{\"table\": \"table_abc\", \"key\": \"col2\", \"value\": 1234.0}",
  "status": "unknown",
  "table": "table_abc",
  "time": "2022-08-16T16:23:31.940600281+08:00",
  "value": 1234
}

```
