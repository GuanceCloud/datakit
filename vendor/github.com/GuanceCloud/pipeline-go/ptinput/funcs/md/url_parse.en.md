### `url_parse()` {#fn-url-parse}

Function prototype: `fn url_parse(key)`

Function description: parse the url whose field name is key.

Function parameters:

- `key`: field name of the url to parse.

Example:

```python
# Data to be processed: {"url": "https://www.baidu.com"}

# process script
json(_, url)
m = url_parse(url)
add_key(scheme, m["scheme"])

# process result
{
     "url": "https://www.baidu.com",
     "scheme": "https"
}
```

The above example extracts its scheme from the url. In addition, it can also extract information such as host, port, path, and Function parameters: carried in the url from the url, as shown in the following example:

```python
# Data to be processed: {"url": "https://www.google.com/search?q=abc&sclient=gws-wiz"}

# process script
json(_, url)
m = url_parse(url)
add_key(sclient, m["params"]["sclient"]) # The Function parameters: carried in the url are saved under the params field
add_key(h, m["host"])
add_key(path, m["path"])

# process result
{
     "url": "https://www.google.com/search?q=abc&sclient=gws-wiz",
     "h": "www.google.com",
     "path": "/search",
     "sclient": "gws-wiz"
}
```
