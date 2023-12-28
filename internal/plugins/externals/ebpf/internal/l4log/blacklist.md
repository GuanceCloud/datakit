# 传输层网络数据过滤功能与语法



过滤器规则示例:

单条规则：

以下规则过滤 ip 为 `1.1.1.1`且端口为 80 的网络数据。(运算符后允许换行)

```py
(ip_saddr == "1.1.1.1" || ip_saddr == "1.1.1.1") &&
     (src_port == 80 || dst_port == 80)
```

多条规则：

规则间使用 `;` 或 `\n` 分隔，满足任意一条规则就进行数据过滤

```py
udp
ip_saddr == "1.1.1.1" && (src_port == 80 || dst_port == 80);
ip_saddr == "10.10.0.1" && (src_port == 80 || dst_port == 80)

ipnet_contains("127.0.0.0/8", ip_saddr); ipv6
```

## 可用于过滤的数据

该过滤器用于对网络数据进行过滤，可比较的数据如下：

| key 名        | 类型 | 描述                                     |
| ------------- | ---- | ---------------------------------------- |
| `tcp`         | bool | 是否为 `TCP` 协议                        |
| `udp`         | bool | 是否为 `UDP` 协议                        |
| `ipv4`        | bool | 是否为 `IPv4` 协议                       |
| `ipv6`        | bool | 是否为 `IPv6` 协议                       |
| `src_port`    | int  | 源端口（以被观测网卡/主机/容器为参考系） |
| `dst_port`    | int  | 目标端口                                 |
| `ip_saddr`   | str  | 源 `IPv4` 网络地址                       |
| `ip_saddr`   | str  | 目标 `IPv4` 网络地址                     |
| `ip6_saddr`   | str  | 源 `IPv6` 网络地址                       |
| `ip6_daddr`   | str  | 目标 `IPv6` 网络地址                     |
| `k8s_src_pod` | str  | 源 `pod` 名                              |
| `k8s_dst_pod` | str  | 目标 `pod` 名                            |

## 运算符和函数

### 运算符

运算符从高往低：

| 优先级 | Op     | 名称               | 结合方向 |
| ------ | ------ | ------------------ | -------- |
| 1      | `()`   | 圆括号             | 左       |
| 2      | `！`   | 逻辑非，一元运算符 | 右       |
| 3      | `!=`   | 不等于             | 左       |
| 3      | `>=`   | 大于等于           | 左       |
| 3      | `>`    | 大于               | 左       |
| 3      | `==`    | 等于               | 左       |
| 3      | `<=`   | 小于等于           | 左       |
| 3      | `<`    | 小于               | 左       |
| 4      | `&&`   | 逻辑与             | 左       |
| 4      | `\|\|` | 逻辑或             | 左       |

### 函数

1. **ipnet_contains**

    函数签名： `fn ipnet_contains(ipnet: str, ipaddr: str) bool`

    描述： 判断地址是否在指定的网段内

    示例：

    ```py
    ipnet_contains("127.0.0.0/8", ip_saddr)
    ```

    如果 ip_saddr 值为 "127.0.0.1"，则该规则返回 `true`，该 TCP 连接数据包/ UDP数据包将被过滤。

2. **has_prefix**

    函数签名： `fn has_prefix(s: str, prefix: str) bool`

    描述： 指定字段是否包含某一前缀

    示例：

    ```py
    has_prefix(k8s_src_pod, "datakit-") || has_prefix(k8s_dst_pod, "datakit-")
    ```

    如果 pod 名为 "datakit-kfez321"，该规则返回 `true`。
