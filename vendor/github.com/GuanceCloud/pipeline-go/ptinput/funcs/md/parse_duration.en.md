### `parse_duration()` {#fn-parse-duration}

Function prototype: `fn parse_duration(key: str)`

Function description: If the value of `key` is a golang duration string (such as `123ms`), then `key` will be automatically parsed into an integer in nanoseconds

The current duration units in golang are as follows:

- `ns` nanoseconds
- `us/µs` microseconds
- `ms` milliseconds
- `s` seconds
- `m` minutes
- `h` hours

Function parameters:

- `key`: the field to be parsed

Example:

```python
# assume abc = "3.5s"
parse_duration(abc) # result abc = 3500000000

# Support negative numbers: abc = "-3.5s"
parse_duration(abc) # result abc = -3500000000

# support floating point: abc = "-2.3s"
parse_duration(abc) # result abc = -2300000000
```
