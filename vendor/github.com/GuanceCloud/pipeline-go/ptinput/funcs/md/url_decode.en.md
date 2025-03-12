### `url_decode()` {#fn-url-decode}

Function prototype: `fn url_decode(key: str)`

Function description: parse the URL in the extracted `key` into plain text

Function parameters:

- `key`: a `key` that has been extracted

Example:

```python
# Data to be processed: {"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"}

# process script
json(_, url) url_decode(url)

# process result
{
   "message": "{"url":"http%3a%2f%2fwww.baidu.com%2fs%3fwd%3d%e6%b5%8b%e8%af%95"}",
   "url": "http://www.baidu.com/s?wd=test"
}
```
