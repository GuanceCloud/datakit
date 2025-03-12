### `b64dec()` {#fn-b64dec}

Function prototype: `fn b64dec(key: str)`

Function description: Base64 decodes the string data obtained on the specified field

Function parameters:

- `key`: fields to extract

Example:

```python
# input data {"str": "aGVsbG8sIHdvcmxk"}
json(_, `str`)
b64enc(`str`)

# result
# {
#   "str": "hello, world"
# }
```
