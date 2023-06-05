### `uppercase()` {#fn-uppercase}

Function prototype: `fn uppercase(key: str)`

Function description: Convert the content in the extracted key to uppercase

Function parameters:

- `key`: Specify the extracted field name to be converted, and convert the content of `key` to uppercase

Example:

```python
# Data to be processed: {"first": "hello","second":2,"third":"aBC","forth":true}

# process script
json(_, first) uppercase(first)

# process result
{
    "first": "HELLO"
}
```
