### `value_type()` {#fn-value-type}

函数原型：`fn value_type(val) str`

函数说明：获取变量的值的类型，返回值范围 ["int", "float", "bool", "str", "list", "map", ""], 若值为 nil，则返回空字符串

参数：

- `val`: 待判断类型的值

示例：

输入：

```json
{"a":{"first": [2.2, 1.1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}
```

脚本：

```python
d = load_json(_)

if value_type(d) == "map" && "a" in d  {
    add_key("val_type", value_type(d["a"]))
}
```

输出：

```json
// Fields
{
  "message": "{\"a\":{\"first\": [2.2, 1.1], \"ff\": \"[2.2, 1.1]\",\"second\":2,\"third\":\"aBC\",\"forth\":true},\"age\":47}",
  "val_type": "map"
}
```
