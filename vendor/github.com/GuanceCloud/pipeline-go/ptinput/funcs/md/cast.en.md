### `cast()` {#fn-cast}

Function prototype: `fn cast(key, dst_type: str)`

Function description: Convert the key value to the specified type

Function parameters:

- `key`: key name
- `type`：The target type of conversion, support `\"str\", \"float\", \"int\", \"bool\"`

Example:

```python
# input data: {"first": 1,"second":2,"third":"aBC","forth":true}

# script
json(_, first) 
cast(first, "str")

# result
{
  "first": "1"
}
```
