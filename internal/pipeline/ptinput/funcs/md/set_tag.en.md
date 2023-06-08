### `set_tag()` {#fn-set-tag}

Function prototype: `fn set_tag(key, value: str)`

Function description: mark the specified field as tag output, after setting as tag, other functions can still operate on the variable. If the key set as a tag is a field that has been cut out, it will not appear in the field, so as to avoid the same name of the cut out field key as the tag key on the existing data

Function parameters:

- `key`: the field to be tagged
- `value`: can be a string literal or a variable

```python
# in << {"str": "13789123014"}
set_tag(str)
json(_, str) # str == "13789123014"
replace(str, "(1[0-9]{2})[0-9]{4}([0-9]{4})", "$1****$2")
# Extracted data(drop: false, cost: 49.248µs):
# {
# "message": "{\"str\": \"13789123014\", \"str_b\": \"3\"}",
# "str#": "137****3014"
# }
# * The character `#` is only the tag whose field is tag when datakit --pl <path> --txt <str> output display

# in << {"str_a": "2", "str_b": "3"}
json(_, str_a)
set_tag(str_a, "3") # str_a == 3
# Extracted data(drop: false, cost: 30.069µs):
# {
# "message": "{\"str_a\": \"2\", \"str_b\": \"3\"}",
# "str_a#": "3"
# }


# in << {"str_a": "2", "str_b": "3"}
json(_, str_a)
json(_, str_b)
set_tag(str_a, str_b) # str_a == str_b == "3"
# Extracted data(drop: false, cost: 32.903µs):
# {
# "message": "{\"str_a\": \"2\", \"str_b\": \"3\"}",
# "str_a#": "3",
# "str_b": "3"
# }
```
