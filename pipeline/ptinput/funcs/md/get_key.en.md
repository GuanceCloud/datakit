### `get_key()` {#fn-get-key}

Function prototype: `fn get_key(key)`

Function description: Read the value of key from the input point

Function parameters:

- `key_name`: key name

Example:

```python
add_key("city", "shanghai")

# Here you can directly access the value of the key with the same name in point through city
if city == "shanghai" {
  add_key("city_1", city)
}

# Due to the right associativity of assignment, get the value whose key is "city" first,
# Then create a variable named city
city = city + " --- ningbo" + " --- " +
    "hangzhou" + " --- suzhou ---" + ""

# get_key gets the value of "city" from point
# If there is a variable named city, it cannot be obtained directly from point
if city != get_key("city") {
  add_key("city_2", city)
}

# result
"""
{
  "city": "shanghai",
  "city_1": "shanghai",
  "city_2": "shanghai --- ningbo --- hangzhou --- suzhou ---"
}
"""
```
