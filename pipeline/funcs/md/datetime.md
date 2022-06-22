
### `datetime()` {#fn-datetime}

函数原型：`datetime(key=required, precision=required, fmt=required)`

函数说明：将时间戳转成指定日期格式

函数参数

- `key`: 已经提取的时间戳 (必选参数)
- `precision`：输入的时间戳精度(s, ms)
- `fmt`：日期格式，时间格式, 支持以下模版

```python
ANSIC       = "Mon Jan _2 15:04:05 2006"
UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
RFC822      = "02 Jan 06 15:04 MST"
RFC822Z     = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
RFC3339     = "2006-01-02T15:04:05Z07:00"
RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
Kitchen     = "3:04PM"
```

示例:

```python
# 待处理数据:
#    {
#        "a":{
#            "timestamp": "1610960605000",
#            "second":2
#        },
#        "age":47
#    }

# 处理脚本
json(_, a.timestamp) datetime(a.timestamp, 'ms', 'RFC3339')
```
