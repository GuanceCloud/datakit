### `query_refer_table()` {#fn-query-refer-table}

Function prototype: `fn query_refer_table(table_name: str, key: str, value)`

Function description: Query the external reference table through the specified key, and append all the columns of the first row of the query result to field.

Function parameters:

- `table_name`: the name of the table to be looked up
- `key`: column name
- `value`: the value corresponding to the column

Example:

```python
# extract table name, column name, column value from input
json(_, table)
json(_, key)
json(_, value)

# Query and append the data of the current column, which is added to the data as a field by default
query_refer_table(table, key, value)

```

Result:

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
