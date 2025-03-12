### `slice_string()` {#fn_slice_string}

Function prototype: `fn slice_string(name: str, start: int, end: int) -> str`

Function description: Returns the substring of the string from index start to end.

Function Parameters:

- `name`: The string to be sliced
- `start`: The starting index of the substring (inclusive)
- `end`: The ending index of the substring (exclusive)

Example:

```python
substring = slice_string("15384073392", 0, 3)
# substring will be "153"
```