### `drop_key()` {#fn-drop-key}

Function prototype: `fn drop_key(key)`

Function description: Delete key

Function parameters:

- `key`: key to be deleted

Example：

```python
# data = "{\"age\": 17, \"name\": \"zhangsan\", \"height\": 180}"

json(_, age,)
json(_, name)
json(_, height)
drop_key(height)

# result
# {
#     "age": 17,
#     "name": "zhangsan"
# }
```
