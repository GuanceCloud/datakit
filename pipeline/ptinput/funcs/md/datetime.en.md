### `datetime()` {#fn-datetime}

Function prototype: `fn datetime(key, precision: str, fmt: str)`

Function description: Convert timestamp to specified date format

Function parameters:

- `key`: Extracted timestamp (required parameter)
- `precision`: Input timestamp precision (s, ms)
- `fmt`：Date format, time format, support the following templates

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
# input data:
#    {
#        "a":{
#            "timestamp": "1610960605000",
#            "second":2
#        },
#        "age":47
#    }

# script
json(_, a.timestamp) datetime(a.timestamp, 'ms', 'RFC3339')
```
