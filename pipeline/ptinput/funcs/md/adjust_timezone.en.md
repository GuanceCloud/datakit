### `adjust_timezone()` {#fn-adjust-timezone}

Function prototype: `fn adjust_timezone(key: int, minute: int)`

Function parameters:

- `key`: Nanosecond timestamp, such as the timestamp obtained by the `default_time(time)` function
- `minute`: The return value allows the number of minutes (integer) beyond the current time, the value range is [0, 15], the default value is 2 minutes

Function description: Make the difference between the incoming timestamp minus the timestamp of the function execution time within (-60+minute, minute] minutes; it is not applicable to data whose time difference exceeds this range, otherwise it will result in wrong data being obtained. Calculation process:

1. Add hours to the value of key to make it within the current hour
2. At this time, calculate the difference between the two minutes. The value range of the two minutes is [0, 60), and the difference range is between (-60,0] and [0, 60)
3. If the difference is less than or equal to -60 + minute, add 1 hour, and if the difference is greater than minute, subtract 1 hour
4. The default value of minute is 2, and the range of the difference is allowed to be (-58, 2], if it is 11:10 at this time, the log time is 3:12:00.001, and the final result is 10:12:00.001; if at this time is 11:59:1.000, the log time is 3:01:1.000, and the final result is 12:01:1.000

Example:

```json
# input data 1 
{
    "time":"11 Jul 2022 12:49:20.937", 
    "second":2,
    "third":"abc",
    "forth":true
}
```

Script：

```python
json(_, time)      # Extract the time field (if the time zone in the container is UTC+0000)
default_time(time) # Convert the extracted time field into a timestamp
                   # (Use local time zone UTC+0800/UTC+0900... parsing for data without time zone)
adjust_timezone(time)
                   # Automatically (re)select time zone, calibrate time offset

```

Execute `datakit pipeline -P <name>.p -F <input_file_name>  --date`:

```json
# output 1
{
  "message": "{\n    \"time\":\"11 Jul 2022 12:49:20.937\",\n    \"second\":2,\n    \"third\":\"abc\",\n    \"forth\":true\n}",
  "status": "unknown",
  "time": "2022-07-11T20:49:20.937+08:00"
}
```

local time: `2022-07-11T20:55:10.521+08:00`

The times obtained by using only `default_time` and parsing according to the default local time zone (UTC+8) are:

- Output result of input 1： `2022-07-11T12:49:20.937+08:00`

After using `adjust_timezone` will get:

- Output result of input 1： `2022-07-11T20:49:20.937+08:00`
