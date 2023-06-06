### `user_agent()` {#fn-user-agent}

函数原型：`fn user_agent(key: str)`

函数说明：对指定字段上获取客户端信息

函数参数

- `key`: 待提取字段

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

json(_, userAgent) user_agent(userAgent)
```
