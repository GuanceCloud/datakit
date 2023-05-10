### `sql_cover()` {#fn-sql-cover}

函数原型：`fn sql_cover(sql_test: str)`

函数说明：脱敏 SQL 语句

```python
# in << {"select abc from def where x > 3 and y < 5"}
sql_cover(_)

# Extracted data(drop: false, cost: 33.279µs):
# {
#   "message": "select abc from def where x > ? and y < ?"
# }
```
