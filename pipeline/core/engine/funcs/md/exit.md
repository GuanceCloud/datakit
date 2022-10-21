### `exit()` {#fn-exit}

函数原型：`fn exit()`

函数说明：结束当前一条日志的解析，若未调用函数 drop() 仍会输出已经解析的部分

```python
# in << {"str_a": "2", "str_b": "3"}
json(_, str_a)
if str_a == "2"{
  exit()
}
json(_, str_b)

# Extracted data(drop: false, cost: 48.233µs):
# {
#   "message": "{\"str_a\": \"2\", \"str_b\": \"3\"}",
#   "str_a": "2"
# }
```

