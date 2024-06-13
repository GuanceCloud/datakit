# bpf-netlog 采集插件

## L4 和 L7 网络日志

### 公共字段

**常规标签**：

- src_ip
  - 类型： string
  - 描述：本地计算机的 IP 地址，等价于 netstat 输出的 local ip addr；**不表示** client ip 或 server ip
- dst_ip
  - 类型：string
  - 描述：套接字所连接的目标计算机的 IP 地址；可视为取自 Linux Socket 相关函数，即
    `int connect(int sockfd, struct sockaddr * addr, socklen_t addrlen)` 传入的 addr 或
    `int accept(int sockfd, struct sockaddr * addr, socklen_t * addrlen)` 执行后的 addr；
    等价于 netstat 的 foreign ip addr
- src_port
  - 类型：string
  - 描述：本地计算机的端口，与 src_ip 共同组成一个网络地址
- dst_port:
  - 类型：string
  - 描述：目标端口，与 dst_ip 共同组曾一个网络地址
- l4_proto
  - 类型：string
  - 描述：传输层协议
- nic_mac
  - 类型：string
  - 描述：网卡 MAC 地址
- nic_name
  - 类型：string
  - 描述：网卡名
- netns
  - 类型：string
  - 描述：网络命名空间，模板： `NS(<device id>:<inode number>)`
- vni_id
  - 类型：string
  - 描述：vni id
- vxlan_packet
  - 类型：string
  - 描述：是否为 vxlan 协议数据包
- inner_traceid
  - 类型：string
  - 描述：关联该被采集网卡上的某个 tcp 连接的 4 层和 7 层网络日志数据
- host_network
  - 类型：string
  - 描述：是否为主机网络
- virtual_nic
  - 类型：string
  - 描述：是否为虚拟网卡
- direction
  - 类型：string
  - 描述：标识 src 是否为传输层 tcp 连接的发起方，或者是否为应用层的服务端，对于 L4 log，其值除了 `incoming`, `outgoing` 外还可能是 `unknown`

**网卡的 K8s 标签**：

添加以下标签的前提是主机上有部署 K8s/K3s，不同于标签`src_k8s_<xxx>`，这些**标签仅与 K8s Pod 内的网卡相绑定**，故要目标容器有不同于主机网络的 Linux network namespace，即 Pod 资源的 K8s yaml 中不配置 host network。

- k8s_namespace
  - 类型：string
  - 描述：K8s 的命名空间名称
- k8s_pod_name
  - 类型：string
  - 描述：K8s Pod 名称
- k8s_container_name
  - 类型：string
  - 描述：K8s 容器名

**网络连接的 K8s 公共字段**

如果 ip 和 port 能对应上 K8s 资源（使用 host network 的 pod 作为客户端时通常无法标记），将追加 K8s 相关标签

- sub_source
  - 类型：string
  - 描述：默认值 `K8s`
- src_k8s_namespace
  - 类型：string
  - 描述：源 K8s namespace
- src_k8s_pod_name
  - 类型：string
- src_k8s_service_name
  - 类型：string
- src_k8s_deployment_name
  - 类型：string
- dst_k8s_namespace
  - 类型：string
  - 描述：目标对应的 K8s Namespace
- dst_k8s_pod_name
  - 类型：string
- dst_k8s_service_name
  - 类型：string
- dst_k8s_deployment_name
  - 类型：string

### L4 网络相关字段

**字段**：

- chunk_id
  - 类型： int
  - 描述： 一个 tcp 连接数据包信息将被分成几段上传，每一段有一个 chunk id
- tx_seq_min
  - 类型：uint32
  - 描述：当前 chunk，src（出网卡数据包）这一侧的 tcp 序列号最小值
- tx_seq_max
  - 类型：uint32
- rx_seq_min
  - 类型：uint32
  - 描述：当前 chunk，dsr（出网卡数据包）这一侧的 tcp 序列号最小值
- rx_seq_max
  - 类型：uint32
- message
  - 类型：string
  - 描述：此次变更将导致 tcp_series 中 tx/rx 的 seq 要加上 tx/rx_seq_pos，time 加上 time_pos，mac 根据 mac_map 进行映射
    ```json
    {
        "l4_proto": "tcp",
        "tcp": {
            "chunk_id": ...,
            "mac_map": {...},
            "tx_seq_pos": ...,
            "rx_seq_pos": ...,
            "time_pos": ...,
            "tcp_series_col_name": ...,
            "tcp_series": ...,
        }
    }
    ```

- `tcp_series`/`tcp_3(4)whs` 对应 `tcp_series_col_name` 列表参考字段

    | 名称             | 类型   | 描述                                            |
    | ---------------- | ------ | ----------------------------------------------- |
    | txrx             | str    | 数据包进出网卡的方向，值 `tx` 或 `rx`           |
    | flags            | str    | tcp flag，值如 `SYN`, `SYN\|ACK`, `PSH\|ACK` 等 |
    | src_mac          | str    | 源网卡 MAC 地址                                 |
    | dst_mac          | str    | 目标网卡 MAC 地址                               |
    | seq              | uint32 | tcp seq                                         |
    | ack_seq          | uint32 | tcp ack seq                                     |
    | tcp_payload_size | int    | tcp payload size                                |
    | win              | uint32 | tcp window size                                 |
    | ts               | int64  | 捕捉到报文的 unix nano timestamp                |

### L7 网络相关字段

标签：

- l7_traceid
  - 类型：string
  - 描述：应用层请求跟踪 id；结合网卡标签，可跟踪请求经过的各个网卡
- trace_id
  - 类型：string
  - 描述：APM trace id
- parent_id
  - 类型：string
  - 描述：APM parent id
- l7_proto
  - 类型：string
  - 描述：应用层网络协议
- http_path
  - 类型：string
  - 描述：http path
- http_method
  - 类型：string
  - 描述：http method
- http_status_code
  - 类型：string
  - 描述：http status code

字段：

- tx_seq
  - 类型：int
  - 描述：对应 l4log 的 tx seq
- rx_seq
  - 类型：int
  - 描述：对应 l4log 的 rx seq

- req_seq
  - 类型：string
  - 描述：请求的 tcp 序列号，diection 为 outgoing 则对应 l4log 的 tx seq，否则为 rx。
- resp_seq
  - 类型：string
  - 描述：响应的 tcp 序列号，diection 为 outgoing 则对应 l4log 的 rx seq，否则为 tx。
