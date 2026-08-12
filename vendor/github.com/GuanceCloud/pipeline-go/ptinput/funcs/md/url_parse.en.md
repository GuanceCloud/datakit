### `url_parse()` {#fn-url-parse}

Function prototype: `fn url_parse(key, prefix: str = "")`

Function description: parse the url whose field name is key.

Function parameters:

- `key`: field name of the url to parse.
- `prefix`: Optional. Adds a custom prefix to the fixed keys (`scheme`, `host`, `port`, `path`, `params`) of the returned map to avoid conflicts with existing fields. It can also be passed as a named parameter, e.g. `url_parse(url, prefix="up_")`.

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

Example with prefix:

```python
# Data to be processed: {"url": "https://www.baidu.com"}

# process script
json(_, url)
m = url_parse(url, "up_")
add_key(scheme, m["up_scheme"])
add_key(h, m["up_host"])

# process result
{
     "url": "https://www.baidu.com",
     "scheme": "https",
     "h": "www.baidu.com"
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
