### `format_int()` {#fn-format-int}

Function prototype: `fn format_int(val: int, base: int) str`

Function description: Converts a numeric value to a numeric string in the specified base.

Function parameters:

- `val`: The number to be converted.
- `base`: Base, ranging from 2 to 36; when the base is greater than 10, lowercase letters a to z are used to represent values 10 and later.

Example:

```python
# script0
a = 7665324064912355185
b = format_int(a, 16)
if b != "6a60b39fd95aaf71" {
    add_key(abc, b)
} else {
    add_key(abc, "ok")
}

# result
'''
{
    "abc": "ok"
}
'''

# script1
a = "7665324064912355185"
b = format_int(parse_int(a, 10), 16)
if b != "6a60b39fd95aaf71" {
    add_key(abc, b)
} else {
    add_key(abc, "ok")
}

# result
'''
{
    "abc": "ok"
}
'''
```
