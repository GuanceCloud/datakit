### `url_parse()` {#fn-url-parse}

函数原型：`fn url_parse(key, prefix: str = "")`

函数说明：解析字段名称为 key 的 url。

函数参数

- `key`: 要解析的 url 的字段名称。
- `prefix`: 可选参数，为返回 map 中固定的 key（`scheme`、`host`、`port`、`path`、`params`）添加自定义前缀，避免与已有字段产生冲突。支持以命名参数形式传入，如 `url_parse(url, prefix="up_")`

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

添加前缀示例：

```python
# 待处理数据：{"url": "https://www.baidu.com"}

# 处理脚本
json(_, url)
m = url_parse(url, "up_")
add_key(scheme, m["up_scheme"])
add_key(h, m["up_host"])

# 处理结果
{
    "url": "https://www.baidu.com",
    "scheme": "https",
    "h": "www.baidu.com"
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
