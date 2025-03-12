### `cidr()` {#fn-cidr}

Function prototype: `fn cidr(ip: str, prefix: str) bool`

Function description: Determine whether the IP is in a CIDR block

Function parameters:

- `ip`: IP address
- `prefix`： IP prefix, such as `192.0.2.1/24`

Example:

```python
# script
ip = "192.0.2.233"
if cidr(ip, "192.0.2.1/24") {
    add_key(ip_prefix, "192.0.2.1/24")
}

# result
{
  "ip_prefix": "192.0.2.1/24"
}
```
