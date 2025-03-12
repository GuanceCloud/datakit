### `user_agent()` {#fn-user-agent}

Function prototype: `fn user_agent(key: str)`

Function description: Obtain client information on the specified field

Function parameters:

- `key`: the field to be extracted

`user_agent()` will generate multiple fields, such as:

- `os`: operating system
- `browser`: browser

Example:

```python
# data to be processed
# {
# "userAgent" : "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36",
# "second" : 2,
# "third" : "abc",
# "forth" : true
# }

json(_, userAgent) user_agent(userAgent)
```
