### `timestamp()` {#fn-timestamp}

Function prototype: `fn timestamp(precision: str = "ns") -> int`

Function description: 返回当前 Unix 时间戳，默认精度为 ns

Function parameters:

- `precision`: 时间戳精度，取值范围为 "ns", "us", "ns", "s", 默认值 "ns"。

Example:


```python
# process script
add_key(time_now_record, timestamp())

datetime(time_now_record, "ns", 
    "%Y-%m-%d %H:%M:%S", "UTC")


# process result
{
  "time_now_record": "2023-03-07 10:41:12"
}

```


```python
# process script
add_key(time_now_record, timestamp())

datetime(time_now_record, "ns", 
    "%Y-%m-%d %H:%M:%S", "Asia/Shanghai")


# process result
{
  "time_now_record": "2023-03-07 18:41:49"
}
```


```python
# process script
add_key(time_now_record, timestamp("ms"))


# process result
{
  "time_now_record": 1678185980578
}
```
