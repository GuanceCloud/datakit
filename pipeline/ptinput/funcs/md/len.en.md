### `len()` {#fn-len}

Function prototype: `fn len(val: str|map|list) int`

Function description: Calculate the number of bytes in string, the number of elements in map and list.

Function parameters:

- `val`: Can be map, list or string

Example:

```python
# example 1
add_key(abc, len("abc"))
# result
{
 "abc": 3,
}

# example 2
add_key(abc, len(["abc"]))
# result
{
  "abc": 1,
}
```
