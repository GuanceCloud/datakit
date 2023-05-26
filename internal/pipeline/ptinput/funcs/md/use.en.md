### `use()` {#fn-use}

Function prototype: `fn use(name: str)`

Parameter:

- `name`: script name, such as abp.p

Function description: call other scripts, all current data can be accessed in the called script

Example:

```python
# Data to be processed: {"ip":"1.2.3.4"}

# Process script a.p
use(\"b.p\")

# Process script b.p
json(_, ip)
geoip (ip)

# Execute the processing result of script a.p
{
   "city" : "Brisbane",
   "country" : "AU",
   "ip" : "1.2.3.4",
   "province" : "Queensland",
   "isp" : "unknown"
   "message" : "{\"ip\": \"1.2.3.4\"}",
}
```
