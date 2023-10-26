### `geoip()` {#fn-geoip}

Function prototype: `fn geoip(ip: str)`

Function description: Append more IP information to IP. `geoip()` will generate additional fields, such as:

- `isp`: operator
- `city`: city
- `province`: province
- `country`: country

Function parameters:

- `ip`: The extracted IP field supports both IPv4 and IPv6

Example:

```python
# input data: {"ip":"1.2.3.4"}

# script
json(_, ip)
geoip(ip)

# result
{
  "city"     : "Brisbane",
  "country"  : "AU",
  "ip"       : "1.2.3.4",
  "province" : "Queensland",
  "isp"      : "unknown"
  "message"  : "{\"ip\": \"1.2.3.4\"}",
}
```
