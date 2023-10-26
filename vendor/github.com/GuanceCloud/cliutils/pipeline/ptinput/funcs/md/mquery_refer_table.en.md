### `mquery_refer_table()` {#fn-mquery-refer-table}

Function prototype: `fn mquery_refer_table(table_name: str, keys: list, values: list)`

Function description: Query the external reference table by specifying multiple keys, and append all columns of the first row of the query result to field.

Function parameters:

- `table_name`: the name of the table to be looked up
- `keys`: a list of multiple column names
- `values`: the values corresponding to each column

Example:

```python
json(_, table)
json(_, key)
json(_, value)

# Query and append the data of the current column, which is added to the data as a field by default
mquery_refer_table(table, values=[value, false], keys=[key, "col4"])

# result

# {
#   "col": "ab",
#   "col2": 1234,
#   "col3": 1235,
#   "col4": false,
#   "key": "col2",
#   "message": "{\"table\": \"table_abc\", \"key\": \"col2\", \"value\": 1234.0}",
#   "status": "unknown",
#   "table": "table_abc",
#   "time": "2022-08-16T16:23:31.940600281+08:00",
#   "value": 1234
# }

```
