
### `duration_precision()` {#fn-duration-precision}

函数原型：`duration_precision(key=required, old_precision=require, new_precision=require)`

函数说明：进行 duration 精度的转换，通过参数指定当前精度和目标精度。支持在 s, ms, us, ns 间转换。

```python
# in << {"ts":12345}
json(_, ts)
cast(ts, "int")
duration_precision(ts, "ms", "ns")

# Extracted data(drop: false, cost: 33.279µs):
# {
#   "message": "{\"ts\":12345}",
#   "ts": 12345000000
# }
```

