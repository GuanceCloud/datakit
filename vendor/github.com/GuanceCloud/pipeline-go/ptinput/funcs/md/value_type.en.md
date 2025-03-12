### `value_type()` {#fn-value-type}

Function prototype: `fn value_type(val) str`

Function description: Obtain the type of the variable's value and return the value range ["int", "float", "bool", "str", "list", "map", "]. If the value is nil, return an empty string.

Function parameters:

- `val`: The value of the type to be determined.

Example:


Input:

```json
{"a":{"first": [2.2, 1.1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}
```

Script:

```python
d = load_json(_)

if value_type(d) == "map" && "a" in d  {
    add_key("val_type", value_type(d["a"]))
}
```

Output:

```json
// Fields
{
  "message": "{\"a\":{\"first\": [2.2, 1.1], \"ff\": \"[2.2, 1.1]\",\"second\":2,\"third\":\"aBC\",\"forth\":true},\"age\":47}",
  "val_type": "map"
}
```
