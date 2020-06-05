# Thu Jun  4 09:57:50 UTC 2020

-  采集器配置文件目录调整

如下采集器的配置目录做了调整：

|采集器名称  | 原配置路径 | 新配置路径 |
| ----       | ------     | --------   |
|`iptables`  | `conf.d/iptables/iptables.conf` | `confd/network/iptables.conf` |
|`ping_check`| `conf.d/ping_check/ping_check.conf` | `conf.d/network/ping_check.conf` |
|`tcp_udp_check`| `conf.d/tcp_udp_check/tcp_udp_check.conf` | `conf.d/network/tcp_udp_check.conf` |
|`nat`       | `conf.d/nat/nat.conf` | `conf.d/network/nat.conf` |
|`socket_listener` | `conf.d/socket_listener/socket_listener.conf` | `conf.d/network/socket_listener.conf` |
|`nsq_consumer`| `conf.d/nsq_consumer/nsq_consumer.conf` | `conf.d/nsq/nsq_consumer.conf` |
|`kube_inventory`| `conf.d/kube_inventory/kube_inventory.conf` | `conf.d/k8s/kube_inventory.conf` |
|`kubernetes`| `conf.d/kubernetes/kubernetes.conf` | `conf.d/k8s/kubernetes.conf`
|`kafka_consumer`| `conf.d/kafka_consumer/kafka_consumer.conf` | `conf.d/kafka/kafka_consumer.conf` |
|`amqp_consumer`| `conf.d/amqp_consumer/amqp_consumer.conf` | `conf.d/amqp/amqp_consumer.conf` |

- 安装目录变更
	- Windows 64bit 默认安装位置从原 `C:\Program Files (x86)\Forethougt\datakit` 改成  `C:\Program Files\DataFlux\datakit`
	- Windows 32bit 默认安装位置从原 `C:\Program Files (x86)\Forethougt\datakit` 改成  `C:\Program Files (x86)\DataFlux\datakit`
	- Linux 默认安装目录从 `/usr/local/cloudcare/forethougt/datakit` 改成 `/usr/local/cloudcare/DataFlux/datakit`

- 暂时移除了以下采集器支持（因动态库依赖不便于多平台同时发布）
	- `oracle_monitor`：依赖一个连接 Oracle C 语言动态库
	- 以下三个依赖 libpcap 动态库：
		- `netPacket`
		- `scanport`
		- `tracerouter`

- `conf.d` 目录不再支持可配置，只能固定在`<datakit安装目录/conf.d>`，如 `/usr/local/cloudcare/DataFlux/datakit/conf.d`
- 通过 `./datakit -tree` 可查看支持的采集器列表，各个平台可能有所不同：

```
=========== datakit collectors ==============
lighttpd
  |-- lighttpd
object
  |-- host
	...
mongodb
  |-- replication
  |-- mongodb_oplog
prometheus
  |-- prometheus
aliyun
  |-- aliyunlog
  |-- aliyunsecurity
  |-- aliyunrdsslowLog
  |-- aliyunactiontrail
  |-- aliyuncost
  |-- aliyunddos
  |-- aliyuncdn
  |-- aliyunprice
  |-- aliyuncms
network
  |-- httpstat
  |-- coredns
	...
=========== telegraf collectors ==============
snmp
  |-- snmp
network
  |-- socket_listener
  |-- tcp_udp_check
  |-- iptables
  |-- nats
  |-- ping_check
  |-- network
rabbitmq
  |-- rabbitmq
amqp
  |-- amqp
  |-- amqp_consumer
	...
```
