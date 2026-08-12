### `user_agent()` {#fn-user-agent}

函数原型：`fn user_agent(key: str, prefix: str = "")`

函数说明：对指定字段上获取客户端信息

函数参数

- `key`: 待提取字段
- `prefix`: 可选参数，为生成的字段添加自定义前缀，避免与已有字段产生冲突。支持以命名参数形式传入，如 `user_agent(userAgent, prefix="ua_")`

`user_agent()` 会生产多个字段，如：

- `os`: 操作系统
- `browser`: 浏览器

示例：

```python
# 待处理数据
#    {
#        "userAgent" : "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36",
#        "second"    : 2,
#        "third"     : "abc",
#        "forth"     : true
#    }

json(_, userAgent)
user_agent(userAgent)
```

添加前缀示例：

```python
json(_, userAgent)
user_agent(userAgent, "ua_")
# 生成的字段为 ua_isMobile、ua_os、ua_browser 等
```
