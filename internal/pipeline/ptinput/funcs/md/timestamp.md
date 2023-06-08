### `timestamp()` {#fn-timestamp}

函数原型：`fn timestamp(precision: str = "ns") -> int`

函数说明：返回当前 Unix 时间戳，默认精度为 ns

函数参数：

- `precision`: 时间戳精度，取值范围为 "ns", "us", "ns", "s", 默认值 "ns"。

示例：

```python
# 处理脚本
add_key(time_now_record, timestamp())

datetime(time_now_record, "ns", 
    "%Y-%m-%d %H:%M:%S", "UTC")


# 处理结果
{
  "time_now_record": "2023-03-07 10:41:12"
}

```

```python
# 处理脚本
add_key(time_now_record, timestamp())

datetime(time_now_record, "ns", 
    "%Y-%m-%d %H:%M:%S", "Asia/Shanghai")


# 处理结果
{
  "time_now_record": "2023-03-07 18:41:49"
}
```

```python
# 处理脚本
add_key(time_now_record, timestamp("ms"))


# 处理结果
{
  "time_now_record": 1678185980578
}
```
