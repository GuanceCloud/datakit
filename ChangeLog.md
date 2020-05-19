# Mon May 18 03:34:44 UTC 2020

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
