### `load_json()` {#fn-load-json}

函数原型：`fn load_json(val: str) nil|bool|float|map|list`

函数说明：将 JSON 字符串转换成 map、list、nil、bool、float 的其中一种，可通过 index 表达式取值及修改值。若反序列化失败，也返回 nil，而不是终止脚本运行。

参数：

- `val`: 要求是 string 类型的数据。

示例：

```python
# _: {"a":{"first": [2.2, 1.1], "ff": "[2.2, 1.1]","second":2,"third":"aBC","forth":true},"age":47}
abc = load_json(_)

add_key(abc, abc["a"]["first"][-1])

abc["a"]["first"][-1] = 11

# 需要将堆栈上的数据同步到 point 中
add_key(abc, abc["a"]["first"][-1])

add_key(len_abc, len(abc))

add_key(len_abc, len(load_json(abc["a"]["ff"])))
```
