### `b64enc()` {#fn-b64enc}

Function prototype: `fn b64enc(key: str)`

Function description: Base64 encode the string data obtained on the specified field

Function parameters:

- `key`: key name

Example:

```python
# input data {"str": "hello, world"}
json(_, `str`)
b64enc(`str`)

# result
# {
#   "str": "aGVsbG8sIHdvcmxk"
# }
```
