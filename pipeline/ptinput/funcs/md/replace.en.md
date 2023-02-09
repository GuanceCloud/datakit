### `replace()` {#fn-replace}

Function prototype: `fn replace(key: str, regex: str, replace_str: str)`

Function description: Replace the string data obtained on the specified field according to regular rules

Function parameters:

- `key`: the field to be extracted
- `regex`: regular expression
- `replace_str`: string to replace

Example:

```python
# Phone number: {"str": "13789123014"}
json(_, str)
replace(str, "(1[0-9]{2})[0-9]{4}([0-9]{4})", "$1****$2")

# English name {"str": "zhang san"}
json(_, str)
replace(str, "([a-z]*) \\w*", "$1 ***")

# ID number {"str": "362201200005302565"}
json(_, str)
replace(str, "([1-9]{4})[0-9]{10}([0-9]{4})", "$1**********$2")

# Chinese name {"str": "Little Aka"}
json(_, str)
replace(str, '([\u4e00-\u9fa5])[\u4e00-\u9fa5]([\u4e00-\u9fa5])', "$1＊$2")
```
