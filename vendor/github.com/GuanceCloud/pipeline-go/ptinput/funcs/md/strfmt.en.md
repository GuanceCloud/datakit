### `strfmt()` {#fn-strfmt}

Function prototype: `fn strfmt(key, fmt: str, args ...: int|float|bool|str|list|map|nil)`

Function description: Format the content of the field specified by the extracted `arg1, arg2, ...` according to `fmt`, and write the formatted content into the `key` field

Function parameters:

- `key`: Specify the field name of the formatted data to be written
- `fmt`: format string template
- `args`: Variable Function parameters:, which can be multiple extracted field names to be formatted

Example:

```python
# Data to be processed: {"a":{"first":2.3,"second":2,"third":"abc","forth":true},"age":47}

# process script
json(_, a.second)
json(_, a.thrid)
cast(a. second, "int")
json(_, a.forth)
strfmt(bb, "%v %s %v", a.second, a.thrid, a.forth)
```
