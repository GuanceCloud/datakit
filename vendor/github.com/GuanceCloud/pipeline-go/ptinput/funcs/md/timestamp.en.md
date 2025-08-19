### `timestamp()` {#fn-timestamp}

Function prototype: `fn timestamp(precision: str = "ns") -> int`

Function description: Returns the current Unix timestamp, with a default precision of ns.

Function parameters:

- `precision`: Timestamp precision, the value range is "ns", "us", "ns", "s", the default value is "ns".

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
