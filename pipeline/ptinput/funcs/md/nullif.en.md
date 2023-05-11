### `nullif()` {#fn-nullif}

Function prototype: `fn nullif(key, value)`

Function description: If the content of the field specified by the extracted `key` is equal to the value of `value`, delete this field

Function parameters:


- `key`: specified field
- `value`: target value

Example:

```python
# input data: {"first": 1,"second":2,"third":"aBC","forth":true}

# script
json(_, first) json(_, second) nullif(first, "1")

# result
{
    "second":2
}
```

> Note: This feature can be implemented with `if/else` semantics:

```python
if first == "1" {
    drop_key(first)
}
```

