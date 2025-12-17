### `grok()` {#fn-grok}

Function prototype: `fn grok(input: str, pattern: str, trim_space: bool = true) bool`

Function description: Extract the contents of the text string `input` by `pattern`, and return true when pattern matches input successfully, otherwise return false.

Function parameters:

- `input`：The text to be extracted can be the original text (`_`) or a `key` after the initial extraction
- `pattern`: grok expression, the data type of the specified key is supported in the expression: bool, float, int, string (corresponding to Pipeline's str, can also be written as str), the default is string
- `trim_space`: Delete the leading and trailing blank characters in the extracted characters, the default value is true

```python
grok(_, pattern)    #Use the entered text directly as raw data
grok(key, pattern)  # For a key that has been extracted before, do grok again
```

Example:

```python
# input data: "12/01/2021 21:13:14.123"

# script
add_pattern("_second", "(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)")
add_pattern("_minute", "(?:[0-5][0-9])")
add_pattern("_hour", "(?:2[0123]|[01]?[0-9])")
add_pattern("time", "([^0-9]?)%{_hour:hour:string}:%{_minute:minute:int}(?::%{_second:second:float})([^0-9]?)")

grok_match_ok = grok(_, "%{DATE_US:date} %{time}")

add_key(grok_match_ok)

# result
{
  "date": "12/01/2021",
  "hour": "21",
  "message": "12/01/2021 21:13:14.123",
  "minute": 13,
  "second": 14.123
}

{
  "date": "12/01/2021",
  "grok_match_ok": true,
  "hour": "21",
  "message": "12/01/2021 21:13:14.123",
  "minute": 13,
  "second": 14.123,
  "status": "unknown",
  "time": 1665994187473917724
}
```
