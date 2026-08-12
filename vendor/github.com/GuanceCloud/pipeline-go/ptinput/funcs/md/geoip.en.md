### `geoip()` {#fn-geoip}

Function prototype: `fn geoip(ip: str, prefix: str = "")`

Function description: Append more IP information to IP. `geoip()` will generate additional fields, such as:

- `isp`: operator
- `city`: city
- `province`: province
- `country`: country

Function parameters:

- `ip`: The extracted IP field supports both IPv4 and IPv6
- `prefix`: Optional. Adds a custom prefix to the generated fields to avoid conflicts with existing fields. It can also be passed as a named parameter, e.g. `geoip(ip, prefix="geo_")`.

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
  "isp"      : "unknown",
  "message"  : "{\"ip\": \"1.2.3.4\"}"
}
```

Example with prefix:

```python
# input data: {"ip":"1.2.3.4"}

# script
json(_, ip)
geoip(ip, "geo_")

# result
{
  "geo_city"     : "Brisbane",
  "geo_country"  : "AU",
  "geo_province" : "Queensland",
  "geo_isp"      : "unknown",
  "ip"           : "1.2.3.4"
}
```
