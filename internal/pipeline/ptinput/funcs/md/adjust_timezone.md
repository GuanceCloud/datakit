### `adjust_timezone()` {#fn-adjust-timezone}

函数原型：`fn adjust_timezone(key: int, minute: int)`

函数参数

- `key`: 纳秒时间戳，如 `default_time(time)` 函数处理后得到的时间戳
- `minute`: 返回值允许超出当前时间的分钟数（整数），取值范围 [0, 15], 默认值为 2 分钟

函数说明：使得传入的时间戳减去函数执行时刻的时间戳的差值在（-60+minute, minute] 分钟内；不适用于时间差超出此范围的数据，否则将导致获取到错误的数据。计算流程：

1. 为 key 的值加上数小时使其处于当前小时内
2. 此时计算两者分钟差，两者分钟数值范围为 [0, 60)，差值范围在 (-60,0] 和 [0, 60)
3. 差值小于等于 -60 + minute 的加 1 小时，大于 minute 的减 1 小时
4. minute 默认值为 2，则差的范围允许在 (-58, 2]，若此时为 11:10，日志时间为 3:12:00.001，最终结果为 10:12:00.001；若此时为 11:59:1.000, 日志时间为 3:01:1.000，最终结果为 12:01:1.000

示例：

```json
# 输入 1 
{
    "time":"11 Jul 2022 12:49:20.937", 
    "second":2,
    "third":"abc",
    "forth":true
}
```

脚本：

```python
json(_, time)      # 提取 time 字段 (若容器中时区 UTC+0000)
default_time(time) # 将提取到的 time 字段转换成时间戳
                   # (对无时区数据使用本地时区 UTC+0800/UTC+0900..。解析)
adjust_timezone(time)
                   # 自动(重新)选择时区，校准时间偏差
```

执行 `datakit pipeline -P <name>.p -F <input_file_name>  --date`:

```json
# 输出 1
{
  "message": "{\n    \"time\":\"11 Jul 2022 12:49:20.937\",\n    \"second\":2,\n    \"third\":\"abc\",\n    \"forth\":true\n}",
  "status": "unknown",
  "time": "2022-07-11T20:49:20.937+08:00"
}
```

本机时间：`2022-07-11T20:55:10.521+08:00`

仅使用 `default_time` 按照默认本机时区（UTC+8）解析得到的时间分别为：

- 输入 1 结果： `2022-07-11T12:49:20.937+08:00`

使用 `adjust_timezone` 后将得到：

- 输入 1 结果： `2022-07-11T20:49:20.937+08:00`
