### `strlen()` {#fn-strlen}

Function prototype: `fn strlen(val: str) int`

Function description: Calculate the number of characters in a string, not the number of bytes.

Parameters:

- `val`: input string

Example:

```python
add_key("len_char", strlen("hello 你好"))
add_key("len_byte", len("hello 你好"))
```

Output:

```json
{
"len_char": 8,
"len_byte": 12
}
```
