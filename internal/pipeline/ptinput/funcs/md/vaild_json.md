### `vaild_json()` {#fn-vaild_json}

函数原型：`fn vaild_json(val: str) bool`

函数说明：判断是否为一个有效的 JSON 字符串。

参数：

- `val`: 要求是 string 类型的数据。

示例：

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

结果：

```json
{
  "a": "nil",
  "b": "[1,2,3]",
  "c": "{\"a\":1}",
  "d": "invaild json",
}
```
