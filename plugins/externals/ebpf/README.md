# ebpf 采集器

DataKit 的该 ebpf 外部采集器的网络相关数据采集和内核结构体元素的偏移量推导基于 [tcptracer-bpf](https://github.com/weaveworks/tcptracer-bpf) 和 [datadog-agent network pkg](https://github.com/DataDog/datadog-agent/tree/main/pkg/network) 开发。

shell 命令执行记录采集支持采集终端执行的 bash 命令

网络数据采集支持采集 OSI 模型 Layer 4 传输层的 TCP/UDP 网络流信息、Layer 7 应用层的 dns 请求数据，指标集为 netflow 和 dnsflow

**数据指标集:**

- *bash*:
  - 标签
      |标签名|描述|
      |-|-|
      |host|host name|
      |source|固定值: bash|
  - 指标列表
      |指标|描述|数据类型|单位|
      |-|-|-|-|
      |cmd|bash 命令|string|-|
      |message|单条 bash 执行记录|string|-|
      |pid|bash 进程的 pid|string|-|
      |user|执行 bash 命令的用户|string|-|
- *netflow*:
  - 标签
      |标签名|描述|
      |-|-|
      |direction|传输方向 (incoming/outgoing)|
      |dst_domain|目标域名|
      |dst_ip|目标 IP|
      |dst_ip_type|目标 IP 类型 (other/private/multicast)|
      |dst_port|目标端口|
      |family|TCP/IP 协议族 (IPv4/IPv6)|
      |host|主机名|
      |pid|进程号|
      |source|固定值: netflow|
      |src_ip|源 IP|
      |src_ip_type|源 IP 类型 (other/private/multicast)|
      |src_port|源端口, 临时端口(32768 ~ 60999)聚合后的值为 *|
      |transport|传输协议 (udp/tcp)|
  - 指标
      |指标|描述|数据类型|单位|
      |-|-|-|-|
      |bytes_read|读取字节数|int|Byte|
      |bytes_written|写入字节数|int|Byte|
      |retransmits|重传次数|int|count|
      |rtt|TCP Latency|int|μs|
      |rtt_var|TCP Jitter|int|μs|
      |tcp_closed|TCP 关闭次数|int|count|
      |tcp_established|TCP 建立连接次数|int|count|
- *dnsflow*:
  - 标签
      |标签名|描述|
      |-|-|
      |dst_ip|DNS server address|
      |dst_port|DNS server port|
      |family|TCP/IP 协议族 (IPv4/IPv6)|
      |host|host name|
      |source|固定值: dnsflow|
      |src_ip|DNS client address|
      |src_port|DNS client port|
      |transport|传输协议 (udp/tcp)|
  - 指标列表
      |指标|描述|数据类型|单位|
      |-|-|-|-|
      |rcode|DNS 响应码: 0 - NoError, 1 - FormErr, 2 - ServFail, 3 - NXDomain, 4 - NotImp, 5 - Refused, ...|int|-|
      |resp_time|DNS 请求的响应时间间隔|int|ns|
      |timeout|DNS 请求超时|bool|-|
