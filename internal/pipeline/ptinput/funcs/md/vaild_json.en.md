### `vaild_json()` {#fn-vaild_json}

Function prototype: `fn vaild_json(val: str) bool`

Function description: Determine if it is a valid JSON string.

Function parameters:

- `val`: Requires data of type string.

Example:

```python
a = "null"
if vaild_json(a) { # true
    if load_json(a) == nil {
        add_key("a", "nil")
    }
}

b = "[1, 2, 3]"
if vaild_json(b) { # true
    add_key("b", load_json(b))
}

c = "{\"a\": 1}"
if vaild_json(c) { # true
    add_key("c", load_json(c))
}

d = "???{\"d\": 1}"
if vaild_json(d) { # true
    add_key("d", load_json(c))
} else {
    add_key("d", "invaild json")
}
```

Result:

```json
{
  "a": "nil",
  "b": "[1,2,3]",
  "c": "{\"a\":1}",
  "d": "invaild json",
}
```
