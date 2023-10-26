### `rename()` {#fn-rename}

Function prototype: `fn rename(new_key, old_key)`

Function description: Rename the extracted fields

Function parameters:

- `new_key`: new field name
- `old_key`: the extracted field name

Example:

```python
# Rename the extracted abc field to abc1
rename('abc1', abc)

# or

rename(abc1, abc)
```

```python
# Data to be processed: {"info": {"age": 17, "name": "zhangsan", "height": 180}}

# process script
json(_, info.name, "name")

# process result
{
   "message": "{\"info\": {\"age\": 17, \"name\": \"zhangsan\", \"height\": 180}}",
   "zhangsan": {
     "age": 17,
     "height": 180,
     "Name": "zhangsan"
   }
}
```
