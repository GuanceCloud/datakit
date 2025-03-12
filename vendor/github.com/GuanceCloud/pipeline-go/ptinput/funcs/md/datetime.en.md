### `datetime()` {#fn-datetime}

Function prototype: `fn datetime(key, precision: str, fmt: str, tz: str = "")`

Function description: Convert timestamp to specified date format

Function parameters:

- `key`: Extracted timestamp (required parameter)
- `precision`: Input timestamp precision (s, ms, us, ns)
- `fmt`: date format, provides built-in date format and supports custom date format
- `tz`: time zone (optional parameter), convert the timestamp to the time in the specified time zone, the default time zone of the host is used

Built-in date formats:

| Built-in format | date                                  | description               |
| -               | -                                     | -                         |
| "ANSI-C"        | "Mon Jan _2 15:04:05 2006"            |                           |
| "UnixDate"      | "Mon Jan _2 15:04:05 MST 2006"        |                           |
| "RubyDate"      | "Mon Jan 02 15:04:05 -0700 2006"      |                           |
| "RFC822"        | "02 Jan 06 15:04 MST"                 |                           |
| "RFC822Z"       | "02 Jan 06 15:04 -0700"               | RFC822 with numeric zone  |
| "RFC850"        | "Monday, 02-Jan-06 15:04:05 MST"      |                           |
| "RFC1123"       | "Mon, 02 Jan 2006 15:04:05 MST"       |                           |
| "RFC1123Z"      | "Mon, 02 Jan 2006 15:04:05 -0700"     | RFC1123 with numeric zone |
| "RFC3339"       | "2006-01-02T15:04:05Z07:00"           |                           |
| "RFC3339Nano"   | "2006-01-02T15:04:05.999999999Z07:00" |                           |
| "Kitchen"       | "3:04PM"                              |                           |


Custom date format:

The output date format can be customized through the combination of placeholders

| character | example | description |
| - | - | - |
| a | %a | week abbreviation, such as `Wed` |
| A | %A | The full letter of the week, such as `Wednesday`|
| b | %b | month abbreviation, such as `Mar` |
| B | %B | The full letter of the month, such as `March` |
| C | %c | century, current year divided by 100 |
| **d** | %d | day of the month; range `[01, 31]` |
| e | %e | day of the month; range `[1, 31]`, pad with spaces |
| **H** | %H | hour, using 24-hour clock; range `[00, 23]` |
| I | %I | hour, using 12-hour clock; range `[01, 12]` |
| j | %j | day of the year, range `[001, 365]` |
| k | %k | hour, using 24-hour clock; range `[0, 23]` |
| l | %l | hour, using 12-hour clock; range `[1, 12]`, padding with spaces |
| **m** | %m | month, range `[01, 12]` |
| **M** | %M | minutes, range `[00, 59]` |
| n | %n | represents a newline character `\n` |
| p | %p | `AM` or `PM` |
| P | %P | `am` or `pm` |
| s | %s | seconds since 1970-01-01 00:00:00 UTC |
| **S** | %S | seconds, range `[00, 60]` |
| t | %t | represents the tab character `\t` |
| u | %u | day of the week, Monday is 1, range `[1, 7]` |
| w | %w | day of the week, 0 for Sunday, range `[0, 6]` |
| y | %y | year in range `[00, 99]` |
| **Y** | %Y | decimal representation of the year|
| **z** | %z | RFC 822/ISO 8601:1988 style time zone (e.g. `-0600` or `+0800` etc.) |
| Z | %Z | time zone abbreviation, such as `CST` |
| % | %% | represents the character `%` |

Example:

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
json(_, a.timestamp)
datetime(a.timestamp, 'ms', 'RFC3339')
```


```python
# script
ts = timestamp()
datetime(ts, 'ns', fmt='%Y-%m-%d %H:%M:%S', tz="UTC")

# output
{
  "ts": "2023-03-08 06:43:39"
}
```

```python
# script
ts = timestamp()
datetime(ts, 'ns', '%m/%d/%y  %H:%M:%S %z', "Asia/Tokyo")

# output
{
  "ts": "03/08/23  15:44:59 +0900"
}
```
