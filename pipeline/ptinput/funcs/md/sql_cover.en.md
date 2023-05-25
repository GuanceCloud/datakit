### `sql_cover()` {#fn-sql-cover}

Function prototype: `fn sql_cover(sql_test: str)`

Function description: desensitized SQL statement

Example:

```python
# in << {"select abc from def where x > 3 and y < 5"}
sql_cover(_)

# Extracted data(drop: false, cost: 33.279µs):
# {
# "message": "select abc from def where x > ? and y < ?"
# }
```
