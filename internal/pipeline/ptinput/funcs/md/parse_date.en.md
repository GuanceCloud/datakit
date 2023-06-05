### `parse_date()` {#fn-parse-date}

Function prototype: `fn parse_date(key: str, yy: str, MM: str, dd: str, hh: str, mm: str, ss: str, ms: str, zone: str)`

Function description: Convert the value of each part of the incoming date field into a timestamp

Function parameters:

- `key`: newly inserted field
- `yy` : Year numeric string, supports four or two digit strings, if it is an empty string, the current year will be used when processing
- `MM`: month string, supports numbers, English, English abbreviation
- `dd`: day string
- `hh`: hour string
- `mm`: minute string
- `ss`: seconds string
- `ms`: milliseconds string
- `us`: microseconds string
- `ns`: string of nanoseconds
- `zone`: time zone string, in the form of "+8" or \"Asia/Shanghai\"

Example:

```python
parse_date(aa, "2021", "May", "12", "10", "10", "34", "", "Asia/Shanghai") # Result aa=1620785434000000000

parse_date(aa, "2021", "12", "12", "10", "10", "34", "", "Asia/Shanghai") # result aa=1639275034000000000

parse_date(aa, "2021", "12", "12", "10", "10", "34", "100", "Asia/Shanghai") # Result aa=1639275034000000100

parse_date(aa, "20", "February", "12", "10", "10", "34", "", "+8") result aa=1581473434000000000
```
