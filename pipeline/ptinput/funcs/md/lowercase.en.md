### `lowercase()` {#fn-lowercase}

Function prototype: `fn lowercase(key: str)`

Function description: Convert the content of the extracted key to lowercase

Function parameters:

- `key`: Specify the extracted field name to be converted

Example:

```python
# input data: {"first": "HeLLo","second":2,"third":"aBC","forth":true}

# script
json(_, first) lowercase(first)

# result
{
		"first": "hello"
}
```
