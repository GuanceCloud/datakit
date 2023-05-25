### `cidr()` {#fn-cidr}

函数原型：`fn cidr(ip: str, prefix: str) bool`

函数说明： 判断 IP 是否在某个 CIDR 块

函数参数

- `ip`: IP 地址
- `prefix`： IP 前缀，如 `192.0.2.1/24`

示例：

```python
# 待处理数据：

# 处理脚本

ip = "192.0.2.233"
if cidr(ip, "192.0.2.1/24") {
    add_key(ip_prefix, "192.0.2.1/24")
}

# 处理结果
{
  "ip_prefix": "192.0.2.1/24"
}
```
