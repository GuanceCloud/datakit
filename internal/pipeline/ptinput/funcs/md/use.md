### `use()` {#fn-use}

函数原型：`fn use(name: str)`

参数：

- `name`: 脚本名，如 abp.p

函数说明：调用其他脚本，可在被调用的脚本访问当前的所有数据
示例：

```python
# 待处理数据：{"ip":"1.2.3.4"}

# 处理脚本 a.p
use(\"b.p\")

# 处理脚本 b.p
json(_, ip)
geoip(ip)

# 执行脚本 a.p 的处理结果
{
  "city"     : "Brisbane",
  "country"  : "AU",
  "ip"       : "1.2.3.4",
  "province" : "Queensland",
  "isp"      : "unknown"
  "message"  : "{\"ip\": \"1.2.3.4\"}",
}
```
