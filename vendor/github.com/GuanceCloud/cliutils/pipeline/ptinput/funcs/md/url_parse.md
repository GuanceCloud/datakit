### `url_parse()` {#fn-url-parse}

函数原型：`fn url_parse(key)`

函数说明：解析字段名称为 key 的 url。

函数参数

- `key`: 要解析的 url 的字段名称。

示例：

```python
# 待处理数据：{"url": "https://www.baidu.com"}

# 处理脚本
json(_, url)
m = url_parse(url)
add_key(scheme, m["scheme"])

# 处理结果
{
    "url": "https://www.baidu.com",
    "scheme": "https"
}
```

上述示例从 url 提取了其 scheme，除此以外，还能从 url 提取出 host, port, path, 以及 url 中携带的参数等信息，如下例子所示：

```python
# 待处理数据：{"url": "https://www.google.com/search?q=abc&sclient=gws-wiz"}

# 处理脚本
json(_, url)
m = url_parse(url)
add_key(sclient, m["params"]["sclient"])    # url 中携带的参数被保存在 params 字段下
add_key(h, m["host"])
add_key(path, m["path"])

# 处理结果
{
    "url": "https://www.google.com/search?q=abc&sclient=gws-wiz",
    "h": "www.google.com",
    "path": "/search",
    "sclient": "gws-wiz"
}
```
