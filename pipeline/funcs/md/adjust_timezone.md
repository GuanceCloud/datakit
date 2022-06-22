
### `adjust_timezone()` {#fn-adjust-timezone}

函数原型：`adjust_timezone(key=required)`

函数说明：自动选择时区，校准时间戳。用于校准日志中的时间格式不带时区信息，且与 pipeline 时间处理函数默认的本地时区不一致时使得时间戳出现数小时的偏差，适用于时间偏差小于24小时

函数参数

- `key`: 纳秒时间戳，如 `default_time(time)` 函数处理后得到的时间戳

示例:

```python
# 原始 json
{
    "time":"10 Dec 2021 03:49:20.937",
    "second":2,
    "third":"abc",
    "forth":true
}

# pipeline 脚本
json(_, time)      # 提取 time 字段 (若容器中时区 UTC+0000)
default_time(time) # 将提取到的 time 字段转换成时间戳
                   # (对无时区数据使用本地时区 UTC+0800/UTC+0900...解析)
adjust_timezone(time)
                   # 自动(重新)选择时区，校准时间偏差
# 处理结果
{
  "time": 1639108160937000000,
}
```

