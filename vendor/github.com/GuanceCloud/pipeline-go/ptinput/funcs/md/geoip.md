### `geoip()` {#fn-geoip}

函数原型：`fn geoip(ip: str, prefix: str = "")`

函数说明：在 IP 上追加更多 IP 信息。 `geoip()` 会额外产生多个字段，如：

- `isp`: 运营商
- `city`: 城市
- `province`: 省份
- `country`: 国家

参数：

- `ip`: 已经提取出来的 IP 字段，支持 IPv4 和 IPv6
- `prefix`: 可选参数，为生成的字段添加自定义前缀，避免与已有字段产生冲突。支持以命名参数形式传入，如 `geoip(ip, prefix="geo_")`

示例：

```python
# 待处理数据：{"ip":"1.2.3.4"}

# 处理脚本
json(_, ip)
geoip(ip)

# 处理结果
{
  "city"     : "Brisbane",
  "country"  : "AU",
  "ip"       : "1.2.3.4",
  "province" : "Queensland",
  "isp"      : "unknown",
  "message"  : "{\"ip\": \"1.2.3.4\"}"
}
```

添加前缀示例：

```python
# 待处理数据：{"ip":"1.2.3.4"}

# 处理脚本
json(_, ip)
geoip(ip, "geo_")

# 处理结果
{
  "geo_city"     : "Brisbane",
  "geo_country"  : "AU",
  "geo_province" : "Queensland",
  "geo_isp"      : "unknown",
  "ip"           : "1.2.3.4"
}
```
